package server

import (
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/dropbox"
	"github.com/MaxBlaushild/poltergeist/pkg/googledrive"
	"github.com/MaxBlaushild/poltergeist/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type server struct {
	authClient        auth.Client
	dbClient          db.DbClient
	googleDriveClient googledrive.Client
	dropboxClient     dropbox.Client
}

type Server interface {
	ListenAndServe(port string)
}

func NewServer(
	authClient auth.Client,
	dbClient db.DbClient,
	googleDriveClient googledrive.Client,
	dropboxClient dropbox.Client,
) Server {
	return &server{
		authClient:        authClient,
		dbClient:          dbClient,
		googleDriveClient: googleDriveClient,
		dropboxClient:     dropboxClient,
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
	r.GET("/travel-angels/documents/user/:userId", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetDocumentsByUserID))
	r.POST("/travel-angels/documents/parse", middleware.WithAuthenticationWithoutLocation(s.authClient, s.ParseDocument))
	r.GET("/travel-angels/google-drive/status", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetGoogleDriveStatus))
	r.GET("/travel-angels/google-drive/auth", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetGoogleDriveAuth))
	r.GET("/travel-angels/google-drive/callback", s.GoogleDriveCallback)
	r.POST("/travel-angels/google-drive/revoke", middleware.WithAuthenticationWithoutLocation(s.authClient, s.RevokeGoogleDrive))
	r.POST("/travel-angels/google-drive/files/:fileId/share", middleware.WithAuthenticationWithoutLocation(s.authClient, s.ShareGoogleDriveFile))
	r.POST("/travel-angels/google-drive/files/:fileId/grant-permissions", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GrantGoogleDrivePermissions))
	r.GET("/travel-angels/google-drive/files", middleware.WithAuthenticationWithoutLocation(s.authClient, s.ListGoogleDriveFiles))
	r.GET("/travel-angels/dropbox/auth", middleware.WithAuthenticationWithoutLocation(s.authClient, s.GetDropboxAuth))
	r.GET("/travel-angels/dropbox/callback", s.DropboxCallback)
	r.POST("/travel-angels/dropbox/revoke", middleware.WithAuthenticationWithoutLocation(s.authClient, s.RevokeDropbox))
	r.POST("/travel-angels/dropbox/files/:path/share", middleware.WithAuthenticationWithoutLocation(s.authClient, s.ShareDropboxFile))
	r.POST("/travel-angels/dropbox/files/:path/create-shared-link", middleware.WithAuthenticationWithoutLocation(s.authClient, s.CreateDropboxSharedLink))
	r.GET("/travel-angels/dropbox/files", middleware.WithAuthenticationWithoutLocation(s.authClient, s.ListDropboxFiles))

	r.Run(fmt.Sprintf(":%s", port))
}
