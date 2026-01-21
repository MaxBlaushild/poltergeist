package pkg

import (
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

// NewServerFromDependencies creates a new verifiable-sn server with minimal dependencies
func NewServerFromDependencies(
	authClient auth.Client,
	dbClient db.DbClient,
) Server {
	awsClient := aws.NewAWSClient("us-east-1")
	// Initialize cert client with empty CA key (will generate new CA)
	certClient, err := cert.NewClient("")
	if err != nil {
		panic(err)
	}
	// Ethereum transactor client is optional - pass nil if not configured
	return server.NewServer(authClient, dbClient, awsClient, certClient, nil, "")
}
