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

const challengeImagePromptTemplate = `Create an image in the style of retro 16-bit RPG pixel art.
Use bold outlines, limited flat colors, and minimal dithering.
Apply 2-3 shading tones per area, with crisp, blocky edges.
Keep the result clean, simple, and non-photorealistic.
Avoid gradients, text, logos, and UI overlays.

Render a fantasy scene illustrating this challenge prompt:
%s

Zone context: %s.

Framing:
- Mid-shot environmental scene with a readable focal subject.
- Slight isometric or 3/4 perspective.
- Match the visual language used for scenario and monster art.`

type GenerateChallengeImageProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewGenerateChallengeImageProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
	awsClient aws.AWSClient,
) GenerateChallengeImageProcessor {
	log.Println("Initializing GenerateChallengeImageProcessor")
	return GenerateChallengeImageProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *GenerateChallengeImageProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate challenge image task: %v", task.Type())

	var payload jobs.GenerateChallengeImageTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	challenge, err := p.dbClient.Challenge().FindByID(ctx, payload.ChallengeID)
	if err != nil {
		return fmt.Errorf("failed to find challenge: %w", err)
	}
	if challenge == nil {
		return fmt.Errorf("challenge not found")
	}

	zoneName := "Unknown Zone"
	if strings.TrimSpace(challenge.Zone.Name) != "" {
		zoneName = strings.TrimSpace(challenge.Zone.Name)
	}

	prompt := fmt.Sprintf(
		challengeImagePromptTemplate,
		truncate(strings.TrimSpace(challenge.Question), 650),
		truncate(zoneName, 120),
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

	imageURL, err := p.uploadChallengeImage(ctx, payload.ChallengeID.String(), imageBytes)
	if err != nil {
		return err
	}

	challenge.ImageURL = imageURL
	challenge.ThumbnailURL = imageURL
	if err := p.dbClient.Challenge().Update(ctx, payload.ChallengeID, challenge); err != nil {
		return fmt.Errorf("failed to update challenge image urls: %w", err)
	}

	return nil
}

func (p *GenerateChallengeImageProcessor) uploadChallengeImage(ctx context.Context, challengeID string, imageBytes []byte) (string, error) {
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

	imageName := fmt.Sprintf("challenges/%s-%d.%s", challengeID, time.Now().UnixNano(), imageExtension)
	return p.awsClient.UploadImageToS3("crew-points-of-interest", imageName, imageBytes)
}
