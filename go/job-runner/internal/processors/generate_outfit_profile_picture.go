package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
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
	Render %s of the person described below wearing %s.
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

	selfieUrl := p.resolveSelfieURL(payload.SelfieUrl)
	appearanceDescriptor := p.inferAppearanceDescriptor(ctx, selfieUrl)
	frontPrompt := buildOutfitPrompt(outfitName, appearanceDescriptor, false)
	frontImageBytes, err := p.generateOutfitPortrait(selfieUrl, frontPrompt)
	if err != nil {
		return p.markOutfitFailed(ctx, gen.ID, err)
	}

	backPrompt := buildOutfitPrompt(outfitName, appearanceDescriptor, true)
	backImageBytes, err := p.generateOutfitPortrait(selfieUrl, backPrompt)
	if err != nil {
		return p.markOutfitFailed(ctx, gen.ID, err)
	}

	frontImageURL, err := p.uploadOutfitImage(ctx, payload.UserID.String(), frontImageBytes)
	if err != nil {
		log.Printf("Failed to upload front-facing outfit image: %v", err)
		return p.markOutfitFailed(ctx, gen.ID, err)
	}

	backImageURL, err := p.uploadOutfitImage(ctx, payload.UserID.String(), backImageBytes)
	if err != nil {
		log.Printf("Failed to upload back-facing outfit image: %v", err)
		return p.markOutfitFailed(ctx, gen.ID, err)
	}

	if err := p.dbClient.User().UpdateProfilePictureUrl(ctx, payload.UserID, frontImageURL); err != nil {
		log.Printf("Failed to update user profile picture: %v", err)
		return p.markOutfitFailed(ctx, gen.ID, err)
	}

	clearedErr := ""
	completeUpdate := &models.OutfitProfileGeneration{
		Status:                models.OutfitGenerationStatusComplete,
		ErrorMessage:          &clearedErr,
		ProfilePictureUrl:     &frontImageURL,
		BackProfilePictureUrl: &backImageURL,
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

func (p *GenerateOutfitProfilePictureProcessor) inferAppearanceDescriptor(ctx context.Context, selfieURL string) string {
	if strings.TrimSpace(selfieURL) == "" {
		return ""
	}
	answer, err := p.deepPriestClient.PetitionTheFountWithImage(&deep_priest.QuestionWithImage{
		Question: "Describe the person's hair and key facial features in a short sentence (hair length, color, texture; facial hair; face shape; eyes; glasses). If no hair, say bald.",
		Image:    selfieURL,
	})
	if err != nil || answer == nil {
		return ""
	}
	desc := strings.TrimSpace(answer.Answer)
	if desc == "" {
		return ""
	}
	return desc
}

func buildOutfitPrompt(outfitName, appearanceDescriptor string, backView bool) string {
	viewDescription := "a front-facing, shoulders-up character portrait"
	if backView {
		viewDescription = "a back-facing, shoulders-up character portrait showing only the back of the head and shoulders"
	}

	prompt := fmt.Sprintf(outfitPromptTemplate, viewDescription, outfitName)
	if appearanceDescriptor != "" {
		if backView {
			prompt = fmt.Sprintf("%s\nMatch the same person from this appearance description by hair shape, length, color, and texture: %s.", prompt, appearanceDescriptor)
		} else {
			prompt = fmt.Sprintf("%s\nMatch the person's hair and key facial features exactly. Description: %s.", prompt, appearanceDescriptor)
		}
	} else {
		prompt = fmt.Sprintf("%s\nMatch the person's hair exactly; if bald or shaved, keep them bald (no hair).", prompt)
	}

	if backView {
		prompt = fmt.Sprintf("%s\nThe character must face away from the viewer. Do not show the face, eyes, nose, or mouth.", prompt)
	}

	return prompt
}

func (p *GenerateOutfitProfilePictureProcessor) generateOutfitPortrait(selfieURL, prompt string) ([]byte, error) {
	genRequest := deep_priest.GenerateImageRequest{
		Prompt: prompt,
		Model:  "gpt-image-1",
		N:      1,
		Size:   genSize,
	}
	deep_priest.ApplyGenerateImageDefaults(&genRequest)
	resp, err := p.deepPriestClient.GenerateImage(genRequest)
	if err != nil {
		log.Printf("Failed to generate outfit image, falling back to edit: %v", err)
		editRequest := deep_priest.EditImageRequest{
			Prompt:   prompt,
			ImageUrl: selfieURL,
			Model:    "dall-e-2",
			N:        1,
			Size:     genSize,
		}
		deep_priest.ApplyEditImageDefaults(&editRequest)
		// The edit endpoint does not accept these fields; ensure they're unset.
		editRequest.Quality = ""
		editRequest.ResponseFormat = ""
		resp, err = p.deepPriestClient.EditImage(editRequest)
		if err != nil {
			log.Printf("Fallback edit failed: %v", err)
			return nil, err
		}
	}

	candidates, err := decodeBase64Candidates(resp)
	if err != nil {
		log.Printf("Failed to decode outfit candidates: %v", err)
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no image candidates returned")
	}

	pp, _, err := EnforcePixelLook(candidates[0], iconSize, quantColors, upscaleOutput, true)
	if err != nil {
		log.Printf("Failed post-process outfit image: %v", err)
		return nil, err
	}

	return pp, nil
}

func (p *GenerateOutfitProfilePictureProcessor) resolveSelfieURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	// If already presigned, keep as-is.
	if strings.Contains(raw, "X-Amz-") || strings.Contains(raw, "X-Amz-Signature") {
		return raw
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return raw
	}
	bucket, key := parseS3BucketKey(parsed)
	if bucket == "" || key == "" {
		return raw
	}
	signed, err := p.awsClient.GeneratePresignedURL(bucket, key, time.Hour)
	if err != nil {
		log.Printf("Failed to presign selfie url: %v", err)
		return raw
	}
	return signed
}

func parseS3BucketKey(u *url.URL) (string, string) {
	host := u.Hostname()
	path := strings.TrimPrefix(u.Path, "/")
	if host == "" || path == "" {
		return "", ""
	}

	if strings.HasPrefix(host, "s3.") || host == "s3.amazonaws.com" {
		parts := strings.Split(path, "/")
		if len(parts) < 2 {
			return "", ""
		}
		return parts[0], strings.Join(parts[1:], "/")
	}

	if idx := strings.Index(host, ".s3."); idx > 0 {
		return host[:idx], path
	}

	if strings.HasSuffix(host, ".s3.amazonaws.com") {
		bucket := strings.TrimSuffix(host, ".s3.amazonaws.com")
		return bucket, path
	}

	return "", ""
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
