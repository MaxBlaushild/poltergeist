package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const playerContextKey = "vampirePlayer"

// instanceIDParam parses the :instanceId path param every instance-scoped
// route carries.
func instanceIDParam(ctx *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(ctx.Param("instanceId"))
}

// withPlayer authenticates a player from their real signed-in account
// (same Bearer-token mechanism as withUser/withInstanceAdmin) and resolves
// which character they play in the instance named in the URL — the
// VampirePlayer row created when they accepted their invite. Replaces the
// old opaque per-character token/sigil: there is no longer any way to reach
// a character's packet without a real account that actually holds it.
func (s *server) withPlayer(ctx *gin.Context) {
	instanceID, err := instanceIDParam(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid instance id"})
		return
	}
	user := s.authenticateUser(ctx)
	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sign in required"})
		return
	}
	player, err := s.dbClient.Vampire().GetPlayerByUserAndInstance(ctx, instanceID, user.ID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if player == nil || !player.Active {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "you don't have a character in this Toast"})
		return
	}
	instance, err := s.dbClient.Vampire().GetInstanceByID(ctx, instanceID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if instance == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	ctx.Set(playerContextKey, player)
	ctx.Set(mysteryIDContextKey, instance.MysteryID)
	ctx.Next()
}

func playerFromContext(ctx *gin.Context) *models.VampirePlayer {
	v, ok := ctx.Get(playerContextKey)
	if !ok {
		return nil
	}
	player, _ := v.(*models.VampirePlayer)
	return player
}

const (
	currentUserContextKey = "vampireUser"
	instanceIDContextKey  = "vampireInstanceId"
	gmNameContextKey      = "vampireGMName"
	// mysteryIDContextKey is set by withPlayer/withInstanceAdmin (both
	// already resolve the instance row) so quiz handlers don't need a
	// separate GetInstanceByID call of their own — see mysteryIDFromContext.
	mysteryIDContextKey = "vampireMysteryId"
)

// authenticateUser validates the request's Bearer token against the shared
// authenticator (the same mechanism every other domain in this repo uses —
// see pkg/middleware.WithAuthenticationWithoutLocation) and returns the
// signed-in user, or nil if the request isn't authenticated.
func (s *server) authenticateUser(ctx *gin.Context) *models.User {
	header := ctx.Request.Header.Get("Authorization")
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil
	}
	user, err := s.authClient.VerifyToken(ctx, &auth.VerifyTokenRequest{Token: parts[1]})
	if err != nil {
		return nil
	}
	return user
}

// withUser requires a signed-in account — used by the platform routes that
// aren't scoped to any one instance yet (creating a Toast, listing "My
// Toasts", accepting a Co-Host invite).
func (s *server) withUser(ctx *gin.Context) {
	user := s.authenticateUser(ctx)
	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sign in required"})
		return
	}
	ctx.Set(currentUserContextKey, user)
	ctx.Next()
}

func userFromContext(ctx *gin.Context) *models.User {
	v, ok := ctx.Get(currentUserContextKey)
	if !ok {
		return nil
	}
	user, _ := v.(*models.User)
	return user
}

// withInstanceAdmin guards the admin surface: the caller must be signed in
// AND be a Host or Co-Host of the instance named in the URL. Replaces the
// old shared-passcode gate — every mutation is now attributable to a real
// account instead of a typed-in name.
func (s *server) withInstanceAdmin(ctx *gin.Context) {
	instanceID, err := instanceIDParam(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid instance id"})
		return
	}
	user := s.authenticateUser(ctx)
	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sign in required"})
		return
	}
	admin, err := s.dbClient.Vampire().GetInstanceAdmin(ctx, instanceID, user.ID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if admin == nil {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "you are not a Host or Co-Host of this Toast"})
		return
	}
	instance, err := s.dbClient.Vampire().GetInstanceByID(ctx, instanceID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if instance == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	ctx.Set(currentUserContextKey, user)
	ctx.Set(instanceIDContextKey, instanceID)
	ctx.Set(mysteryIDContextKey, instance.MysteryID)
	ctx.Set(gmNameContextKey, user.Name)
	ctx.Next()
}

func instanceIDFromContext(ctx *gin.Context) uuid.UUID {
	v, _ := ctx.Get(instanceIDContextKey)
	id, _ := v.(uuid.UUID)
	return id
}

// mysteryIDFromContext returns the current instance's mystery — set by
// withPlayer (player routes) or withInstanceAdmin (gm routes), both of
// which already load the instance row for their own checks.
func mysteryIDFromContext(ctx *gin.Context) uuid.UUID {
	v, _ := ctx.Get(mysteryIDContextKey)
	id, _ := v.(uuid.UUID)
	return id
}

func gmNameFromContext(ctx *gin.Context) string {
	v, ok := ctx.Get(gmNameContextKey)
	if !ok {
		return "unknown"
	}
	name, _ := v.(string)
	return name
}

// withSuperUser guards the shared content library editor (/admin/*) — the
// only accounts allowed to edit characters, houses, items, and quiz
// questions. Separate from, and stricter than, withInstanceAdmin: a Host or
// Co-Host on some instance is not automatically a super user.
func (s *server) withSuperUser(ctx *gin.Context) {
	user := s.authenticateUser(ctx)
	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sign in required"})
		return
	}
	ok, err := s.dbClient.Vampire().IsSuperUser(ctx, user.ID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "you don't have access to the shared content library"})
		return
	}
	ctx.Set(currentUserContextKey, user)
	ctx.Next()
}

// logSuperUser writes an audit entry to the shared-library log, attributed
// to the signed-in super user. Failures are swallowed, same as logGM.
func (s *server) logSuperUser(ctx *gin.Context, action string, payload map[string]interface{}) {
	user := userFromContext(ctx)
	if user == nil {
		return
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		raw = []byte("{}")
	}
	_ = s.dbClient.Vampire().LogSuperUserAction(ctx, user.ID, action, raw)
}
