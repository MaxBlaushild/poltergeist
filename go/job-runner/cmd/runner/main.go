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

const (
	defaultPolymarketTradesURL  = "https://clob.polymarket.com/data/trades"
	defaultPolymarketBaseURL    = "https://clob.polymarket.com"
	defaultPolymarketTradesPath = "/data/trades"
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
	processRecurringQuestsProcessor := processors.NewProcessRecurringQuestsProcessor(dbClient)
	cleanupOrphanedQuestActionsProcessor := processors.NewCleanupOrphanedQuestActionsProcessor(dbClient)
	createProfilePictureProcessor := processors.NewCreateProfilePictureProcessor(dbClient, deepPriestClient, awsClient)
	generateOutfitProfilePictureProcessor := processors.NewGenerateOutfitProfilePictureProcessor(dbClient, deepPriestClient, awsClient)
	generateInventoryItemImageProcessor := processors.NewGenerateInventoryItemImageProcessor(dbClient, deepPriestClient, awsClient)
	generateSpellIconProcessor := processors.NewGenerateSpellIconProcessor(dbClient, deepPriestClient, awsClient)
	generateMonsterImageProcessor := processors.NewGenerateMonsterImageProcessor(dbClient, deepPriestClient, awsClient)
	generateMonsterTemplateImageProcessor := processors.NewGenerateMonsterTemplateImageProcessor(dbClient, deepPriestClient, awsClient)
	generateCharacterImageProcessor := processors.NewGenerateCharacterImageProcessor(dbClient, deepPriestClient, awsClient, client)
	generatePointOfInterestImageProcessor := processors.NewGeneratePointOfInterestImageProcessor(dbClient, locationSeederClient, client)
	generateScenarioImageProcessor := processors.NewGenerateScenarioImageProcessor(dbClient, deepPriestClient, awsClient)
	generateScenarioProcessor := processors.NewGenerateScenarioProcessor(dbClient, deepPriestClient, client)
	generateImageThumbnailProcessor := processors.NewGenerateImageThumbnailProcessor(dbClient, awsClient)
	queueThumbnailBackfillProcessor := processors.NewQueueThumbnailBackfillProcessor(dbClient, client)
	seedTreasureChestsProcessor := processors.NewSeedTreasureChestsProcessor(dbClient)
	calculateTrendingDestinationsProcessor := processors.NewCalculateTrendingDestinationsProcessor(dbClient)
	importPointOfInterestProcessor := processors.NewImportPointOfInterestProcessor(dbClient, locationSeederClient, client)
	importZonesForMetroProcessor := processors.NewImportZonesForMetroProcessor(dbClient)
	seedZoneDraftProcessor := processors.NewSeedZoneDraftProcessor(dbClient, googlemapsClient, deepPriestClient)
	applyZoneSeedDraftProcessor := processors.NewApplyZoneSeedDraftProcessor(dbClient, locationSeederClient, deepPriestClient, client)
	shuffleZoneSeedChallengeProcessor := processors.NewShuffleZoneSeedChallengeProcessor(dbClient)
	shuffleQuestNodeChallengeProcessor := processors.NewShuffleQuestNodeChallengeProcessor(dbClient, deepPriestClient)

	logPolymarketConfiguration(cfg)
	polymarketConfigHint := buildPolymarketConfigHint(cfg)

	polymarketClient := polymarket.NewClient(polymarket.ClientConfig{
		BaseURL:       defaultPolymarketBaseURL,
		TradesPath:    defaultPolymarketTradesPath,
		TradesURL:     defaultPolymarketTradesURL,
		APIKey:        cfg.Secret.PolymarketAPIKey,
		APISecret:     cfg.Secret.PolymarketAPISecret,
		APIPassphrase: cfg.Secret.PolymarketAPIPassphrase,
		Address:       cfg.Secret.PolymarketAddress,
	})
	log.Printf("Polymarket client initialized with fixed endpoint trades_url=%q", defaultPolymarketTradesURL)

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
		polymarketConfigHint,
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
	mux.Handle(jobs.ProcessRecurringQuestsTaskType, &processRecurringQuestsProcessor)
	mux.Handle(jobs.CleanupOrphanedQuestActionsTaskType, &cleanupOrphanedQuestActionsProcessor)
	mux.Handle(jobs.CreateProfilePictureTaskType, &createProfilePictureProcessor)
	mux.Handle(jobs.GenerateOutfitProfilePictureTaskType, &generateOutfitProfilePictureProcessor)
	mux.Handle(jobs.GenerateInventoryItemImageTaskType, &generateInventoryItemImageProcessor)
	mux.Handle(jobs.GenerateSpellIconTaskType, &generateSpellIconProcessor)
	mux.Handle(jobs.GenerateMonsterImageTaskType, &generateMonsterImageProcessor)
	mux.Handle(jobs.GenerateMonsterTemplateImageTaskType, &generateMonsterTemplateImageProcessor)
	mux.Handle(jobs.GenerateCharacterImageTaskType, &generateCharacterImageProcessor)
	mux.Handle(jobs.GeneratePointOfInterestImageTaskType, &generatePointOfInterestImageProcessor)
	mux.Handle(jobs.GenerateScenarioImageTaskType, &generateScenarioImageProcessor)
	mux.Handle(jobs.GenerateScenarioTaskType, &generateScenarioProcessor)
	mux.Handle(jobs.GenerateImageThumbnailTaskType, &generateImageThumbnailProcessor)
	mux.Handle(jobs.QueueThumbnailBackfillTaskType, &queueThumbnailBackfillProcessor)
	mux.Handle(jobs.SeedTreasureChestsTaskType, &seedTreasureChestsProcessor)
	mux.Handle(jobs.CalculateTrendingDestinationsTaskType, &calculateTrendingDestinationsProcessor)
	mux.Handle(jobs.ImportPointOfInterestTaskType, importPointOfInterestProcessor)
	mux.Handle(jobs.ImportZonesForMetroTaskType, importZonesForMetroProcessor)
	mux.Handle(jobs.SeedZoneDraftTaskType, &seedZoneDraftProcessor)
	mux.Handle(jobs.ApplyZoneSeedDraftTaskType, &applyZoneSeedDraftProcessor)
	mux.Handle(jobs.ShuffleZoneSeedChallengeTaskType, &shuffleZoneSeedChallengeProcessor)
	mux.Handle(jobs.ShuffleQuestNodeChallengeTaskType, &shuffleQuestNodeChallengeProcessor)
	mux.Handle(jobs.MonitorPolymarketTradesTaskType, monitorPolymarketTradesProcessor)
	if checkBlockchainTransactionsProcessor != nil {
		mux.Handle(jobs.CheckBlockchainTransactionsTaskType, checkBlockchainTransactionsProcessor)
	}

	scheduler := asynq.NewScheduler(redisConnOpt, &asynq.SchedulerOpts{})

	if _, err = scheduler.Register("@daily", asynq.NewTask(jobs.QueueQuestGenerationsTaskType, nil)); err != nil {
		log.Fatalf("could not register the schedule: %v", err)
	}

	if _, err = scheduler.Register("@every 15m", asynq.NewTask(jobs.ProcessRecurringQuestsTaskType, nil)); err != nil {
		log.Fatalf("could not register the recurring quest schedule: %v", err)
	}

	if _, err = scheduler.Register("@every 1h", asynq.NewTask(jobs.CleanupOrphanedQuestActionsTaskType, nil)); err != nil {
		log.Fatalf("could not register the orphaned quest action cleanup schedule: %v", err)
	}

	if _, err = scheduler.Register("@weekly", asynq.NewTask(jobs.SeedTreasureChestsTaskType, nil)); err != nil {
		log.Fatalf("could not register the schedule: %v", err)
	}

	if _, err = scheduler.Register("@every 6h", asynq.NewTask(jobs.CalculateTrendingDestinationsTaskType, nil)); err != nil {
		log.Fatalf("could not register the schedule: %v", err)
	}

	if _, err = scheduler.Register("@daily", asynq.NewTask(jobs.QueueThumbnailBackfillTaskType, nil)); err != nil {
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

func logPolymarketConfiguration(cfg *config.Config) {
	log.Printf(
		"Polymarket config: endpoint_source=fixed_constants trades_url=%q base_url=%q trades_path=%q legacy_trades_url_env_set=%t legacy_base_url_env_set=%t legacy_trades_path_env_set=%t alert_to_set=%t alert_from_set=%t limit=%d notional_threshold=%.2f size_threshold=%.2f api_key_set=%t api_secret_set=%t api_passphrase_set=%t address_set=%t",
		defaultPolymarketTradesURL,
		defaultPolymarketBaseURL,
		defaultPolymarketTradesPath,
		cfg.Public.PolymarketTradesURL != "",
		cfg.Public.PolymarketBaseURL != "",
		cfg.Public.PolymarketTradesPath != "",
		cfg.Public.PolymarketAlertToNumber != "",
		cfg.Public.PolymarketAlertFromNumber != "",
		cfg.Public.PolymarketTradesLimit,
		cfg.Public.PolymarketSuspiciousNotionalThreshold,
		cfg.Public.PolymarketSuspiciousSizeThreshold,
		cfg.Secret.PolymarketAPIKey != "",
		cfg.Secret.PolymarketAPISecret != "",
		cfg.Secret.PolymarketAPIPassphrase != "",
		cfg.Secret.PolymarketAddress != "",
	)

	missingL2 := make([]string, 0, 4)
	if cfg.Secret.PolymarketAPIKey == "" {
		missingL2 = append(missingL2, "POLYMARKET_API_KEY")
	}
	if cfg.Secret.PolymarketAPISecret == "" {
		missingL2 = append(missingL2, "POLYMARKET_API_SECRET")
	}
	if cfg.Secret.PolymarketAPIPassphrase == "" {
		missingL2 = append(missingL2, "POLYMARKET_API_PASSPHRASE")
	}
	if cfg.Secret.PolymarketAddress == "" {
		missingL2 = append(missingL2, "POLYMARKET_ADDRESS")
	}
	if len(missingL2) > 0 {
		log.Printf("Polymarket L2 credentials incomplete; missing=%v", missingL2)
	}
}

func buildPolymarketConfigHint(cfg *config.Config) string {
	return fmt.Sprintf(
		"endpoint_source=fixed_constants trades_url=%q base_url=%q trades_path=%q api_key_set=%t api_secret_set=%t api_passphrase_set=%t address_set=%t",
		defaultPolymarketTradesURL,
		defaultPolymarketBaseURL,
		defaultPolymarketTradesPath,
		cfg.Secret.PolymarketAPIKey != "",
		cfg.Secret.PolymarketAPISecret != "",
		cfg.Secret.PolymarketAPIPassphrase != "",
		cfg.Secret.PolymarketAddress != "",
	)
}
