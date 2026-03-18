package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

const baseStructureLevelImagePromptTemplate = `
Create art direction for a fantasy MMORPG base room.

Room context:
- room name: %s
- room category: %s
- room description: %s
- room effect type: %s
- room level: %d of %d
- progression cue: %s

Rules:
- Match the established discovered point-of-interest / base image style:
  - retro 16-bit fantasy RPG pixel art
  - crisp outlines
  - readable silhouette
  - limited palette
  - polished but game-ready
- Show a room, hall, workshop, chamber, study, shrine, or yard that clearly belongs inside a growing adventurer base.
- The room should look modest at low levels and more elaborate, fortified, or refined at higher levels.
- Keep it specific to the room's role and description instead of generic housing art.
- No text, no logos, no UI, no modern objects.
- Prefer a clean background and a centered composition suitable for a management screen card.
`

type GenerateBaseStructureLevelImageProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
	asyncClient      *asynq.Client
}

func NewGenerateBaseStructureLevelImageProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
	awsClient aws.AWSClient,
	asyncClient *asynq.Client,
) GenerateBaseStructureLevelImageProcessor {
	log.Println("Initializing GenerateBaseStructureLevelImageProcessor")
	return GenerateBaseStructureLevelImageProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
		asyncClient:      asyncClient,
	}
}

func (p *GenerateBaseStructureLevelImageProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate base structure level image task: %v", task.Type())

	var payload jobs.GenerateBaseStructureLevelImageTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	if payload.StructureDefinitionID == uuid.Nil {
		return fmt.Errorf("structureDefinitionId is required")
	}
	if payload.Level <= 0 {
		return fmt.Errorf("level must be positive")
	}

	definition, err := p.dbClient.BaseStructureDefinition().FindByID(ctx, payload.StructureDefinitionID)
	if err != nil {
		return fmt.Errorf("failed to load base structure definition: %w", err)
	}
	if definition == nil {
		return fmt.Errorf("base structure definition not found")
	}

	visual, err := p.findOrCreateVisual(ctx, definition.ID, payload.Level)
	if err != nil {
		return err
	}

	visual.ImageGenerationStatus = models.BaseStructureImageGenerationStatusInProgress
	visual.ImageGenerationError = nil
	if err := p.dbClient.BaseStructureLevelVisual().Upsert(ctx, visual); err != nil {
		return fmt.Errorf("failed to mark base structure level visual in progress: %w", err)
	}

	prompt := fmt.Sprintf(
		baseStructureLevelImagePromptTemplate,
		strings.TrimSpace(definition.Name),
		strings.TrimSpace(definition.Category),
		strings.TrimSpace(definition.Description),
		strings.TrimSpace(string(definition.EffectType)),
		payload.Level,
		maxInt(definition.MaxLevel, 1),
		baseStructureLevelProgressionCue(payload.Level, definition.MaxLevel),
	)

	request := deep_priest.GenerateImageRequest{Prompt: prompt}
	deep_priest.ApplyGenerateImageDefaults(&request)

	imagePayload, err := p.deepPriestClient.GenerateImage(request)
	if err != nil {
		return p.markFailed(ctx, visual, err)
	}

	imageBytes, err := decodeCharacterImagePayload(imagePayload)
	if err != nil {
		return p.markFailed(ctx, visual, err)
	}

	imageURL, err := p.uploadImage(definition.ID, payload.Level, imageBytes)
	if err != nil {
		return p.markFailed(ctx, visual, err)
	}

	visual.ImageURL = imageURL
	visual.ThumbnailURL = ""
	visual.ImageGenerationStatus = models.BaseStructureImageGenerationStatusComplete
	visual.ImageGenerationError = nil
	if err := p.dbClient.BaseStructureLevelVisual().Upsert(ctx, visual); err != nil {
		return fmt.Errorf("failed to update base structure level visual image: %w", err)
	}

	p.enqueueThumbnailTask(visual.ID, imageURL)

	return nil
}

