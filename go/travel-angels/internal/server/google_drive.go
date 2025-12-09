package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/travel-angels/internal/parser"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ShareFileRequest struct {
	Email string `json:"email" binding:"required"`
	Role  string `json:"role" binding:"required"` // reader, writer, commenter
}

type GrantPermissionRequest struct {
	PermissionType string `json:"permissionType" binding:"required"` // user, domain
}

type ImportDocumentRequest struct {
	FileID     string                    `json:"fileId" binding:"required"`
	ImportType string                    `json:"importType" binding:"required"` // "import" or "reference"
	Locations  []DocumentLocationRequest `json:"locations"`
}

func (s *server) GetGoogleDriveStatus(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check if user has a Google Drive token
	_, err = s.googleDriveClient.GetToken(ctx, user.ID.String())
	connected := err == nil

	ctx.JSON(http.StatusOK, gin.H{
		"connected": connected,
	})
}

func (s *server) GetGoogleDriveAuth(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Generate state with userID for CSRF protection
	state := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("user:%s:time:%d", user.ID.String(), time.Now().Unix())))

	authURL, err := s.googleDriveClient.GetAuthURL(state)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate auth URL: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"authUrl": authURL,
		"state":   state,
	})
}

func (s *server) GoogleDriveCallback(ctx *gin.Context) {
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
	tokenResp, err := s.googleDriveClient.ExchangeCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to exchange authorization code: " + err.Error(),
		})
		return
	}

	// Store tokens in database
	token := &models.GoogleDriveToken{
		ID:           uuid.New(),
		UserID:       userID,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    tokenResp.ExpiresAt,
		TokenType:    tokenResp.TokenType,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.dbClient.GoogleDriveToken().Create(ctx, token); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to store tokens: " + err.Error(),
		})
		return
	}

	// Redirect to app or frontend success page
	redirectURI := ctx.Query("redirect_uri")
	if redirectURI != "" {
		ctx.Redirect(http.StatusFound, redirectURI+"?success=true&service=google-drive")
		return
	}

	// Default: redirect to app deep link
	appRedirectURL := fmt.Sprintf("travelangels://oauth-callback?success=true&service=google-drive")
	ctx.Redirect(http.StatusFound, appRedirectURL)
}

func (s *server) RevokeGoogleDrive(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Delete tokens from database
	if err := s.dbClient.GoogleDriveToken().Delete(ctx, user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to revoke access: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Google Drive access revoked successfully",
	})
}

func (s *server) ShareGoogleDriveFile(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	fileID := ctx.Param("fileId")
	if fileID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "fileId is required",
		})
		return
	}

	var req ShareFileRequest
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
	if err := s.googleDriveClient.ShareFile(ctx, user.ID.String(), fileID, req.Email, req.Role); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to share file: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "File shared successfully",
		"fileId":  fileID,
		"email":   req.Email,
		"role":    req.Role,
	})
}

func (s *server) GrantGoogleDrivePermissions(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	fileID := ctx.Param("fileId")
	if fileID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "fileId is required",
		})
		return
	}

	var req GrantPermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Grant permission to Travel Angels service account
	if err := s.googleDriveClient.CreatePermission(ctx, user.ID.String(), fileID, req.PermissionType); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to grant permissions: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":        "Permissions granted to Travel Angels successfully",
		"fileId":         fileID,
		"permissionType": req.PermissionType,
	})
}

func (s *server) ListGoogleDriveFiles(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Parse query parameters
	pageSize := 10 // default
	if pageSizeStr := ctx.Query("pageSize"); pageSizeStr != "" {
		if parsed, err := strconv.Atoi(pageSizeStr); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	pageToken := ctx.Query("pageToken")
	query := ctx.Query("q")

	// List files
	files, err := s.googleDriveClient.ListFiles(ctx, user.ID.String(), pageSize, pageToken, query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list files: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, files)
}

func (s *server) ImportGoogleDriveDocument(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var req ImportDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate importType
	if req.ImportType != "import" && req.ImportType != "reference" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "importType must be 'import' or 'reference'",
		})
		return
	}

	// Get file info from Google Drive
	file, err := s.googleDriveClient.GetFile(ctx, user.ID.String(), req.FileID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get file: " + err.Error(),
		})
		return
	}

	// Determine provider based on mimeType
	var provider models.CloudDocumentProvider
	switch file.MimeType {
	case "application/vnd.google-apps.document":
		provider = models.CloudDocumentProviderGoogleDocs
	case "application/vnd.google-apps.spreadsheet":
		provider = models.CloudDocumentProviderGoogleSheets
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "file must be a Google Doc or Sheet",
		})
		return
	}

	// Create document
	document := &models.Document{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Title:     file.Name,
		Provider:  provider,
		UserID:    user.ID,
		Link:      &file.WebViewLink,
	}

	// If importing, download HTML and convert to markdown
	if req.ImportType == "import" {
		htmlBytes, err := s.googleDriveClient.ExportFileAsHTML(ctx, user.ID.String(), req.FileID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to export file as HTML: " + err.Error(),
			})
			return
		}

		// Convert HTML to markdown
		documentParser := parser.NewDocumentParser()
		parsedDoc, err := documentParser.ParseHTML(htmlBytes)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to convert HTML to markdown: " + err.Error(),
			})
			return
		}

		document.Content = &parsedDoc.Content
	} else {
		// For reference, leave Content as nil
		document.Content = nil
	}

	// Create document in database
	createdDocument, err := s.dbClient.Document().Create(ctx, document, []uuid.UUID{}, []string{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create document: " + err.Error(),
		})
		return
	}

	// Create document locations if provided
	if len(req.Locations) > 0 {
		documentLocationHandler := s.dbClient.DocumentLocation()
		for _, locReq := range req.Locations {
			// Validate location type
			locationType := models.LocationType(locReq.Type)
			switch locationType {
			case models.LocationTypeCity, models.LocationTypeCountry, models.LocationTypeContinent:
				// Valid type
			default:
				// Default to city if invalid
				locationType = models.LocationTypeCity
			}

			location := &models.DocumentLocation{
				DocumentID:       createdDocument.ID,
				PlaceID:          locReq.PlaceID,
				Name:             locReq.Name,
				FormattedAddress: locReq.FormattedAddress,
				Latitude:         locReq.Latitude,
				Longitude:        locReq.Longitude,
				LocationType:     locationType,
			}

			if err := documentLocationHandler.Create(ctx, location); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error": "failed to create document location: " + err.Error(),
				})
				return
			}
		}
	}

	// Reload document with locations
	createdDocument, err = s.dbClient.Document().FindByID(ctx, createdDocument.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve created document",
		})
		return
	}

	ctx.JSON(http.StatusCreated, createdDocument)
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
