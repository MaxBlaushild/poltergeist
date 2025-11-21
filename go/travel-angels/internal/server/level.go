package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *server) GetLevel(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	level, err := s.dbClient.UserLevel().FindOrCreateForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	level.ExperienceToNextLevel = level.XPToNextLevel()

	ctx.JSON(http.StatusOK, level)
}
