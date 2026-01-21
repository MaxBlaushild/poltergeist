package main

import (
	ethereum_transactor "github.com/MaxBlaushild/poltergeist/pkg/ethereum_transactor"
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/cert"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/verifiable-sn/internal/config"
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

	var ethereumTransactorClient ethereum_transactor.Client
	if cfg.Public.EthereumTransactorURL != "" {
		ethereumTransactorClient = ethereum_transactor.NewClient(cfg.Public.EthereumTransactorURL)
	}

	server.NewServer(authClient, dbClient, awsClient, certClient, ethereumTransactorClient, cfg.Public.C2PAContractAddress).ListenAndServe("8087")
}
