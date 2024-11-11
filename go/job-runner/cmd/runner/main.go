package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MaxBlaushild/job-runner/internal/config"
	"github.com/MaxBlaushild/job-runner/internal/processors"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/imagine"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
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

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.Public.RedisUrl})
	defer client.Close()

	imageGenerationService := imagine.NewClient(cfg.Secret.ImagineApiKey)

	pollImageGenerationProcessor := processors.NewPollImageGenerationProcessor(dbClient, imageGenerationService)
	queuePollImageGenerationProcessor := processors.NewQueuePollImageGenerationProcessor(dbClient, imageGenerationService, client)
	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.Handle(jobs.PollImageGenerationTaskType, &pollImageGenerationProcessor)
	mux.Handle(jobs.QueuePollImageGenerationTaskType, &queuePollImageGenerationProcessor)

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
