package pkg

import (
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
	return server.NewServer(authClient, dbClient)
}

// NewServerFromDependencies creates a new vampire-ascendancy server with minimal
// dependencies. It exists so core can compose the server the same way it does the
// other folded-in modules. vampire-ascendancy only needs the shared auth and db
// clients, so there is no extra config to thread through.
func NewServerFromDependencies(
	authClient auth.Client,
	dbClient db.DbClient,
) Server {
	return server.NewServer(authClient, dbClient)
}
