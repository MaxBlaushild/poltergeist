package pkg

import (
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/billing"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/dropbox"
	"github.com/MaxBlaushild/poltergeist/pkg/googledrive"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/travel-angels/internal/server"
	"github.com/gin-gonic/gin"
)

// Server interface for travel-angels server
type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r *gin.Engine)
}

// CoreConfig is an interface for core configuration
// It allows travel-angels to accept any config implementation that provides these getters
type CoreConfig interface {
	GetDbHost() string
	GetDbUser() string
	GetDbPort() string
	GetDbName() string
	GetDbPassword() string
	GetGoogleMapsApiKey() string
	GetGoogleDriveClientID() string
	GetGoogleDriveClientSecret() string
	GetGoogleDriveRedirectURI() string
	GetDropboxClientID() string
	GetDropboxClientSecret() string
	GetDropboxRedirectURI() string
	GetBaseURL() string
}

// NewServerFromDependencies creates a new travel-angels server with minimal dependencies
// It initializes all internal dependencies internally using the provided core config
func NewServerFromDependencies(
	authClient auth.Client,
	dbClient db.DbClient,
	coreConfig CoreConfig,
) Server {
	// Initialize clients using core config
	awsClient := aws.NewAWSClient("us-east-1")
	billingClient := billing.NewClient()
	googleMapsClient := googlemaps.NewClient(coreConfig.GetGoogleMapsApiKey())

	googleDriveClient := googledrive.NewClient(googledrive.ClientConfig{
		ClientID:     coreConfig.GetGoogleDriveClientID(),
		ClientSecret: coreConfig.GetGoogleDriveClientSecret(),
		RedirectURI:  coreConfig.GetGoogleDriveRedirectURI(),
	}, dbClient)

	dropboxClient := dropbox.NewClient(dropbox.ClientConfig{
		ClientID:     coreConfig.GetDropboxClientID(),
		ClientSecret: coreConfig.GetDropboxClientSecret(),
		RedirectURI:  coreConfig.GetDropboxRedirectURI(),
	}, dbClient)

	baseURL := coreConfig.GetBaseURL()
	if baseURL == "" {
		baseURL = "http://localhost:8083"
	}

	return server.NewServer(
		authClient,
		dbClient,
		googleDriveClient,
		dropboxClient,
		awsClient,
		billingClient,
		googleMapsClient,
		baseURL,
	)
}

// NewServer creates a new travel-angels server with all dependencies provided
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
	return server.NewServer(
		authClient,
		dbClient,
		googleDriveClient,
		dropboxClient,
		awsClient,
		billingClient,
		googleMapsClient,
		baseURL,
	)
}
