package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MaxBlaushild/poltergeist/final-fete/internal/config"
	"github.com/MaxBlaushild/poltergeist/final-fete/internal/gameengine"
	"github.com/MaxBlaushild/poltergeist/final-fete/internal/server"
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/hue"
)

func main() {
	config, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		panic(err)
	}

	authClient := auth.NewClient()
	dbClient, err := db.NewClient(db.ClientConfig{
		Name:     config.Public.DbName,
		Host:     config.Public.DbHost,
		Port:     config.Public.DbPort,
		User:     config.Public.DbUser,
		Password: config.Secret.DbPassword,
	})
	if err != nil {
		panic(err)
	}

	// Initialize Hue OAuth client if credentials are provided
	var hueOAuthClient hue.OAuthClient
	if config.Secret.HueClientID != "" && config.Secret.HueClientSecret != "" && config.Public.HueRedirectURI != "" {
		hueOAuthClient = hue.NewOAuthClient(hue.OAuthClientConfig{
			ClientID:     config.Secret.HueClientID,
			ClientSecret: config.Secret.HueClientSecret,
			RedirectURI:  config.Public.HueRedirectURI,
		})
	}

	// Initialize Hue cloud client using OAuth
	var hueClient hue.Client
	ctx := context.Background()
	if hueOAuthClient != nil {
		// Load refresh token from database
		hueToken, err := dbClient.HueToken().FindLatest(ctx)
		if err != nil {
			log.Printf("Warning: Failed to load Hue token from database: %v", err)
		} else if hueToken != nil && hueToken.RefreshToken != "" {
			// Create token updater callback to persist refreshed tokens
			tokenUpdater := func(accessToken, refreshToken string, expiresAt time.Time) error {
				// Find the token again to get the ID
				latestToken, err := dbClient.HueToken().FindLatest(ctx)
				if err != nil {
					return fmt.Errorf("failed to find token for update: %w", err)
				}
				if latestToken == nil {
					return fmt.Errorf("token not found in database")
				}

				// Update token fields
				latestToken.AccessToken = accessToken
				latestToken.RefreshToken = refreshToken
				latestToken.ExpiresAt = expiresAt

				// Save updated token
				return dbClient.HueToken().Update(ctx, latestToken)
			}

			// Initialize cloud client with OAuth, refresh token, existing access token (if valid), and token updater
			hueClient = hue.NewClientWithOAuth(
				hueOAuthClient,
				hueToken.RefreshToken,
				hueToken.AccessToken,
				hueToken.ExpiresAt,
				tokenUpdater,
				config.Secret.HueApplicationKey,
			)
			log.Println("Hue cloud client initialized successfully")
		} else {
			log.Println("Warning: No Hue refresh token found in database. Hue features will be unavailable. Run OAuth flow first.")
		}
	}

	// Initialize puzzle game engine client
	puzzleGameEngineClient := gameengine.NewUtilityClosetPuzzleClient(dbClient)

	server.NewServer(authClient, dbClient, hueClient, hueOAuthClient, puzzleGameEngineClient).ListenAndServe("8085")
}
