package pkg

import (
	"log"

	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/billing"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/email"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/reef-site/internal/config"
	"github.com/MaxBlaushild/poltergeist/reef-site/internal/server"
	"github.com/gin-gonic/gin"
)

// Server mirrors the shape every other domain in this repo exposes
// (go/vampire-ascendancy/pkg, go/sonar/pkg, ...) so go/core can compose it
// into the single combined router the same way it composes everything else.
type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r *gin.Engine)
}

// NewServerFromDependencies lets core compose reef-site the same way it
// composes the other folded-in modules, sharing the one DB client/connection
// pool rather than opening a second one. reef-site's own config is loaded
// from the process environment here (not threaded through from core) because
// go/reef-site/internal/config is only importable from within this module —
// the same reason go/vampire-ascendancy/pkg reads REDIS_URL via os.Getenv
// internally instead of accepting it as a parameter.
func NewServerFromDependencies(dbClient db.DbClient) Server {
	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		log.Printf("[reef-site] failed to load config from env: %v", err)
		cfg = &config.Config{}
	}

	awsClient := aws.NewAWSClient(cfg.Public.AwsRegion)
	jobsClient := jobs.NewClient(cfg.Public.RedisUrl)
	emailClient := email.NewClient(email.ClientConfig{
		ApiKey:      cfg.Secret.EmailApiKey,
		FromAddress: cfg.Public.EmailFromAddress,
		WebHost:     cfg.Public.BaseURL,
	})

	return server.NewServer(server.Deps{
		DbClient:      dbClient,
		Config:        cfg,
		AwsClient:     awsClient,
		JobsClient:    jobsClient,
		EmailClient:   emailClient,
		BillingClient: billing.NewClient(),
	})
}
