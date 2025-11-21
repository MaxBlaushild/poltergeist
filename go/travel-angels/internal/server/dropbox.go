package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ShareDropboxFileRequest struct {
	Email string `json:"email" binding:"required"`
	Role  string `json:"role" binding:"required"` // viewer, editor
}

func (s *server) GetDropboxAuth(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Generate state with userID for CSRF protection
	state := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("user:%s:time:%d", user.ID.String(), time.Now().Unix())))

	authURL := s.dropboxClient.GetAuthURL(state)

	ctx.JSON(http.StatusOK, gin.H{
		"authUrl": authURL,
		"state":   state,
	})
}

func (s *server) DropboxCallback(ctx *gin.Context) {
	code := ctx.Query("code")
	state := ctx.Query("state")

	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "authorization code is required",
		})
		return
	}

	if state == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "state parameter is required",
		})
		return
	}

	// Decode state to get userID
	decodedState, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid state parameter",
		})
		return
	}

	// Extract userID from state (format: "user:UUID:time:timestamp")
	stateStr := string(decodedState)
	var userID uuid.UUID
	var userIDStr string

	// Parse state to extract userID
	if strings.HasPrefix(stateStr, "user:") {
		parts := strings.Split(stateStr, ":")
		if len(parts) >= 2 {
			userIDStr = parts[1]
		}
	}

	if userIDStr == "" {
		// Fallback: try to get from query param
		userIDStr = ctx.Query("user_id")
	}

	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required in state or query parameter",
		})
		return
	}

	userID, err = uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user_id format",
		})
		return
	}

	// Verify state contains userID (basic validation)
	if !contains(stateStr, userID.String()) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid state - user mismatch",
		})
		return
	}

	// Exchange code for tokens
	tokenResp, err := s.dropboxClient.ExchangeCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to exchange authorization code: " + err.Error(),
		})
		return
	}

	// Store tokens in database
	token := &models.DropboxToken{
		ID:           uuid.New(),
		UserID:       userID,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    tokenResp.ExpiresAt,
		TokenType:    tokenResp.TokenType,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.dbClient.DropboxToken().Create(ctx, token); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to store tokens: " + err.Error(),
		})
		return
	}

	// Redirect to frontend success page or return JSON
	redirectURI := ctx.Query("redirect_uri")
	if redirectURI != "" {
		ctx.Redirect(http.StatusFound, redirectURI+"?success=true")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Dropbox authorization successful",
		"userId":  userID.String(),
	})
}

func (s *server) RevokeDropbox(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Delete tokens from database
	if err := s.dbClient.DropboxToken().Delete(ctx, user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to revoke access: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Dropbox access revoked successfully",
	})
}

func (s *server) ShareDropboxFile(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	filePath := ctx.Param("path")
	if filePath == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "path is required",
		})
		return
	}

	// Decode URL-encoded path
	decodedPath, err := url.PathUnescape(filePath)
	if err != nil {
		decodedPath = filePath // Fallback to original if decode fails
	}

	var req ShareDropboxFileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate email format (basic)
	if req.Email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "email is required",
		})
		return
	}

	// Share file
	if err := s.dropboxClient.ShareFile(ctx, user.ID.String(), decodedPath, req.Email, req.Role); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to share file: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "File shared successfully",
		"path":    decodedPath,
		"email":   req.Email,
		"role":    req.Role,
	})
}

func (s *server) CreateDropboxSharedLink(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	filePath := ctx.Param("path")
	if filePath == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "path is required",
		})
		return
	}

	// Decode URL-encoded path
	decodedPath, err := url.PathUnescape(filePath)
	if err != nil {
		decodedPath = filePath // Fallback to original if decode fails
	}

	// Create shared link
	sharedLink, err := s.dropboxClient.CreateSharedLink(ctx, user.ID.String(), decodedPath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create shared link: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "Shared link created successfully",
		"path":       decodedPath,
		"sharedLink": sharedLink,
	})
}

func (s *server) ListDropboxFiles(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Parse query parameters
	path := ctx.Query("path")
	if path == "" {
		path = "" // Default to root
	}

	recursive := false
	if recursiveStr := ctx.Query("recursive"); recursiveStr != "" {
		if parsed, err := strconv.ParseBool(recursiveStr); err == nil {
			recursive = parsed
		}
	}

	// List files
	files, err := s.dropboxClient.ListFiles(ctx, user.ID.String(), path, recursive)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list files: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, files)
}
