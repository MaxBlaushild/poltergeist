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

const baseStructureLevelTopDownImagePromptTemplate = `
Create art direction for a fantasy MMORPG base room interior tile viewed from directly overhead.

Room context:
- room name: %s
- room category: %s
- room description: %s
- room effect type: %s
- room level: %d of %d
- progression cue: %s

Rules:
- Match the established base grass tile style:
  - retro 16-bit fantasy RPG pixel art
  - strict orthographic top-down view
  - crisp outlines
  - readable silhouette
  - limited palette
  - tile-friendly composition
- This should be the inside of the room, not the outside of a building.
- Think of the interior of a building in an early JRPG: open-ceiling cutaway room, visible floor, furniture and fixtures seen from directly above, walls around the edges.
- Do not show any roof, exterior facade, surrounding grass, outdoor terrain, or outside environment.
- The camera should be straight overhead, like looking down into the room after the roof has been removed. No angle, no perspective, no three-quarter view, no isometric view.
- The entirety of the image should be the room interior itself. Do not show framing margins, empty background, or any space outside the room.
- The boundaries of the image should align with the room's interior walls or room edges so the room fills the full square tile.
- Compose it like a navigable top-down game interior tile: floor, furnishings, workstations, shrine pieces, hearths, shelves, or equipment arranged within the room.
- The room should look modest at low levels and more elaborate, fortified, or refined at higher levels.
- Fill the square with the room interior. No empty studio background, no framing card treatment.
- No text, no logos, no UI, no modern objects.
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
	view := normalizeBaseStructureImageView(payload.View)

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

	setBaseStructureVisualStateInProgress(visual, view)
	if err := p.dbClient.BaseStructureLevelVisual().Upsert(ctx, visual); err != nil {
		return fmt.Errorf("failed to mark base structure level visual in progress: %w", err)
	}

	promptTemplate := baseStructureLevelImagePromptTemplate
	if view == baseStructureImageViewTopDown {
		promptTemplate = baseStructureLevelTopDownImagePromptTemplate
	}
	prompt := buildBaseStructureLevelPrompt(definition, payload.Level, view, promptTemplate)

	request := deep_priest.GenerateImageRequest{Prompt: prompt}
	deep_priest.ApplyGenerateImageDefaults(&request)

	imagePayload, err := p.deepPriestClient.GenerateImage(request)
	if err != nil {
		return p.markFailed(ctx, visual, view, err)
	}

	imageBytes, err := decodeCharacterImagePayload(imagePayload)
	if err != nil {
		return p.markFailed(ctx, visual, view, err)
	}

	imageURL, err := p.uploadImage(definition.ID, payload.Level, view, imageBytes)
	if err != nil {
		return p.markFailed(ctx, visual, view, err)
	}

	setBaseStructureVisualStateComplete(visual, view, imageURL)
	if err := p.dbClient.BaseStructureLevelVisual().Upsert(ctx, visual); err != nil {
		return fmt.Errorf("failed to update base structure level visual image: %w", err)
	}

	p.enqueueThumbnailTask(visual.ID, imageURL, view)

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
		ID:                           uuid.New(),
		CreatedAt:                    now,
		UpdatedAt:                    now,
		StructureDefinitionID:        definitionID,
		Level:                        level,
		ImageGenerationStatus:        models.BaseStructureImageGenerationStatusQueued,
		TopDownImageGenerationStatus: models.BaseStructureImageGenerationStatusNone,
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
	view string,
	err error,
) error {
	if visual != nil {
		setBaseStructureVisualStateFailed(visual, view, err)
		if updateErr := p.dbClient.BaseStructureLevelVisual().Upsert(ctx, visual); updateErr != nil {
			log.Printf("Failed to mark base structure level image generation failed: %v", updateErr)
		}
	}
	return err
}

func (p *GenerateBaseStructureLevelImageProcessor) uploadImage(
	definitionID uuid.UUID,
	level int,
	view string,
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
	imagePrefix := "base-structures"
	if view == baseStructureImageViewTopDown {
		imagePrefix = "base-structures/top-down"
	}
	imageName := fmt.Sprintf("%s/%s-level-%d-%d.%s", imagePrefix, definitionID.String(), level, time.Now().UnixNano(), imageExtension)
	return p.awsClient.UploadImageToS3("crew-points-of-interest", imageName, imageBytes)
}

func (p *GenerateBaseStructureLevelImageProcessor) enqueueThumbnailTask(
	visualID uuid.UUID,
	imageURL string,
	view string,
) {
	if p.asyncClient == nil || visualID == uuid.Nil || strings.TrimSpace(imageURL) == "" {
		return
	}
	entityType := jobs.ThumbnailEntityBaseStructureLevel
	if view == baseStructureImageViewTopDown {
		entityType = jobs.ThumbnailEntityBaseStructureLevelTopDown
	}
	payloadBytes, err := json.Marshal(jobs.GenerateImageThumbnailTaskPayload{
		EntityType: entityType,
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

const (
	baseStructureImageViewCard    = "card"
	baseStructureImageViewTopDown = "top_down"
)

func normalizeBaseStructureImageView(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case baseStructureImageViewTopDown:
		return baseStructureImageViewTopDown
	default:
		return baseStructureImageViewCard
	}
}

func buildBaseStructureLevelPrompt(
	definition *models.BaseStructureDefinition,
	level int,
	view string,
	fallbackTemplate string,
) string {
	if definition == nil {
		return ""
	}
	if view == baseStructureImageViewTopDown {
		if prompt := strings.TrimSpace(definition.TopDownImagePrompt); prompt != "" {
			return prompt
		}
	} else if prompt := strings.TrimSpace(definition.ImagePrompt); prompt != "" {
		return prompt
	}
	return fmt.Sprintf(
		fallbackTemplate,
		strings.TrimSpace(definition.Name),
		strings.TrimSpace(definition.Category),
		strings.TrimSpace(definition.Description),
		strings.TrimSpace(string(definition.EffectType)),
		level,
		maxInt(definition.MaxLevel, 1),
		baseStructureLevelProgressionCue(level, definition.MaxLevel),
	)
}

func setBaseStructureVisualStateInProgress(visual *models.BaseStructureLevelVisual, view string) {
	if view == baseStructureImageViewTopDown {
		visual.TopDownImageGenerationStatus = models.BaseStructureImageGenerationStatusInProgress
		visual.TopDownImageGenerationError = nil
		return
	}
	visual.ImageGenerationStatus = models.BaseStructureImageGenerationStatusInProgress
	visual.ImageGenerationError = nil
}

func setBaseStructureVisualStateComplete(visual *models.BaseStructureLevelVisual, view string, imageURL string) {
	if view == baseStructureImageViewTopDown {
		visual.TopDownImageURL = imageURL
		visual.TopDownThumbnailURL = ""
		visual.TopDownImageGenerationStatus = models.BaseStructureImageGenerationStatusComplete
		visual.TopDownImageGenerationError = nil
		return
	}
	visual.ImageURL = imageURL
	visual.ThumbnailURL = ""
	visual.ImageGenerationStatus = models.BaseStructureImageGenerationStatusComplete
	visual.ImageGenerationError = nil
}

func setBaseStructureVisualStateFailed(visual *models.BaseStructureLevelVisual, view string, err error) {
	errMsg := err.Error()
	if view == baseStructureImageViewTopDown {
		visual.TopDownImageGenerationStatus = models.BaseStructureImageGenerationStatusFailed
		visual.TopDownImageGenerationError = &errMsg
		return
	}
	visual.ImageGenerationStatus = models.BaseStructureImageGenerationStatusFailed
	visual.ImageGenerationError = &errMsg
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
