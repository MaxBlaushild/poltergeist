package server

import (
	"net/http"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateQuickDecisionRequestRequest struct {
	Question string  `json:"question" binding:"required"`
	Option1  string  `json:"option1" binding:"required"`
	Option2  string  `json:"option2" binding:"required"`
	Option3  *string `json:"option3"`
}

func (s *server) CreateQuickDecisionRequest(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var req CreateQuickDecisionRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate that question and at least 2 options are provided
	if req.Question == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "question is required",
		})
		return
	}

	if req.Option1 == "" || req.Option2 == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "at least option1 and option2 are required",
		})
		return
	}

	// Create quick decision request model
	request := &models.QuickDecisionRequest{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		Question:  req.Question,
		Option1:   req.Option1,
		Option2:   req.Option2,
		Option3:   req.Option3,
	}

	// Save via database handler
	createdRequest, err := s.dbClient.QuickDecisionRequest().Create(ctx, request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, createdRequest)
}
