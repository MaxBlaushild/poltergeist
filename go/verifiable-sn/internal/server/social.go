package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

const (
	socialProviderInstagram  = "instagram"
	socialProviderTwitter    = "twitter"
	defaultInstagramAuthURL  = "https://www.facebook.com/v19.0/dialog/oauth"
	defaultInstagramTokenURL = "https://graph.facebook.com/v19.0/oauth/access_token"
	defaultTwitterAuthURL    = "https://twitter.com/i/oauth2/authorize"
	defaultTwitterTokenURL   = "https://api.x.com/2/oauth2/token"
)

type socialAuthResponse struct {
	AuthURL string `json:"authUrl"`
	State   string `json:"state"`
}

type socialAccountResponse struct {
	Provider  string     `json:"provider"`
	AccountID *string    `json:"accountId,omitempty"`
	Username  *string    `json:"username,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type socialPostRequest struct {
	PostID string `json:"postId" binding:"required"`
}

type socialPostFallback struct {
	Text     string `json:"text"`
	MediaURL string `json:"mediaUrl"`
}

func (s *server) ListSocialAccounts(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	accounts, err := s.dbClient.SocialAccount().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]socialAccountResponse, 0, len(accounts))
	for _, account := range accounts {
		resp = append(resp, socialAccountResponse{
			Provider:  account.Provider,
			AccountID: account.AccountID,
			Username:  account.Username,
			ExpiresAt: account.ExpiresAt,
		})
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *server) GetSocialAuthURL(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	provider := strings.ToLower(ctx.Param("provider"))
	oauthConfig, err := s.getOAuthConfig(provider)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	statePayload := fmt.Sprintf("user:%s:provider:%s:time:%d", user.ID.String(), provider, time.Now().Unix())
	state := base64.URLEncoding.EncodeToString([]byte(statePayload))

	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	ctx.JSON(http.StatusOK, socialAuthResponse{
		AuthURL: authURL,
		State:   state,
	})
}

func (s *server) SocialCallback(ctx *gin.Context) {
	code := ctx.Query("code")
	state := ctx.Query("state")
	provider := strings.ToLower(ctx.Param("provider"))

	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "authorization code is required"})
		return
	}
	if state == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "state is required"})
		return
	}

	decodedState, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	stateStr := string(decodedState)
	userID, stateProvider, parseErr := parseSocialState(stateStr)
	if parseErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": parseErr.Error()})
		return
	}
	if provider != stateProvider {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "state provider mismatch"})
		return
	}

	oauthConfig, err := s.getOAuthConfig(provider)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := oauthConfig.Exchange(ctx.Request.Context(), code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("token exchange failed: %v", err)})
		return
	}

	var expiresAt *time.Time
	if !token.Expiry.IsZero() {
		exp := token.Expiry
		expiresAt = &exp
	}

	account := &models.SocialAccount{
		ID:           uuid.New(),
		UserID:       userID,
		Provider:     provider,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    expiresAt,
	}
	if scopeVal := token.Extra("scope"); scopeVal != nil {
		if scopeStr, ok := scopeVal.(string); ok && scopeStr != "" {
			account.Scopes = &scopeStr
		}
	}

	if err := s.dbClient.SocialAccount().Upsert(ctx, account); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to store token: %v", err)})
		return
	}

	redirectURI := ctx.Query("redirect_uri")
	if redirectURI != "" {
		ctx.Redirect(http.StatusFound, redirectURI+"?success=true&provider="+provider)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":  "Social authorization successful",
		"provider": provider,
	})
}

func (s *server) RevokeSocial(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	provider := strings.ToLower(ctx.Param("provider"))
	if !isSupportedProvider(provider) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider"})
		return
	}

	if err := s.dbClient.SocialAccount().DeleteByUserAndProvider(ctx, user.ID, provider); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Social account revoked"})
}

func (s *server) PostToSocial(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	provider := strings.ToLower(ctx.Param("provider"))
	if !isSupportedProvider(provider) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider"})
		return
	}

	var req socialPostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	postID, err := uuid.Parse(req.PostID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	post, err := s.dbClient.Post().FindByID(ctx, postID)
	if err != nil || post == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	if post.UserID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "can only share your own posts"})
		return
	}

	account, err := s.dbClient.SocialAccount().FindByUserAndProvider(ctx, user.ID, provider)
	if err != nil || account == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "account not linked"})
		return
	}

	fallbackText := "Check out my post on Vera!"
	if post.Caption != nil && *post.Caption != "" {
		fallbackText = *post.Caption
	}

	ctx.JSON(http.StatusNotImplemented, gin.H{
		"error": "Social posting is not yet configured for this provider.",
		"fallback": socialPostFallback{
			Text:     fallbackText,
			MediaURL: post.ImageURL,
		},
	})
}

func (s *server) getOAuthConfig(provider string) (*oauth2.Config, error) {
	switch provider {
	case socialProviderInstagram:
		if s.socialConfig.InstagramClientID == "" || s.socialConfig.InstagramClientSecret == "" || s.socialConfig.InstagramRedirectURL == "" {
			return nil, fmt.Errorf("instagram OAuth is not configured")
		}
		authURL := s.socialConfig.InstagramAuthURL
		if authURL == "" {
			authURL = defaultInstagramAuthURL
		}
		tokenURL := s.socialConfig.InstagramTokenURL
		if tokenURL == "" {
			tokenURL = defaultInstagramTokenURL
		}
		scopes := s.socialConfig.InstagramScopes
		if len(scopes) == 0 {
			scopes = []string{"instagram_basic", "instagram_content_publish", "pages_show_list", "pages_read_engagement"}
		}
		return &oauth2.Config{
			ClientID:     s.socialConfig.InstagramClientID,
			ClientSecret: s.socialConfig.InstagramClientSecret,
			RedirectURL:  s.socialConfig.InstagramRedirectURL,
			Scopes:       scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			},
		}, nil
	case socialProviderTwitter:
		if s.socialConfig.TwitterClientID == "" || s.socialConfig.TwitterClientSecret == "" || s.socialConfig.TwitterRedirectURL == "" {
			return nil, fmt.Errorf("twitter OAuth is not configured")
		}
		authURL := s.socialConfig.TwitterAuthURL
		if authURL == "" {
			authURL = defaultTwitterAuthURL
		}
		tokenURL := s.socialConfig.TwitterTokenURL
		if tokenURL == "" {
			tokenURL = defaultTwitterTokenURL
		}
		scopes := s.socialConfig.TwitterScopes
		if len(scopes) == 0 {
			scopes = []string{"tweet.read", "users.read", "tweet.write", "offline.access"}
		}
		return &oauth2.Config{
			ClientID:     s.socialConfig.TwitterClientID,
			ClientSecret: s.socialConfig.TwitterClientSecret,
			RedirectURL:  s.socialConfig.TwitterRedirectURL,
			Scopes:       scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported provider")
	}
}

func isSupportedProvider(provider string) bool {
	return provider == socialProviderInstagram || provider == socialProviderTwitter
}

func parseSocialState(state string) (uuid.UUID, string, error) {
	parts := strings.Split(state, ":")
	if len(parts) < 4 {
		return uuid.Nil, "", fmt.Errorf("invalid state")
	}

	var userIDStr string
	var provider string
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "user" && i+1 < len(parts) {
			userIDStr = parts[i+1]
		}
		if parts[i] == "provider" && i+1 < len(parts) {
			provider = parts[i+1]
		}
	}

	if userIDStr == "" || provider == "" {
		return uuid.Nil, "", fmt.Errorf("invalid state")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("invalid user id")
	}

	return userID, provider, nil
}
