package main

import (
	"github.com/MaxBlaushild/poltergeist/ethereum-transactor/internal/config"
	"github.com/MaxBlaushild/poltergeist/ethereum-transactor/internal/server"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/ethereum"
)

func main() {
	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		panic(err)
	}

	// Validate required config
	if cfg.Secret.PrivateKey == "" {
		panic("PRIVATE_KEY environment variable is required")
	}
	if cfg.Public.RPCURL == "" {
		panic("RPC_URL environment variable is required")
	}
	if cfg.Public.ChainID == 0 {
		panic("CHAIN_ID environment variable is required")
	}

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

	ethereumClient, err := ethereum.NewClient(cfg.Public.RPCURL, cfg.Secret.PrivateKey, cfg.Public.ChainID)
	if err != nil {
		panic(err)
	}

	server.NewServer(dbClient, ethereumClient, cfg.Public.ChainID).ListenAndServe("8088")
}
