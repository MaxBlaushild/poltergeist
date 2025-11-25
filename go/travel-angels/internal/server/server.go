package server

import (
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/billing"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/dropbox"
	"github.com/MaxBlaushild/poltergeist/pkg/googledrive"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type server struct {
	authClient        auth.Client
	dbClient          db.DbClient
	googleDriveClient googledrive.Client
	dropboxClient     dropbox.Client
	awsClient         aws.AWSClient
	billingClient     billing.Client
	googleMapsClient  googlemaps.Client
	baseURL           string
}

type Server interface {
	ListenAndServe(port string)
}

func NewServer(
	authClient auth.Client,
	dbClient db.DbClient,
	googleDriveClient googledrive.Client,
	dropboxClient dropbox.Client,
	awsClient aws.AWSClient,
	billingClient billing.Client,
	googleMapsClient googlemaps.Client,
	baseURL string,
) Server {
	return &server{
		authClient:        authClient,
		dbClient:          dbClient,
		googleDriveClient: googleDriveClient,
		dropboxClient:     dropboxClient,
		awsClient:         awsClient,
		billingClient:     billingClient,
		googleMapsClient:  googleMapsClient,
		baseURL:           baseURL,
	}
}

func (s *server) ListenAndServe(port string) {
	r := gin.Default()

	r.GET("/travel-angels/health", s.GetHealth)
	r.POST("/travel-angels/login", s.login)
	r.POST("/travel-angels/register", s.register)
	r.GET("/travel-angels/whoami", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetWhoami))
	r.GET("/travel-angels/level", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetLevel))
	r.POST("/travel-angels/documents", middleware.WithAuthenticationWithoutLocation(s.authClient, s.CreateDocument))
	r.PUT("/travel-angels/documents/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.UpdateDocument))
	r.DELETE("/travel-angels/documents/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.DeleteDocument))
	r.GET("/travel-angels/documents/user/:userId", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetDocumentsByUserID))
	r.GET("/travel-angels/documents/friends", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetFriendsDocuments))
	r.POST("/travel-angels/documents/parse", middleware.WithAuthenticationWithoutLocation(s.authClient, s.ParseDocument))
	r.GET("/travel-angels/google-drive/status", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetGoogleDriveStatus))
	r.GET("/travel-angels/google-drive/auth", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetGoogleDriveAuth))
	r.GET("/travel-angels/google-drive/callback", s.GoogleDriveCallback)
	r.POST("/travel-angels/google-drive/revoke", middleware.WithAuthenticationWithoutLocation(s.authClient, s.RevokeGoogleDrive))
	r.POST("/travel-angels/google-drive/files/:fileId/share", middleware.WithAuthenticationWithoutLocation(s.authClient, s.ShareGoogleDriveFile))
	r.POST("/travel-angels/google-drive/files/:fileId/grant-permissions", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GrantGoogleDrivePermissions))
	r.GET("/travel-angels/google-drive/files", middleware.WithAuthenticationWithoutLocation(s.authClient, s.ListGoogleDriveFiles))
	r.POST("/travel-angels/google-drive/documents/import", middleware.WithAuthenticationWithoutLocation(s.authClient, s.ImportGoogleDriveDocument))
	r.GET("/travel-angels/friends", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetFriends))
	r.GET("/travel-angels/friend-invites", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetFriendInvites))
	r.POST("/travel-angels/friend-invites/create", middleware.WithAuthenticationWithoutLocation(s.authClient, s.CreateFriendInvite))
	r.POST("/travel-angels/friend-invites/accept", middleware.WithAuthenticationWithoutLocation(s.authClient, s.AcceptFriendInvite))
	r.DELETE("/travel-angels/friend-invites/:id", middleware.WithAuthenticationWithoutLocation(s.authClient, s.DeleteFriendInvite))
	r.GET("/travel-angels/users/search", middleware.WithAuthenticationWithoutLocation(s.authClient, s.SearchUsers))
	r.GET("/travel-angels/users/validate-username", middleware.WithAuthenticationWithoutLocation(s.authClient, s.ValidateUsername))
	r.GET("/travel-angels/location/search", s.SearchLocation)
	r.GET("/travel-angels/dropbox/auth", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetDropboxAuth))
	r.GET("/travel-angels/dropbox/callback", s.DropboxCallback)
	r.POST("/travel-angels/dropbox/revoke", middleware.WithAuthenticationWithoutLocation(s.authClient, s.RevokeDropbox))
	r.POST("/travel-angels/dropbox/files/:path/share", middleware.WithAuthenticationWithoutLocation(s.authClient, s.ShareDropboxFile))
	r.POST("/travel-angels/dropbox/files/:path/create-shared-link", middleware.WithAuthenticationWithoutLocation(s.authClient, s.CreateDropboxSharedLink))
	r.GET("/travel-angels/dropbox/files", middleware.WithAuthenticationWithoutLocation(s.authClient, s.ListDropboxFiles))
	r.POST("/travel-angels/media/uploadUrl", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetPresignedUploadUrl))
	r.POST("/travel-angels/profile", middleware.WithAuthenticationWithoutLocation(s.authClient, s.UpdateProfile))
	r.GET("/travel-angels/credits", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetCredits))
	r.POST("/travel-angels/credits/purchase", middleware.WithAuthenticationWithoutLocation(s.authClient, s.PurchaseCredits))
	r.POST("/travel-angels/credits/webhook", s.HandleCreditsWebhook)
	r.POST("/travel-angels/credits/add", middleware.WithAuthenticationWithoutLocation(s.authClient, s.AddCredits))
	r.POST("/travel-angels/credits/subtract", middleware.WithAuthenticationWithoutLocation(s.authClient, s.SubtractCredits))

	r.Run(fmt.Sprintf(":%s", port))
}
