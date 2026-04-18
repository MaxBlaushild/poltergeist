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
)

const spellIconPromptTemplate = "A retro 16-bit RPG spell icon for %s. School: %s. %s. %s. Arcane energy motif, magical glyph shapes, no characters, no text, no logos, transparent background, centered composition, crisp outlines, limited palette."

// GenerateSpellIconProcessor generates spell icons in the background.
type GenerateSpellIconProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewGenerateSpellIconProcessor(dbClient db.DbClient, deepPriestClient deep_priest.DeepPriest, awsClient aws.AWSClient) GenerateSpellIconProcessor {
	log.Println("Initializing GenerateSpellIconProcessor")
	return GenerateSpellIconProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *GenerateSpellIconProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate spell icon task: %v", task.Type())

	var payload jobs.GenerateSpellIconTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal spell icon task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	spell, err := p.dbClient.Spell().FindByID(ctx, payload.SpellID)
	if err != nil {
		log.Printf("Failed to find spell: %v", err)
		return fmt.Errorf("failed to find spell: %w", err)
	}
	if spell == nil {
		return fmt.Errorf("spell not found")
	}

	if err := p.dbClient.Spell().Update(ctx, payload.SpellID, map[string]interface{}{
		"image_generation_status": models.SpellImageGenerationStatusInProgress,
	}); err != nil {
		log.Printf("Failed to update spell status: %v", err)
		return fmt.Errorf("failed to update spell status: %w", err)
	}

	school := strings.TrimSpace(spell.SchoolOfMagic)
	if school == "" {
		school = "Arcane"
	}
	description := strings.TrimSpace(spell.Description)
	if description == "" {
		description = "A signature magical technique"
	}
	effectText := strings.TrimSpace(spell.EffectText)
	if effectText == "" {
		effectText = "Mystic energy effect"
	}

	prompt := spellIconPrompt(
		strings.TrimSpace(spell.Name),
		description,
		school,
		effectText,
		spell.AbilityType,
		spell.Genre,
	)
	request := deep_priest.GenerateImageRequest{
		Prompt: prompt,
	}
	deep_priest.ApplyGenerateImageDefaults(&request)

	imagePayload, err := p.deepPriestClient.GenerateImage(request)
	if err != nil {
		log.Printf("Failed to generate spell icon: %v", err)
		return p.markFailed(ctx, payload.SpellID, err)
	}

	imageBytes, err := decodeImagePayload(imagePayload)
	if err != nil {
		log.Printf("Failed to decode generated spell icon: %v", err)
		return p.markFailed(ctx, payload.SpellID, err)
	}

	iconURL, err := p.uploadImage(ctx, payload.SpellID, imageBytes)
	if err != nil {
		log.Printf("Failed to upload generated spell icon: %v", err)
		return p.markFailed(ctx, payload.SpellID, err)
	}

	if err := p.dbClient.Spell().Update(ctx, payload.SpellID, map[string]interface{}{
		"icon_url":                iconURL,
		"image_generation_status": models.SpellImageGenerationStatusComplete,
		"image_generation_error":  "",
	}); err != nil {
		log.Printf("Failed to update spell with icon URL: %v", err)
		return fmt.Errorf("failed to update spell icon url: %w", err)
	}

	log.Printf("Spell icon generated successfully for ID: %s", payload.SpellID)
	return nil
}

func (p *GenerateSpellIconProcessor) markFailed(ctx context.Context, spellID uuid.UUID, err error) error {
	errMsg := err.Error()
	if dbErr := p.dbClient.Spell().Update(ctx, spellID, map[string]interface{}{
		"image_generation_status": models.SpellImageGenerationStatusFailed,
		"image_generation_error":  errMsg,
	}); dbErr != nil {
		log.Printf("Failed to mark spell icon generation failed: %v", dbErr)
	}
	return err
}

func (p *GenerateSpellIconProcessor) uploadImage(ctx context.Context, spellID uuid.UUID, imageBytes []byte) (string, error) {
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

	imageName := fmt.Sprintf("spells/%s-%d.%s", spellID.String(), time.Now().UnixNano(), imageExtension)
	return p.awsClient.UploadImageToS3("crew-points-of-interest", imageName, imageBytes)
}
