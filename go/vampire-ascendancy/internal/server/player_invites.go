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
// phone number) to play a specific character in their Toast. Replaces the
// old walk-up "pick your character + sigil" login — the only way into an
// instance as a player is now an accepted invite. See
// MULTI_TENANT_REQUIREMENTS.md.

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
		row := gin.H{
			"id":          inv.ID,
			"guestName":   inv.GuestName,
			"phoneNumber": inv.PhoneNumber,
			"status":      inv.Status,
			"createdAt":   inv.CreatedAt,
		}
		if inv.Character != nil {
			ch := gin.H{"id": inv.Character.ID, "name": inv.Character.Name, "title": inv.Character.Title}
			if inv.Character.House != nil {
				ch["house"] = inv.Character.House.Name
			}
			row["character"] = ch
		}
		out = append(out, row)
	}
	ctx.JSON(http.StatusOK, gin.H{"invites": out})
}

// POST /gm/invites { guestName, phoneNumber, characterId } — invite a real
// person to a character. Sends the RSVP link by SMS.
func (s *server) gmCreateInvite(ctx *gin.Context) {
	instanceID := instanceIDFromContext(ctx)
	user := userFromContext(ctx)
	var body struct {
		GuestName   string `json:"guestName"`
		PhoneNumber string `json:"phoneNumber"`
		CharacterID string `json:"characterId"`
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
	characterID, err := uuid.Parse(body.CharacterID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character id"})
		return
	}

	v := s.dbClient.Vampire()
	character, err := v.GetCharacterByID(ctx, characterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if character == nil || character.RoleType != "player" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unknown character"})
		return
	}

	token, err := genOpaqueToken()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	inv, err := v.CreatePlayerInvite(ctx, instanceID, characterID, guestName, phone, user.ID, token)
	if err != nil {
		conflictOrInternal(ctx, err)
		return
	}

	if err := s.sendInviteSMS(ctx, guestName, phone, character.Name, token); err != nil {
		// The invite exists either way — a GM can Resend once the phone
		// number's fixed, so don't fail the whole request over SMS delivery.
		ctx.JSON(http.StatusOK, gin.H{"id": inv.ID, "warning": "invite created, but the text failed to send: " + err.Error()})
		return
	}

	s.logGM(ctx, "create_invite", map[string]interface{}{"characterId": characterID.String(), "guestName": guestName})
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
	var found *inviteWithCharacterName
	for _, inv := range invites {
		if inv.ID != id {
			continue
		}
		if inv.Status != "pending" {
			ctx.JSON(http.StatusConflict, gin.H{"error": "this invite is no longer pending"})
			return
		}
		name := ""
		if inv.Character != nil {
			name = inv.Character.Name
		}
		found = &inviteWithCharacterName{token: inv.Token, guestName: inv.GuestName, phone: inv.PhoneNumber, characterName: name}
		break
	}
	if found == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "invite not found"})
		return
	}
	if err := s.sendInviteSMS(ctx, found.guestName, found.phone, found.characterName, found.token); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "resend_invite", map[string]interface{}{"id": id.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

type inviteWithCharacterName struct {
	token         string
	guestName     string
	phone         string
	characterName string
}

// sendInviteSMS texts the RSVP link. Kept short (not the full bio — that's
// on the RSVP page itself) since SMS bodies are metered per segment.
func (s *server) sendInviteSMS(ctx *gin.Context, guestName, phone, characterName, token string) error {
	if s.texterClient == nil {
		return fmt.Errorf("texting is not configured")
	}
	link := fmt.Sprintf("%s/rsvp/%s", strings.TrimRight(s.siteURL, "/"), token)
	body := fmt.Sprintf("%s, you're invited to play %s in a vampire murder-mystery night! See your character and RSVP: %s", guestName, characterName, link)
	return s.texterClient.Text(ctx, &texter.Text{
		Body:     body,
		To:       phone,
		From:     s.fromPhone,
		TextType: "vampire-ascendancy-player-invite",
	})
}

// ---- Public (RSVP page) ----

// GET /rsvp/:token — the invite teaser: character name/title/house/bio and
// who it's for. No account required. 404s for an unknown token so it can't
// be used to probe for valid ones.
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
	if inv.Character != nil {
		ch := gin.H{
			"id":           inv.Character.ID,
			"name":         inv.Character.Name,
			"title":        inv.Character.Title,
			"preEventInfo": inv.Character.PreEventInfo,
		}
		if inv.Character.House != nil {
			ch["house"] = inv.Character.House.Name
		}
		resp["character"] = ch
	}
	ctx.JSON(http.StatusOK, resp)
}

// POST /rsvp/:token/decline — no account required; the token is the
// credential. Frees the character up for the admin to re-invite someone else.
func (s *server) declinePlayerInvite(ctx *gin.Context) {
	token := ctx.Param("token")
	if err := s.dbClient.Vampire().DeclinePlayerInvite(ctx, token); err != nil {
		conflictOrInternal(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /rsvp/:token/accept — requires a signed-in account (email/password
// or Google, same sign-up flow as Hosts). Creates the player row and
// returns the instance id so the frontend can drop them straight into it.
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
