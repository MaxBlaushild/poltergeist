package server

import (
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
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

func (s *server) UpdateProfile(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var requestBody struct {
		Username         *string `json:"username"`
		ProfilePictureUrl *string `json:"profilePictureUrl"`
		Category         *string `json:"category"`
		AgeRange         *string `json:"ageRange"`
		Bio              *string `json:"bio"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate username if provided
	if requestBody.Username != nil && !util.ValidateUsername(*requestBody.Username) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid username",
		})
		return
	}

	// Check username uniqueness if provided
	if requestBody.Username != nil {
		existingUser, err := s.dbClient.User().FindByUsername(ctx, *requestBody.Username)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		if existingUser != nil && existingUser.ID != user.ID {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "username already taken",
			})
			return
		}
	}

	// Build updates struct
	updates := models.User{}
	needsUpdate := false

	if requestBody.Username != nil {
		updates.Username = requestBody.Username
		needsUpdate = true
	}

	if requestBody.ProfilePictureUrl != nil {
		updates.ProfilePictureUrl = *requestBody.ProfilePictureUrl
		needsUpdate = true
	}

	if requestBody.Category != nil {
		updates.Category = requestBody.Category
		needsUpdate = true
	}

	if requestBody.AgeRange != nil {
		updates.AgeRange = requestBody.AgeRange
		needsUpdate = true
	}

	if requestBody.Bio != nil {
		updates.Bio = requestBody.Bio
		needsUpdate = true
	}

	if !needsUpdate {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "no fields to update",
		})
		return
	}

	// Update user
	if err := s.dbClient.User().Update(ctx, user.ID, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Refresh user from database to get latest data
	updatedUser, err := s.dbClient.User().FindByID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(200, updatedUser)
}
