package main

import (
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/cert"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	ethereum_transactor "github.com/MaxBlaushild/poltergeist/pkg/ethereum_transactor"
	"github.com/MaxBlaushild/poltergeist/verifiable-sn/internal/config"
	"github.com/MaxBlaushild/poltergeist/verifiable-sn/internal/push"
	"github.com/MaxBlaushild/poltergeist/verifiable-sn/internal/server"
)

func main() {
	authClient := auth.NewClient()
	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		panic(err)
	}

	// Database connection will be provided by environment variables
	// In standalone mode, you'd configure these from env/config
	dbClient, err := db.NewClient(db.ClientConfig{
		Name:     cfg.Public.DbName,
		Host:     cfg.Public.DbHost,
		Port:     cfg.Public.DbPort,
		User:     cfg.Public.DbUser,
		Password: cfg.Secret.DbPassword,
	})
	if err != nil {
		panic(err)
	}

	awsClient := aws.NewAWSClient("us-east-1")

	certClient, err := cert.NewClient(cfg.Secret.CAPrivateKey)
	if err != nil {
		panic(err)
	}

	// Validate required configuration
	if cfg.Public.EthereumTransactorURL == "" {
		panic("ETHEREUM_TRANSACTOR_URL is required")
	}
	if cfg.Public.C2PAContractAddress == "" {
		panic("C2PA_CONTRACT_ADDRESS is required")
	}

	ethereumTransactorClient := ethereum_transactor.NewClient(cfg.Public.EthereumTransactorURL)
	pushClient := push.NewClient()

	parseScopes := func(scopes string) []string {
		if scopes == "" {
			return nil
		}
		parts := strings.FieldsFunc(scopes, func(r rune) bool {
			return r == ',' || r == ' ' || r == '\n' || r == '\t'
		})
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	}

	socialConfig := server.SocialConfig{
		InstagramClientID:     cfg.Public.InstagramClientID,
		InstagramClientSecret: cfg.Secret.InstagramClientSecret,
		InstagramRedirectURL:  cfg.Public.InstagramRedirectURL,
		InstagramAuthURL:      cfg.Public.InstagramAuthURL,
		InstagramTokenURL:     cfg.Public.InstagramTokenURL,
		InstagramScopes:       parseScopes(cfg.Public.InstagramScopes),
		TwitterClientID:       cfg.Public.TwitterClientID,
		TwitterClientSecret:   cfg.Secret.TwitterClientSecret,
		TwitterRedirectURL:    cfg.Public.TwitterRedirectURL,
		TwitterAuthURL:        cfg.Public.TwitterAuthURL,
		TwitterTokenURL:       cfg.Public.TwitterTokenURL,
		TwitterScopes:         parseScopes(cfg.Public.TwitterScopes),
	}

	server.NewServer(
		authClient,
		dbClient,
		awsClient,
		certClient,
		ethereumTransactorClient,
		cfg.Public.C2PAContractAddress,
		pushClient,
		socialConfig,
	).ListenAndServe("8087")
}
