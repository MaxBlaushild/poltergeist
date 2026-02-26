package processors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	thumbnailSize = 192
)

// GenerateImageThumbnailProcessor creates smaller thumbnail images for characters and points of interest.
type GenerateImageThumbnailProcessor struct {
	dbClient  db.DbClient
	awsClient aws.AWSClient
}

func NewGenerateImageThumbnailProcessor(dbClient db.DbClient, awsClient aws.AWSClient) GenerateImageThumbnailProcessor {
	log.Println("Initializing GenerateImageThumbnailProcessor")
	return GenerateImageThumbnailProcessor{
		dbClient:  dbClient,
		awsClient: awsClient,
	}
}

func (p *GenerateImageThumbnailProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate image thumbnail task: %v", task.Type())

	var payload jobs.GenerateImageThumbnailTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	entityType := strings.ToLower(strings.TrimSpace(payload.EntityType))
	sourceUrl := strings.TrimSpace(payload.SourceUrl)

	switch entityType {
	case jobs.ThumbnailEntityCharacter:
		if payload.EntityID == uuid.Nil {
			return fmt.Errorf("missing entity ID")
		}
		if sourceUrl == "" {
			character, err := p.dbClient.Character().FindByID(ctx, payload.EntityID)
			if err != nil {
				return fmt.Errorf("failed to find character: %w", err)
			}
			if character == nil {
				return fmt.Errorf("character not found")
			}
			sourceUrl = strings.TrimSpace(character.DialogueImageURL)
		}
	case jobs.ThumbnailEntityPointOfInterest:
		if payload.EntityID == uuid.Nil {
			return fmt.Errorf("missing entity ID")
		}
		if sourceUrl == "" {
			poi, err := p.dbClient.PointOfInterest().FindByID(ctx, payload.EntityID)
			if err != nil {
				return fmt.Errorf("failed to find point of interest: %w", err)
			}
			if poi == nil {
				return fmt.Errorf("point of interest not found")
			}
			sourceUrl = strings.TrimSpace(poi.ImageUrl)
		}
	case jobs.ThumbnailEntityStatic:
		// Use provided source URL only.
	default:
		return fmt.Errorf("unknown entity type: %s", entityType)
	}

	if sourceUrl == "" {
		return fmt.Errorf("missing source image url")
	}

	imageBytes, err := downloadThumbnailSource(sourceUrl)
	if err != nil {
		return fmt.Errorf("failed to download source image: %w", err)
	}

	decoded, err := imaging.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	thumb := imaging.Fill(decoded, thumbnailSize, thumbnailSize, imaging.Center, imaging.Lanczos)
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, thumb, imaging.PNG, imaging.PNGCompressionLevel(png.BestCompression)); err != nil {
		return fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	key := strings.TrimSpace(payload.DestinationKey)
	if key == "" {
		key = thumbnailKey(entityType, payload.EntityID)
	}
	thumbnailUrl, err := p.awsClient.UploadImageToS3(jobs.ThumbnailBucket, key, buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to upload thumbnail: %w", err)
	}

	switch entityType {
	case jobs.ThumbnailEntityCharacter:
		if err := p.dbClient.Character().Update(ctx, payload.EntityID, &models.Character{
			ThumbnailURL: thumbnailUrl,
		}); err != nil {
			return fmt.Errorf("failed to update character thumbnail: %w", err)
		}
	case jobs.ThumbnailEntityPointOfInterest:
		if err := p.dbClient.PointOfInterest().Update(ctx, payload.EntityID, &models.PointOfInterest{
			ThumbnailURL: thumbnailUrl,
		}); err != nil {
			return fmt.Errorf("failed to update point of interest thumbnail: %w", err)
		}
	case jobs.ThumbnailEntityStatic:
		// No-op: static thumbnail is uploaded only.
	}

	log.Printf("Thumbnail generated for %s %s", entityType, payload.EntityID)
	return nil
}

func downloadThumbnailSource(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("downloaded image was empty")
	}
	return body, nil
}

func thumbnailKey(entityType string, entityID uuid.UUID) string {
	prefix := "unknown"
	switch entityType {
	case jobs.ThumbnailEntityCharacter:
		prefix = "characters"
	case jobs.ThumbnailEntityPointOfInterest:
		prefix = "points-of-interest"
	case jobs.ThumbnailEntityStatic:
		prefix = "static"
	}
	return fmt.Sprintf("thumbnails/%s/%s-%d.png", prefix, entityID.String(), time.Now().UnixNano())
}
