package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

const playerContextKey = "vampirePlayer"

// withPlayer authenticates a player from their opaque per-character token.
// The token is the player's identity — no character id ever comes from the
// client, so a player can only ever reach their own packet.
func (s *server) withPlayer(ctx *gin.Context) {
	token := ctx.GetHeader("X-Player-Token")
	if token == "" {
		token = ctx.Query("token")
	}
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing player token"})
		return
	}

	player, err := s.dbClient.Vampire().GetPlayerByToken(ctx, token)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if player == nil || !player.Active {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid player token"})
		return
	}

	ctx.Set(playerContextKey, player)
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
