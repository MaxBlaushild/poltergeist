package server

import (
	"errors"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

func (s *server) GetAuthenticatedUser(ctx *gin.Context) (*models.User, error) {
	// First, check if user is already set in context by middleware
	if user, exists := ctx.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			return u, nil
		}
	}

	// If not in context, verify token from header
	authorizationHeader := ctx.GetHeader("Authorization")
	if authorizationHeader == "" {
		return nil, errors.New("authorization header required")
	}

	// Extract token from "Bearer {token}" format
	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 {
		return nil, errors.New("invalid authorization header format")
	}

	token := headerParts[1]
	user, err := s.authClient.VerifyToken(ctx, &auth.VerifyTokenRequest{
		Token: token,
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

