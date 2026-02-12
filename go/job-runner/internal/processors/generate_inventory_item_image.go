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

	"github.com/hibiken/asynq"
)

const inventoryItemPromptTemplate = "A retro 16-bit RPG pixel art item icon of %s. %s. Rarity: %s. Centered, crisp outlines, limited colors, no text, no logos, transparent background."

// GenerateInventoryItemImageProcessor generates an inventory item image in the background.
type GenerateInventoryItemImageProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewGenerateInventoryItemImageProcessor(dbClient db.DbClient, deepPriestClient deep_priest.DeepPriest, awsClient aws.AWSClient) GenerateInventoryItemImageProcessor {
	log.Println("Initializing GenerateInventoryItemImageProcessor")
	return GenerateInventoryItemImageProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *GenerateInventoryItemImageProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate inventory item image task: %v", task.Type())

	var payload jobs.GenerateInventoryItemImageTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	item, err := p.dbClient.InventoryItem().FindInventoryItemByID(ctx, payload.InventoryItemID)
	if err != nil {
		log.Printf("Failed to find inventory item: %v", err)
		return fmt.Errorf("failed to find inventory item: %w", err)
	}
	if item == nil {
		return fmt.Errorf("inventory item not found")
	}

	updateStatus := &models.InventoryItem{ImageGenerationStatus: models.InventoryImageGenerationStatusInProgress}
	if err := p.dbClient.InventoryItem().UpdateInventoryItem(ctx, payload.InventoryItemID, updateStatus); err != nil {
		log.Printf("Failed to update inventory item status: %v", err)
		return fmt.Errorf("failed to update inventory item status: %w", err)
	}

	description := strings.TrimSpace(payload.Description)
	if description == "" {
		description = "A unique fantasy item"
	}
	prompt := fmt.Sprintf(inventoryItemPromptTemplate, payload.Name, description, payload.RarityTier)
	request := deep_priest.GenerateImageRequest{
		Prompt: prompt,
	}
	deep_priest.ApplyGenerateImageDefaults(&request)
	imageB64, err := p.deepPriestClient.GenerateImage(request)
	if err != nil {
		log.Printf("Failed to generate inventory item image: %v", err)
		return p.markFailed(ctx, payload.InventoryItemID, err)
	}

	imageBytes, err := decodeImagePayload(imageB64)
	if err != nil {
		log.Printf("Failed to decode generated image: %v", err)
		return p.markFailed(ctx, payload.InventoryItemID, err)
	}

	imageURL, err := p.uploadImage(ctx, payload.InventoryItemID, imageBytes)
	if err != nil {
		log.Printf("Failed to upload generated image: %v", err)
		return p.markFailed(ctx, payload.InventoryItemID, err)
	}

	clearedErr := ""
	completeUpdate := &models.InventoryItem{
		ImageURL:              imageURL,
		ImageGenerationStatus: models.InventoryImageGenerationStatusComplete,
		ImageGenerationError:  &clearedErr,
	}
	if err := p.dbClient.InventoryItem().UpdateInventoryItem(ctx, payload.InventoryItemID, completeUpdate); err != nil {
		log.Printf("Failed to update inventory item with image URL: %v", err)
		return fmt.Errorf("failed to update inventory item: %w", err)
	}

	log.Printf("Inventory item image generated successfully for ID: %d", payload.InventoryItemID)
	return nil
}

func (p *GenerateInventoryItemImageProcessor) markFailed(ctx context.Context, itemID int, err error) error {
	errMsg := err.Error()
	update := &models.InventoryItem{
		ImageGenerationStatus: models.InventoryImageGenerationStatusFailed,
		ImageGenerationError:  &errMsg,
	}
	if dbErr := p.dbClient.InventoryItem().UpdateInventoryItem(ctx, itemID, update); dbErr != nil {
		log.Printf("Failed to mark inventory item generation failed: %v", dbErr)
	}
	return err
}

func decodeImagePayload(encoded string) ([]byte, error) {
	trimmed := strings.TrimSpace(encoded)
	if trimmed == "" {
		return nil, fmt.Errorf("empty image payload")
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return downloadImage(trimmed)
	}

	// Handle JSON array of base64 strings (some backends return this)
	if strings.HasPrefix(trimmed, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
			for _, entry := range arr {
				entry = strings.TrimSpace(entry)
				if entry == "" {
					continue
				}
				return decodeImagePayload(entry)
			}
			return nil, fmt.Errorf("image payload array contained no data")
		}
	}

	// Handle JSON object with data[0].b64_json shape
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
			return decodeImagePayload(payload.Data[0].B64JSON)
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

func downloadImage(url string) ([]byte, error) {
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

func (p *GenerateInventoryItemImageProcessor) uploadImage(ctx context.Context, itemID int, imageBytes []byte) (string, error) {
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

	imageName := fmt.Sprintf("inventory-items/%d-%d.%s", itemID, time.Now().UnixNano(), imageExtension)
	return p.awsClient.UploadImageToS3("crew-points-of-interest", imageName, imageBytes)
}
