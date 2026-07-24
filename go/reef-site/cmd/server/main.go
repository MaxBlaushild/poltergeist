package main

import (
	"log"

	"github.com/MaxBlaushild/poltergeist/pkg/aws"
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
		ApiKey:      cfg.Secret.EmailApiKey,
		FromAddress: cfg.Public.EmailFromAddress,
		WebHost:     cfg.Public.BaseURL,
	})

	log.Println("reef-site listening on :8091")
	server.NewServer(server.Deps{
		DbClient:    dbClient,
		Config:      cfg,
		AwsClient:   awsClient,
		JobsClient:  jobsClient,
		EmailClient: emailClient,
	}).ListenAndServe("8091")
}
