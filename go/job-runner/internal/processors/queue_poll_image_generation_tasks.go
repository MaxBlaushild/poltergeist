package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/imagine"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/hibiken/asynq"
)

type QueueImageGenerationProcessor struct {
	dbClient               db.DbClient
	imageGenerationService imagine.ImagineClient
	asyncClient            *asynq.Client
}

const (
	imageGenerationTimeout           = time.Minute * 5
	QueuePollImageGenerationTaskType = "queue_poll_image_generation"
)

func NewQueuePollImageGenerationProcessor(dbClient db.DbClient, imageGenerationService imagine.ImagineClient, asyncClient *asynq.Client) QueueImageGenerationProcessor {
	return QueueImageGenerationProcessor{
		dbClient:               dbClient,
		imageGenerationService: imageGenerationService,
		asyncClient:            asyncClient,
	}
}

func (p *QueueImageGenerationProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	imgGens, err := p.dbClient.ImageGeneration().FindByState(ctx, models.GenerationStatusPending)
	if err != nil {
		return err
	}

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

	return nil
}
