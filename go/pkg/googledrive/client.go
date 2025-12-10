package googledrive

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type ClientConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type client struct {
	config              *oauth2.Config
	dbClient            db.DbClient
	serviceAccountEmail string // Optional: for granting permissions to Travel Angels
}

type Client interface {
	GetAuthURL(state string) (string, error)
	ExchangeCode(ctx context.Context, code string) (*TokenResponse, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
	ListFiles(ctx context.Context, userID string, pageSize int, pageToken string, query string) (*FileListResponse, error)
	GetFile(ctx context.Context, userID string, fileID string) (*File, error)
	DownloadFile(ctx context.Context, userID string, fileID string) ([]byte, error)
	ExportFileAsHTML(ctx context.Context, userID string, fileID string) ([]byte, error)
	ShareFile(ctx context.Context, userID string, fileID string, email string, role string) error
	CreatePermission(ctx context.Context, userID string, fileID string, permissionType string) error
	GetToken(ctx context.Context, userID string) (*models.GoogleDriveToken, error)
}

func NewClient(config ClientConfig, dbClient db.DbClient) Client {
	if config.RedirectURI == "" {
		panic("Google Drive RedirectURI is required")
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURI,
		Scopes: []string{
			drive.DriveScope,
		},
		Endpoint: google.Endpoint,
	}

	return &client{
		config:   oauthConfig,
		dbClient: dbClient,
	}
}

func (c *client) GetAuthURL(state string) (string, error) {
	if c.config.RedirectURL == "" {
		return "", fmt.Errorf("Google Drive RedirectURI is not configured")
	}
	return c.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce), nil
}

func (c *client) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
	token, err := c.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	if !token.Valid() {
		return nil, fmt.Errorf("invalid token received")
	}

	expiresAt := time.Now().Add(time.Duration(token.Expiry.Unix()-time.Now().Unix()) * time.Second)
	if token.Expiry.After(time.Now()) {
		expiresAt = token.Expiry
	}

	return &TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    int(token.Expiry.Sub(time.Now()).Seconds()),
		ExpiresAt:    expiresAt,
		TokenType:    token.TokenType,
	}, nil
}

func (c *client) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	tokenSource := c.config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(newToken.Expiry.Unix()-time.Now().Unix()) * time.Second)
	if newToken.Expiry.After(time.Now()) {
		expiresAt = newToken.Expiry
	}

	return &TokenResponse{
		AccessToken:  newToken.AccessToken,
		RefreshToken: refreshToken, // Refresh token doesn't change
		ExpiresIn:    int(newToken.Expiry.Sub(time.Now()).Seconds()),
		ExpiresAt:    expiresAt,
		TokenType:    newToken.TokenType,
	}, nil
}

// dbTokenSource wraps an oauth2 TokenSource and updates the database when tokens are refreshed
type dbTokenSource struct {
	baseSource    oauth2.TokenSource
	ctx           context.Context
	userID        string
	dbClient      db.DbClient
	originalToken *models.GoogleDriveToken
}

func (ts *dbTokenSource) Token() (*oauth2.Token, error) {
	// Get token from base source (this will auto-refresh if needed)
	token, err := ts.baseSource.Token()
	if err != nil {
		log.Printf("[dbTokenSource] Failed to get token: %v", err)
		return nil, err
	}

	// Check if token was refreshed (new access token or expiry changed significantly)
	// Use a 1 second tolerance for expiry comparison to account for time precision
	expiryDiff := token.Expiry.Sub(ts.originalToken.ExpiresAt)
	wasRefreshed := token.AccessToken != ts.originalToken.AccessToken ||
		expiryDiff > time.Second || expiryDiff < -time.Second

	if wasRefreshed {
		log.Printf("[dbTokenSource] Token was refreshed, updating database...")
		log.Printf("[dbTokenSource] Old expiry: %v, New expiry: %v", ts.originalToken.ExpiresAt, token.Expiry)

		// Update the original token with new values
		ts.originalToken.AccessToken = token.AccessToken
		ts.originalToken.ExpiresAt = token.Expiry
		ts.originalToken.TokenType = token.TokenType
		ts.originalToken.UpdatedAt = time.Now()

		// Update refresh token if a new one was provided
		if token.RefreshToken != "" && token.RefreshToken != ts.originalToken.RefreshToken {
			log.Printf("[dbTokenSource] Refresh token was also updated")
			ts.originalToken.RefreshToken = token.RefreshToken
		}

		// Persist to database
		if err := ts.dbClient.GoogleDriveToken().Update(ts.ctx, ts.originalToken); err != nil {
			log.Printf("[dbTokenSource] Failed to update token in database: %v", err)
			// Don't fail the request if DB update fails, but log it
			// The token is still valid for this request
		} else {
			log.Printf("[dbTokenSource] Successfully updated token in database")
		}
	}

	return token, nil
}

