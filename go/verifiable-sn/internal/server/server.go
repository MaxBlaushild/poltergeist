package server

import (
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/cert"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	ethereum_transactor "github.com/MaxBlaushild/poltergeist/pkg/ethereum_transactor"
	"github.com/MaxBlaushild/poltergeist/pkg/middleware"
	"github.com/MaxBlaushild/poltergeist/verifiable-sn/internal/push"
	"github.com/gin-gonic/gin"
)

type server struct {
	authClient               auth.Client
	dbClient                 db.DbClient
	awsClient                aws.AWSClient
	certClient               cert.Client
	ethereumTransactorClient ethereum_transactor.Client
	c2PAContractAddress      string
	pushClient               push.Client
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
	pushClient push.Client,
) Server {
	return &server{
		authClient:               authClient,
		dbClient:                 dbClient,
		awsClient:                awsClient,
		certClient:               certClient,
		ethereumTransactorClient: ethereumTransactorClient,
		c2PAContractAddress:      c2PAContractAddress,
		pushClient:               pushClient,
	}
}

func (s *server) SetupRoutes(r *gin.Engine) {
	r.GET("/verifiable-sn/health", s.GetHealth)

	// Authentication routes (no auth required)
	r.POST("/verifiable-sn/login", s.login)
	r.POST("/verifiable-sn/register", s.register)

	// Post routes
	r.POST("/verifiable-sn/posts", middleware.WithAuthenticationWithoutLocation(s.authClient, s.CreatePost))
	r.GET("/verifiable-sn/post-tag-suggestions", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetPostTagSuggestions))
	r.GET("/verifiable-sn/posts/feed", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetFeed))
	r.GET("/verifiable-sn/posts/user/:userId", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetUserPosts))
	r.GET("/verifiable-sn/posts/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetPost))
	r.POST("/verifiable-sn/posts/:id/tags", middleware.WithAuthenticationWithoutLocation(s.authClient, s.AddPostTags))
	r.DELETE("/verifiable-sn/posts/:id/tags", middleware.WithAuthenticationWithoutLocation(s.authClient, s.RemovePostTag))
	r.GET("/verifiable-sn/posts/:id/blockchain-transaction", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetBlockchainTransactionByManifestHash))
	r.DELETE("/verifiable-sn/posts/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.DeletePost))
	r.POST("/verifiable-sn/posts/:id/flag", middleware.WithAuthenticationWithoutLocation(s.authClient, s.FlagPost))
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
	r.POST("/verifiable-sn/albums/:id/tags", middleware.WithAuthenticationWithoutLocation(s.authClient, s.AddAlbumTag))
	r.DELETE("/verifiable-sn/albums/:id/tags", middleware.WithAuthenticationWithoutLocation(s.authClient, s.RemoveAlbumTag))
	r.POST("/verifiable-sn/albums/:id/invite", middleware.WithAuthenticationWithoutLocation(s.authClient, s.InviteToAlbum))
	r.GET("/verifiable-sn/albums/:id/members", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetAlbumMembers))
	r.DELETE("/verifiable-sn/albums/:id/members", middleware.WithAuthenticationWithoutLocation(s.authClient, s.RemoveAlbumMember))
	r.PATCH("/verifiable-sn/albums/:id/members", middleware.WithAuthenticationWithoutLocation(s.authClient, s.UpdateAlbumMemberRole))
	r.GET("/verifiable-sn/albums/:id/invites", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetAlbumPendingInvites))
	r.GET("/verifiable-sn/album-invites", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetAlbumInvites))
	r.POST("/verifiable-sn/album-invites/:inviteId/accept", middleware.WithAuthenticationWithoutLocation(s.authClient, s.AcceptAlbumInvite))
	r.POST("/verifiable-sn/album-invites/:inviteId/reject", middleware.WithAuthenticationWithoutLocation(s.authClient, s.RejectAlbumInvite))

	// Notification routes
	r.GET("/verifiable-sn/notifications", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetNotifications))
	r.PATCH("/verifiable-sn/notifications/read-all", middleware.WithAuthenticationWithoutLocation(s.authClient, s.MarkAllNotificationsRead))
	r.PATCH("/verifiable-sn/notifications/:id/read", middleware.WithAuthenticationWithoutLocation(s.authClient, s.MarkNotificationRead))
	r.POST("/verifiable-sn/device-tokens", middleware.WithAuthenticationWithoutLocation(s.authClient, s.RegisterDeviceToken))

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

	// Admin routes (flagged posts)
	r.GET("/verifiable-sn/admin/flagged-posts", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetFlaggedPosts))
	r.POST("/verifiable-sn/admin/flagged-posts/:id/dismiss", middleware.WithAuthenticationWithoutLocation(s.authClient, s.DismissFlaggedPost))
	r.DELETE("/verifiable-sn/admin/flagged-posts/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.AdminDeletePost))
}

func (s *server) ListenAndServe(port string) {
	r := gin.Default()
	s.SetupRoutes(r)
	r.Run(fmt.Sprintf(":%s", port))
}
