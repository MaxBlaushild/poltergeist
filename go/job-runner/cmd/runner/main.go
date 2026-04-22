package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/MaxBlaushild/job-runner/internal/config"
	"github.com/MaxBlaushild/job-runner/internal/processors"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/locationseeder"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
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
	redisClient := newRedisClient(cfg.Public.RedisUrl)
	defer redisClient.Close()

	awsClient := aws.NewAWSClient("us-east-1")

	googlemapsClient := googlemaps.NewClient(cfg.Secret.GoogleMapsApiKey)
	deepPriestClient := deep_priest.SummonDeepPriest()
	locationSeederClient := locationseeder.NewClient(googlemapsClient, dbClient, deepPriestClient, awsClient)
	dungeonmasterClient := dungeonmaster.NewClient(googlemapsClient, dbClient, deepPriestClient, locationSeederClient, awsClient, client)

	generateQuestForZoneProcessor := processors.NewGenerateQuestForZoneProcessor(dbClient, dungeonmasterClient)
	queueQuestGenerationsProcessor := processors.NewQueueQuestGenerationsProcessor(dbClient, dungeonmasterClient, client)
	processRecurringQuestsProcessor := processors.NewProcessRecurringQuestsProcessor(dbClient)
	processRecurringStandaloneContentProcessor := processors.NewProcessRecurringStandaloneContentProcessor(dbClient)
	cleanupOrphanedQuestActionsProcessor := processors.NewCleanupOrphanedQuestActionsProcessor(dbClient)
	createProfilePictureProcessor := processors.NewCreateProfilePictureProcessor(dbClient, deepPriestClient, awsClient)
	generateOutfitProfilePictureProcessor := processors.NewGenerateOutfitProfilePictureProcessor(dbClient, deepPriestClient, awsClient)
	generateInventoryItemImageProcessor := processors.NewGenerateInventoryItemImageProcessor(dbClient, deepPriestClient, awsClient)
	generateSpellIconProcessor := processors.NewGenerateSpellIconProcessor(dbClient, deepPriestClient, awsClient)
	generateSpellsBulkProcessor := processors.NewGenerateSpellsBulkProcessor(dbClient, redisClient)
	generateSpellProgressionFromPromptProcessor := processors.NewGenerateSpellProgressionFromPromptProcessor(dbClient, redisClient, deepPriestClient)
	rebalanceSpellDamageProcessor := processors.NewRebalanceSpellDamageProcessor(dbClient, redisClient)
	generateMonsterImageProcessor := processors.NewGenerateMonsterImageProcessor(dbClient, deepPriestClient, awsClient)
	generateMonsterTemplateImageProcessor := processors.NewGenerateMonsterTemplateImageProcessor(dbClient, deepPriestClient, awsClient)
	generateMonsterTemplatesBulkProcessor := processors.NewGenerateMonsterTemplatesBulkProcessor(dbClient, redisClient, deepPriestClient)
	refreshMonsterTemplateAffinitiesProcessor := processors.NewRefreshMonsterTemplateAffinitiesProcessor(dbClient, redisClient, deepPriestClient)
	resetMonsterTemplateProgressionsProcessor := processors.NewResetMonsterTemplateProgressionsProcessor(dbClient, redisClient)
	processMainStoryDistrictRunProcessor := processors.NewProcessMainStoryDistrictRunProcessor(dbClient, dungeonmasterClient)
	generateCharacterImageProcessor := processors.NewGenerateCharacterImageProcessor(dbClient, deepPriestClient, awsClient, client)
	generatePointOfInterestImageProcessor := processors.NewGeneratePointOfInterestImageProcessor(dbClient, locationSeederClient, client)
	generateScenarioImageProcessor := processors.NewGenerateScenarioImageProcessor(dbClient, deepPriestClient, awsClient)
	generateExpositionImageProcessor := processors.NewGenerateExpositionImageProcessor(dbClient, deepPriestClient, awsClient)
	generateTutorialImageProcessor := processors.NewGenerateTutorialImageProcessor(dbClient, deepPriestClient, awsClient)
	instantiateTutorialBaseQuestProcessor := processors.NewInstantiateTutorialBaseQuestProcessor(dbClient, dungeonmasterClient)
	generateChallengeImageProcessor := processors.NewGenerateChallengeImageProcessor(dbClient, deepPriestClient, awsClient)
	generateChallengeTemplateImageProcessor := processors.NewGenerateChallengeTemplateImageProcessor(dbClient, deepPriestClient, awsClient)
	generateInventoryItemSuggestionsProcessor := processors.NewGenerateInventoryItemSuggestionsProcessor(dbClient, deepPriestClient)
	generateBaseStructureLevelImageProcessor := processors.NewGenerateBaseStructureLevelImageProcessor(dbClient, deepPriestClient, awsClient, client)
	generateScenarioProcessor := processors.NewGenerateScenarioProcessor(dbClient, deepPriestClient, client)
	generateChallengesProcessor := processors.NewGenerateChallengesProcessor(dbClient, deepPriestClient, client)
	generateScenarioTemplatesProcessor := processors.NewGenerateScenarioTemplatesProcessor(dbClient, deepPriestClient)
	generateChallengeTemplatesProcessor := processors.NewGenerateChallengeTemplatesProcessor(dbClient, deepPriestClient)
	generateLocationArchetypesProcessor := processors.NewGenerateLocationArchetypesProcessor(dbClient, deepPriestClient)
	generateQuestArchetypeSuggestionsProcessor := processors.NewGenerateQuestArchetypeSuggestionsProcessor(dbClient, deepPriestClient)
	generateMainStorySuggestionsProcessor := processors.NewGenerateMainStorySuggestionsProcessor(dbClient, deepPriestClient)
	generateZoneFlavorProcessor := processors.NewGenerateZoneFlavorProcessor(dbClient, deepPriestClient)
	generateZoneTagsProcessor := processors.NewGenerateZoneTagsProcessor(dbClient, deepPriestClient)
	generateZoneKindPatternTileProcessor := processors.NewGenerateZoneKindPatternTileProcessor(dbClient, deepPriestClient, awsClient)
	generateImageThumbnailProcessor := processors.NewGenerateImageThumbnailProcessor(dbClient, awsClient)
	queueThumbnailBackfillProcessor := processors.NewQueueThumbnailBackfillProcessor(dbClient, client)
	seedTreasureChestsProcessor := processors.NewSeedTreasureChestsProcessor(dbClient)
	calculateTrendingDestinationsProcessor := processors.NewCalculateTrendingDestinationsProcessor(dbClient)
	importPointOfInterestProcessor := processors.NewImportPointOfInterestProcessor(dbClient, locationSeederClient, client)
	importZonesForMetroProcessor := processors.NewImportZonesForMetroProcessor(dbClient)
	seedZoneDraftProcessor := processors.NewSeedZoneDraftProcessor(dbClient, googlemapsClient, deepPriestClient)
	seedDistrictProcessor := processors.NewSeedDistrictProcessor(dbClient, deepPriestClient, dungeonmasterClient, locationSeederClient, client)
	applyZoneSeedDraftProcessor := processors.NewApplyZoneSeedDraftProcessor(dbClient, locationSeederClient, deepPriestClient, client)
	shuffleZoneSeedChallengeProcessor := processors.NewShuffleZoneSeedChallengeProcessor(dbClient)
	backfillContentZoneKindsProcessor := processors.NewBackfillContentZoneKindsProcessor(dbClient, redisClient)

	// logPolymarketConfiguration(cfg)
	// polymarketConfigHint := buildPolymarketConfigHint(cfg)

	// polymarketClient := polymarket.NewClient(polymarket.ClientConfig{
	// 	BaseURL:       defaultPolymarketBaseURL,
	// 	TradesPath:    defaultPolymarketTradesPath,
	// 	TradesURL:     defaultPolymarketTradesURL,
	// 	APIKey:        cfg.Secret.PolymarketAPIKey,
	// 	APISecret:     cfg.Secret.PolymarketAPISecret,
	// 	APIPassphrase: cfg.Secret.PolymarketAPIPassphrase,
	// 	Address:       cfg.Secret.PolymarketAddress,
	// })
	// log.Printf("Polymarket client initialized with fixed endpoint trades_url=%q", defaultPolymarketTradesURL)

	// texterClient := texter.NewClient()
	// monitorPolymarketTradesProcessor := processors.NewMonitorPolymarketTradesProcessor(
	// 	dbClient,
	// 	polymarketClient,
	// 	texterClient,
	// 	cfg.Public.PolymarketAlertToNumber,
	// 	cfg.Public.PolymarketAlertFromNumber,
	// 	cfg.Public.PolymarketSuspiciousNotionalThreshold,
	// 	cfg.Public.PolymarketSuspiciousSizeThreshold,
	// 	cfg.Public.PolymarketTradesLimit,
	// 	polymarketConfigHint,
	// )

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
	mux.Handle(jobs.ProcessRecurringStandaloneContentTaskType, &processRecurringStandaloneContentProcessor)
	mux.Handle(jobs.CleanupOrphanedQuestActionsTaskType, &cleanupOrphanedQuestActionsProcessor)
	mux.Handle(jobs.CreateProfilePictureTaskType, &createProfilePictureProcessor)
	mux.Handle(jobs.GenerateOutfitProfilePictureTaskType, &generateOutfitProfilePictureProcessor)
	mux.Handle(jobs.GenerateInventoryItemImageTaskType, &generateInventoryItemImageProcessor)
	mux.Handle(jobs.GenerateSpellIconTaskType, &generateSpellIconProcessor)
	mux.Handle(jobs.GenerateSpellsBulkTaskType, &generateSpellsBulkProcessor)
	mux.Handle(jobs.GenerateSpellProgressionFromPromptTaskType, &generateSpellProgressionFromPromptProcessor)
	mux.Handle(jobs.RebalanceSpellDamageTaskType, &rebalanceSpellDamageProcessor)
	mux.Handle(jobs.GenerateMonsterImageTaskType, &generateMonsterImageProcessor)
	mux.Handle(jobs.GenerateMonsterTemplateImageTaskType, &generateMonsterTemplateImageProcessor)
	mux.Handle(jobs.GenerateMonsterTemplatesBulkTaskType, &generateMonsterTemplatesBulkProcessor)
	mux.Handle(jobs.RefreshMonsterTemplateAffinitiesTaskType, &refreshMonsterTemplateAffinitiesProcessor)
	mux.Handle(jobs.ResetMonsterTemplateProgressionsTaskType, &resetMonsterTemplateProgressionsProcessor)
	mux.Handle(jobs.ProcessMainStoryDistrictRunTaskType, &processMainStoryDistrictRunProcessor)
	mux.Handle(jobs.GenerateCharacterImageTaskType, &generateCharacterImageProcessor)
	mux.Handle(jobs.GeneratePointOfInterestImageTaskType, &generatePointOfInterestImageProcessor)
	mux.Handle(jobs.GenerateScenarioImageTaskType, &generateScenarioImageProcessor)
	mux.Handle(jobs.GenerateExpositionImageTaskType, &generateExpositionImageProcessor)
	mux.Handle(jobs.GenerateTutorialImageTaskType, &generateTutorialImageProcessor)
	mux.Handle(jobs.InstantiateTutorialBaseQuestTaskType, &instantiateTutorialBaseQuestProcessor)
	mux.Handle(jobs.GenerateChallengeImageTaskType, &generateChallengeImageProcessor)
	mux.Handle(jobs.GenerateChallengeTemplateImageTaskType, &generateChallengeTemplateImageProcessor)
	mux.Handle(jobs.GenerateInventoryItemSuggestionsTaskType, &generateInventoryItemSuggestionsProcessor)
	mux.Handle(jobs.GenerateBaseStructureLevelImageTaskType, &generateBaseStructureLevelImageProcessor)
	mux.Handle(jobs.GenerateBaseStructureLevelTopDownImageTaskType, &generateBaseStructureLevelImageProcessor)
	mux.Handle(jobs.GenerateScenarioTaskType, &generateScenarioProcessor)
	mux.Handle(jobs.GenerateChallengesTaskType, &generateChallengesProcessor)
	mux.Handle(jobs.GenerateScenarioTemplatesTaskType, &generateScenarioTemplatesProcessor)
	mux.Handle(jobs.GenerateChallengeTemplatesTaskType, &generateChallengeTemplatesProcessor)
	mux.Handle(jobs.GenerateLocationArchetypesTaskType, &generateLocationArchetypesProcessor)
	mux.Handle(jobs.GenerateQuestArchetypeSuggestionsTaskType, &generateQuestArchetypeSuggestionsProcessor)
	mux.Handle(jobs.GenerateMainStorySuggestionsTaskType, &generateMainStorySuggestionsProcessor)
	mux.Handle(jobs.GenerateZoneFlavorTaskType, &generateZoneFlavorProcessor)
	mux.Handle(jobs.GenerateZoneTagsTaskType, &generateZoneTagsProcessor)
	mux.Handle(jobs.GenerateZoneKindPatternTileTaskType, &generateZoneKindPatternTileProcessor)
	mux.Handle(jobs.GenerateBaseDescriptionTaskType, asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		log.Printf("Discarding legacy task %s because player base flavor generation is disabled", t.Type())
		return nil
	}))
	mux.Handle(jobs.GenerateImageThumbnailTaskType, &generateImageThumbnailProcessor)
	mux.Handle(jobs.QueueThumbnailBackfillTaskType, &queueThumbnailBackfillProcessor)
	mux.Handle(jobs.SeedTreasureChestsTaskType, &seedTreasureChestsProcessor)
	mux.Handle(jobs.CalculateTrendingDestinationsTaskType, &calculateTrendingDestinationsProcessor)
	mux.Handle(jobs.ImportPointOfInterestTaskType, importPointOfInterestProcessor)
	mux.Handle(jobs.ImportZonesForMetroTaskType, importZonesForMetroProcessor)
	mux.Handle(jobs.SeedZoneDraftTaskType, &seedZoneDraftProcessor)
	mux.Handle(jobs.SeedDistrictTaskType, &seedDistrictProcessor)
	mux.Handle(jobs.ApplyZoneSeedDraftTaskType, &applyZoneSeedDraftProcessor)
	mux.Handle(jobs.ShuffleZoneSeedChallengeTaskType, &shuffleZoneSeedChallengeProcessor)
	mux.Handle(jobs.BackfillContentZoneKindsTaskType, &backfillContentZoneKindsProcessor)
	mux.Handle(jobs.MonitorPolymarketTradesTaskType, asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		log.Printf("Discarding legacy task %s because Polymarket monitoring is disabled", t.Type())
		return nil
	}))
	mux.Handle(jobs.CheckBlockchainTransactionsTaskType, asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		log.Printf("Discarding task %s because blockchain transaction polling is temporarily disabled", t.Type())
		return nil
	}))

	scheduler := asynq.NewScheduler(redisConnOpt, &asynq.SchedulerOpts{})

	if _, err = scheduler.Register("@daily", asynq.NewTask(jobs.QueueQuestGenerationsTaskType, nil)); err != nil {
		log.Fatalf("could not register the schedule: %v", err)
	}

	if _, err = scheduler.Register("@every 15m", asynq.NewTask(jobs.ProcessRecurringQuestsTaskType, nil)); err != nil {
		log.Fatalf("could not register the recurring quest schedule: %v", err)
	}

	if _, err = scheduler.Register("@every 15m", asynq.NewTask(jobs.ProcessRecurringStandaloneContentTaskType, nil)); err != nil {
		log.Fatalf("could not register the recurring standalone content schedule: %v", err)
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

	// if _, err = scheduler.Register("@every 1m", asynq.NewTask(jobs.MonitorPolymarketTradesTaskType, nil)); err != nil {
	// 	log.Fatalf("could not register the polymarket trades monitor schedule: %v", err)
	// }

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

// func logPolymarketConfiguration(cfg *config.Config) {
// 	log.Printf(
// 		"Polymarket config: endpoint_source=fixed_constants trades_url=%q base_url=%q trades_path=%q legacy_trades_url_env_set=%t legacy_base_url_env_set=%t legacy_trades_path_env_set=%t alert_to_set=%t alert_from_set=%t limit=%d notional_threshold=%.2f size_threshold=%.2f api_key_set=%t api_secret_set=%t api_passphrase_set=%t address_set=%t",
// 		defaultPolymarketTradesURL,
// 		defaultPolymarketBaseURL,
// 		defaultPolymarketTradesPath,
// 		cfg.Public.PolymarketTradesURL != "",
// 		cfg.Public.PolymarketBaseURL != "",
// 		cfg.Public.PolymarketTradesPath != "",
// 		cfg.Public.PolymarketAlertToNumber != "",
// 		cfg.Public.PolymarketAlertFromNumber != "",
// 		cfg.Public.PolymarketTradesLimit,
// 		cfg.Public.PolymarketSuspiciousNotionalThreshold,
// 		cfg.Public.PolymarketSuspiciousSizeThreshold,
// 		cfg.Secret.PolymarketAPIKey != "",
// 		cfg.Secret.PolymarketAPISecret != "",
// 		cfg.Secret.PolymarketAPIPassphrase != "",
// 		cfg.Secret.PolymarketAddress != "",
// 	)

// 	missingL2 := make([]string, 0, 4)
// 	if cfg.Secret.PolymarketAPIKey == "" {
// 		missingL2 = append(missingL2, "POLYMARKET_API_KEY")
// 	}
// 	if cfg.Secret.PolymarketAPISecret == "" {
// 		missingL2 = append(missingL2, "POLYMARKET_API_SECRET")
// 	}
// 	if cfg.Secret.PolymarketAPIPassphrase == "" {
// 		missingL2 = append(missingL2, "POLYMARKET_API_PASSPHRASE")
// 	}
// 	if cfg.Secret.PolymarketAddress == "" {
// 		missingL2 = append(missingL2, "POLYMARKET_ADDRESS")
// 	}
// 	if len(missingL2) > 0 {
// 		log.Printf("Polymarket L2 credentials incomplete; missing=%v", missingL2)
// }
// }

func newRedisClient(redisURL string) *redis.Client {
	trimmed := strings.TrimSpace(redisURL)
	if trimmed == "" {
		log.Fatal("redis URL is required")
	}

	var (
		opt *redis.Options
		err error
	)
	if strings.HasPrefix(trimmed, "redis://") || strings.HasPrefix(trimmed, "rediss://") {
		opt, err = redis.ParseURL(trimmed)
		if err != nil {
			log.Fatalf("could not parse redis url: %v", err)
		}
	} else {
		opt = &redis.Options{Addr: trimmed}
	}

	return redis.NewClient(opt)
}

// func buildPolymarketConfigHint(cfg *config.Config) string {
// 	return fmt.Sprintf(
// 		"endpoint_source=fixed_constants trades_url=%q base_url=%q trades_path=%q api_key_set=%t api_secret_set=%t api_passphrase_set=%t address_set=%t",
// 		defaultPolymarketTradesURL,
// 		defaultPolymarketBaseURL,
// 		defaultPolymarketTradesPath,
// 		cfg.Secret.PolymarketAPIKey != "",
// 		cfg.Secret.PolymarketAPISecret != "",
// 		cfg.Secret.PolymarketAPIPassphrase != "",
// 		cfg.Secret.PolymarketAddress != "",
// 	)
// }
