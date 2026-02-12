package processors

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

const characterPromptTemplate = "A retro 16-bit RPG pixel art character portrait of %s. %s. Centered, shoulders-up, crisp outlines, limited colors, no text, no logos, clean background."

// GenerateCharacterImageProcessor generates a character dialogue/thumbnail image in the background.
type GenerateCharacterImageProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewGenerateCharacterImageProcessor(dbClient db.DbClient, deepPriestClient deep_priest.DeepPriest, awsClient aws.AWSClient) GenerateCharacterImageProcessor {
	log.Println("Initializing GenerateCharacterImageProcessor")
	return GenerateCharacterImageProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *GenerateCharacterImageProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate character image task: %v", task.Type())

	var payload jobs.GenerateCharacterImageTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	character, err := p.dbClient.Character().FindByID(ctx, payload.CharacterID)
	if err != nil {
		log.Printf("Failed to find character: %v", err)
		return fmt.Errorf("failed to find character: %w", err)
	}
	if character == nil {
		return fmt.Errorf("character not found")
	}

	updateStatus := &models.Character{ImageGenerationStatus: models.CharacterImageGenerationStatusInProgress}
	if err := p.dbClient.Character().Update(ctx, payload.CharacterID, updateStatus); err != nil {
		log.Printf("Failed to update character status: %v", err)
		return fmt.Errorf("failed to update character status: %w", err)
	}

	description := strings.TrimSpace(payload.Description)
	if description == "" {
		description = "A memorable fantasy hero"
	}
	prompt := fmt.Sprintf(characterPromptTemplate, payload.Name, description)
	request := deep_priest.GenerateImageRequest{
		Prompt: prompt,
	}
	deep_priest.ApplyGenerateImageDefaults(&request)
	imageB64, err := p.deepPriestClient.GenerateImage(request)
	if err != nil {
		log.Printf("Failed to generate character image: %v", err)
		return p.markFailed(ctx, payload.CharacterID, err)
	}

	imageBytes, err := decodeCharacterImagePayload(imageB64)
	if err != nil {
		log.Printf("Failed to decode generated image: %v", err)
		return p.markFailed(ctx, payload.CharacterID, err)
	}

	imageURL, err := p.uploadImage(ctx, payload.CharacterID, imageBytes)
	if err != nil {
		log.Printf("Failed to upload generated image: %v", err)
		return p.markFailed(ctx, payload.CharacterID, err)
	}

	clearedErr := ""
	completeUpdate := &models.Character{
		DialogueImageURL:      imageURL,
		ThumbnailURL:          imageURL,
		ImageGenerationStatus: models.CharacterImageGenerationStatusComplete,
		ImageGenerationError:  &clearedErr,
	}
	if err := p.dbClient.Character().Update(ctx, payload.CharacterID, completeUpdate); err != nil {
		log.Printf("Failed to update character with image URL: %v", err)
		return fmt.Errorf("failed to update character: %w", err)
	}

	log.Printf("Character image generated successfully for ID: %s", payload.CharacterID)
	return nil
}

func (p *GenerateCharacterImageProcessor) markFailed(ctx context.Context, characterID uuid.UUID, err error) error {
	errMsg := err.Error()
	update := &models.Character{
		ImageGenerationStatus: models.CharacterImageGenerationStatusFailed,
		ImageGenerationError:  &errMsg,
	}
	if dbErr := p.dbClient.Character().Update(ctx, characterID, update); dbErr != nil {
		log.Printf("Failed to mark character generation failed: %v", dbErr)
	}
	return err
}

func decodeCharacterImagePayload(encoded string) ([]byte, error) {
	trimmed := strings.TrimSpace(encoded)
	if trimmed == "" {
		return nil, fmt.Errorf("empty image payload")
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return downloadCharacterImage(trimmed)
	}

	if strings.HasPrefix(trimmed, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
			for _, entry := range arr {
				entry = strings.TrimSpace(entry)
				if entry == "" {
					continue
				}
				return decodeCharacterImagePayload(entry)
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
			return decodeCharacterImagePayload(payload.Data[0].B64JSON)
		}
	}

	if strings.HasPrefix(trimmed, "data:") {
		if comma := strings.Index(trimmed, ","); comma != -1 {
			trimmed = trimmed[comma+1:]
		}
	}
	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return nil, err
	}
	if len(decoded) == 0 {
		return nil, fmt.Errorf("decoded image was empty")
	}
	return decoded, nil
}

func downloadCharacterImage(url string) ([]byte, error) {
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

func (p *GenerateCharacterImageProcessor) uploadImage(ctx context.Context, characterID uuid.UUID, imageBytes []byte) (string, error) {
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

	imageName := fmt.Sprintf("characters/%s-%d.%s", characterID.String(), time.Now().UnixNano(), imageExtension)
	return p.awsClient.UploadImageToS3("crew-profile-icons", imageName, imageBytes)
}
