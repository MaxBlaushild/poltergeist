package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
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

const outfitPromptTemplate = `
	Create an image in the style of retro 16-bit RPG pixel art.
	Use bold outlines, limited flat colors, and minimal dithering.
	Apply 2–3 shading tones per area, with crisp, blocky edges.
	Keep the result clean, simple, and non-photorealistic.
	Avoid gradients, text, and logos.
	Render the person from the reference selfie as a shoulders-up character portrait wearing %s.
	Clean, light background.`

type GenerateOutfitProfilePictureProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewGenerateOutfitProfilePictureProcessor(dbClient db.DbClient, deepPriestClient deep_priest.DeepPriest, awsClient aws.AWSClient) GenerateOutfitProfilePictureProcessor {
	log.Println("Initializing GenerateOutfitProfilePictureProcessor")
	return GenerateOutfitProfilePictureProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *GenerateOutfitProfilePictureProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate outfit profile picture task: %v", task.Type())

	var payload jobs.GenerateOutfitProfilePictureTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	gen, err := p.dbClient.OutfitProfileGeneration().FindByID(ctx, payload.GenerationID)
	if err != nil {
		log.Printf("Failed to find outfit generation: %v", err)
		return fmt.Errorf("failed to find outfit generation: %w", err)
	}

	updateInProgress := &models.OutfitProfileGeneration{
		Status: models.OutfitGenerationStatusInProgress,
	}
	if err := p.dbClient.OutfitProfileGeneration().Update(ctx, gen.ID, updateInProgress); err != nil {
		log.Printf("Failed to update outfit generation status: %v", err)
		return fmt.Errorf("failed to update outfit generation status: %w", err)
	}

	outfitName := strings.TrimSpace(payload.OutfitName)
	if outfitName == "" {
		outfitName = "a fantasy adventurer outfit"
	}

	editRequest := deep_priest.EditImageRequest{
		Prompt:   fmt.Sprintf(outfitPromptTemplate, outfitName),
		ImageUrl: payload.SelfieUrl,
		Model:    "gpt-image-1",
		N:        1,
		Size:     genSize,
	}
	deep_priest.ApplyEditImageDefaults(&editRequest)
	resp, err := p.deepPriestClient.EditImage(editRequest)
	if err != nil {
		log.Printf("Failed to generate outfit profile picture: %v", err)
		return p.markOutfitFailed(ctx, gen.ID, err)
	}

	candidates, err := decodeBase64Candidates(resp)
	if err != nil {
		log.Printf("Failed to decode outfit candidates: %v", err)
		return p.markOutfitFailed(ctx, gen.ID, err)
	}
	if len(candidates) == 0 {
		return p.markOutfitFailed(ctx, gen.ID, fmt.Errorf("no image candidates returned"))
	}

	pp, _, err := EnforcePixelLook(candidates[0], iconSize, quantColors, upscaleOutput, true)
	if err != nil {
		log.Printf("Failed post-process outfit image: %v", err)
		return p.markOutfitFailed(ctx, gen.ID, err)
	}

	imageURL, err := p.uploadOutfitImage(ctx, payload.UserID.String(), pp)
	if err != nil {
		log.Printf("Failed to upload outfit image: %v", err)
		return p.markOutfitFailed(ctx, gen.ID, err)
	}

	if err := p.dbClient.User().UpdateProfilePictureUrl(ctx, payload.UserID, imageURL); err != nil {
		log.Printf("Failed to update user profile picture: %v", err)
		return p.markOutfitFailed(ctx, gen.ID, err)
	}

	clearedErr := ""
	completeUpdate := &models.OutfitProfileGeneration{
		Status:            models.OutfitGenerationStatusComplete,
		ErrorMessage:      &clearedErr,
		ProfilePictureUrl: &imageURL,
	}
	if err := p.dbClient.OutfitProfileGeneration().Update(ctx, gen.ID, completeUpdate); err != nil {
		log.Printf("Failed to update outfit generation complete: %v", err)
		return fmt.Errorf("failed to update outfit generation: %w", err)
	}

	if err := p.dbClient.InventoryItem().UseInventoryItem(ctx, payload.OwnedInventoryItemID); err != nil {
		log.Printf("Failed to consume outfit item: %v", err)
		return p.markOutfitFailed(ctx, gen.ID, err)
	}

	log.Printf("Outfit profile picture generated successfully for user %s", payload.UserID)
	return nil
}

func (p *GenerateOutfitProfilePictureProcessor) markOutfitFailed(ctx context.Context, genID uuid.UUID, err error) error {
	errMsg := err.Error()
	update := &models.OutfitProfileGeneration{
		Status:       models.OutfitGenerationStatusFailed,
		ErrorMessage: &errMsg,
	}
	if dbErr := p.dbClient.OutfitProfileGeneration().Update(ctx, genID, update); dbErr != nil {
		log.Printf("Failed to mark outfit generation failed: %v", dbErr)
	}
	return err
}

func (p *GenerateOutfitProfilePictureProcessor) uploadOutfitImage(ctx context.Context, userID string, imageBytes []byte) (string, error) {
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

	timestamp := strconv.FormatInt(time.Now().UnixNano(), 16)
	imageName := timestamp + "-" + userID + "." + imageExtension

	imageUrl, err := p.awsClient.UploadImageToS3("crew-profile-icons", imageName, imageBytes)
	if err != nil {
		return "", err
	}

	return imageUrl, nil
}
