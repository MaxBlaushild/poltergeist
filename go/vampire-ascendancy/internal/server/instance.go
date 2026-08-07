package server

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// conflictOrInternal writes a 409 for a *db.ConflictError (a business-rule
// guard, e.g. "can't remove the Host") or a 500 for anything else.
func conflictOrInternal(ctx *gin.Context, err error) {
	if ce, ok := err.(*db.ConflictError); ok {
		ctx.JSON(http.StatusConflict, gin.H{"error": ce.Message})
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func genOpaqueToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GET /vampire-ascendancy/mysteries — every active mystery, for the "Host
// a Toast" picker. Just enough to choose from (name + summary); the full
// editor payload (lore, beats, quiz, secrets) is super-user-only, under
// /admin/mysteries.
func (s *server) listActiveMysteries(ctx *gin.Context) {
	mysteries, err := s.dbClient.Vampire().ListMysteries(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(mysteries))
	for _, m := range mysteries {
		if !m.Active {
			continue
		}
		out = append(out, gin.H{"id": m.ID, "name": m.Name, "summary": m.Summary})
	}
	ctx.JSON(http.StatusOK, gin.H{"mysteries": out})
}

// ---- Instances ("Toasts") — platform routes, not scoped to any instance ----

// POST /vampire-ascendancy/instances — "Host a Toast". The caller becomes
// its Host and the instance starts fully populated with the shared items
// library (the Host trims it down from the Content tab); mysteryId is
// chosen once here and never changes (see MYSTERY_REQUIREMENTS.md).
func (s *server) createInstance(ctx *gin.Context) {
	user := userFromContext(ctx)
	var body struct {
		Name      string `json:"name"`
		MysteryID string `json:"mysteryId"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "give this Toast a name"})
		return
	}
	mysteryID, err := uuid.Parse(body.MysteryID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "choose a mystery"})
		return
	}

	v := s.dbClient.Vampire()
	mystery, err := v.GetMysteryByID(ctx, mysteryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if mystery == nil || !mystery.Active {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unknown or inactive mystery"})
		return
	}

	inst, err := v.CreateInstance(ctx, name, &user.ID, mysteryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := v.AddInstanceAdmin(ctx, inst.ID, user.ID, models.InstanceAdminRoleOwner); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := v.SeedInstanceLibrary(ctx, inst.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Seed the game-state row up front so GET /gm/state never has to
	// lazily create it on first read.
	if _, err := v.GetGameState(ctx, inst.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": inst.ID, "name": inst.Name})
}

// GET /vampire-ascendancy/instances — "My Toasts": every instance the
// signed-in user administers (Host/Co-Host) *or* plays a character in.
// Each row is tagged with a "kind" so the dashboard knows where to send
// them — an admin row links to the GM console, a player row links to the
// player app and carries the character (with portrait) for a preview
// card. An instance where the account is both (rare — a Co-Host who's
// also been invited as a player) is shown once, as admin, since that's
// the more privileged surface.
func (s *server) listMyInstances(ctx *gin.Context) {
	user := userFromContext(ctx)
	v := s.dbClient.Vampire()

	adminInstances, err := v.ListInstancesForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(adminInstances))
	seen := make(map[uuid.UUID]bool, len(adminInstances))
	for _, inst := range adminInstances {
		role := models.InstanceAdminRoleAdmin
		if admin, _ := v.GetInstanceAdmin(ctx, inst.ID, user.ID); admin != nil {
			role = admin.Role
		}
		seen[inst.ID] = true
		out = append(out, gin.H{
			"id":        inst.ID,
			"name":      inst.Name,
			"status":    inst.Status,
			"createdAt": inst.CreatedAt,
			"kind":      "admin",
			"role":      role, // "owner" -> Host, "admin" -> Co-Host
		})
	}

	players, err := v.ListPlayerInstancesForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, p := range players {
		if seen[p.InstanceID] {
			continue
		}
		seen[p.InstanceID] = true
		inst, err := v.GetInstanceByID(ctx, p.InstanceID)
		if err != nil || inst == nil {
			continue
		}
		row := gin.H{
			"id":        inst.ID,
			"name":      inst.Name,
			"status":    inst.Status,
			"createdAt": inst.CreatedAt,
			"kind":      "player",
		}
		if p.Character != nil {
			ch := gin.H{"id": p.Character.ID, "name": p.Character.Name, "title": p.Character.Title}
			if p.Character.House != nil {
				ch["house"] = p.Character.House.Name
			}
			if ic, _ := v.GetInstanceCharacter(ctx, p.InstanceID, p.Character.ID); ic != nil && ic.ImageURL != "" {
				ch["imageUrl"] = ic.ImageURL
			}
			row["character"] = ch
		}
		out = append(out, row)
	}

	ctx.JSON(http.StatusOK, gin.H{"instances": out})
}

// POST /vampire-ascendancy/invites/:token/accept — accept a "join as
// Co-Host" invite. Requires the caller to already be signed in (email or
// Google) — if they don't have an account yet, the frontend sends them
// through sign-up first and retries.
func (s *server) acceptInstanceAdminInvite(ctx *gin.Context) {
	user := userFromContext(ctx)
	token := ctx.Param("token")
	instanceID, err := s.dbClient.Vampire().AcceptInstanceAdminInvite(ctx, token, user.ID)
	if err != nil {
		conflictOrInternal(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"instanceId": instanceID})
}

// ---- Administrators ("Host" / "Co-Host") — gm routes ----

// GET /gm/admins — the Co-Hosts tab's roster, Host first, plus pending invites.
func (s *server) gmListAdmins(ctx *gin.Context) {
	instanceID := instanceIDFromContext(ctx)
	v := s.dbClient.Vampire()
	admins, err := v.ListInstanceAdmins(ctx, instanceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	invites, err := v.ListPendingInstanceAdminInvites(ctx, instanceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	adminOut := make([]gin.H, 0, len(admins))
	for _, a := range admins {
		row := gin.H{"userId": a.UserID, "role": a.Role}
		if a.User != nil {
			row["name"] = a.User.Name
			row["email"] = a.User.Email
		}
		adminOut = append(adminOut, row)
	}
	inviteOut := make([]gin.H, 0, len(invites))
	for _, inv := range invites {
		inviteOut = append(inviteOut, gin.H{"id": inv.ID, "email": inv.Email, "expiresAt": inv.ExpiresAt})
	}
	ctx.JSON(http.StatusOK, gin.H{"admins": adminOut, "pendingInvites": inviteOut})
}

// POST /gm/admins/invite — "Invite a Co-Host". Any current Host or Co-Host
// may invite another. Returns the invite token; the frontend builds the
// shareable accept link (`/accept-invite/<token>`) and the admin sends it —
// there's no automated email delivery yet.
func (s *server) gmInviteAdmin(ctx *gin.Context) {
	instanceID := instanceIDFromContext(ctx)
	user := userFromContext(ctx)
	var body struct {
		Email string `json:"email"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	email := strings.TrimSpace(strings.ToLower(body.Email))
	if email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "an email is required"})
		return
	}

	token, err := genOpaqueToken()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	inv, err := s.dbClient.Vampire().CreateInstanceAdminInvite(ctx, instanceID, email, user.ID, token, time.Now().Add(7*24*time.Hour))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "invite_admin", map[string]interface{}{"email": email})
	ctx.JSON(http.StatusOK, gin.H{"id": inv.ID, "token": inv.Token, "expiresAt": inv.ExpiresAt})
}

