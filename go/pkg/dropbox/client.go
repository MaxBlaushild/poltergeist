package dropbox

import (
	"context"
	"fmt"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"golang.org/x/oauth2"
)

type ClientConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type client struct {
	config   *oauth2.Config
	dbClient db.DbClient
}

type Client interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*TokenResponse, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
	ListFiles(ctx context.Context, userID string, path string, recursive bool) (*FileListResponse, error)
	GetFile(ctx context.Context, userID string, filePath string) (*File, error)
	DownloadFile(ctx context.Context, userID string, filePath string) ([]byte, error)
	ShareFile(ctx context.Context, userID string, filePath string, email string, role string) error
	CreateSharedLink(ctx context.Context, userID string, filePath string) (string, error)
	GetToken(ctx context.Context, userID string) (*models.DropboxToken, error)
}

func NewClient(config ClientConfig, dbClient db.DbClient) Client {
	// Dropbox OAuth 2.0 endpoint
	dropboxEndpoint := oauth2.Endpoint{
		AuthURL:  "https://www.dropbox.com/oauth2/authorize",
		TokenURL: "https://api.dropbox.com/oauth2/token",
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURI,
		Scopes: []string{
			"files.content.read",
			"files.content.write",
			"sharing.write",
		},
		Endpoint: dropboxEndpoint,
	}

	return &client{
		config:   oauthConfig,
		dbClient: dbClient,
	}
}

func (c *client) GetAuthURL(state string) string {
	return c.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("token_access_type", "offline"))
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

func (c *client) getDropboxClient(ctx context.Context, userID string) error {
	return nil
	// token, err := c.GetToken(ctx, userID)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get token: %w", err)
	// }

	// // Check if token needs refresh
	// if time.Now().After(token.ExpiresAt.Add(-5 * time.Minute)) {
	// 	tokenResp, err := c.RefreshAccessToken(ctx, token.RefreshToken)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to refresh token: %w", err)
	// 	}

	// 	// Update token in database
	// 	token.AccessToken = tokenResp.AccessToken
	// 	token.ExpiresAt = tokenResp.ExpiresAt
	// 	if err := c.dbClient.DropboxToken().Update(ctx, token); err != nil {
	// 		return nil, fmt.Errorf("failed to update token: %w", err)
	// 	}
	// }

	// // Create Dropbox client with access token
	// config := dropboxclient.Config{
	// 	Token: token.AccessToken,
	// }

	// return dropboxclient.Client(config), nil
}

func (c *client) ListFiles(ctx context.Context, userID string, path string, recursive bool) (*FileListResponse, error) {
	// config, err := c.getDropboxClient(ctx, userID)
	// if err != nil {
	// 	return nil, err
	// }

	// dbx := dropboxclient.New(config)
	// filesClient := files.New(dbx)

	// listArg := files.NewListFolderArg(path)
	// if recursive {
	// 	listArg.Recursive = true
	// }

	// result, err := filesClient.ListFolder(listArg)
	// if err != nil {
	// 	// If path is not found, try listing root
	// 	if path != "" {
	// 		listArg.Path = ""
	// 		result, err = filesClient.ListFolder(listArg)
	// 		if err != nil {
	// 			return nil, fmt.Errorf("failed to list files: %w", err)
	// 		}
	// 	} else {
	// 		return nil, fmt.Errorf("failed to list files: %w", err)
	// 	}
	// }

	// filesList := make([]File, 0, len(result.Entries))
	// for _, entry := range result.Entries {
	// 	file := File{
	// 		ID:          entry.Id,
	// 		Name:        entry.Name,
	// 		PathLower:   entry.PathLower,
	// 		PathDisplay: entry.PathDisplay,
	// 		Rev:         entry.Rev,
	// 	}

	// 	switch meta := entry.(type) {
	// 	case *files.FileMetadata:
	// 		file.IsFolder = false
	// 		file.Size = int64(meta.Size)
	// 		if meta.ClientModified != nil {
	// 			file.ClientModified = time.Time(*meta.ClientModified)
	// 		}
	// 		if meta.ServerModified != nil {
	// 			file.ServerModified = time.Time(*meta.ServerModified)
	// 		}
	// 		if meta.ContentHash != nil {
	// 			file.ContentHash = *meta.ContentHash
	// 		}
	// 	case *files.FolderMetadata:
	// 		file.IsFolder = true
	// 	}

	// 	filesList = append(filesList, file)
	// }

	// return &FileListResponse{
	// 	Files:   filesList,
	// 	HasMore: result.HasMore,
	// 	Cursor:  result.Cursor,
	// }, nil
	return nil, nil
}

