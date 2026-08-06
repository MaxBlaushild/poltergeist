package pkg

import (
	"log"

	"github.com/MaxBlaushild/poltergeist/bgi-site/internal/config"
	"github.com/MaxBlaushild/poltergeist/bgi-site/internal/server"
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/billing"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/email"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/gin-gonic/gin"
)

// Server mirrors go/reef-site/pkg's shape so go/core can compose bgi-site
// into the single combined router the same way it composes every other
// domain (see go/bgi-site/PLATFORM_FINDINGS.md).
type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r *gin.Engine)
}

// NewServerFromDependencies lets core compose bgi-site the same way it
// composes reef-site, sharing the one DB client/connection pool. bgi-site's
// own config is loaded from the process environment here, not threaded
// through from core, for the same reason reef-site's does — its internal
// config package is only importable from within this module.
func NewServerFromDependencies(dbClient db.DbClient) Server {
	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		log.Printf("[bgi-site] failed to load config from env: %v", err)
		cfg = &config.Config{}
	}

	awsClient := aws.NewAWSClient(cfg.Public.AwsRegion)
	jobsClient := jobs.NewClient(cfg.Public.RedisUrl)
	emailClient := email.NewClient(email.ClientConfig{
		AccountSid:  cfg.Secret.TwilioAccountSid,
		AuthToken:   cfg.Secret.TwilioAuthToken,
		FromAddress: cfg.Public.EmailFromAddress,
		FromName:    "bgi",
		WebHost:     cfg.Public.SiteURL,
	})

	return server.NewServer(server.Deps{
		DbClient:      dbClient,
		Config:        cfg,
		AwsClient:     awsClient,
		JobsClient:    jobsClient,
		EmailClient:   emailClient,
		BillingClient: billing.NewClient(),
		AuthClient:    auth.NewClient(),
	})
}
