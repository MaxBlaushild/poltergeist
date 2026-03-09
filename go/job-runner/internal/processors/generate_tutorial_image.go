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
	"github.com/hibiken/asynq"
)

const tutorialScenarioImagePromptTemplate = `Create an image in the style of retro 16-bit RPG pixel art.
Use bold outlines, limited flat colors, and minimal dithering.
Apply 2-3 shading tones per area, with crisp, blocky edges.
Keep the result clean, simple, and non-photorealistic.
Avoid gradients, text, logos, and UI overlays.

Render a fantasy scene illustrating this scenario prompt:
%s

Zone context: %s.

Framing:
- Mid-shot environmental scene with a readable focal subject.
- Slight isometric or 3/4 perspective.
- Match the visual language used for character portraits, outfit portraits, and POI images.`

const tutorialScenarioImageZoneContext = "Unclaimed Streets tutorial starting point"

type GenerateTutorialImageProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewGenerateTutorialImageProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
	awsClient aws.AWSClient,
) GenerateTutorialImageProcessor {
	log.Println("Initializing GenerateTutorialImageProcessor")
	return GenerateTutorialImageProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *GenerateTutorialImageProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate tutorial image task: %v", task.Type())

	var payload jobs.GenerateTutorialImageTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal tutorial image task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	config, err := p.dbClient.Tutorial().GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load tutorial config: %w", err)
	}

	config.ImageGenerationStatus = models.TutorialImageGenerationStatusInProgress
	config.ImageGenerationError = nil
	if _, err := p.dbClient.Tutorial().UpsertConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to mark tutorial image generation in progress: %w", err)
	}

	scenarioPrompt := strings.TrimSpace(payload.ScenarioPrompt)
	if scenarioPrompt == "" {
		scenarioPrompt = strings.TrimSpace(config.ScenarioPrompt)
	}
	if scenarioPrompt == "" {
		return p.markFailed(ctx, fmt.Errorf("scenario prompt is required"))
	}

	prompt := fmt.Sprintf(
		tutorialScenarioImagePromptTemplate,
		truncateTutorialPrompt(scenarioPrompt, 650),
		truncateTutorialPrompt(tutorialScenarioImageZoneContext, 120),
	)

	request := deep_priest.GenerateImageRequest{Prompt: prompt}
	deep_priest.ApplyGenerateImageDefaults(&request)

	imagePayload, err := p.deepPriestClient.GenerateImage(request)
	if err != nil {
		log.Printf("Failed to generate tutorial image: %v", err)
		return p.markFailed(ctx, err)
	}

	imageBytes, err := decodeImagePayload(imagePayload)
	if err != nil {
		log.Printf("Failed to decode tutorial image payload: %v", err)
		return p.markFailed(ctx, err)
	}

	imageURL, err := p.uploadImage(imageBytes)
	if err != nil {
		log.Printf("Failed to upload tutorial image: %v", err)
		return p.markFailed(ctx, err)
	}

	config, err = p.dbClient.Tutorial().GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to reload tutorial config: %w", err)
	}
	config.ScenarioImageURL = imageURL
	config.ImageGenerationStatus = models.TutorialImageGenerationStatusComplete
	config.ImageGenerationError = nil
	if _, err := p.dbClient.Tutorial().UpsertConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to save tutorial image url: %w", err)
	}

	return nil
}

func (p *GenerateTutorialImageProcessor) markFailed(ctx context.Context, err error) error {
	config, findErr := p.dbClient.Tutorial().GetConfig(ctx)
	if findErr == nil && config != nil {
		errMsg := err.Error()
		config.ImageGenerationStatus = models.TutorialImageGenerationStatusFailed
		config.ImageGenerationError = &errMsg
		if _, upsertErr := p.dbClient.Tutorial().UpsertConfig(ctx, config); upsertErr != nil {
			log.Printf("Failed to mark tutorial image generation failed: %v", upsertErr)
		}
	}
	return err
}

func (p *GenerateTutorialImageProcessor) uploadImage(imageBytes []byte) (string, error) {
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

	imageName := fmt.Sprintf("tutorials/scenario-%d.%s", time.Now().UnixNano(), imageExtension)
	return p.awsClient.UploadImageToS3("crew-points-of-interest", imageName, imageBytes)
}

func truncateTutorialPrompt(value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if max <= 0 || len(trimmed) <= max {
		return trimmed
	}
	return strings.TrimSpace(trimmed[:max]) + "..."
}
