package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MaxBlaushild/job-runner/internal/config"
	"github.com/MaxBlaushild/job-runner/internal/processors"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster"
	"github.com/MaxBlaushild/poltergeist/pkg/ethereum"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/locationseeder"
	"github.com/MaxBlaushild/poltergeist/pkg/polymarket"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"

	"github.com/hibiken/asynq"
)

func main() {
	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		log.Fatalf("could not get config: %v", err)
	}

	fmt.Println(cfg.Public.RedisUrl)

	dbClient, err := db.NewClient(db.ClientConfig{
		Host:     cfg.Public.DbHost,
		User:     cfg.Public.DbUser,
		Port:     cfg.Public.DbPort,
		Name:     cfg.Public.DbName,
		Password: cfg.Secret.DbPassword,
	})
	if err != nil {
		log.Fatalf("could not connect to db: %v", err)
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Public.RedisUrl},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ShutdownTimeout: 5 * time.Minute, // Allow tasks up to 5 minutes to complete during shutdown
		},
	)

	redisConnOpt := asynq.RedisClientOpt{Addr: cfg.Public.RedisUrl}
	client := asynq.NewClient(redisConnOpt)
	defer client.Close()

	awsClient := aws.NewAWSClient("us-east-1")

	googlemapsClient := googlemaps.NewClient(cfg.Secret.GoogleMapsApiKey)
	deepPriestClient := deep_priest.SummonDeepPriest()
	locationSeederClient := locationseeder.NewClient(googlemapsClient, dbClient, deepPriestClient, awsClient)
	dungeonmasterClient := dungeonmaster.NewClient(googlemapsClient, dbClient, deepPriestClient, locationSeederClient, awsClient)

	generateQuestForZoneProcessor := processors.NewGenerateQuestForZoneProcessor(dbClient, dungeonmasterClient)
	queueQuestGenerationsProcessor := processors.NewQueueQuestGenerationsProcessor(dbClient, dungeonmasterClient, client)
	createProfilePictureProcessor := processors.NewCreateProfilePictureProcessor(dbClient, deepPriestClient, awsClient)
	generateInventoryItemImageProcessor := processors.NewGenerateInventoryItemImageProcessor(dbClient, deepPriestClient, awsClient)
	generateCharacterImageProcessor := processors.NewGenerateCharacterImageProcessor(dbClient, deepPriestClient, awsClient)
	seedTreasureChestsProcessor := processors.NewSeedTreasureChestsProcessor(dbClient)
	calculateTrendingDestinationsProcessor := processors.NewCalculateTrendingDestinationsProcessor(dbClient)
	importPointOfInterestProcessor := processors.NewImportPointOfInterestProcessor(dbClient, locationSeederClient)

	var polymarketClient polymarket.Client
	if cfg.Public.PolymarketTradesURL != "" || cfg.Public.PolymarketBaseURL != "" {
		polymarketClient = polymarket.NewClient(polymarket.ClientConfig{
			BaseURL:    cfg.Public.PolymarketBaseURL,
			TradesPath: cfg.Public.PolymarketTradesPath,
			TradesURL:  cfg.Public.PolymarketTradesURL,
			APIKey:     cfg.Secret.PolymarketAPIKey,
		})
	}

	texterClient := texter.NewClient()
	monitorPolymarketTradesProcessor := processors.NewMonitorPolymarketTradesProcessor(
		dbClient,
		polymarketClient,
		texterClient,
		cfg.Public.PolymarketAlertToNumber,
		cfg.Public.PolymarketAlertFromNumber,
		cfg.Public.PolymarketSuspiciousNotionalThreshold,
		cfg.Public.PolymarketSuspiciousSizeThreshold,
		cfg.Public.PolymarketTradesLimit,
	)

	// Initialize Ethereum client for blockchain transaction checking (read-only)
	var checkBlockchainTransactionsProcessor *processors.CheckBlockchainTransactionsProcessor
	if cfg.Public.RPCURL != "" && cfg.Public.ChainID != 0 {
		ethereumClient, err := ethereum.NewReadOnlyClient(cfg.Public.RPCURL, cfg.Public.ChainID)
		if err != nil {
			log.Printf("Warning: Failed to create Ethereum client for transaction checking: %v", err)
		} else {
			checkBlockchainTransactionsProcessor = new(processors.CheckBlockchainTransactionsProcessor)
			*checkBlockchainTransactionsProcessor = processors.NewCheckBlockchainTransactionsProcessor(dbClient, ethereumClient)
		}
	}

	mux := asynq.NewServeMux()

	// Add error logging middleware to each handler
	mux.Use(func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			err := h.ProcessTask(ctx, t)
			if err != nil {
				log.Printf("Failed to process task %s: %v\nStack trace:\n%+v", t.Type(), err, err)
			}
			return err
		})
	})

	mux.Handle(jobs.GenerateQuestForZoneTaskType, &generateQuestForZoneProcessor)
	mux.Handle(jobs.QueueQuestGenerationsTaskType, &queueQuestGenerationsProcessor)
	mux.Handle(jobs.CreateProfilePictureTaskType, &createProfilePictureProcessor)
	mux.Handle(jobs.GenerateInventoryItemImageTaskType, &generateInventoryItemImageProcessor)
	mux.Handle(jobs.GenerateCharacterImageTaskType, &generateCharacterImageProcessor)
	mux.Handle(jobs.SeedTreasureChestsTaskType, &seedTreasureChestsProcessor)
	mux.Handle(jobs.CalculateTrendingDestinationsTaskType, &calculateTrendingDestinationsProcessor)
	mux.Handle(jobs.ImportPointOfInterestTaskType, importPointOfInterestProcessor)
	mux.Handle(jobs.MonitorPolymarketTradesTaskType, monitorPolymarketTradesProcessor)
	if checkBlockchainTransactionsProcessor != nil {
		mux.Handle(jobs.CheckBlockchainTransactionsTaskType, checkBlockchainTransactionsProcessor)
	}

	scheduler := asynq.NewScheduler(redisConnOpt, &asynq.SchedulerOpts{})

	if _, err = scheduler.Register("@daily", asynq.NewTask(jobs.QueueQuestGenerationsTaskType, nil)); err != nil {
		log.Fatalf("could not register the schedule: %v", err)
	}

	if _, err = scheduler.Register("@weekly", asynq.NewTask(jobs.SeedTreasureChestsTaskType, nil)); err != nil {
		log.Fatalf("could not register the schedule: %v", err)
	}

	if _, err = scheduler.Register("@every 6h", asynq.NewTask(jobs.CalculateTrendingDestinationsTaskType, nil)); err != nil {
		log.Fatalf("could not register the schedule: %v", err)
	}

	if checkBlockchainTransactionsProcessor != nil {
		if _, err = scheduler.Register("@every 15s", asynq.NewTask(jobs.CheckBlockchainTransactionsTaskType, nil)); err != nil {
			log.Fatalf("could not register the blockchain transactions check schedule: %v", err)
		}
	}

	if _, err = scheduler.Register("@every 1m", asynq.NewTask(jobs.MonitorPolymarketTradesTaskType, nil)); err != nil {
		log.Fatalf("could not register the polymarket trades monitor schedule: %v", err)
	}

	go func() {
		if err := scheduler.Run(); err != nil {
			log.Fatalf("could not run scheduler: %v", err)
		}
	}()

	// Start health check server
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		if err := http.ListenAndServe(":9013", nil); err != nil {
			log.Fatalf("could not start health check server: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		<-sigChan
		log.Println("Received shutdown signal. Draining queue...")

		// Stop accepting new tasks and wait for in-progress tasks to complete
		srv.Shutdown()

		// Stop the scheduler
		scheduler.Shutdown()
	}()

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
