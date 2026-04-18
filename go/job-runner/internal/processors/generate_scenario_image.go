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

const scenarioImagePromptTemplate = `Create an image in the style of retro 16-bit RPG pixel art.
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

type GenerateScenarioImageProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewGenerateScenarioImageProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
	awsClient aws.AWSClient,
) GenerateScenarioImageProcessor {
	log.Println("Initializing GenerateScenarioImageProcessor")
	return GenerateScenarioImageProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *GenerateScenarioImageProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate scenario image task: %v", task.Type())

	var payload jobs.GenerateScenarioImageTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	scenario, err := p.dbClient.Scenario().FindByID(ctx, payload.ScenarioID)
	if err != nil {
		return fmt.Errorf("failed to find scenario: %w", err)
	}
	if scenario == nil {
		return fmt.Errorf("scenario not found")
	}

	zoneName := "Unknown Zone"
	if strings.TrimSpace(scenario.Zone.Name) != "" {
		zoneName = strings.TrimSpace(scenario.Zone.Name)
	}

	prompt := buildScenarioImagePrompt(scenario, zoneName)
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

	imageURL, err := p.uploadScenarioImage(ctx, payload.ScenarioID.String(), imageBytes)
	if err != nil {
		return err
	}

	scenario.ImageURL = imageURL
	scenario.ThumbnailURL = imageURL
	if err := p.dbClient.Scenario().Update(ctx, payload.ScenarioID, scenario); err != nil {
		return fmt.Errorf("failed to update scenario image urls: %w", err)
	}

	return nil
}

func (p *GenerateScenarioImageProcessor) uploadScenarioImage(ctx context.Context, scenarioID string, imageBytes []byte) (string, error) {
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

	imageName := fmt.Sprintf("scenarios/%s-%d.%s", scenarioID, time.Now().UnixNano(), imageExtension)
	return p.awsClient.UploadImageToS3("crew-points-of-interest", imageName, imageBytes)
}

func buildScenarioImagePrompt(
	scenario *models.Scenario,
	zoneName string,
) string {
	base := fmt.Sprintf(
		scenarioImagePromptTemplate,
		truncate(strings.TrimSpace(scenario.Prompt), 650),
		truncate(zoneName, 120),
	)
	if scenario == nil || isBaselineFantasyScenarioGenre(scenario.Genre) {
		return base
	}
	direction := scenarioGenreImageDirection(scenario.Genre)
	if direction == "" {
		return base
	}
	return strings.TrimSpace(base + "\n\nGenre direction: " + direction)
}
