package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"math"
	"math/rand"
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

const baseDescriptionPromptTemplate = `
You are writing a short base description for a fantasy MMORPG-style game.

Base context:
- owner display: %s
- latitude: %.6f
- longitude: %.6f
- regional cue: %s
- inspiration motifs: %s
- variation seed: %s

Return JSON only:
{
  "description": "exactly 2 sentences"
}

Rules:
- Write exactly 2 sentences.
- The tone should feel proud, hopeful, and optimistic, as if this base is a hard-won foothold with promise ahead.
- Use the location as inspiration for the imagined terrain, weather, approach, skyline, or surrounding district.
- Make the wording feel specific and unique for this base rather than generic housing copy.
- Do not mention GPS, coordinates, apps, maps, latitude, longitude, or modern infrastructure.
- Do not mention the variation seed directly.
- Avoid direct second-person instructions.
`

const baseImagePromptTemplate = `
You are designing art direction for a player-built base in a fantasy MMORPG.

Base context:
- owner display: %s
- latitude: %.6f
- longitude: %.6f
- regional cue: %s
- inspiration motifs: %s
- variation seed: %s

Return JSON only:
{
  "text": "single concise image prompt"
}

Rules:
- The image should match the aesthetic of retro 16-bit RPG pixel art point of interest images:
  - Crisp outlines, limited color palette, clean background
  - Centered subject, readable silhouette
  - Exterior view, 3/4 angle or slight isometric perspective
  - No text, no logos, no UI
- Depict a proud, well-kept adventurer base, camp, lodge, hall, or fortified homestead.
- Use the location and motifs as inspiration for terrain, weather, materials, skyline, vegetation, and approach.
- Keep the result specific and varied, not a generic cottage every time.
- Do not mention coordinates, GPS, apps, maps, or modern infrastructure.
- Do not mention the variation seed directly.
`

type baseDescriptionGenerationResponse struct {
	Description string `json:"description"`
}

type baseImagePromptResponse struct {
	Text string `json:"text"`
}

type GenerateBaseDescriptionProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
	asyncClient      *asynq.Client
}

func NewGenerateBaseDescriptionProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
	awsClient aws.AWSClient,
	asyncClient *asynq.Client,
) GenerateBaseDescriptionProcessor {
	log.Println("Initializing GenerateBaseDescriptionProcessor")
	return GenerateBaseDescriptionProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
		asyncClient:      asyncClient,
	}
}

func (p *GenerateBaseDescriptionProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate base description task: %v", task.Type())

	var payload jobs.GenerateBaseDescriptionTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := p.dbClient.BaseDescriptionGenerationJob().FindByID(ctx, payload.JobID)
	if err != nil {
		return err
	}
	if job == nil {
		log.Printf("Base description generation job %s not found", payload.JobID)
		return nil
	}

	job.Status = models.BaseDescriptionGenerationStatusInProgress
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.BaseDescriptionGenerationJob().Update(ctx, job); err != nil {
		return err
	}

	if err := p.generateDescription(ctx, job); err != nil {
		return p.failBaseDescriptionGenerationJob(ctx, job, err)
	}

	return nil
}

func (p *GenerateBaseDescriptionProcessor) generateDescription(ctx context.Context, job *models.BaseDescriptionGenerationJob) error {
	base, err := p.dbClient.Base().FindByID(ctx, job.BaseID)
	if err != nil {
		return fmt.Errorf("failed to load base: %w", err)
	}
	if base == nil {
		return fmt.Errorf("base not found")
	}

	ownerDisplay := "an adventurer"
	if base.User.Username != nil && strings.TrimSpace(*base.User.Username) != "" {
		ownerDisplay = "@" + strings.TrimSpace(*base.User.Username)
	} else if strings.TrimSpace(base.User.Name) != "" {
		ownerDisplay = strings.TrimSpace(base.User.Name)
	}

	seed := job.ID.String()
	motifs := strings.Join(selectBaseDescriptionMotifs(seed), ", ")
	prompt := fmt.Sprintf(
		baseDescriptionPromptTemplate,
		ownerDisplay,
		base.Latitude,
		base.Longitude,
		baseRegionalCue(base.Latitude),
		motifs,
		seed,
	)

	answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return fmt.Errorf("failed to generate base description: %w", err)
	}

	var generated baseDescriptionGenerationResponse
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), &generated); err != nil {
		return fmt.Errorf("failed to parse generated base description payload: %w", err)
	}

	description := sanitizeBaseDescription(generated.Description)
	if description == "" {
		return fmt.Errorf("generated base description was empty")
	}

	imageURL, err := p.generateBaseImage(ctx, base, ownerDisplay, motifs, seed)
	if err != nil {
		return fmt.Errorf("failed to generate base image: %w", err)
	}

	if err := p.dbClient.Base().UpdateFlavor(ctx, base.ID, description, imageURL); err != nil {
		return fmt.Errorf("failed to update base flavor: %w", err)
	}

	p.enqueueThumbnailTask(base.ID, imageURL)

	job.Status = models.BaseDescriptionGenerationStatusCompleted
	job.GeneratedDescription = &description
	job.GeneratedImageURL = &imageURL
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.BaseDescriptionGenerationJob().Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update base description generation job: %w", err)
	}

	return nil
}

