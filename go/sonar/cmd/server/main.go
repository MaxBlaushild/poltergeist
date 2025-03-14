package main

import (
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/mapbox"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/MaxBlaushild/poltergeist/pkg/useapi"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/charicturist"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/chat"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/config"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/judge"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/quartermaster"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/server"
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

	deepPriest := deep_priest.SummonDeepPriest()
	texterClient := texter.NewClient()
	authClient := auth.NewClient()
	awsClient := aws.NewAWSClient("us-east-1")
	judgeClient := judge.NewClient(awsClient, dbClient, deepPriest)
	quartermaster := quartermaster.NewClient(dbClient)
	chatClient := chat.NewClient(dbClient, quartermaster)
	useApiClient := useapi.NewClient(cfg.Secret.UseApiKey)
	charicturist := charicturist.NewClient(useApiClient, dbClient)
	mapboxClient := mapbox.NewClient(cfg.Secret.MapboxApiKey)
	s := server.NewServer(authClient, texterClient, dbClient, cfg, awsClient, judgeClient, quartermaster, chatClient, charicturist, mapboxClient)

	s.ListenAndServe("8042")
}
