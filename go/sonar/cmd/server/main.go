package main

import (
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/liveness"
	"github.com/MaxBlaushild/poltergeist/pkg/locationseeder"
	"github.com/MaxBlaushild/poltergeist/pkg/mapbox"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	"github.com/MaxBlaushild/poltergeist/pkg/useapi"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/charicturist"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/chat"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/config"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/gameengine"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/judge"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/quartermaster"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/questlog"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/search"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/server"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
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
	questlogClient := questlog.NewClient(dbClient)
	googlemapsClient := googlemaps.NewClient(cfg.Secret.GoogleMapsApiKey)
	locationSeeder := locationseeder.NewClient(googlemapsClient, dbClient, deepPriest, awsClient)
	dungeonmaster := dungeonmaster.NewClient(googlemapsClient, dbClient, deepPriest, locationSeeder, awsClient)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Public.RedisUrl,
		Password: "",
		DB:       0,
	})
	asyncClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.Public.RedisUrl})
	searchClient := search.NewSearchClient(dbClient, deepPriest)
	gameEngineClient := gameengine.NewGameEngineClient(dbClient, judgeClient, quartermaster, chatClient)
	livenessClient := liveness.NewClient(redisClient)
	s := server.NewServer(
		authClient,
		texterClient,
		dbClient,
		cfg,
		awsClient,
		judgeClient,
		quartermaster,
		chatClient,
		charicturist,
		mapboxClient,
		questlogClient,
		locationSeeder,
		googlemapsClient,
		dungeonmaster,
		asyncClient,
		redisClient,
		searchClient,
		gameEngineClient,
		livenessClient,
	)

	s.ListenAndServe("8042")
}
