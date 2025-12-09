package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	// Check for teamId query parameter and add user to team if provided
	teamIdStr := ctx.Query("teamId")
	if teamIdStr != "" {
		teamID, err := uuid.Parse(teamIdStr)
		if err == nil {
			// Add user to team, but don't fail login if this fails
			if err := s.dbClient.FeteTeam().AddUserToTeam(ctx, teamID, authenticateResponse.User.ID); err != nil {
				log.Printf("Warning: Failed to add user %s to team %s during login: %v", authenticateResponse.User.ID, teamID, err)
				// Continue with login even if team join fails
			}
		} else {
			log.Printf("Warning: Invalid teamId query parameter during login: %s", teamIdStr)
		}
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

	// For final-fete, registration only requires phone number and code
	// Set name to empty string if not provided
	if requestBody.Name == "" {
		requestBody.Name = ""
	}

	authenticateResponse, err := s.authClient.RegisterByText(ctx, &requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check for teamId query parameter and add user to team if provided
	teamIdStr := ctx.Query("teamId")
	if teamIdStr != "" {
		teamID, err := uuid.Parse(teamIdStr)
		if err == nil {
			// Add user to team, but don't fail registration if this fails
			if err := s.dbClient.FeteTeam().AddUserToTeam(ctx, teamID, authenticateResponse.User.ID); err != nil {
				log.Printf("Warning: Failed to add user %s to team %s during registration: %v", authenticateResponse.User.ID, teamID, err)
				// Continue with registration even if team join fails
			}
		} else {
			log.Printf("Warning: Invalid teamId query parameter during registration: %s", teamIdStr)
		}
	}

	ctx.JSON(200, gin.H{
		"user":  authenticateResponse.User,
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

