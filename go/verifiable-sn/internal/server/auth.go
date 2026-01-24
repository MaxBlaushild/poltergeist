package server

import (
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/gin-gonic/gin"
)

func (s *server) login(ctx *gin.Context) {
	var requestBody auth.LoginByTextRequest

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Format phone number to ensure it starts with +
	requestBody.PhoneNumber = formatPhoneNumber(requestBody.PhoneNumber)

	authenticateResponse, err := s.authClient.LoginByText(ctx, &requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"user":  authenticateResponse.User,
		"token": authenticateResponse.Token,
	})
}

func (s *server) register(ctx *gin.Context) {
	var requestBody auth.RegisterByTextRequest

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Format phone number to ensure it starts with +
	requestBody.PhoneNumber = formatPhoneNumber(requestBody.PhoneNumber)

	// Set name to empty string if not provided
	if requestBody.Name == "" {
		requestBody.Name = ""
	}

	// Validate username if provided
	if requestBody.Username != nil && !util.ValidateUsername(*requestBody.Username) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid username",
		})
		return
	}

	authenticateResponse, err := s.authClient.RegisterByText(ctx, &requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Refresh user from database to get latest data including profile picture URL
	refreshedUser, err := s.dbClient.User().FindByID(ctx, authenticateResponse.User.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"user":  refreshedUser,
		"token": authenticateResponse.Token,
	})
}

// formatPhoneNumber ensures phone number starts with +
func formatPhoneNumber(phoneNumber string) string {
	// Remove all non-digit characters except +
	cleaned := strings.ReplaceAll(phoneNumber, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")

	// If it doesn't start with +, add it
	if !strings.HasPrefix(cleaned, "+") {
		cleaned = "+" + cleaned
	}

	return cleaned
}
