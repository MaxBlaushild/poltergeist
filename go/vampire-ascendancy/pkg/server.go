package pkg

import (
	"os"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
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
	texterClient texter.Client,
) Server {
	// REDIS_URL/PHONE_NUMBER/VAMPIRE_SITE_URL are process env vars in the
	// composed (core) deployment. If REDIS_URL is unset, grading enqueue is
	// disabled and the GM sees a clear error; PHONE_NUMBER/VAMPIRE_SITE_URL
	// are needed for player-invite SMS to go out with a working link.
	return server.NewServer(authClient, dbClient, texterClient, os.Getenv("REDIS_URL"), os.Getenv("PHONE_NUMBER"), os.Getenv("VAMPIRE_SITE_URL"))
}

// NewServerFromDependencies creates a new vampire-ascendancy server with minimal
// dependencies. It exists so core can compose the server the same way it does the
// other folded-in modules (see sonar.NewServerFromDependencies for the same
// texterClient-as-dependency pattern).
func NewServerFromDependencies(
	authClient auth.Client,
	dbClient db.DbClient,
	texterClient texter.Client,
) Server {
	return server.NewServer(authClient, dbClient, texterClient, os.Getenv("REDIS_URL"), os.Getenv("PHONE_NUMBER"), os.Getenv("VAMPIRE_SITE_URL"))
}
