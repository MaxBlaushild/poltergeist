package processors

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"cosmossdk.io/errors"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/useapi"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	imageUpscaleTimeout = time.Minute * 5
)

type PollImageUpscaleProcessor struct {
	dbClient     db.DbClient
	useApiClient useapi.Client
	awsClient    aws.AWSClient
}

type PollImageUpscaleTaskPayload struct {
	ID string `json:"id"`
}

func NewPollImageUpscaleProcessor(dbClient db.DbClient, useApiClient useapi.Client, awsClient aws.AWSClient) PollImageUpscaleProcessor {
	return PollImageUpscaleProcessor{
		dbClient:     dbClient,
		useApiClient: useApiClient,
		awsClient:    awsClient,
	}
}

func (p *PollImageUpscaleProcessor) processImage(ctx context.Context, url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.Wrap(err, "error getting image")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Wrap(err, "error getting image: non-200 status code")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "error reading image body")
	}

	key := uuid.New().String() + ".png"

	s3Url, err := p.awsClient.UploadImageToS3("crew-profile-icons", key, body)
	if err != nil {
		return "", errors.Wrap(err, "error uploading image to S3")
	}

	return s3Url, nil
}

func (p *PollImageUpscaleProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload PollImageUpscaleTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	spew.Dump(payload)

	uuidID, err := uuid.Parse(payload.ID)
	if err != nil {
		return err
	}

	imgGen, err := p.dbClient.ImageGeneration().FindByID(ctx, uuidID)
	if err != nil {
		return err
	}

	if imgGen.Status != models.GenerateImageOptions {
		return nil
	}

	if imgGen.CreatedAt.Add(imageUpscaleTimeout).Before(time.Now()) {
		if imgGen.OptionOne == nil && imgGen.OptionTwo == nil && imgGen.OptionThree == nil && imgGen.OptionFour == nil {
			if err := p.dbClient.ImageGeneration().UpdateState(ctx, imgGen.ID, models.GenerationStatusFailed); err != nil {
				return errors.Wrap(err, "error updating image generation state to failed")
			}
		}

		if err := p.dbClient.ImageGeneration().UpdateState(ctx, imgGen.ID, models.GenerationStatusComplete); err != nil {
			return errors.Wrap(err, "error updating image generation state to complete")
		}

		return nil
	}

	if imgGen.OptionOne == nil {
		if err := p.dbClient.ImageGeneration().UpdateState(ctx, imgGen.ID, models.GenerationStatusFailed); err != nil {
			return errors.Wrap(err, "error updating image generation state to failed")
		}

		return nil
	}

	if !strings.HasPrefix(*imgGen.OptionOne, "http://") && !strings.HasPrefix(*imgGen.OptionOne, "https://") {
		upscaleResponse, err := p.useApiClient.CheckUpscaleImageStatus(ctx, *imgGen.OptionOne)
		if err != nil {
			return errors.Wrap(err, "error upscaling image")
		}

		if upscaleResponse.Status == "done" {
			s3Url, err := p.processImage(ctx, upscaleResponse.Result.URL)
			if err != nil {
				return errors.Wrap(err, "error processing image")
			}

			if err := p.dbClient.ImageGeneration().Updates(ctx, imgGen.ID, &models.ImageGeneration{
				OptionOne: &s3Url,
			}); err != nil {
				return errors.Wrap(err, "error updating image generation options")
			}

			if err := p.dbClient.User().UpdateProfilePictureUrl(ctx, imgGen.UserID, upscaleResponse.Result.URL); err != nil {
				return errors.Wrap(err, "error updating user image one")
			}
		} else {
			return nil
		}
	}

	if !strings.HasPrefix(*imgGen.OptionTwo, "http://") && !strings.HasPrefix(*imgGen.OptionTwo, "https://") {
		upscaleResponse, err := p.useApiClient.CheckUpscaleImageStatus(ctx, *imgGen.OptionTwo)
		if err != nil {
			return errors.Wrap(err, "error upscaling image")
		}

		if upscaleResponse.Status == "done" {
			s3Url, err := p.processImage(ctx, upscaleResponse.Result.URL)
			if err != nil {
				return errors.Wrap(err, "error processing image")
			}

			if err := p.dbClient.ImageGeneration().Updates(ctx, imgGen.ID, &models.ImageGeneration{
				OptionTwo: &s3Url,
			}); err != nil {
				return errors.Wrap(err, "error updating image generation options")
			}
		} else {
			return nil
		}
	}

	if !strings.HasPrefix(*imgGen.OptionThree, "http://") && !strings.HasPrefix(*imgGen.OptionThree, "https://") {
		upscaleResponse, err := p.useApiClient.CheckUpscaleImageStatus(ctx, *imgGen.OptionThree)
		if err != nil {
			return errors.Wrap(err, "error upscaling image")
		}

		if upscaleResponse.Status == "done" {
			s3Url, err := p.processImage(ctx, upscaleResponse.Result.URL)
			if err != nil {
				return errors.Wrap(err, "error processing image")
			}

			if err := p.dbClient.ImageGeneration().Updates(ctx, imgGen.ID, &models.ImageGeneration{
				OptionThree: &s3Url,
			}); err != nil {
				return errors.Wrap(err, "error updating image generation options")
			}
		} else {
			return nil
		}
	}

	if !strings.HasPrefix(*imgGen.OptionFour, "http://") && !strings.HasPrefix(*imgGen.OptionFour, "https://") {
		upscaleResponse, err := p.useApiClient.CheckUpscaleImageStatus(ctx, *imgGen.OptionFour)
		if err != nil {
			return errors.Wrap(err, "error upscaling image")
		}

		if upscaleResponse.Status == "done" {
			s3Url, err := p.processImage(ctx, upscaleResponse.Result.URL)
			if err != nil {
				return errors.Wrap(err, "error processing image")
			}

			if err := p.dbClient.ImageGeneration().Updates(ctx, imgGen.ID, &models.ImageGeneration{
				OptionFour: &s3Url,
			}); err != nil {
				return errors.Wrap(err, "error updating image generation options")
			}
		} else {
			return nil
		}
	}

	if err := p.dbClient.ImageGeneration().UpdateState(ctx, imgGen.ID, models.GenerationStatusComplete); err != nil {
		return errors.Wrap(err, "error updating image generation state to completed")
	}

	return nil
}
