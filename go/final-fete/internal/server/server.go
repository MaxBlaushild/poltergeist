package server

import (
	"fmt"

	"github.com/MaxBlaushild/poltergeist/final-fete/internal/gameengine"
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/hue"
	"github.com/MaxBlaushild/poltergeist/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type server struct {
	authClient             auth.Client
	dbClient               db.DbClient
	hueClient              hue.Client
	hueOAuthClient         hue.OAuthClient
	puzzleGameEngineClient gameengine.UtilityClosetPuzzleClient
}

type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r *gin.Engine)
}

func NewServer(
	authClient auth.Client,
	dbClient db.DbClient,
	hueClient hue.Client,
	hueOAuthClient hue.OAuthClient,
	puzzleGameEngineClient gameengine.UtilityClosetPuzzleClient,
) Server {
	return &server{
		authClient:             authClient,
		dbClient:               dbClient,
		hueClient:              hueClient,
		hueOAuthClient:         hueOAuthClient,
		puzzleGameEngineClient: puzzleGameEngineClient,
	}
}

func (s *server) SetupRoutes(r *gin.Engine) {
	r.GET("/final-fete/health", s.GetHealth)

	// Authentication routes (no auth required)
	r.POST("/final-fete/login", s.login)
	r.POST("/final-fete/register", s.register)

	// Fete Rooms routes
	r.GET("/final-fete/rooms", middleware.WithAuthenticationWithoutLocation(s.authClient, s.getAllFeteRooms))
	r.GET("/final-fete/rooms/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.getFeteRoom))
	r.POST("/final-fete/rooms", middleware.WithAuthenticationWithoutLocation(s.authClient, s.createFeteRoom))
	r.PUT("/final-fete/rooms/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.updateFeteRoom))
	r.DELETE("/final-fete/rooms/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.deleteFeteRoom))
	r.POST("/final-fete/rooms/:id/toggle", middleware.WithAuthenticationWithoutLocation(s.authClient, s.toggleFeteRoom))
	r.POST("/final-fete/rooms/:id/unlock", middleware.WithAuthenticationWithoutLocation(s.authClient, s.unlockFeteRoom))

	// Fete Teams routes
	r.GET("/final-fete/teams", middleware.WithAuthenticationWithoutLocation(s.authClient, s.getAllFeteTeams))
	r.GET("/final-fete/teams/current", middleware.WithAuthenticationWithoutLocation(s.authClient, s.getCurrentUserTeam))
	r.GET("/final-fete/teams/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.getFeteTeam))
	r.POST("/final-fete/teams", middleware.WithAuthenticationWithoutLocation(s.authClient, s.createFeteTeam))
	r.PUT("/final-fete/teams/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.updateFeteTeam))
	r.DELETE("/final-fete/teams/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.deleteFeteTeam))
	r.GET("/final-fete/teams/:id/users", middleware.WithAuthenticationWithoutLocation(s.authClient, s.getTeamUsers))
	r.POST("/final-fete/teams/:id/users", middleware.WithAuthenticationWithoutLocation(s.authClient, s.addUserToTeam))
	r.DELETE("/final-fete/teams/:id/users/:userId", middleware.WithAuthenticationWithoutLocation(s.authClient, s.removeUserFromTeam))
	r.GET("/final-fete/users/search", middleware.WithAuthenticationWithoutLocation(s.authClient, s.searchUsers))

	// Fete Room Linked List Teams routes
	r.GET("/final-fete/room-linked-list-teams", middleware.WithAuthenticationWithoutLocation(s.authClient, s.getAllFeteRoomLinkedListTeams))
	r.GET("/final-fete/room-linked-list-teams/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.getFeteRoomLinkedListTeam))
	r.POST("/final-fete/room-linked-list-teams", middleware.WithAuthenticationWithoutLocation(s.authClient, s.createFeteRoomLinkedListTeam))
	r.PUT("/final-fete/room-linked-list-teams/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.updateFeteRoomLinkedListTeam))
	r.DELETE("/final-fete/room-linked-list-teams/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.deleteFeteRoomLinkedListTeam))

	// Fete Room Teams routes
	r.GET("/final-fete/room-teams", middleware.WithAuthenticationWithoutLocation(s.authClient, s.getAllFeteRoomTeams))
	r.GET("/final-fete/room-teams/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.getFeteRoomTeam))
	r.POST("/final-fete/room-teams", middleware.WithAuthenticationWithoutLocation(s.authClient, s.createFeteRoomTeam))
	r.DELETE("/final-fete/room-teams/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.deleteFeteRoomTeam))

	// Hue Lights routes
	r.GET("/final-fete/hue-lights", middleware.WithAuthenticationWithoutLocation(s.authClient, s.getAllHueLights))

	// Hue OAuth routes (callback doesn't require auth)
	r.GET("/final-fete/hue-oauth/callback", s.hueOAuthCallback)

	// Utility Closet Puzzle routes
	r.GET("/final-fete/utility-closet-puzzle", s.GetPuzzleState)
	r.PUT("/final-fete/utility-closet-puzzle", s.UpdatePuzzle)
	r.GET("/final-fete/utility-closet-puzzle/press/:slot", middleware.WithAntiviralCookie(s.PressButton))
	r.POST("/final-fete/utility-closet-puzzle/reset", s.ResetPuzzle)

	// Utility Closet Puzzle Admin CRUD routes
	r.GET("/final-fete/admin/utility-closet-puzzle", s.AdminGetPuzzleState)
	r.POST("/final-fete/admin/utility-closet-puzzle", s.AdminCreatePuzzle)
	r.PUT("/final-fete/admin/utility-closet-puzzle", s.AdminUpdatePuzzle)
	r.DELETE("/final-fete/admin/utility-closet-puzzle", s.AdminDeletePuzzle)
	r.POST("/final-fete/utility-closet-puzzle/toggle-achievement", s.ToggleAchievement)
}

func (s *server) ListenAndServe(port string) {
	r := gin.Default()
	s.SetupRoutes(r)
	r.Run(fmt.Sprintf(":%s", port))
}
