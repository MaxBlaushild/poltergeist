package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *server) listInsiderTrades(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.Query("limit"))
	offset, _ := strconv.Atoi(ctx.Query("offset"))

	trades, err := s.dbClient.InsiderTrade().List(ctx, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, trades)
}
