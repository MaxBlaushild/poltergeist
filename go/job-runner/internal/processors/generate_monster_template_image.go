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

const monsterTemplateImagePromptTemplate = "A retro 16-bit RPG monster portrait of %s. %s. Base attributes: STR %d, DEX %d, CON %d, INT %d, WIS %d, CHA %d. Aggressive fantasy creature, centered composition, no text, no logos, no frame, readable silhouette, crisp outlines, limited palette."

// GenerateMonsterTemplateImageProcessor generates monster template art in the background.
type GenerateMonsterTemplateImageProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewGenerateMonsterTemplateImageProcessor(dbClient db.DbClient, deepPriestClient deep_priest.DeepPriest, awsClient aws.AWSClient) GenerateMonsterTemplateImageProcessor {
	log.Println("Initializing GenerateMonsterTemplateImageProcessor")
	return GenerateMonsterTemplateImageProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *GenerateMonsterTemplateImageProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate monster template image task: %v", task.Type())

	var payload jobs.GenerateMonsterTemplateImageTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal monster template image task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	template, err := p.dbClient.MonsterTemplate().FindByID(ctx, payload.MonsterTemplateID)
	if err != nil {
		log.Printf("Failed to find monster template: %v", err)
		return fmt.Errorf("failed to find monster template: %w", err)
	}
	if template == nil {
		return fmt.Errorf("monster template not found")
	}

	template.ImageGenerationStatus = models.MonsterTemplateImageGenerationStatusInProgress
	emptyError := ""
	template.ImageGenerationError = &emptyError
	if err := p.dbClient.MonsterTemplate().Update(ctx, payload.MonsterTemplateID, template); err != nil {
		log.Printf("Failed to update monster template status: %v", err)
		return fmt.Errorf("failed to update monster template status: %w", err)
	}

	templateName := strings.TrimSpace(template.Name)
	if templateName == "" {
		templateName = "Unknown Monster"
	}
	description := strings.TrimSpace(template.Description)
	if description == "" {
		description = "A menacing creature encountered in the wild"
	}
	prompt := fmt.Sprintf(
		monsterTemplateImagePromptTemplate,
		templateName,
		description,
		clampMinInt(template.BaseStrength, 1),
		clampMinInt(template.BaseDexterity, 1),
		clampMinInt(template.BaseConstitution, 1),
		clampMinInt(template.BaseIntelligence, 1),
		clampMinInt(template.BaseWisdom, 1),
		clampMinInt(template.BaseCharisma, 1),
	)

	request := deep_priest.GenerateImageRequest{Prompt: prompt}
	deep_priest.ApplyGenerateImageDefaults(&request)

	imagePayload, err := p.deepPriestClient.GenerateImage(request)
	if err != nil {
		log.Printf("Failed to generate monster template image: %v", err)
		return p.markFailed(ctx, payload.MonsterTemplateID, err)
	}

	imageBytes, err := decodeImagePayload(imagePayload)
	if err != nil {
		log.Printf("Failed to decode monster template image payload: %v", err)
		return p.markFailed(ctx, payload.MonsterTemplateID, err)
	}

	imageURL, err := p.uploadImage(ctx, payload.MonsterTemplateID, imageBytes)
	if err != nil {
		log.Printf("Failed to upload monster template image: %v", err)
		return p.markFailed(ctx, payload.MonsterTemplateID, err)
	}

	template.ImageURL = imageURL
	template.ThumbnailURL = imageURL
	template.ImageGenerationStatus = models.MonsterTemplateImageGenerationStatusComplete
	template.ImageGenerationError = &emptyError
	if err := p.dbClient.MonsterTemplate().Update(ctx, payload.MonsterTemplateID, template); err != nil {
		log.Printf("Failed to update monster template image URLs: %v", err)
		return fmt.Errorf("failed to update monster template image urls: %w", err)
	}

	log.Printf("Monster template image generated successfully for ID: %s", payload.MonsterTemplateID)
	return nil
}

func (p *GenerateMonsterTemplateImageProcessor) markFailed(ctx context.Context, templateID uuid.UUID, err error) error {
	template, findErr := p.dbClient.MonsterTemplate().FindByID(ctx, templateID)
	if findErr == nil && template != nil {
		errMsg := err.Error()
		template.ImageGenerationStatus = models.MonsterTemplateImageGenerationStatusFailed
		template.ImageGenerationError = &errMsg
		if updateErr := p.dbClient.MonsterTemplate().Update(ctx, templateID, template); updateErr != nil {
			log.Printf("Failed to mark monster template image generation failed: %v", updateErr)
		}
	}
	return err
}

func (p *GenerateMonsterTemplateImageProcessor) uploadImage(ctx context.Context, templateID uuid.UUID, imageBytes []byte) (string, error) {
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

	imageName := fmt.Sprintf("monster-templates/%s-%d.%s", templateID.String(), time.Now().UnixNano(), imageExtension)
	return p.awsClient.UploadImageToS3("crew-points-of-interest", imageName, imageBytes)
}

func clampMinInt(value int, minimum int) int {
	if value < minimum {
		return minimum
	}
	return value
}
