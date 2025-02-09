package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/MaxBlaushild/job-runner/internal/config"
	"github.com/MaxBlaushild/job-runner/internal/processors"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/useapi"
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
		},
	)

	redisConnOpt := asynq.RedisClientOpt{Addr: cfg.Public.RedisUrl}
	client := asynq.NewClient(redisConnOpt)
	defer client.Close()

	awsClient := aws.NewAWSClient("us-east-1")

	useApiService := useapi.NewClient(cfg.Secret.UseApiKey)

	pollImageGenerationProcessor := processors.NewPollImageGenerationProcessor(dbClient, useApiService)
	pollImageUpscaleProcessor := processors.NewPollImageUpscaleProcessor(dbClient, useApiService, awsClient)
	queuePollImageGenerationProcessor := processors.NewQueuePollImageGenerationProcessor(dbClient, useApiService, client)

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

	mux.Handle(jobs.PollImageGenerationTaskType, &pollImageGenerationProcessor)
	mux.Handle(jobs.PollImageUpscaleTaskType, &pollImageUpscaleProcessor)
	mux.Handle(jobs.QueuePollImageGenerationTaskType, &queuePollImageGenerationProcessor)

	scheduler := asynq.NewScheduler(redisConnOpt, &asynq.SchedulerOpts{})

	// Schedule the task to run every 30 seconds.
	if _, err = scheduler.Register("@every 30s", asynq.NewTask(jobs.QueuePollImageGenerationTaskType, nil)); err != nil {
		log.Fatalf("could not register the schedule: %v", err)
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

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
