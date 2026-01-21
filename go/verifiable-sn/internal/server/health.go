package server

import (
	"github.com/gin-gonic/gin"
)

func (s *server) GetHealth(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "ok",
	})
}

