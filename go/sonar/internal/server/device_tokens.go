package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *server) registerDeviceToken(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		log.Printf("[push][register-token] unauthorized request: %v", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var req struct {
		Token    string `json:"token" binding:"required"`
		Platform string `json:"platform" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("[push][register-token] user=%s invalid payload: %v", user.ID, err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "token and platform are required",
		})
		return
	}

	token := strings.TrimSpace(req.Token)
	platform := strings.ToLower(strings.TrimSpace(req.Platform))
	if token == "" {
		log.Printf("[push][register-token] user=%s missing token (platform=%q)", user.ID, platform)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "token and platform are required",
		})
		return
	}

	switch platform {
	case "ios", "android", "web":
	default:
		log.Printf(
			"[push][register-token] user=%s invalid platform=%q token=%s",
			user.ID,
			platform,
			tokenPreview(token),
		)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "platform must be one of ios, android, web",
		})
		return
	}
	log.Printf(
		"[push][register-token] user=%s platform=%s token=%s len=%d",
		user.ID,
		platform,
		tokenPreview(token),
		len(token),
	)

	if err := s.dbClient.UserDeviceToken().Upsert(
		ctx.Request.Context(),
		user.ID,
		token,
		platform,
	); err != nil {
		log.Printf("[push][register-token] user=%s upsert failed: %v", user.ID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	log.Printf("[push][register-token] user=%s upsert succeeded", user.ID)

	ctx.Status(http.StatusNoContent)
}
