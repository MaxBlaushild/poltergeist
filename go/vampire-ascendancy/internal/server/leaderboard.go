package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getLeaderboard returns House Favor standings — always visible to players, even
// before content is unlocked. It is the authoritative live standing.
func (s *server) getLeaderboard(ctx *gin.Context) {
	standings, err := s.dbClient.Vampire().Leaderboard(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"standings": standings})
}

// POST /gm/hf — award or deduct House Favor. Appends to the ledger; the
// leaderboard is the running sum.
func (s *server) gmAwardHouseFavor(ctx *gin.Context) {
	var body struct {
		HouseID string `json:"houseId"`
		Delta   int    `json:"delta"`
		Reason  string `json:"reason"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	houseID, err := uuid.Parse(body.HouseID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid house id"})
		return
	}
	if body.Delta == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "delta must be non-zero"})
		return
	}

	gmName := gmNameFromContext(ctx)
	if err := s.dbClient.Vampire().AddHouseFavor(ctx, &models.VampireHouseFavorLedger{
		HouseID: houseID,
		Delta:   body.Delta,
		Reason:  body.Reason,
		GMName:  gmName,
		Source:  "manual",
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logGM(ctx, "award_house_favor", map[string]interface{}{
		"houseId": body.HouseID,
		"delta":   body.Delta,
		"reason":  body.Reason,
	})

	standings, err := s.dbClient.Vampire().Leaderboard(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"standings": standings})
}

// getHouses lists the houses (for GM award dropdowns).
func (s *server) getHouses(ctx *gin.Context) {
	houses, err := s.dbClient.Vampire().ListHouses(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"houses": houses})
}
