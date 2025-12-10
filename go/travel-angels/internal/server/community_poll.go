package server

import (
	"net/http"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateCommunityPollRequest struct {
	Question string   `json:"question" binding:"required"`
	Options  []string `json:"options" binding:"required"`
}

func (s *server) CreateCommunityPoll(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var req CreateCommunityPollRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate that question is provided
	if req.Question == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "question is required",
		})
		return
	}

	// Validate that 3-10 options are provided
	if len(req.Options) < 3 || len(req.Options) > 10 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "options must contain between 3 and 10 items",
		})
		return
	}

	// Validate that all options are non-empty
	for _, option := range req.Options {
		if option == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "all options must be non-empty",
			})
			return
		}
	}

	// Create community poll model
	poll := &models.CommunityPoll{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		Question:  req.Question,
		Options:   models.StringArray(req.Options),
	}

	// Save via database handler
	createdPoll, err := s.dbClient.CommunityPoll().Create(ctx, poll)
	if err != nil {
		// Log the error for debugging
		gin.DefaultErrorWriter.Write([]byte("Error creating community poll: " + err.Error() + "\n"))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create community poll: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, createdPoll)
}
