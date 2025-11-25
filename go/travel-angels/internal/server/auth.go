package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/gin-gonic/gin"
)

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

	ctx.JSON(200, gin.H{
		"user":  authenticateResponse.User,
		"token": authenticateResponse.Token,
	})
}

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

	payload := gin.H{
		"user":  authenticateResponse.User,
		"token": authenticateResponse.Token,
	}

	ctx.JSON(200, payload)
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

func (s *server) GetPresignedUploadUrl(ctx *gin.Context) {
	var requestBody struct {
		Bucket string `binding:"required" json:"bucket"`
		Key    string `binding:"required" json:"key"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	url, err := s.awsClient.GeneratePresignedUploadURL(requestBody.Bucket, requestBody.Key, time.Hour)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}

func (s *server) ValidateUsername(ctx *gin.Context) {
	username := ctx.Query("username")
	if username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "username is required",
		})
		return
	}

	// Validate username format
	if !util.ValidateUsername(username) {
		ctx.JSON(http.StatusOK, gin.H{
			"valid":   false,
			"message": "Invalid username format. Username must contain only letters and numbers.",
		})
		return
	}

	// Check uniqueness
	user, err := s.dbClient.User().FindByUsername(ctx, username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if user != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"valid":   false,
			"message": "Username already taken.",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"valid": true,
	})
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
		Username          *string  `json:"username"`
		ProfilePictureUrl *string  `json:"profilePictureUrl"`
		DateOfBirth       *string  `json:"dateOfBirth"`
		Gender            *string  `json:"gender"`
		Latitude          *float64 `json:"latitude"`
		Longitude         *float64 `json:"longitude"`
		LocationAddress   *string  `json:"locationAddress"`
		Bio               *string  `json:"bio"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Update username if provided
	if requestBody.Username != nil {
		if !util.ValidateUsername(*requestBody.Username) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid username",
			})
			return
		}

		// Check uniqueness (excluding current user)
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

		if err := s.dbClient.User().SetUsername(ctx, user.ID, *requestBody.Username); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	// Update profile picture URL if provided
	if requestBody.ProfilePictureUrl != nil {
		if err := s.dbClient.User().UpdateProfilePictureUrl(ctx, user.ID, *requestBody.ProfilePictureUrl); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	// Update demographic and location fields if provided
	updates := models.User{}
	needsUpdate := false

	if requestBody.DateOfBirth != nil {
		parsed, err := time.Parse("2006-01-02", *requestBody.DateOfBirth)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid date of birth format",
			})
			return
		}
		updates.DateOfBirth = &parsed
		needsUpdate = true
	}

	if requestBody.Gender != nil {
		updates.Gender = requestBody.Gender
		needsUpdate = true
	}

	if requestBody.Latitude != nil {
		updates.Latitude = requestBody.Latitude
		needsUpdate = true
	}

	if requestBody.Longitude != nil {
		updates.Longitude = requestBody.Longitude
		needsUpdate = true
	}

	if requestBody.LocationAddress != nil {
		updates.LocationAddress = requestBody.LocationAddress
		needsUpdate = true
	}

	if requestBody.Bio != nil {
		updates.Bio = requestBody.Bio
		needsUpdate = true
	}

	if needsUpdate {
		if err := s.dbClient.User().Update(ctx, user.ID, updates); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	// Fetch updated user
	updatedUser, err := s.dbClient.User().FindByID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, updatedUser)
}
