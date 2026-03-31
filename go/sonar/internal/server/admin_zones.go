package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *server) getAdminZones(ctx *gin.Context) {
	summaries, err := s.dbClient.Zone().FindAdminSummaries(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, summaries)
}
