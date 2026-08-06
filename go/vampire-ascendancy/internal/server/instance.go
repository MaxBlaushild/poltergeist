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

// ---- Instances ("Toasts") — platform routes, not scoped to any instance ----

// POST /vampire-ascendancy/instances — "Host a Toast". The caller becomes
// its Host and the instance starts fully populated from the shared content
// library (everything included; the Host trims it down from the Content
// tab).
func (s *server) createInstance(ctx *gin.Context) {
	user := userFromContext(ctx)
	var body struct {
		Name string `json:"name"`
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

	v := s.dbClient.Vampire()
	inst, err := v.CreateInstance(ctx, name, &user.ID)
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
// signed-in user Hosts or Co-Hosts.
func (s *server) listMyInstances(ctx *gin.Context) {
	user := userFromContext(ctx)
	v := s.dbClient.Vampire()
	instances, err := v.ListInstancesForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(instances))
	for _, inst := range instances {
		role := models.InstanceAdminRoleAdmin
		if admin, _ := v.GetInstanceAdmin(ctx, inst.ID, user.ID); admin != nil {
			role = admin.Role
		}
		out = append(out, gin.H{
			"id":        inst.ID,
			"name":      inst.Name,
			"status":    inst.Status,
			"createdAt": inst.CreatedAt,
			"role":      role, // "owner" -> Host, "admin" -> Co-Host
		})
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

// GET /gm/library/items
func (s *server) gmListLibraryItems(ctx *gin.Context) {
	rows, err := s.dbClient.Vampire().ListLibraryItems(ctx, instanceIDFromContext(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": rows})
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

// GET /gm/library/quiz-questions
func (s *server) gmListLibraryQuizQuestions(ctx *gin.Context) {
	rows, err := s.dbClient.Vampire().ListLibraryQuizQuestions(ctx, instanceIDFromContext(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"questions": rows})
}

// PUT /gm/library/quiz-questions/:id { included: bool }
func (s *server) gmSetQuizQuestionIncluded(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}
	var body struct {
		Included bool `json:"included"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Vampire().SetQuizQuestionIncluded(ctx, instanceIDFromContext(ctx), id, body.Included); err != nil {
		conflictOrInternal(ctx, err)
		return
	}
	s.logGM(ctx, "set_quiz_question_included", map[string]interface{}{"questionId": id.String(), "included": body.Included})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
