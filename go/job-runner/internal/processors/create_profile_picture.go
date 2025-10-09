package processors

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/hibiken/asynq"
)

const prompt = `
Re-create this selfie as a pixelated retro video-game profile portrait, keeping the same face structure, expression, pose, and framing as the original photo.
Maintain recognizable likeness and proportions of the person.
Render in 8-bit or 16-bit RPG style with bold outlines, blocky shading, and a limited color palette reminiscent of classic SNES-era fantasy RPGs.
Replace modern clothing with generic adventurer gear such as a tunic, cloak, or light armor.
Keep lighting direction and general composition consistent with the input image.
Background should be simple or transparent.
The final result should look like a stylized, non-photorealistic pixel-art avatar, clean and crisp, with no smooth gradients or painterly details.
`

type CreateProfilePictureProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewCreateProfilePictureProcessor(dbClient db.DbClient, deepPriestClient deep_priest.DeepPriest, awsClient aws.AWSClient) CreateProfilePictureProcessor {
	log.Println("Initializing CreateProfilePictureProcessor")
	return CreateProfilePictureProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *CreateProfilePictureProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing create profile picture task: %v", task.Type())

	var payload jobs.CreateProfilePictureTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Processing create profile picture task: %v", payload)

	user, err := p.dbClient.User().FindByID(ctx, payload.UserID)
	if err != nil {
		log.Printf("Failed to find user: %v", err)
		return fmt.Errorf("failed to find user: %w", err)
	}

	log.Printf("Generating profile picture for user ID: %v", payload.UserID)
	base64JSON, err := p.deepPriestClient.EditImage(deep_priest.EditImageRequest{
		Prompt:         prompt,
		ImageUrl:       payload.ProfilePictureUrl,
		Model:          "gpt-image-1",
		N:              1,
		Quality:        "standard",
		Size:           "1024x1024",
		User:           "poltergeist",
		ResponseFormat: "b64_json",
	})
	if err != nil {
		log.Printf("Failed to generate profile picture: %v", err)
		return fmt.Errorf("failed to generate profile picture: %w", err)
	}

	url, err := p.UploadImage(ctx, user.ID.String(), base64JSON)
	if err != nil {
		log.Printf("Failed to download image: %v", err)
		return fmt.Errorf("failed to download image: %w", err)
	}

	log.Printf("Profile picture generated successfully for user ID: %v", payload.UserID)

	err = p.dbClient.User().UpdateProfilePictureUrl(ctx, payload.UserID, url)
	if err != nil {
		log.Printf("Failed to update user profile picture URL: %v", err)
		return fmt.Errorf("failed to update user profile picture URL: %w", err)
	}

	log.Printf("Profile picture updated successfully for user ID: %v", payload.UserID)
	return nil
}

func (p *CreateProfilePictureProcessor) UploadImage(ctx context.Context, userID string, base64Image string) (string, error) {
	imageBytes, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		return "", err
	}

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