func (p *GenerateBaseDescriptionProcessor) generateBaseImage(
	ctx context.Context,
	base *models.Base,
	ownerDisplay string,
	motifs string,
	seed string,
) (string, error) {
	promptRequest := fmt.Sprintf(
		baseImagePromptTemplate,
		ownerDisplay,
		base.Latitude,
		base.Longitude,
		baseRegionalCue(base.Latitude),
		motifs,
		seed,
	)
	answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: promptRequest})
	if err != nil {
		return "", err
	}

	var generatedPrompt baseImagePromptResponse
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), &generatedPrompt); err != nil {
		return "", err
	}
	imagePrompt := strings.TrimSpace(generatedPrompt.Text)
	if imagePrompt == "" {
		return "", fmt.Errorf("generated base image prompt was empty")
	}

	request := deep_priest.GenerateImageRequest{Prompt: imagePrompt}
	deep_priest.ApplyGenerateImageDefaults(&request)
	imagePayload, err := p.deepPriestClient.GenerateImage(request)
	if err != nil {
		return "", err
	}

	imageBytes, err := decodeCharacterImagePayload(imagePayload)
	if err != nil {
		return "", err
	}

	return p.uploadImage(base.ID, imageBytes)
}

func (p *GenerateBaseDescriptionProcessor) uploadImage(baseID uuid.UUID, imageBytes []byte) (string, error) {
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

	imageName := fmt.Sprintf("bases/%s-%d.%s", baseID.String(), time.Now().UnixNano(), imageExtension)
	return p.awsClient.UploadImageToS3(jobs.ThumbnailBucket, imageName, imageBytes)
}

func (p *GenerateBaseDescriptionProcessor) enqueueThumbnailTask(baseID uuid.UUID, imageURL string) {
	if p.asyncClient == nil || strings.TrimSpace(imageURL) == "" {
		return
	}
	payload := jobs.GenerateImageThumbnailTaskPayload{
		EntityType: jobs.ThumbnailEntityBase,
		EntityID:   baseID,
		SourceUrl:  imageURL,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal base thumbnail task payload: %v", err)
		return
	}
	if _, err := p.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateImageThumbnailTaskType, payloadBytes)); err != nil {
		log.Printf("Failed to enqueue base thumbnail task: %v", err)
	}
}

func baseRegionalCue(latitude float64) string {
	absLat := math.Abs(latitude)
	switch {
	case absLat < 12:
		return "lush frontier near warm lowlands"
	case absLat < 28:
		return "temperate marches with long roads and open sky"
	case absLat < 45:
		return "seasonal borderland shaped by weather and watchfulness"
	case absLat < 60:
		return "cool high-country edge with brisk air and stubborn stone"
	default:
		return "windswept far reach where only determined settlers endure"
	}
}

func selectBaseDescriptionMotifs(seed string) []string {
	pools := [][]string{
		{"hearthglow", "watchfire", "bannered gate", "garden wall", "stone stoop"},
		{"rising smoke", "lantern light", "fresh-cut timber", "weathered stone", "bright pennants"},
		{"new ambition", "earned shelter", "steady growth", "quiet pride", "hopeful frontier"},
	}

	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(seed))
	rng := rand.New(rand.NewSource(int64(hasher.Sum64())))

	selected := make([]string, 0, len(pools))
	for _, pool := range pools {
		if len(pool) == 0 {
			continue
		}
		selected = append(selected, pool[rng.Intn(len(pool))])
	}
	return selected
}

func sanitizeBaseDescription(value string) string {
	cleaned := strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if cleaned == "" {
		return ""
	}
	return truncateToSentenceCount(cleaned, 2)
}

func truncateToSentenceCount(value string, sentenceCount int) string {
	if sentenceCount <= 0 || value == "" {
		return strings.TrimSpace(value)
	}
	count := 0
	for idx, r := range value {
		switch r {
		case '.', '!', '?':
			count++
			if count >= sentenceCount {
				return strings.TrimSpace(value[:idx+1])
			}
		}
	}
	return strings.TrimSpace(value)
}

func (p *GenerateBaseDescriptionProcessor) failBaseDescriptionGenerationJob(
	ctx context.Context,
	job *models.BaseDescriptionGenerationJob,
	err error,
) error {
	message := err.Error()
	job.Status = models.BaseDescriptionGenerationStatusFailed
	job.ErrorMessage = &message
	job.UpdatedAt = time.Now()
	if updateErr := p.dbClient.BaseDescriptionGenerationJob().Update(ctx, job); updateErr != nil {
		log.Printf("Failed to mark base description generation job %s as failed: %v", job.ID, updateErr)
	}
	return err
}
