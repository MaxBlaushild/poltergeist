package server

import (
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/cert"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	ethereum_transactor "github.com/MaxBlaushild/poltergeist/pkg/ethereum_transactor"
	"github.com/MaxBlaushild/poltergeist/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type server struct {
	authClient            auth.Client
	dbClient              db.DbClient
	awsClient             aws.AWSClient
	certClient            cert.Client
	ethereumTransactorClient ethereum_transactor.Client
	c2PAContractAddress   string
}

type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r *gin.Engine)
}

func NewServer(
	authClient auth.Client,
	dbClient db.DbClient,
	awsClient aws.AWSClient,
	certClient cert.Client,
	ethereumTransactorClient ethereum_transactor.Client,
	c2PAContractAddress string,
) Server {
	return &server{
		authClient:              authClient,
		dbClient:                dbClient,
		awsClient:                awsClient,
		certClient:               certClient,
		ethereumTransactorClient: ethereumTransactorClient,
		c2PAContractAddress:      c2PAContractAddress,
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
	r.GET("/verifiable-sn/posts/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetPost))
	r.GET("/verifiable-sn/posts/:id/blockchain-transaction", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetBlockchainTransactionByManifestHash))
	r.DELETE("/verifiable-sn/posts/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.DeletePost))
	r.POST("/verifiable-sn/posts/:id/reactions", middleware.WithAuthenticationWithoutLocation(s.authClient, s.CreateReaction))
	r.DELETE("/verifiable-sn/posts/:id/reactions", middleware.WithAuthenticationWithoutLocation(s.authClient, s.DeleteReaction))
	r.GET("/verifiable-sn/posts/:id/comments", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetComments))
	r.POST("/verifiable-sn/posts/:id/comments", middleware.WithAuthenticationWithoutLocation(s.authClient, s.CreateComment))
	r.DELETE("/verifiable-sn/posts/:id/comments/:commentId", middleware.WithAuthenticationWithoutLocation(s.authClient, s.DeleteComment))

	// Album routes
	r.POST("/verifiable-sn/albums", middleware.WithAuthenticationWithoutLocation(s.authClient, s.CreateAlbum))
	r.GET("/verifiable-sn/albums", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetAlbums))
	r.GET("/verifiable-sn/albums/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetAlbum))
	r.DELETE("/verifiable-sn/albums/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.DeleteAlbum))

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
	r.PUT("/verifiable-sn/users/profile", middleware.WithAuthenticationWithoutLocation(s.authClient, s.UpdateProfile))

	// Certificate routes
	r.GET("/verifiable-sn/certificate/check", middleware.WithAuthenticationWithoutLocation(s.authClient, s.CheckCertificate))
	r.POST("/verifiable-sn/certificate/enroll", middleware.WithAuthenticationWithoutLocation(s.authClient, s.EnrollCertificate))
	r.GET("/verifiable-sn/certificate", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetCertificate))
	r.GET("/verifiable-sn/certificate/user/:userId", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetUserCertificate))
}

func (s *server) ListenAndServe(port string) {
	r := gin.Default()
	s.SetupRoutes(r)
	r.Run(fmt.Sprintf(":%s", port))
}
