package pkg

import (
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
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
	return server.NewServer(authClient, dbClient, awsClient)
}
