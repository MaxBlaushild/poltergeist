package pkg

import (
	"os"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/vampire-ascendancy/internal/server"
	"github.com/gin-gonic/gin"
)

// Server interface for the vampire-ascendancy server
type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r *gin.Engine)
}

// NewServer creates a new vampire-ascendancy server
func NewServer(
	authClient auth.Client,
	dbClient db.DbClient,
) Server {
	// REDIS_URL is a process env var in the composed (core) deployment; if unset,
	// grading enqueue is disabled and the GM sees a clear error.
	return server.NewServer(authClient, dbClient, os.Getenv("REDIS_URL"))
}

// NewServerFromDependencies creates a new vampire-ascendancy server with minimal
// dependencies. It exists so core can compose the server the same way it does the
// other folded-in modules. vampire-ascendancy only needs the shared auth and db
// clients, so there is no extra config to thread through.
func NewServerFromDependencies(
	authClient auth.Client,
	dbClient db.DbClient,
) Server {
	// REDIS_URL is a process env var in the composed (core) deployment; if unset,
	// grading enqueue is disabled and the GM sees a clear error.
	return server.NewServer(authClient, dbClient, os.Getenv("REDIS_URL"))
}
