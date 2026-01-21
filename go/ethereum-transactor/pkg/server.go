package pkg

import (
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/ethereum"
	"github.com/MaxBlaushild/poltergeist/ethereum-transactor/internal/server"
)

// Server interface for ethereum-transactor server
type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r interface{})
}

// NewServerFromDependencies creates a new ethereum-transactor server with minimal dependencies
func NewServerFromDependencies(
	dbClient db.DbClient,
	ethereumClient ethereum.EthereumClient,
	chainID int64,
) Server {
	return server.NewServer(dbClient, ethereumClient, chainID)
}

// NewServer creates a new ethereum-transactor server with all dependencies provided
func NewServer(
	dbClient db.DbClient,
	ethereumClient ethereum.EthereumClient,
	chainID int64,
) Server {
	return server.NewServer(dbClient, ethereumClient, chainID)
}

