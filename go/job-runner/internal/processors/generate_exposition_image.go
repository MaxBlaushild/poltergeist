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
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/hibiken/asynq"
)

const expositionImagePromptTemplate = `Create an image in the style of retro 16-bit RPG pixel art.
Use bold outlines, limited flat colors, and minimal dithering.
Apply 2-3 shading tones per area, with crisp, blocky edges.
Keep the result clean, simple, and non-photorealistic.
Avoid gradients, text, logos, speech bubbles, and UI overlays.

Primary visual direction:
%s

Dialogue mood and scene context:
%s

Zone context: %s.

Framing:
- Mid-shot environmental scene with a readable focal subject.
- Slight isometric or 3/4 perspective.
- Suggest a conversation or revelation scene with one or more fantasy characters present.
- Match the visual language used for scenario, challenge, and monster art.`

type GenerateExpositionImageProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewGenerateExpositionImageProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
	awsClient aws.AWSClient,
) GenerateExpositionImageProcessor {
	log.Println("Initializing GenerateExpositionImageProcessor")
	return GenerateExpositionImageProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *GenerateExpositionImageProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate exposition image task: %v", task.Type())

	var payload jobs.GenerateExpositionImageTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	exposition, err := p.dbClient.Exposition().FindByID(ctx, payload.ExpositionID)
	if err != nil {
		return fmt.Errorf("failed to find exposition: %w", err)
	}
	if exposition == nil {
		return fmt.Errorf("exposition not found")
	}

	zoneName := "Unknown Zone"
	if strings.TrimSpace(exposition.Zone.Name) != "" {
		zoneName = strings.TrimSpace(exposition.Zone.Name)
	}

	description := strings.TrimSpace(exposition.Description)
	if description == "" {
		description = strings.TrimSpace(exposition.Title)
	}

	dialogueLines := make([]string, 0, len(exposition.Dialogue))
	for _, message := range exposition.Dialogue {
		text := strings.TrimSpace(message.Text)
		if text == "" {
			continue
		}
		dialogueLines = append(dialogueLines, text)
		if len(dialogueLines) >= 3 {
			break
		}
	}
	dialogueSummary := strings.Join(dialogueLines, " ")
	if dialogueSummary == "" {
		dialogueSummary = description
	}

	prompt := fmt.Sprintf(
		expositionImagePromptTemplate,
		truncate(description, 650),
		truncate(dialogueSummary, 400),
		truncate(zoneName, 120),
	)
	request := deep_priest.GenerateImageRequest{Prompt: prompt}
	deep_priest.ApplyGenerateImageDefaults(&request)
	imageB64, err := p.deepPriestClient.GenerateImage(request)
	if err != nil {
		return err
	}

	imageBytes, err := decodeImagePayload(imageB64)
	if err != nil {
		return err
	}

	imageURL, err := p.uploadExpositionImage(ctx, payload.ExpositionID.String(), imageBytes)
	if err != nil {
		return err
	}

	exposition.ImageURL = imageURL
	exposition.ThumbnailURL = imageURL
	exposition.UpdatedAt = time.Now()
	if err := p.dbClient.Exposition().Update(ctx, payload.ExpositionID, exposition); err != nil {
		return fmt.Errorf("failed to update exposition image urls: %w", err)
	}

	return nil
}

func (p *GenerateExpositionImageProcessor) uploadExpositionImage(
	ctx context.Context,
	expositionID string,
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

	imageName := fmt.Sprintf("expositions/%s-%d.%s", expositionID, time.Now().UnixNano(), imageExtension)
	return p.awsClient.UploadImageToS3("crew-points-of-interest", imageName, imageBytes)
}
