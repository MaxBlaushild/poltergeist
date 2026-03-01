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

const monsterImagePromptTemplate = "A retro 16-bit RPG monster portrait of %s. %s. Zone context: %s. Aggressive fantasy creature, centered composition, no text, no logos, no frame, readable silhouette, crisp outlines, limited palette."

// GenerateMonsterImageProcessor generates monster art in the background.
type GenerateMonsterImageProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewGenerateMonsterImageProcessor(dbClient db.DbClient, deepPriestClient deep_priest.DeepPriest, awsClient aws.AWSClient) GenerateMonsterImageProcessor {
	log.Println("Initializing GenerateMonsterImageProcessor")
	return GenerateMonsterImageProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *GenerateMonsterImageProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate monster image task: %v", task.Type())

	var payload jobs.GenerateMonsterImageTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal monster image task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	monster, err := p.dbClient.Monster().FindByID(ctx, payload.MonsterID)
	if err != nil {
		log.Printf("Failed to find monster: %v", err)
		return fmt.Errorf("failed to find monster: %w", err)
	}
	if monster == nil {
		return fmt.Errorf("monster not found")
	}

	monster.ImageGenerationStatus = models.MonsterImageGenerationStatusInProgress
	emptyError := ""
	monster.ImageGenerationError = &emptyError
	if err := p.dbClient.Monster().Update(ctx, payload.MonsterID, monster); err != nil {
		log.Printf("Failed to update monster status: %v", err)
		return fmt.Errorf("failed to update monster status: %w", err)
	}

	description := strings.TrimSpace(monster.Description)
	if description == "" {
		if monster.Template != nil {
			description = strings.TrimSpace(monster.Template.Description)
		}
		if description == "" {
			description = "A menacing creature encountered in the wild"
		}
	}
	zoneName := strings.TrimSpace(monster.Zone.Name)
	if zoneName == "" {
		zoneName = "Unknown Zone"
	}
	monsterName := strings.TrimSpace(monster.Name)
	if monsterName == "" && monster.Template != nil {
		monsterName = strings.TrimSpace(monster.Template.Name)
	}
	if monsterName == "" {
		monsterName = "Unknown Monster"
	}

	prompt := fmt.Sprintf(monsterImagePromptTemplate, monsterName, description, zoneName)
	request := deep_priest.GenerateImageRequest{Prompt: prompt}
	deep_priest.ApplyGenerateImageDefaults(&request)

	imagePayload, err := p.deepPriestClient.GenerateImage(request)
	if err != nil {
		log.Printf("Failed to generate monster image: %v", err)
		return p.markFailed(ctx, payload.MonsterID, err)
	}

	imageBytes, err := decodeImagePayload(imagePayload)
	if err != nil {
		log.Printf("Failed to decode monster image payload: %v", err)
		return p.markFailed(ctx, payload.MonsterID, err)
	}

	imageURL, err := p.uploadImage(ctx, payload.MonsterID, imageBytes)
	if err != nil {
		log.Printf("Failed to upload monster image: %v", err)
		return p.markFailed(ctx, payload.MonsterID, err)
	}

	monster.ImageURL = imageURL
	monster.ThumbnailURL = imageURL
	monster.ImageGenerationStatus = models.MonsterImageGenerationStatusComplete
	monster.ImageGenerationError = &emptyError
	if err := p.dbClient.Monster().Update(ctx, payload.MonsterID, monster); err != nil {
		log.Printf("Failed to update monster image URLs: %v", err)
		return fmt.Errorf("failed to update monster image urls: %w", err)
	}

	log.Printf("Monster image generated successfully for ID: %s", payload.MonsterID)
	return nil
}

func (p *GenerateMonsterImageProcessor) markFailed(ctx context.Context, monsterID uuid.UUID, err error) error {
	monster, findErr := p.dbClient.Monster().FindByID(ctx, monsterID)
	if findErr == nil && monster != nil {
		errMsg := err.Error()
		monster.ImageGenerationStatus = models.MonsterImageGenerationStatusFailed
		monster.ImageGenerationError = &errMsg
		if updateErr := p.dbClient.Monster().Update(ctx, monsterID, monster); updateErr != nil {
			log.Printf("Failed to mark monster image generation failed: %v", updateErr)
		}
	}
	return err
}

func (p *GenerateMonsterImageProcessor) uploadImage(ctx context.Context, monsterID uuid.UUID, imageBytes []byte) (string, error) {
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

	imageName := fmt.Sprintf("monsters/%s-%d.%s", monsterID.String(), time.Now().UnixNano(), imageExtension)
	return p.awsClient.UploadImageToS3("crew-points-of-interest", imageName, imageBytes)
}
