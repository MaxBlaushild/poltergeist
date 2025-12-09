package hue

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

// OAuthClientConfig holds OAuth configuration
type OAuthClientConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// OAuthClient provides OAuth2 authentication for Philips Hue cloud API
type OAuthClient interface {
	GetAuthURL(state string) (string, error)
	ExchangeCode(ctx context.Context, code string) (*TokenResponse, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
}

// TokenResponse represents OAuth token response
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

type oauthClient struct {
	config *oauth2.Config
}

// NewOAuthClient creates a new OAuth client for Philips Hue
func NewOAuthClient(config OAuthClientConfig) OAuthClient {
	if config.RedirectURI == "" {
		panic("Hue OAuth RedirectURI is required")
	}

	// Philips Hue OAuth endpoints
	hueEndpoint := oauth2.Endpoint{
		AuthURL:  "https://api.meethue.com/v2/oauth2/authorize",
		TokenURL: "https://api.meethue.com/v2/oauth2/token",
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURI,
		Scopes: []string{
			"hue:api", // Standard Hue API scope
		},
		Endpoint: hueEndpoint,
	}

	return &oauthClient{
		config: oauthConfig,
	}
}

func (c *oauthClient) GetAuthURL(state string) (string, error) {
	if c.config.RedirectURL == "" {
		return "", fmt.Errorf("Hue OAuth RedirectURI is not configured")
	}
	return c.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce), nil
}

func (c *oauthClient) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
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
		ExpiresIn:    int(time.Until(token.Expiry).Seconds()),
		ExpiresAt:    expiresAt,
		TokenType:    token.TokenType,
	}, nil
}

func (c *oauthClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	// Validate refresh token
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token is empty")
	}

	// Create token source with refresh token
	tokenSource := c.config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	// Get new token (this will automatically use the refresh token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Validate the new token
	if newToken.AccessToken == "" {
		return nil, fmt.Errorf("received empty access token from refresh")
	}

	// Calculate expiry time
	expiresAt := newToken.Expiry
	if expiresAt.IsZero() || !expiresAt.After(time.Now()) {
		// If expiry is not set or already expired, set a default (8 hours from now)
		expiresAt = time.Now().Add(8 * time.Hour)
	}

	// Use new refresh token if provided, otherwise keep the old one
	// According to Hue docs, each refresh may return a new refresh token
	newRefreshToken := refreshToken
	if newToken.RefreshToken != "" && newToken.RefreshToken != refreshToken {
		newRefreshToken = newToken.RefreshToken
	}

	tokenType := newToken.TokenType
	if tokenType == "" {
		tokenType = "Bearer" // Default token type
	}

	return &TokenResponse{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int(time.Until(expiresAt).Seconds()),
		ExpiresAt:    expiresAt,
		TokenType:    tokenType,
	}, nil
}
