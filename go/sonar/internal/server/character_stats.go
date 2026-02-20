package server

import (
	"errors"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

type characterStatsResponse struct {
	Strength      int `json:"strength"`
	Dexterity     int `json:"dexterity"`
	Constitution  int `json:"constitution"`
	Intelligence  int `json:"intelligence"`
	Wisdom        int `json:"wisdom"`
	Charisma      int `json:"charisma"`
	UnspentPoints int `json:"unspentPoints"`
	Level         int `json:"level"`
}

type characterStatsAllocationRequest struct {
	Allocations map[string]int `json:"allocations"`
}

func (s *server) getCharacterStats(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	userLevel, err := s.dbClient.UserLevel().FindOrCreateForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stats, err := s.dbClient.UserCharacterStats().EnsureLevelPoints(ctx, user.ID, userLevel.Level)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, characterStatsResponseFrom(stats, userLevel.Level))
}

func (s *server) allocateCharacterStats(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var req characterStatsAllocationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil || len(req.Allocations) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "allocations required"})
		return
	}

	userLevel, err := s.dbClient.UserLevel().FindOrCreateForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stats, err := s.dbClient.UserCharacterStats().ApplyAllocations(ctx, user.ID, userLevel.Level, req.Allocations)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrNoStatAllocations),
			errors.Is(err, db.ErrInvalidStatAllocation),
			errors.Is(err, db.ErrInsufficientStatPoints):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, characterStatsResponseFrom(stats, userLevel.Level))
}

func characterStatsResponseFrom(stats *models.UserCharacterStats, level int) characterStatsResponse {
	return characterStatsResponse{
		Strength:      stats.Strength,
		Dexterity:     stats.Dexterity,
		Constitution:  stats.Constitution,
		Intelligence:  stats.Intelligence,
		Wisdom:        stats.Wisdom,
		Charisma:      stats.Charisma,
		UnspentPoints: stats.UnspentPoints,
		Level:         level,
	}
}
