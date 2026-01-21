package pkg

import (
	"github.com/MaxBlaushild/poltergeist/final-fete/internal/gameengine"
	"github.com/MaxBlaushild/poltergeist/final-fete/internal/server"
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/hue"
	"github.com/gin-gonic/gin"
)

// Server interface for final-fete server
type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r *gin.Engine)
}

// UtilityClosetPuzzleClient interface for the puzzle game engine
type UtilityClosetPuzzleClient = gameengine.UtilityClosetPuzzleClient

// NewServer creates a new final-fete server
func NewServer(
	authClient auth.Client,
	dbClient db.DbClient,
	hueClient hue.Client,
	hueOAuthClient hue.OAuthClient,
	puzzleGameEngineClient UtilityClosetPuzzleClient,
) Server {
	return server.NewServer(authClient, dbClient, hueClient, hueOAuthClient, puzzleGameEngineClient)
}

func NewServerFromDependencies(
	authClient auth.Client,
	dbClient db.DbClient,
	hueClient hue.Client,
	hueOAuthClient hue.OAuthClient,
) Server {
	puzzleGameEngineClient := gameengine.NewUtilityClosetPuzzleClient(dbClient)
	return NewServer(authClient, dbClient, hueClient, hueOAuthClient, puzzleGameEngineClient)
}
