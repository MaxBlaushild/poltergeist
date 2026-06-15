package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// POST /gm/bt — record a Blood Token award/deduction for a player. Blood Tokens
// are physical; this is the recorded backup tally (e.g. for winning a physical
// game). Does not affect House Favor.
func (s *server) gmAwardBloodTokens(ctx *gin.Context) {
	var body struct {
		PlayerID string `json:"playerId"`
		Delta    int    `json:"delta"`
		Reason   string `json:"reason"`
		Source   string `json:"source"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	playerID, err := uuid.Parse(body.PlayerID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}
	if body.Delta == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "delta must be non-zero"})
		return
	}
	source := body.Source
	if source == "" {
		source = "physical_game"
	}

	gmName := gmNameFromContext(ctx)
	if err := s.dbClient.Vampire().AddBloodTokens(ctx, &models.VampireBloodTokenLog{
		PlayerID: playerID,
		Delta:    body.Delta,
		Reason:   body.Reason,
		Source:   source,
		GMName:   gmName,
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logGM(ctx, "award_blood_tokens", map[string]interface{}{
		"playerId": body.PlayerID,
		"delta":    body.Delta,
		"reason":   body.Reason,
	})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
