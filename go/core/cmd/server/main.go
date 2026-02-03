package main

import (
	"log"

	"github.com/MaxBlaushild/core/internal/config"
	"github.com/MaxBlaushild/core/internal/server"
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	sonar "github.com/MaxBlaushild/poltergeist/sonar/pkg"
	travelangels "github.com/MaxBlaushild/poltergeist/travel-angels/pkg"
	verifiablesn "github.com/MaxBlaushild/poltergeist/verifiable-sn/pkg"
)

func main() {
	// Load shared configuration
	cfg := config.NewConfigFromEnv()

	texterClient := texter.NewClient()

	// Initialize auth client
	authClient := auth.NewClient()

	// Initialize database client using shared config
	dbClient, err := db.NewClient(db.ClientConfig{
		Name:     cfg.Public.DbName,
		Host:     cfg.Public.DbHost,
		Port:     cfg.Public.DbPort,
		User:     cfg.Public.DbUser,
		Password: cfg.Secret.DbPassword,
	})
	if err != nil {
		log.Printf("Warning: Failed to initialize database client: %v. Routes will not be available.", err)
		panic(err)
	}

	// Initialize Hue OAuth client if credentials are provided
	// var hueOAuthClient hue.OAuthClient
	// if cfg.Secret.HueClientID != "" && cfg.Secret.HueClientSecret != "" && cfg.Public.HueRedirectURI != "" {
	// 	hueOAuthClient = hue.NewOAuthClient(hue.OAuthClientConfig{
	// 		ClientID:     cfg.Secret.HueClientID,
	// 		ClientSecret: cfg.Secret.HueClientSecret,
	// 		RedirectURI:  cfg.Public.HueRedirectURI,
	// 	})
	// }

	// Initialize Hue cloud client using OAuth
	// var hueClient hue.Client
	// ctx := context.Background()
	// if hueOAuthClient != nil {
	// 	// Load refresh token from database
	// 	hueToken, err := dbClient.HueToken().FindLatest(ctx)
	// 	if err != nil {
	// 		log.Printf("Warning: Failed to load Hue token from database: %v", err)
	// 	} else if hueToken != nil && hueToken.RefreshToken != "" {
	// 		// Create token updater callback to persist refreshed tokens
	// 		tokenUpdater := func(accessToken, refreshToken string, expiresAt time.Time) error {
	// 			// Find the token again to get the ID
	// 			latestToken, err := dbClient.HueToken().FindLatest(ctx)
	// 			if err != nil {
	// 				return fmt.Errorf("failed to find token for update: %w", err)
	// 			}
	// 			if latestToken == nil {
	// 				return fmt.Errorf("token not found in database")
	// 			}

	// 			// Update token fields
	// 			latestToken.AccessToken = accessToken
	// 			latestToken.RefreshToken = refreshToken
	// 			latestToken.ExpiresAt = expiresAt

	// 			// Save updated token
	// 			return dbClient.HueToken().Update(ctx, latestToken)
	// 		}

	// 		// Initialize cloud client with OAuth, refresh token, existing access token (if valid), and token updater
	// 		hueClient = hue.NewClientWithOAuth(
	// 			hueOAuthClient,
	// 			hueToken.RefreshToken,
	// 			hueToken.AccessToken,
	// 			hueToken.ExpiresAt,
	// 			tokenUpdater,
	// 			cfg.Secret.HueApplicationKey,
	// 		)
	// 		log.Println("Hue cloud client initialized successfully")
	// 	} else {
	// 		log.Println("Warning: No Hue refresh token found in database. Hue features will be unavailable. Run OAuth flow first.")
	// 	}
	// }

	// finalFeteServer := finalfete.NewServerFromDependencies(authClient, dbClient, hueClient, hueOAuthClient)
	sonarServer := sonar.NewServerFromDependencies(authClient, texterClient, dbClient, cfg)
	travelAngelsServer := travelangels.NewServerFromDependencies(authClient, dbClient, cfg)
	verifiableSnServer := verifiablesn.NewServerFromDependencies(authClient, dbClient, cfg)
	srv := server.NewServer(sonarServer, travelAngelsServer, verifiableSnServer, texterClient)
	srv.ListenAndServe("8080")
}
