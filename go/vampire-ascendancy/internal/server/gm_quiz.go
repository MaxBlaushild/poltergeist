package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// POST /gm/quiz/open — open or close the end quiz for all players.
func (s *server) gmSetQuizOpen(ctx *gin.Context) {
	var body struct {
		Open bool `json:"open"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	state, err := s.dbClient.Vampire().UpdateGameState(ctx, map[string]interface{}{
		"quiz_open": body.Open,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logGM(ctx, "set_quiz_open", map[string]interface{}{"open": body.Open})
	ctx.JSON(http.StatusOK, gameStateResponse(state))
}

// GET /gm/quiz/submissions — all quiz answers, for reading open-ended responses
// and adjudicating.
func (s *server) gmListQuizSubmissions(ctx *gin.Context) {
	details, err := s.dbClient.Vampire().ListQuizSubmissionsDetailed(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"submissions": details})
}
