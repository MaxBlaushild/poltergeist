package server

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *server) sendTestPushToCurrentUser(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		log.Printf("[push][test] unauthorized request: %v", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}
	log.Printf("[push][test] request received for user=%s", user.ID)

	var req struct {
		DelaySeconds int `json:"delaySeconds"`
	}
	if ctx.Request.ContentLength > 0 {
		if err := ctx.ShouldBindJSON(&req); err != nil && err != io.EOF {
			log.Printf("[push][test] invalid payload for user=%s: %v", user.ID, err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid payload",
			})
			return
		}
	}
	if req.DelaySeconds < 0 || req.DelaySeconds > 30 {
		log.Printf(
			"[push][test] invalid delay for user=%s: %d",
			user.ID,
			req.DelaySeconds,
		)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "delaySeconds must be between 0 and 30",
		})
		return
	}
	if req.DelaySeconds > 0 {
		log.Printf(
			"[push][test] delaying send for user=%s by %ds",
			user.ID,
			req.DelaySeconds,
		)
		select {
		case <-time.After(time.Duration(req.DelaySeconds) * time.Second):
		case <-ctx.Request.Context().Done():
			log.Printf("[push][test] request cancelled during delay for user=%s", user.ID)
			return
		}
	}

	if s.pushClient == nil {
		log.Printf("[push][test] push client not configured for user=%s", user.ID)
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "push client not configured",
		})
		return
	}

	tokens, err := s.dbClient.UserDeviceToken().FindByUserID(ctx.Request.Context(), user.ID)
	if err != nil {
		log.Printf("[push][test] failed to fetch tokens for user=%s: %v", user.ID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch registered device tokens",
		})
		return
	}
	log.Printf("[push][test] found %d token(s) for user=%s", len(tokens), user.ID)
	if len(tokens) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "no device tokens registered for current user",
		})
		return
	}

	data := map[string]string{
		"type":   "push_test",
		"sentAt": time.Now().UTC().Format(time.RFC3339),
	}
	title := "Unclaimed Streets"
	body := "Test push notification delivered successfully."

	sentCount := 0
	failedCount := 0
	for _, token := range tokens {
		if err := s.pushClient.Send(ctx.Request.Context(), token.Token, title, body, data); err != nil {
			failedCount++
			log.Printf(
				"[push][test] send failed user=%s platform=%s token=%s: %v",
				user.ID,
				token.Platform,
				tokenPreview(token.Token),
				err,
			)
			continue
		}
		sentCount++
		log.Printf(
			"[push][test] send succeeded user=%s platform=%s token=%s",
			user.ID,
			token.Platform,
			tokenPreview(token.Token),
		)
	}
	log.Printf(
		"[push][test] send complete user=%s sent=%d failed=%d total=%d",
		user.ID,
		sentCount,
		failedCount,
		len(tokens),
	)

	if sentCount == 0 {
		ctx.JSON(http.StatusBadGateway, gin.H{
			"error":  "failed to send push to registered tokens",
			"failed": failedCount,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"sent":             sentCount,
		"failed":           failedCount,
		"tokens":           len(tokens),
		"delayedBySeconds": req.DelaySeconds,
	})
}
