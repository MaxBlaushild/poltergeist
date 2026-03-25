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

const challengeTemplateImagePromptTemplate = `Create an image in the style of retro 16-bit RPG pixel art.
Use bold outlines, limited flat colors, and minimal dithering.
Apply 2-3 shading tones per area, with crisp, blocky edges.
Keep the result clean, simple, and non-photorealistic.
Avoid gradients, text, logos, and UI overlays.

Primary visual direction (highest priority; use this heavily):
%s

Challenge objective (secondary context):
%s

Location archetype context: %s.

Framing:
- Mid-shot environmental scene with a readable focal subject.
- Slight isometric or 3/4 perspective.
- Match the visual language used for scenario and monster art.`

type GenerateChallengeTemplateImageProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewGenerateChallengeTemplateImageProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
	awsClient aws.AWSClient,
) GenerateChallengeTemplateImageProcessor {
	log.Println("Initializing GenerateChallengeTemplateImageProcessor")
	return GenerateChallengeTemplateImageProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *GenerateChallengeTemplateImageProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate challenge template image task: %v", task.Type())

	var payload jobs.GenerateChallengeTemplateImageTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	template, err := p.dbClient.ChallengeTemplate().FindByID(ctx, payload.ChallengeTemplateID)
	if err != nil {
		return fmt.Errorf("failed to find challenge template: %w", err)
	}
	if template == nil {
		return fmt.Errorf("challenge template not found")
	}

	locationArchetypeName := "Unknown Location Archetype"
	if template.LocationArchetype != nil && strings.TrimSpace(template.LocationArchetype.Name) != "" {
		locationArchetypeName = strings.TrimSpace(template.LocationArchetype.Name)
	}
	description := strings.TrimSpace(template.Description)
	if description == "" {
		description = strings.TrimSpace(template.Question)
	}

	prompt := fmt.Sprintf(
		challengeTemplateImagePromptTemplate,
		truncate(description, 650),
		truncate(strings.TrimSpace(template.Question), 350),
		truncate(locationArchetypeName, 120),
	)
	request := deep_priest.GenerateImageRequest{
		Prompt: prompt,
	}
	deep_priest.ApplyGenerateImageDefaults(&request)
	imageB64, err := p.deepPriestClient.GenerateImage(request)
	if err != nil {
		return err
	}

	imageBytes, err := decodeImagePayload(imageB64)
	if err != nil {
		return err
	}

	imageURL, err := p.uploadChallengeTemplateImage(ctx, payload.ChallengeTemplateID.String(), imageBytes)
	if err != nil {
		return err
	}

	template.ImageURL = imageURL
	template.ThumbnailURL = imageURL
	if err := p.dbClient.ChallengeTemplate().Update(ctx, payload.ChallengeTemplateID, template); err != nil {
		return fmt.Errorf("failed to update challenge template image urls: %w", err)
	}

	return nil
}

func (p *GenerateChallengeTemplateImageProcessor) uploadChallengeTemplateImage(ctx context.Context, templateID string, imageBytes []byte) (string, error) {
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

	imageName := fmt.Sprintf("challenge-templates/%s-%d.%s", templateID, time.Now().UnixNano(), imageExtension)
	return p.awsClient.UploadImageToS3("crew-points-of-interest", imageName, imageBytes)
}
