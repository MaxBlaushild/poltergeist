package server

import "github.com/gin-gonic/gin"

func (s *server) getHealth(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"status": "ok",
	})
}
