package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Player invites: a Host/Co-Host invites a specific real person (name +
// phone number) to join their Toast as a player. Character-agnostic — the
// invite doesn't name a character; the invitee picks their own, from the
// instance's curated pool (see character_pool.go), after accepting. Which
// characters are even offerable is curated separately, on the Character
// Pool tab. Replaces the old walk-up "pick your character + sigil" login,
// and the version of this flow where a Host assigned the character up
// front. See MULTI_TENANT_REQUIREMENTS.md and MYSTERY_REQUIREMENTS.md.

// ---- Admin (gm-scoped): the Invites tab ----

// GET /gm/invites — every invite for this Toast (any status), newest first.
func (s *server) gmListInvites(ctx *gin.Context) {
	invites, err := s.dbClient.Vampire().ListPlayerInvites(ctx, instanceIDFromContext(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(invites))
	for _, inv := range invites {
		out = append(out, gin.H{
			"id":          inv.ID,
			"guestName":   inv.GuestName,
			"phoneNumber": inv.PhoneNumber,
			"status":      inv.Status,
			"createdAt":   inv.CreatedAt,
		})
	}
	ctx.JSON(http.StatusOK, gin.H{"invites": out})
}

// POST /gm/invites { guestName, phoneNumber } — invite a real person to
// join as a player; they choose their own character after accepting. Sends
// the RSVP link by SMS.
func (s *server) gmCreateInvite(ctx *gin.Context) {
	instanceID := instanceIDFromContext(ctx)
	user := userFromContext(ctx)
	var body struct {
		GuestName   string `json:"guestName"`
		PhoneNumber string `json:"phoneNumber"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	guestName := strings.TrimSpace(body.GuestName)
	phone := strings.TrimSpace(body.PhoneNumber)
	if guestName == "" || phone == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "a name and phone number are required"})
		return
	}

	v := s.dbClient.Vampire()
	token, err := genOpaqueToken()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	inv, err := v.CreatePlayerInvite(ctx, instanceID, guestName, phone, user.ID, token)
	if err != nil {
		conflictOrInternal(ctx, err)
		return
	}

	if err := s.sendInviteSMS(ctx, guestName, phone, token); err != nil {
		// The invite exists either way — a GM can Resend once the phone
		// number's fixed, so don't fail the whole request over SMS delivery.
		ctx.JSON(http.StatusOK, gin.H{"id": inv.ID, "warning": "invite created, but the text failed to send: " + err.Error()})
		return
	}

	s.logGM(ctx, "create_invite", map[string]interface{}{"guestName": guestName})
	ctx.JSON(http.StatusOK, gin.H{"id": inv.ID})
}

// DELETE /gm/invites/:id — revoke a pending invite (or clear a stale
// accepted/declined one from the list).
func (s *server) gmDeleteInvite(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite id"})
		return
	}
	if err := s.dbClient.Vampire().DeletePlayerInvite(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "delete_invite", map[string]interface{}{"id": id.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /gm/invites/:id/resend — re-send the same RSVP link by SMS (e.g. the
// guest lost the text, or the number was mistyped and just got fixed via a
// fresh invite).
func (s *server) gmResendInvite(ctx *gin.Context) {
	instanceID := instanceIDFromContext(ctx)
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite id"})
		return
	}
	invites, err := s.dbClient.Vampire().ListPlayerInvites(ctx, instanceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var found *inviteToResend
	for _, inv := range invites {
		if inv.ID != id {
			continue
		}
		if inv.Status != "pending" {
			ctx.JSON(http.StatusConflict, gin.H{"error": "this invite is no longer pending"})
			return
		}
		found = &inviteToResend{token: inv.Token, guestName: inv.GuestName, phone: inv.PhoneNumber}
		break
	}
	if found == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "invite not found"})
		return
	}
	if err := s.sendInviteSMS(ctx, found.guestName, found.phone, found.token); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "resend_invite", map[string]interface{}{"id": id.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

type inviteToResend struct {
	token     string
	guestName string
	phone     string
}

// sendInviteSMS texts the RSVP link. No character to name anymore — that's
// chosen by the invitee after accepting, not known at invite time.
func (s *server) sendInviteSMS(ctx *gin.Context, guestName, phone, token string) error {
	if s.texterClient == nil {
		return fmt.Errorf("texting is not configured")
	}
	link := fmt.Sprintf("%s/rsvp/%s", strings.TrimRight(s.siteURL, "/"), token)
	body := fmt.Sprintf("%s, you're invited to a vampire murder-mystery night! RSVP and choose your character: %s", guestName, link)
	return s.texterClient.Text(ctx, &texter.Text{
		Body:     body,
		To:       phone,
		From:     s.fromPhone,
		TextType: "vampire-ascendancy-player-invite",
	})
}

// ---- Public (RSVP page) ----

// GET /rsvp/:token — the invite teaser: who it's for and which Toast. No
// character to show yet (chosen after accepting) — no account required.
// 404s for an unknown token so it can't be used to probe for valid ones.
func (s *server) getPlayerInvite(ctx *gin.Context) {
	token := ctx.Param("token")
	inv, err := s.dbClient.Vampire().GetPlayerInviteByToken(ctx, token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if inv == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "invite not found"})
		return
	}
	instance, err := s.dbClient.Vampire().GetInstanceByID(ctx, inv.InstanceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := gin.H{
		"guestName":  inv.GuestName,
		"status":     inv.Status,
		"instanceId": inv.InstanceID,
	}
	if instance != nil {
		resp["instanceName"] = instance.Name
	}
	ctx.JSON(http.StatusOK, resp)
}

// POST /rsvp/:token/decline — no account required; the token is the
// credential. Frees up an invite slot for the host to invite someone else.
func (s *server) declinePlayerInvite(ctx *gin.Context) {
	token := ctx.Param("token")
	if err := s.dbClient.Vampire().DeclinePlayerInvite(ctx, token); err != nil {
		conflictOrInternal(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /rsvp/:token/accept — requires a signed-in account (email/password
// or Google, same sign-up flow as Hosts). Creates the player row (with no
// character yet — see character_self_select.go for the picker that follows)
// and returns the instance id so the frontend can drop them straight into it.
func (s *server) acceptPlayerInvite(ctx *gin.Context) {
	token := ctx.Param("token")
	user := userFromContext(ctx)
	player, err := s.dbClient.Vampire().AcceptPlayerInvite(ctx, token, user.ID)
	if err != nil {
		conflictOrInternal(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"instanceId": player.InstanceID})
}
