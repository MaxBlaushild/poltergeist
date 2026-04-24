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
	"github.com/hibiken/asynq"
)

type GenerateZoneShroudPatternTileProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewGenerateZoneShroudPatternTileProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
	awsClient aws.AWSClient,
) GenerateZoneShroudPatternTileProcessor {
	log.Println("Initializing GenerateZoneShroudPatternTileProcessor")
	return GenerateZoneShroudPatternTileProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *GenerateZoneShroudPatternTileProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate zone shroud pattern tile task: %v", task.Type())

	var payload jobs.GenerateZoneShroudPatternTileTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	config, err := p.dbClient.ZoneShroudConfig().Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to load zone shroud config: %w", err)
	}

	prompt := strings.TrimSpace(payload.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(config.PatternTilePrompt)
	}
	if prompt == "" {
		prompt = "Create a seamless repeating square fog-of-war texture tile for a fantasy RPG world map. Keep it atmospheric, readable, tileable, and free of text or borders."
	}

	config.PatternTilePrompt = prompt
	config.PatternTileGenerationStatus = models.ZoneKindPatternTileGenerationStatusInProgress
	config.PatternTileGenerationError = ""
	if _, err := p.dbClient.ZoneShroudConfig().Upsert(ctx, config); err != nil {
		return fmt.Errorf("failed to mark zone shroud tile generation in progress: %w", err)
	}

	request := deep_priest.GenerateImageRequest{Prompt: prompt}
	deep_priest.ApplyGenerateImageDefaults(&request)

	imagePayload, err := p.deepPriestClient.GenerateImage(request)
	if err != nil {
		return p.markFailed(ctx, config, err)
	}

	imageBytes, err := decodeImagePayload(imagePayload)
	if err != nil {
		return p.markFailed(ctx, config, err)
	}

	tileBytes, err := prepareZoneKindPatternTile(imageBytes)
	if err != nil {
		return p.markFailed(ctx, config, err)
	}

	imageURL, err := p.uploadImage(tileBytes)
	if err != nil {
		return p.markFailed(ctx, config, err)
	}

	config.PatternTileURL = imageURL
	config.PatternTilePrompt = prompt
	config.PatternTileGenerationStatus = models.ZoneKindPatternTileGenerationStatusComplete
	config.PatternTileGenerationError = ""
	if _, err := p.dbClient.ZoneShroudConfig().Upsert(ctx, config); err != nil {
		return fmt.Errorf("failed to update zone shroud tile url: %w", err)
	}

	return nil
}

func (p *GenerateZoneShroudPatternTileProcessor) markFailed(
	ctx context.Context,
	config *models.ZoneShroudConfig,
	err error,
) error {
	if config != nil {
		config.PatternTileGenerationStatus = models.ZoneKindPatternTileGenerationStatusFailed
		config.PatternTileGenerationError = strings.TrimSpace(err.Error())
		if _, updateErr := p.dbClient.ZoneShroudConfig().Upsert(ctx, config); updateErr != nil {
			log.Printf("Failed to mark zone shroud pattern tile generation failed: %v", updateErr)
		}
	}
	return err
}

func (p *GenerateZoneShroudPatternTileProcessor) uploadImage(imageBytes []byte) (string, error) {
	key := fmt.Sprintf("zone-shroud-patterns/shroud-%d.png", time.Now().UnixNano())
	return p.awsClient.UploadImageToS3(zoneKindPatternTileBucket, key, imageBytes)
}