func (c *client) getDriveService(ctx context.Context, userID string) (*drive.Service, error) {
	log.Printf("[GoogleDriveClient.getDriveService] Getting token for userID: %s", userID)
	token, err := c.GetToken(ctx, userID)
	if err != nil {
		log.Printf("[GoogleDriveClient.getDriveService] Failed to get token: %v", err)
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	log.Printf("[GoogleDriveClient.getDriveService] Successfully retrieved token, expires at: %v", token.ExpiresAt)

	// Create oauth2 token from database token
	oauthToken := &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.ExpiresAt,
		TokenType:    token.TokenType,
	}

	// Create base token source that will handle automatic refresh
	baseTokenSource := c.config.TokenSource(ctx, oauthToken)

	// Wrap it with our dbTokenSource to persist refreshes
	dbTokenSource := &dbTokenSource{
		baseSource:    baseTokenSource,
		ctx:           ctx,
		userID:        userID,
		dbClient:      c.dbClient,
		originalToken: token,
	}

	log.Printf("[GoogleDriveClient.getDriveService] Creating drive service with auto-refresh token source...")
	driveService, err := drive.NewService(ctx, option.WithTokenSource(dbTokenSource))
	if err != nil {
		log.Printf("[GoogleDriveClient.getDriveService] Failed to create drive service: %v", err)
		return nil, fmt.Errorf("failed to create drive service: %w", err)
	}
	log.Printf("[GoogleDriveClient.getDriveService] Successfully created drive service")

	return driveService, nil
}

