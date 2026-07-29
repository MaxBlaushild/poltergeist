package main

import (
	"log"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/billing"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/email"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/reef-site/internal/config"
	"github.com/MaxBlaushild/poltergeist/reef-site/internal/server"
)

func main() {
	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		panic(err)
	}

	dbClient, err := db.NewClient(db.ClientConfig{
		Name:     cfg.Public.DbName,
		Host:     cfg.Public.DbHost,
		Port:     cfg.Public.DbPort,
		User:     cfg.Public.DbUser,
		Password: cfg.Secret.DbPassword,
	})
	if err != nil {
		panic(err)
	}

	awsClient := aws.NewAWSClient(cfg.Public.AwsRegion)
	jobsClient := jobs.NewClient(cfg.Public.RedisUrl)
	emailClient := email.NewClient(email.ClientConfig{
		AccountSid:  cfg.Secret.TwilioAccountSid,
		AuthToken:   cfg.Secret.TwilioAuthToken,
		FromAddress: cfg.Public.EmailFromAddress,
		FromName:    "reef",
		WebHost:     cfg.Public.SiteURL,
	})

	log.Println("reef-site listening on :8091")
	server.NewServer(server.Deps{
		DbClient:      dbClient,
		Config:        cfg,
		AwsClient:     awsClient,
		JobsClient:    jobsClient,
		EmailClient:   emailClient,
		BillingClient: billing.NewClient(),
		AuthClient:    auth.NewClient(),
	}).ListenAndServe("8091")
}
