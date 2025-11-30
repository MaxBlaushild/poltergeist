package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/final-fete/internal/config"
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

	// Initialize Hue client if credentials are provided
	var hueClient hue.Client
	if config.Secret.HueBridgeHostname != "" && config.Secret.HueBridgeUsername != "" {
		ctx := context.Background()
		hueClient = hue.NewClient()
		hostname := config.Secret.HueBridgeHostname
		// Ensure hostname has https:// prefix if it doesn't already
		if !strings.HasPrefix(hostname, "http://") && !strings.HasPrefix(hostname, "https://") {
			hostname = fmt.Sprintf("https://%s", hostname)
		}
		err = hueClient.Connect(ctx, hostname, config.Secret.HueBridgeUsername)
		if err != nil {
			panic(fmt.Errorf("failed to connect to Hue bridge: %w", err))
		}
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

	server.NewServer(authClient, dbClient, hueClient, hueOAuthClient).ListenAndServe("8085")
}
