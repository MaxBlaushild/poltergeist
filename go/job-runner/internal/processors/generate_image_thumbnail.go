package processors

import (
	"bytes"
	"context"
	"encoding/base64"
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

// GenerateImageThumbnailProcessor creates smaller thumbnail images for characters, points of interest, bases, and base room art.
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
	case jobs.ThumbnailEntityBase:
		if payload.EntityID == uuid.Nil {
			return fmt.Errorf("missing entity ID")
		}
		if sourceUrl == "" {
			base, err := p.dbClient.Base().FindByID(ctx, payload.EntityID)
			if err != nil {
				return fmt.Errorf("failed to find base: %w", err)
			}
			if base == nil {
				return fmt.Errorf("base not found")
			}
			sourceUrl = strings.TrimSpace(base.ImageURL)
		}
	case jobs.ThumbnailEntityBaseStructureLevel:
		if payload.EntityID == uuid.Nil {
			return fmt.Errorf("missing entity ID")
		}
		if sourceUrl == "" {
			visual, err := p.dbClient.BaseStructureLevelVisual().FindByID(ctx, payload.EntityID)
			if err != nil {
				return fmt.Errorf("failed to find base structure level visual: %w", err)
			}
			if visual == nil {
				return fmt.Errorf("base structure level visual not found")
			}
			sourceUrl = strings.TrimSpace(visual.ImageURL)
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
	case jobs.ThumbnailEntityBase:
		if err := p.dbClient.Base().UpdateThumbnailURL(ctx, payload.EntityID, thumbnailUrl); err != nil {
			return fmt.Errorf("failed to update base thumbnail: %w", err)
		}
	case jobs.ThumbnailEntityBaseStructureLevel:
		if err := p.dbClient.BaseStructureLevelVisual().UpdateThumbnailURL(ctx, payload.EntityID, thumbnailUrl); err != nil {
			return fmt.Errorf("failed to update base structure level thumbnail: %w", err)
		}
	case jobs.ThumbnailEntityStatic:
		// No-op: static thumbnail is uploaded only.
	}

	log.Printf("Thumbnail generated for %s %s", entityType, payload.EntityID)
	return nil
}

func downloadThumbnailSource(source string) ([]byte, error) {
	trimmed := strings.TrimSpace(source)
	if trimmed == "" {
		return nil, fmt.Errorf("missing source image url")
	}

	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return downloadThumbnailSourceURL(trimmed)
	}

	if strings.HasPrefix(trimmed, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
			for _, entry := range arr {
				entry = strings.TrimSpace(entry)
				if entry == "" {
					continue
				}
				return downloadThumbnailSource(entry)
			}
			return nil, fmt.Errorf("image payload array contained no data")
		}
	}

	if strings.HasPrefix(trimmed, "{") {
		var payload struct {
			Data []struct {
				B64JSON string `json:"b64_json"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(trimmed), &payload); err == nil {
			if len(payload.Data) == 0 || strings.TrimSpace(payload.Data[0].B64JSON) == "" {
				return nil, fmt.Errorf("image payload object contained no data")
			}
			return downloadThumbnailSource(payload.Data[0].B64JSON)
		}
	}

	if strings.HasPrefix(trimmed, "data:") {
		if comma := strings.Index(trimmed, ","); comma != -1 {
			trimmed = trimmed[comma+1:]
		}
	}

	return decodeBase64ThumbnailSource(trimmed)
}

func downloadThumbnailSourceURL(url string) ([]byte, error) {
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

func decodeBase64ThumbnailSource(raw string) ([]byte, error) {
	for _, encoding := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		decoded, err := encoding.DecodeString(raw)
		if err != nil {
			continue
		}
		if len(decoded) == 0 {
			return nil, fmt.Errorf("decoded image was empty")
		}
		return decoded, nil
	}
	return nil, fmt.Errorf("failed to decode image payload as base64")
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