func (c *client) GetFile(ctx context.Context, userID string, filePath string) (*File, error) {
	// config, err := c.getDropboxClient(ctx, userID)
	// if err != nil {
	// 	return nil, err
	// }

	// dbx := dropboxclient.New(config)
	// filesClient := files.New(dbx)

	// metadata, err := filesClient.GetMetadata(files.NewGetMetadataArg(filePath))
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get file metadata: %w", err)
	// }

	// file := File{
	// 	ID:          metadata.Id,
	// 	Name:        metadata.Name,
	// 	PathLower:   metadata.PathLower,
	// 	PathDisplay: metadata.PathDisplay,
	// 	Rev:         metadata.Rev,
	// }

	// switch meta := metadata.(type) {
	// case *files.FileMetadata:
	// 	file.IsFolder = false
	// 	file.Size = int64(meta.Size)
	// 	if meta.ClientModified != nil {
	// 		file.ClientModified = time.Time(*meta.ClientModified)
	// 	}
	// 	if meta.ServerModified != nil {
	// 		file.ServerModified = time.Time(*meta.ServerModified)
	// 	}
	// 	if meta.ContentHash != nil {
	// 		file.ContentHash = *meta.ContentHash
	// 	}
	// case *files.FolderMetadata:
	// 	file.IsFolder = true
	// }

	// return &file, nil
	return nil, nil
}

func (c *client) DownloadFile(ctx context.Context, userID string, filePath string) ([]byte, error) {
	// config, err := c.getDropboxClient(ctx, userID)
	// if err != nil {
	// 	return nil, err
	// }

	// dbx := dropboxclient.New(config)
	// filesClient := files.New(dbx)

	// downloadArg := files.NewDownloadArg(filePath)
	// _, content, err := filesClient.Download(downloadArg)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to download file: %w", err)
	// }
	// defer content.Close()

	// data, err := io.ReadAll(content)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to read file content: %w", err)
	// }

	return nil, nil
}

func (c *client) ShareFile(ctx context.Context, userID string, filePath string, email string, role string) error {
	// config, err := c.getDropboxClient(ctx, userID)
	// if err != nil {
	// 	return err
	// }

	// dbx := dropboxclient.New(config)
	// sharingClient := sharing.New(dbx)

	// // Validate role
	// validRoles := map[string]bool{"viewer": true, "editor": true}
	// if !validRoles[role] {
	// 	return fmt.Errorf("invalid role: %s. Must be viewer or editor", role)
	// }

	// // Determine access level
	// accessLevel := sharing.NewAccessLevelViewer()
	// if role == "editor" {
	// 	accessLevel = sharing.NewAccessLevelEditor()
	// }

	// // Create member selector by email
	// memberSelector := sharing.NewMemberSelectorEmail(email)

	// // Create add member request
	// addMember := sharing.NewAddMember(memberSelector)
	// addMember.AccessLevel = accessLevel

	// // Try to share as folder first
	// addFolderMemberArg := sharing.NewAddFolderMemberArg(filePath, []*sharing.AddMember{addMember}, false)

	// err = sharingClient.AddFolderMember(addFolderMemberArg)
	// if err != nil {
	// 	// If it's not a folder, try to share the folder that contains it
	// 	// For now, return the error
	// 	return fmt.Errorf("failed to share file/folder: %w", err)
	// }

	return nil
}

func (c *client) CreateSharedLink(ctx context.Context, userID string, filePath string) (string, error) {
	// config, err := c.getDropboxClient(ctx, userID)
	// if err != nil {
	// 	return "", err
	// }

	// dbx := dropboxclient.New(config)
	// sharingClient := sharing.New(dbx)

	// createSharedLinkArg := sharing.NewCreateSharedLinkArg(filePath)
	// createSharedLinkArg.ShortUrl = &[]bool{false}[0]

	// sharedLinkMetadata, err := sharingClient.CreateSharedLink(createSharedLinkArg)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to create shared link: %w", err)
	// }

	// return sharedLinkMetadata.Url, nil
	return "", nil
}

func (c *client) GetToken(ctx context.Context, userID string) (*models.DropboxToken, error) {
	// userUUID, err := uuid.Parse(userID)
	// if err != nil {
	// 	return nil, fmt.Errorf("invalid user ID: %w", err)
	// }

	// token, err := c.dbClient.DropboxToken().FindByUserID(ctx, userUUID)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to find token: %w", err)
	// }

	// return token, nil
	return nil, nil
}
