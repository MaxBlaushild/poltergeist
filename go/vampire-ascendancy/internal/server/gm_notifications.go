package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var validScopes = map[string]bool{"all": true, "house": true, "player": true}

// POST /gm/notifications — push a full-screen broadcast to everyone, one house,
// or one player. Becomes the active notification.
func (s *server) gmPushNotification(ctx *gin.Context) {
	var body struct {
		Title    string `json:"title"`
		Body     string `json:"body"`
		Scope    string `json:"scope"`
		TargetID string `json:"targetId"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validScopes[body.Scope] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid scope"})
		return
	}
	if body.Body == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "message body required"})
		return
	}

	var target *uuid.UUID
	if body.Scope != "all" {
		id, err := uuid.Parse(body.TargetID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "a target is required for this scope"})
			return
		}
		target = &id
	}

	notif := &models.VampireNotification{
		Title:     body.Title,
		Body:      body.Body,
		Scope:     body.Scope,
		TargetID:  target,
		CreatedBy: gmNameFromContext(ctx),
		Active:    true,
	}
	if err := s.dbClient.Vampire().CreateNotification(ctx, notif); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.dbClient.Vampire().UpdateGameState(ctx, map[string]interface{}{
		"active_notification_id": notif.ID,
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logGM(ctx, "push_notification", map[string]interface{}{
		"scope":    body.Scope,
		"targetId": body.TargetID,
		"title":    body.Title,
	})
	ctx.JSON(http.StatusOK, gin.H{"id": notif.ID})
}

// POST /gm/notifications/clear — dismiss the active broadcast for everyone.
func (s *server) gmClearNotifications(ctx *gin.Context) {
	if err := s.dbClient.Vampire().DeactivateAllNotifications(ctx); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.dbClient.Vampire().UpdateGameState(ctx, map[string]interface{}{
		"active_notification_id": nil,
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.logGM(ctx, "clear_notifications", map[string]interface{}{})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
