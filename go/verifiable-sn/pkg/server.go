package pkg

import (
	ethereum_transactor "github.com/MaxBlaushild/poltergeist/pkg/ethereum_transactor"
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/cert"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/verifiable-sn/internal/server"
	"github.com/gin-gonic/gin"
)

// Server interface for verifiable-sn server
type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r *gin.Engine)
}

// CoreConfig interface for accessing core configuration
type CoreConfig interface {
	GetEthereumTransactorURL() string
	GetC2PAContractAddress() string
}

// NewServerFromDependencies creates a new verifiable-sn server with minimal dependencies
func NewServerFromDependencies(
	authClient auth.Client,
	dbClient db.DbClient,
	coreConfig CoreConfig,
) Server {
	awsClient := aws.NewAWSClient("us-east-1")
	// Initialize cert client with empty CA key (will generate new CA)
	certClient, err := cert.NewClient("")
	if err != nil {
		panic(err)
	}

	// Validate required configuration
	ethereumTransactorURL := coreConfig.GetEthereumTransactorURL()
	c2PAContractAddress := coreConfig.GetC2PAContractAddress()
	if ethereumTransactorURL == "" {
		panic("ETHEREUM_TRANSACTOR_URL is required")
	}
	if c2PAContractAddress == "" {
		panic("C2PA_CONTRACT_ADDRESS is required")
	}

	ethereumTransactorClient := ethereum_transactor.NewClient(ethereumTransactorURL)
	return server.NewServer(authClient, dbClient, awsClient, certClient, ethereumTransactorClient, c2PAContractAddress)
}