func (p *GenerateBaseStructureLevelImageProcessor) findOrCreateVisual(
	ctx context.Context,
	definitionID uuid.UUID,
	level int,
) (*models.BaseStructureLevelVisual, error) {
	visual, err := p.dbClient.BaseStructureLevelVisual().FindByDefinitionIDAndLevel(ctx, definitionID, level)
	if err == nil && visual != nil {
		return visual, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to load base structure level visual: %w", err)
	}
	now := time.Now()
	visual = &models.BaseStructureLevelVisual{
		ID:                    uuid.New(),
		CreatedAt:             now,
		UpdatedAt:             now,
		StructureDefinitionID: definitionID,
		Level:                 level,
		ImageGenerationStatus: models.BaseStructureImageGenerationStatusQueued,
	}
	if err := p.dbClient.BaseStructureLevelVisual().Upsert(ctx, visual); err != nil {
		return nil, fmt.Errorf("failed to create base structure level visual: %w", err)
	}
	created, err := p.dbClient.BaseStructureLevelVisual().FindByDefinitionIDAndLevel(ctx, definitionID, level)
	if err != nil {
		return nil, fmt.Errorf("failed to reload base structure level visual: %w", err)
	}
	return created, nil
}

func (p *GenerateBaseStructureLevelImageProcessor) markFailed(
	ctx context.Context,
	visual *models.BaseStructureLevelVisual,
	err error,
) error {
	if visual != nil {
		errMsg := err.Error()
		visual.ImageGenerationStatus = models.BaseStructureImageGenerationStatusFailed
		visual.ImageGenerationError = &errMsg
		if updateErr := p.dbClient.BaseStructureLevelVisual().Upsert(ctx, visual); updateErr != nil {
			log.Printf("Failed to mark base structure level image generation failed: %v", updateErr)
		}
	}
	return err
}

func (p *GenerateBaseStructureLevelImageProcessor) uploadImage(
	definitionID uuid.UUID,
	level int,
	imageBytes []byte,
) (string, error) {
	if len(imageBytes) == 0 {
		return "", fmt.Errorf("no image data provided")
	}
	imageFormat, err := util.DetectImageFormat(imageBytes)
	if err != nil {
		return "", err
	}
	imageExtension, err := util.GetImageExtension(imageFormat)
	if err != nil {
		return "", err
	}
	imageName := fmt.Sprintf(
		"base-structures/%s-level-%d-%d.%s",
		definitionID.String(),
		level,
		time.Now().UnixNano(),
		imageExtension,
	)
	return p.awsClient.UploadImageToS3("crew-points-of-interest", imageName, imageBytes)
}

func (p *GenerateBaseStructureLevelImageProcessor) enqueueThumbnailTask(
	visualID uuid.UUID,
	imageURL string,
) {
	if p.asyncClient == nil || visualID == uuid.Nil || strings.TrimSpace(imageURL) == "" {
		return
	}
	payloadBytes, err := json.Marshal(jobs.GenerateImageThumbnailTaskPayload{
		EntityType: jobs.ThumbnailEntityBaseStructureLevel,
		EntityID:   visualID,
		SourceUrl:  imageURL,
	})
	if err != nil {
		log.Printf("Failed to marshal base structure level thumbnail task payload: %v", err)
		return
	}
	if _, err := p.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateImageThumbnailTaskType, payloadBytes)); err != nil {
		log.Printf("Failed to enqueue base structure level thumbnail task: %v", err)
	}
}

func baseStructureLevelProgressionCue(level int, maxLevel int) string {
	switch {
	case maxLevel <= 1 || level >= maxLevel:
		return "fully realized, prestigious, and battle-ready"
	case level <= 1:
		return "newly built, practical, and humble"
	case level == maxLevel-1:
		return "well-developed, sturdy, and close to masterwork"
	default:
		return "expanded, capable, and clearly improving"
	}
}
