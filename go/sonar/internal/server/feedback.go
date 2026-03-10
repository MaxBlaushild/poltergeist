package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *server) submitFeedback(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		Message string  `json:"message"`
		Route   string  `json:"route"`
		ZoneID  *string `json:"zoneId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message := strings.TrimSpace(requestBody.Message)
	if message == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	route := strings.TrimSpace(requestBody.Route)
	if route == "" {
		route = "/"
	}

	var zoneID *uuid.UUID
	if requestBody.ZoneID != nil {
		trimmed := strings.TrimSpace(*requestBody.ZoneID)
		if trimmed != "" {
			parsed, err := uuid.Parse(trimmed)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zoneId"})
				return
			}
			zoneID = &parsed
		}
	}

	item := &models.FeedbackItem{
		UserID:  user.ID,
		ZoneID:  zoneID,
		Route:   route,
		Message: message,
	}
	if err := s.dbClient.FeedbackItem().Create(ctx, item); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

func (s *server) listFeedbackItems(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.Query("limit"))

	items, err := s.dbClient.FeedbackItem().ListRecent(ctx, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, items)
}
