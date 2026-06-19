package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

func gameStateResponse(state *models.VampireGameState) gin.H {
	return gin.H{
		"currentAct":           state.CurrentAct,
		"contentUnlocked":      state.ContentUnlocked,
		"quizPart1Open":        state.QuizPart1Open,
		"quizPart2Open":        state.QuizPart2Open,
		"quizPart1OpenedAt":    state.QuizPart1OpenedAt,
		"activeNotificationId": state.ActiveNotificationID,
	}
}

// getState is a lightweight poll endpoint for players: the current game state
// plus any active broadcast notifications. The client polls this to react to
// unlocks, act changes, and announcements without a manual refresh.
func (s *server) getState(ctx *gin.Context) {
	state, err := s.dbClient.Vampire().GetGameState(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	notifs, err := s.dbClient.Vampire().ListActiveNotifications(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"gameState":     gameStateResponse(state),
		"notifications": notifs,
	})
}