func (c *client) ListFiles(ctx context.Context, userID string, pageSize int, pageToken string, query string) (*FileListResponse, error) {
	log.Printf("[GoogleDriveClient.ListFiles] Starting - userID: %s, pageSize: %d, pageToken: %s, query: %s", userID, pageSize, pageToken, query)

	driveService, err := c.getDriveService(ctx, userID)
	if err != nil {
		log.Printf("[GoogleDriveClient.ListFiles] Failed to get drive service: %v", err)
		return nil, err
	}
	log.Printf("[GoogleDriveClient.ListFiles] Successfully created drive service")

	listCall := driveService.Files.List().
		PageSize(int64(pageSize)).
		Fields("nextPageToken, files(id, name, mimeType, size, createdTime, modifiedTime, webViewLink, webContentLink, owners, shared)")

	if query != "" {
		listCall = listCall.Q(query)
		log.Printf("[GoogleDriveClient.ListFiles] Added query filter: %s", query)
	}

	if pageToken != "" {
		listCall = listCall.PageToken(pageToken)
		log.Printf("[GoogleDriveClient.ListFiles] Added page token")
	}

	log.Printf("[GoogleDriveClient.ListFiles] Executing Google Drive API call...")
	files, err := listCall.Do()
	if err != nil {
		log.Printf("[GoogleDriveClient.ListFiles] Google Drive API error: %v", err)
		log.Printf("[GoogleDriveClient.ListFiles] Error type: %T", err)
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	log.Printf("[GoogleDriveClient.ListFiles] Successfully received %d files from Google Drive", len(files.Files))

	result := &FileListResponse{
		Files:         make([]File, 0, len(files.Files)),
		NextPageToken: files.NextPageToken,
	}

	for _, file := range files.Files {
		createdTime, _ := time.Parse(time.RFC3339, file.CreatedTime)
		modifiedTime, _ := time.Parse(time.RFC3339, file.ModifiedTime)

		owners := make([]Owner, 0, len(file.Owners))
		for _, owner := range file.Owners {
			owners = append(owners, Owner{
				DisplayName:  owner.DisplayName,
				EmailAddress: owner.EmailAddress,
			})
		}

		result.Files = append(result.Files, File{
			ID:             file.Id,
			Name:           file.Name,
			MimeType:       file.MimeType,
			Size:           file.Size,
			CreatedTime:    createdTime,
			ModifiedTime:   modifiedTime,
			WebViewLink:    file.WebViewLink,
			WebContentLink: file.WebContentLink,
			Owners:         owners,
			Shared:         file.Shared,
		})
	}

	return result, nil
}

func (c *client) GetFile(ctx context.Context, userID string, fileID string) (*File, error) {
	driveService, err := c.getDriveService(ctx, userID)
	if err != nil {
		return nil, err
	}

	file, err := driveService.Files.Get(fileID).
		Fields("id, name, mimeType, size, createdTime, modifiedTime, webViewLink, webContentLink, owners, shared").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	createdTime, _ := time.Parse(time.RFC3339, file.CreatedTime)
	modifiedTime, _ := time.Parse(time.RFC3339, file.ModifiedTime)

	owners := make([]Owner, 0, len(file.Owners))
	for _, owner := range file.Owners {
		owners = append(owners, Owner{
			DisplayName:  owner.DisplayName,
			EmailAddress: owner.EmailAddress,
		})
	}

	return &File{
		ID:             file.Id,
		Name:           file.Name,
		MimeType:       file.MimeType,
		Size:           file.Size,
		CreatedTime:    createdTime,
		ModifiedTime:   modifiedTime,
		WebViewLink:    file.WebViewLink,
		WebContentLink: file.WebContentLink,
		Owners:         owners,
		Shared:         file.Shared,
	}, nil
}

func (c *client) DownloadFile(ctx context.Context, userID string, fileID string) ([]byte, error) {
	driveService, err := c.getDriveService(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp, err := driveService.Files.Get(fileID).Download()
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return data, nil
}

func (c *client) ExportFileAsHTML(ctx context.Context, userID string, fileID string) ([]byte, error) {
	driveService, err := c.getDriveService(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Export Google Docs/Sheets as HTML
	resp, err := driveService.Files.Export(fileID, "text/html").Download()
	if err != nil {
		return nil, fmt.Errorf("failed to export file as HTML: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read exported file content: %w", err)
	}

	return data, nil
}

func (c *client) ShareFile(ctx context.Context, userID string, fileID string, email string, role string) error {
	driveService, err := c.getDriveService(ctx, userID)
	if err != nil {
		return err
	}

	// Validate role
	validRoles := map[string]bool{"reader": true, "writer": true, "commenter": true}
	if !validRoles[role] {
		return fmt.Errorf("invalid role: %s. Must be reader, writer, or commenter", role)
	}

	permission := &drive.Permission{
		Type:         "user",
		Role:         role,
		EmailAddress: email,
	}

	_, err = driveService.Permissions.Create(fileID, permission).
		SendNotificationEmail(false).
		Do()
	if err != nil {
		return fmt.Errorf("failed to share file: %w", err)
	}

	return nil
}

func (c *client) CreatePermission(ctx context.Context, userID string, fileID string, permissionType string) error {
	driveService, err := c.getDriveService(ctx, userID)
	if err != nil {
		return err
	}

	if c.serviceAccountEmail == "" {
		return fmt.Errorf("service account email not configured")
	}

	permission := &drive.Permission{
		Type:         permissionType, // "user" or "domain"
		Role:         "reader",
		EmailAddress: c.serviceAccountEmail,
	}

	_, err = driveService.Permissions.Create(fileID, permission).
		SendNotificationEmail(false).
		Do()
	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

func (c *client) GetToken(ctx context.Context, userID string) (*models.GoogleDriveToken, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	token, err := c.dbClient.GoogleDriveToken().FindByUserID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to find token: %w", err)
	}

	return token, nil
}