// DELETE /gm/admins/invites/:id — revoke a pending invite.
func (s *server) gmDeleteAdminInvite(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite id"})
		return
	}
	if err := s.dbClient.Vampire().DeleteInstanceAdminInvite(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "revoke_admin_invite", map[string]interface{}{"id": id.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /gm/admins/:userId — remove a Co-Host. Any current administrator
// may remove another, EXCEPT the Host — that returns 409 (see
// MULTI_TENANT_REQUIREMENTS.md's "Admin management rights": without this
// guardrail two admins could remove each other down to zero and lock
// everyone out with no recovery path). The Host must hand off hosting
// first.
func (s *server) gmRemoveAdmin(ctx *gin.Context) {
	instanceID := instanceIDFromContext(ctx)
	userID, err := uuid.Parse(ctx.Param("userId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := s.dbClient.Vampire().RemoveInstanceAdmin(ctx, instanceID, userID); err != nil {
		conflictOrInternal(ctx, err)
		return
	}
	s.logGM(ctx, "remove_admin", map[string]interface{}{"userId": userID.String()})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /gm/admins/transfer — "Make [name] the Host". Only the current Host
// can call this; toUserId must already be a Co-Host.
func (s *server) gmTransferOwnership(ctx *gin.Context) {
	instanceID := instanceIDFromContext(ctx)
	user := userFromContext(ctx)
	var body struct {
		ToUserID string `json:"toUserId"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	toUserID, err := uuid.Parse(body.ToUserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := s.dbClient.Vampire().TransferInstanceOwnership(ctx, instanceID, user.ID, toUserID); err != nil {
		conflictOrInternal(ctx, err)
		return
	}
	s.logGM(ctx, "transfer_ownership", map[string]interface{}{"toUserId": body.ToUserID})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---- Content library (Content tab): items / quiz questions ----
//
// Characters have no inclusion toggle here — which characters are "in" an
// instance is now implicit (someone has an accepted invite to play them; see
// player_invites.go), not a separate step. Items and quiz questions are
// still real catalog entries a Toast can be trimmed to.

type libraryItemWithPhoto struct {
	db.LibraryItem
	HasPhoto bool `json:"hasPhoto"`
}

// GET /gm/library/items — every catalog item (included or not), with the
// same detail (description, effect, photo) as the Items tab's assign
// picker, so the Content tab's include/exclude toggle can show the same
// rich card instead of a bare name.
func (s *server) gmListLibraryItems(ctx *gin.Context) {
	v := s.dbClient.Vampire()
	rows, err := v.ListLibraryItems(ctx, instanceIDFromContext(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	photoIDs, _ := v.ItemPhotoIDs(ctx)
	has := make(map[string]bool, len(photoIDs))
	for _, id := range photoIDs {
		has[id.String()] = true
	}
	out := make([]libraryItemWithPhoto, 0, len(rows))
	for _, r := range rows {
		out = append(out, libraryItemWithPhoto{LibraryItem: r, HasPhoto: has[r.ID.String()]})
	}
	ctx.JSON(http.StatusOK, gin.H{"items": out})
}

// PUT /gm/library/items/:id { included: bool }
func (s *server) gmSetItemIncluded(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}
	var body struct {
		Included bool `json:"included"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Vampire().SetItemIncluded(ctx, instanceIDFromContext(ctx), id, body.Included); err != nil {
		conflictOrInternal(ctx, err)
		return
	}
	s.logGM(ctx, "set_item_included", map[string]interface{}{"itemId": id.String(), "included": body.Included})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

