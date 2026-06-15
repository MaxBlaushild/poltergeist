package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// submitMission records the player's answer for one of their own missions and
// marks it submitted. Requires content to be unlocked, and enforces that the
// mission belongs to the player's character.
func (s *server) submitMission(ctx *gin.Context) {
	player := playerFromContext(ctx)

	state, err := s.dbClient.Vampire().GetGameState(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !state.ContentUnlocked {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "missions are not available yet"})
		return
	}
	if player.CharacterID == nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "no character assigned"})
		return
	}

	missionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid mission id"})
		return
	}

	mission, err := s.dbClient.Vampire().GetMissionByID(ctx, missionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Guard: a player may only submit their own character's missions.
	if mission == nil || mission.CharacterID != *player.CharacterID {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "mission not found"})
		return
	}

	var body struct {
		Answer string `json:"answer"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, err := s.dbClient.Vampire().UpsertMissionSubmission(ctx, player.ID, missionID, body.Answer)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":       sub.Status,
		"playerAnswer": sub.PlayerAnswer,
		"awardedBt":    sub.AwardedBT,
	})
}
