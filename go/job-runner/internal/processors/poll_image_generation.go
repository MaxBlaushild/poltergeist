package processors

import (
	"context"
	"encoding/json"
	"time"

	"cosmossdk.io/errors"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/useapi"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type PollImageGenerationProcessor struct {
	dbClient     db.DbClient
	useApiClient useapi.Client
}

type PollImageGenerationTaskPayload struct {
	ID string `json:"id"`
}

func NewPollImageGenerationProcessor(dbClient db.DbClient, useApiClient useapi.Client) PollImageGenerationProcessor {
	return PollImageGenerationProcessor{
		dbClient:     dbClient,
		useApiClient: useApiClient,
	}
}

func (p *PollImageGenerationProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload PollImageGenerationTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	spew.Dump("payload")
	spew.Dump(payload)

	uuidID, err := uuid.Parse(payload.ID)
	if err != nil {
		return err
	}

	imgGen, err := p.dbClient.ImageGeneration().FindByID(ctx, uuidID)
	if err != nil {
		return err
	}

	if imgGen.Status != models.GenerationStatusPending {
		return nil
	}

	response, err := p.useApiClient.CheckImageGenerationOptionsStatus(ctx, imgGen.GenerationID)
	if err != nil {
		return errors.Wrap(err, "error getting image generation status")
	}

	if response.Status == "done" {
		if imgGen.OptionOne == nil {
			upscaleResponse, err := p.useApiClient.UpscaleImage(ctx, imgGen.GenerationID, 1)
			if err != nil {
				spew.Dump("upscale err")
				spew.Dump(err)
				return errors.Wrap(err, "error upscaling image")
			}

			if err := p.dbClient.ImageGeneration().Updates(ctx, imgGen.ID, &models.ImageGeneration{
				OptionOne: &upscaleResponse.Hash,
			}); err != nil {
				return errors.Wrap(err, "error updating image generation options")
			}
		}

		if imgGen.OptionTwo == nil {
			upscaleResponse, err := p.useApiClient.UpscaleImage(ctx, imgGen.GenerationID, 2)
			if err != nil {
				spew.Dump("upscale err")
				spew.Dump(err)
				return errors.Wrap(err, "error upscaling image")
			}

			if err := p.dbClient.ImageGeneration().Updates(ctx, imgGen.ID, &models.ImageGeneration{
				OptionTwo: &upscaleResponse.Hash,
			}); err != nil {
				return errors.Wrap(err, "error updating image generation options")
			}
		}

		if imgGen.OptionThree == nil {
			upscaleResponse, err := p.useApiClient.UpscaleImage(ctx, imgGen.GenerationID, 3)
			if err != nil {
				return errors.Wrap(err, "error upscaling image")
			}

			if err := p.dbClient.ImageGeneration().Updates(ctx, imgGen.ID, &models.ImageGeneration{
				OptionThree: &upscaleResponse.Hash,
			}); err != nil {
				return errors.Wrap(err, "error updating image generation options")
			}
		}

		if imgGen.OptionFour == nil {
			upscaleResponse, err := p.useApiClient.UpscaleImage(ctx, imgGen.GenerationID, 4)
			if err != nil {
				return errors.Wrap(err, "error upscaling image")
			}

			if err := p.dbClient.ImageGeneration().Updates(ctx, imgGen.ID, &models.ImageGeneration{
				OptionFour: &upscaleResponse.Hash,
			}); err != nil {
				return errors.Wrap(err, "error updating image generation options")
			}
		}

		if err := p.dbClient.ImageGeneration().UpdateState(ctx, imgGen.ID, models.GenerateImageOptions); err != nil {
			return errors.Wrap(err, "error updating image generation state to completed")
		}
	} else if response.Status == "failed" {
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
