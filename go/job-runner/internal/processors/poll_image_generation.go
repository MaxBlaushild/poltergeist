package processors

import (
	"context"
	"encoding/json"
	"time"

	"cosmossdk.io/errors"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/imagine"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type PollImageGenerationProcessor struct {
	dbClient               db.DbClient
	imageGenerationService imagine.ImagineClient
}

type PollImageGenerationTaskPayload struct {
	ID string `json:"id"`
}

func NewPollImageGenerationProcessor(dbClient db.DbClient, imageGenerationService imagine.ImagineClient) PollImageGenerationProcessor {
	return PollImageGenerationProcessor{
		dbClient:               dbClient,
		imageGenerationService: imageGenerationService,
	}
}

func (p *PollImageGenerationProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload PollImageGenerationTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	uuidID, err := uuid.Parse(payload.ID)
	if err != nil {
		return err
	}

	imgGen, err := p.dbClient.ImageGeneration().FindByID(ctx, uuidID)
	if err != nil {
		return err
	}

	response, err := p.imageGenerationService.GetImageGenerationStatus(ctx, imgGen.GenerationID)
	if err != nil {
		return errors.Wrap(err, "error getting image generation status")
	}

	if response.Data.Status == "completed" {
		if err := p.dbClient.ImageGeneration().UpdateState(ctx, imgGen.ID, models.GenerationStatusComplete); err != nil {
			return errors.Wrap(err, "error updating image generation state to completed")
		}

		if err := p.dbClient.ImageGeneration().SetOptions(ctx, imgGen.ID, response.Data.UpscaledURLs); err != nil {
			return errors.Wrap(err, "error setting image generation options")
		}

		if err := p.dbClient.User().UpdateProfilePictureUrl(ctx, imgGen.UserID, response.Data.UpscaledURLs[0]); err != nil {
			return errors.Wrap(err, "error updating user profile picture url")
		}
	} else if response.Data.Status == "failed" {
		if err := p.dbClient.ImageGeneration().UpdateState(ctx, imgGen.ID, models.GenerationStatusFailed); err != nil {
			return errors.Wrap(err, "error updating image generation state to failed")
		}
	} else if imgGen.CreatedAt.Add(imageGenerationTimeout).Before(time.Now()) {
		if err := p.dbClient.ImageGeneration().UpdateState(ctx, imgGen.ID, models.GenerationStatusFailed); err != nil {
			return errors.Wrap(err, "error updating image generation state to failed")
		}
	}

	return nil
}
