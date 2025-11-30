package server

import (
	"net/http"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *server) hueOAuthCallback(ctx *gin.Context) {
	code := ctx.Query("code")
	_ = ctx.Query("state") // State parameter (could be validated in future)

	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "authorization code is required",
		})
		return
	}

	if s.hueOAuthClient == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Hue OAuth client not configured",
		})
		return
	}

	// Exchange code for tokens
	tokenResp, err := s.hueOAuthClient.ExchangeCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to exchange authorization code: " + err.Error(),
		})
		return
	}

	// Store tokens in database
	token := &models.HueToken{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		UserID:       nil, // No user in final-fete context
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    tokenResp.ExpiresAt,
		TokenType:    tokenResp.TokenType,
	}

	// Check if there's an existing token (without user), update it instead
	existingToken, err := s.dbClient.HueToken().FindLatest(ctx)
	if err == nil && existingToken != nil && existingToken.UserID == nil {
		// Update existing token
		existingToken.AccessToken = token.AccessToken
		existingToken.RefreshToken = token.RefreshToken
		existingToken.ExpiresAt = token.ExpiresAt
		existingToken.TokenType = token.TokenType
		existingToken.UpdatedAt = time.Now()

		if err := s.dbClient.HueToken().Update(ctx, existingToken); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to update token: " + err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message":      "Hue OAuth authorization successful - token updated",
			"refreshToken": tokenResp.RefreshToken,
			"expiresAt":    tokenResp.ExpiresAt,
		})
		return
	}

	// Create new token
	if err := s.dbClient.HueToken().Create(ctx, token); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to store token: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "Hue OAuth authorization successful",
		"refreshToken": tokenResp.RefreshToken,
		"expiresAt":    tokenResp.ExpiresAt,
	})
}

