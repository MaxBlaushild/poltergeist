package server

import (
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type server struct {
	authClient auth.Client
	dbClient   db.DbClient
	awsClient  aws.AWSClient
}

type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r *gin.Engine)
}

func NewServer(
	authClient auth.Client,
	dbClient db.DbClient,
	awsClient aws.AWSClient,
) Server {
	return &server{
		authClient: authClient,
		dbClient:   dbClient,
		awsClient:  awsClient,
	}
}

func (s *server) SetupRoutes(r *gin.Engine) {
	r.GET("/verifiable-sn/health", s.GetHealth)

	// Authentication routes (no auth required)
	r.POST("/verifiable-sn/login", s.login)
	r.POST("/verifiable-sn/register", s.register)

	// Post routes
	r.POST("/verifiable-sn/posts", middleware.WithAuthenticationWithoutLocation(s.authClient, s.CreatePost))
	r.GET("/verifiable-sn/posts/feed", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetFeed))
	r.GET("/verifiable-sn/posts/user/:userId", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetUserPosts))
	r.DELETE("/verifiable-sn/posts/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.DeletePost))

	// Media routes
	r.POST("/verifiable-sn/media/uploadUrl", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetPresignedUploadUrl))

	// Friend routes
	r.GET("/verifiable-sn/friends", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetFriends))
	r.GET("/verifiable-sn/friend-invites", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetFriendInvites))
	r.POST("/verifiable-sn/friend-invites/create", middleware.WithAuthenticationWithoutLocation(s.authClient, s.CreateFriendInvite))
	r.POST("/verifiable-sn/friend-invites/accept", middleware.WithAuthenticationWithoutLocation(s.authClient, s.AcceptFriendInvite))
	r.DELETE("/verifiable-sn/friend-invites/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.DeleteFriendInvite))

	// User routes
	r.GET("/verifiable-sn/users/search", middleware.WithAuthenticationWithoutLocation(s.authClient, s.SearchUsers))
}

func (s *server) ListenAndServe(port string) {
	r := gin.Default()
	s.SetupRoutes(r)
	r.Run(fmt.Sprintf(":%s", port))
}
