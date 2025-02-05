package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/useapi"
	"github.com/davecgh/go-spew/spew"
	"github.com/hibiken/asynq"
)

type QueueImageGenerationProcessor struct {
	dbClient     db.DbClient
	useApiClient useapi.Client
	asyncClient  *asynq.Client
}

const (
	imageGenerationTimeout = time.Minute * 5
)

func NewQueuePollImageGenerationProcessor(dbClient db.DbClient, useApiClient useapi.Client, asyncClient *asynq.Client) QueueImageGenerationProcessor {
	return QueueImageGenerationProcessor{
		dbClient:     dbClient,
		useApiClient: useApiClient,
		asyncClient:  asyncClient,
	}
}

func (p *QueueImageGenerationProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	imgGens, err := p.dbClient.ImageGeneration().FindByState(ctx, models.GenerationStatusPending)
	if err != nil {
		return err
	}

	spew.Dump("imgGens")
	spew.Dump(len(imgGens))
	for _, imgGen := range imgGens {
		payload, err := json.Marshal(PollImageGenerationTaskPayload{
			ID: imgGen.ID.String(),
		})
		if err != nil {
			return err
		}
		if _, err := p.asyncClient.Enqueue(asynq.NewTask(jobs.PollImageGenerationTaskType, payload)); err != nil {
			fmt.Errorf("error enqueuing poll image generation task: %w", err)
		}
	}

	upscaleImgGens, err := p.dbClient.ImageGeneration().FindByState(ctx, models.GenerateImageOptions)
	if err != nil {
		return err
	}

	spew.Dump("upscaleImgGens")
	spew.Dump(len(upscaleImgGens))

	for _, gen := range upscaleImgGens {
		payload, err := json.Marshal(PollImageUpscaleTaskPayload{
			ID: gen.ID.String(),
		})
		if err != nil {
			return err
		}
		if _, err := p.asyncClient.Enqueue(asynq.NewTask(jobs.PollImageUpscaleTaskType, payload)); err != nil {
			fmt.Errorf("error enqueuing poll image upscale task: %w", err)
		}
	}

	return nil
}
