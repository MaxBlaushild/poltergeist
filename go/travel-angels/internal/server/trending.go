package server

import "github.com/gin-gonic/gin"

func (s *server) GetTrendingDestinations(ctx *gin.Context) {
	if s.library == nil {
		ctx.JSON(500, gin.H{"error": "Library unavailable"})
		return
	}
	s.library.Trending(ctx)
}
