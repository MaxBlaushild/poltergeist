package processors

import (
	"context"
	"encoding/base64"
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
	"github.com/MaxBlaushild/poltergeist/pkg/locationseeder"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type ApplyZoneSeedDraftProcessor struct {
	dbClient       db.DbClient
	locationSeeder locationseeder.Client
	deepPriest     deep_priest.DeepPriest
	awsClient      aws.AWSClient
	asyncClient    *asynq.Client
}

func NewApplyZoneSeedDraftProcessor(
	dbClient db.DbClient,
	locationSeeder locationseeder.Client,
	deepPriest deep_priest.DeepPriest,
	awsClient aws.AWSClient,
	asyncClient *asynq.Client,
) ApplyZoneSeedDraftProcessor {
	log.Println("Initializing ApplyZoneSeedDraftProcessor")
	return ApplyZoneSeedDraftProcessor{
		dbClient:       dbClient,
		locationSeeder: locationSeeder,
		deepPriest:     deepPriest,
		awsClient:      awsClient,
		asyncClient:    asyncClient,
	}
}

func (p *ApplyZoneSeedDraftProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing apply zone seed draft task: %v", task.Type())

	var payload jobs.ApplyZoneSeedDraftTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := p.dbClient.ZoneSeedJob().FindByID(ctx, payload.JobID)
	if err != nil {
		log.Printf("Failed to fetch zone seed job: %v", err)
		return err
	}
	if job == nil {
		log.Printf("Zone seed job not found: %v", payload.JobID)
		return nil
	}
	if job.Status == models.ZoneSeedStatusApplied {
		log.Printf("Zone seed job already applied: %v", payload.JobID)
		return nil
	}
	if job.Status != models.ZoneSeedStatusApproved && job.Status != models.ZoneSeedStatusApplying {
		log.Printf("Zone seed job not approved for apply: %v (status=%s)", payload.JobID, job.Status)
		return nil
	}

	job.Status = models.ZoneSeedStatusApplying
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ZoneSeedJob().Update(ctx, job); err != nil {
		log.Printf("Failed to update zone seed job status: %v", err)
		return err
	}

	zone, err := p.dbClient.Zone().FindByID(ctx, job.ZoneID)
	if err != nil {
		return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("failed to find zone: %w", err))
	}
	if zone == nil {
		return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("zone not found"))
	}

	if strings.TrimSpace(job.Draft.ZoneDescription) != "" {
		if err := p.dbClient.Zone().UpdateNameAndDescription(
			ctx,
			zone.ID,
			zone.Name,
			job.Draft.ZoneDescription,
		); err != nil {
			return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("failed to update zone description: %w", err))
		}
	}

	for _, draftPOI := range job.Draft.PointsOfInterest {
		if _, err := p.ensurePointOfInterest(ctx, zone, draftPOI.PlaceID); err != nil {
			return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("failed to import point of interest: %w", err))
		}
	}

	characterByDraftID := map[uuid.UUID]*models.Character{}
	for _, draftCharacter := range job.Draft.Characters {
		poi, err := p.ensurePointOfInterest(ctx, zone, draftCharacter.PlaceID)
		if err != nil {
			return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("failed to ensure character POI: %w", err))
		}
		character, err := p.createCharacterFromDraft(ctx, zone, poi, draftCharacter)
		if err != nil {
			return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("failed to create character: %w", err))
		}
		characterByDraftID[draftCharacter.DraftID] = character
	}

	for _, draftQuest := range job.Draft.Quests {
		poi, err := p.ensurePointOfInterest(ctx, zone, draftQuest.PlaceID)
		if err != nil {
			return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("failed to ensure quest POI: %w", err))
		}
		character := characterByDraftID[draftQuest.QuestGiverDraftID]
		if character == nil {
			for _, fallback := range characterByDraftID {
				character = fallback
				break
			}
		}
		if character == nil {
			return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("no quest giver available"))
		}

		if err := p.createQuestFromDraft(ctx, zone, poi, character, draftQuest); err != nil {
			return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("failed to create quest: %w", err))
		}
	}

	job.Status = models.ZoneSeedStatusApplied
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ZoneSeedJob().Update(ctx, job); err != nil {
		log.Printf("Failed to update zone seed job: %v", err)
		return err
	}

	log.Printf("Zone seed job %v applied", job.ID)
	return nil
}

func (p *ApplyZoneSeedDraftProcessor) failApplyZoneSeedJob(ctx context.Context, job *models.ZoneSeedJob, err error) error {
	msg := err.Error()
	job.Status = models.ZoneSeedStatusFailed
	job.ErrorMessage = &msg
	job.UpdatedAt = time.Now()
	if updateErr := p.dbClient.ZoneSeedJob().Update(ctx, job); updateErr != nil {
		log.Printf("Failed to mark zone seed job failed: %v", updateErr)
	}
	return err
}

func (p *ApplyZoneSeedDraftProcessor) ensurePointOfInterest(
	ctx context.Context,
	zone *models.Zone,
	placeID string,
) (*models.PointOfInterest, error) {
	if strings.TrimSpace(placeID) == "" {
		return nil, nil
	}
	existing, err := p.dbClient.PointOfInterest().FindByGoogleMapsPlaceID(ctx, placeID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	poi, err := p.locationSeeder.ImportPlace(ctx, placeID, *zone)
	if err != nil {
		return nil, err
	}
	return poi, nil
}

func (p *ApplyZoneSeedDraftProcessor) createCharacterFromDraft(
	ctx context.Context,
	zone *models.Zone,
	poi *models.PointOfInterest,
	draft models.ZoneSeedCharacterDraft,
) (*models.Character, error) {
	startLat, startLng := zone.Latitude, zone.Longitude
	if poi != nil {
		lat, lng, err := parsePointOfInterestCoords(poi)
		if err == nil {
			startLat = lat
			startLng = lng
		}
	}

	movementPattern := &models.MovementPattern{
		MovementPatternType: models.MovementPatternStatic,
		ZoneID:              &zone.ID,
		StartingLatitude:    startLat,
		StartingLongitude:   startLng,
		Path:                models.LocationPath{},
	}
	if err := p.dbClient.MovementPattern().Create(ctx, movementPattern); err != nil {
		return nil, err
	}

	character := &models.Character{
		Name:                  draft.Name,
		Description:           draft.Description,
		PointOfInterestID:     nil,
		MovementPatternID:     movementPattern.ID,
		MovementPattern:       *movementPattern,
		ImageGenerationStatus: models.CharacterImageGenerationStatusQueued,
	}
	if poi != nil {
		character.PointOfInterestID = &poi.ID
	}

	if err := p.dbClient.Character().Create(ctx, character); err != nil {
		return nil, err
	}

	payloadBytes, err := json.Marshal(jobs.GenerateCharacterImageTaskPayload{
		CharacterID: character.ID,
		Name:        character.Name,
		Description: character.Description,
	})
	if err != nil {
		return nil, err
	}

	if _, err := p.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateCharacterImageTaskType, payloadBytes)); err != nil {
		errMsg := err.Error()
		_ = p.dbClient.Character().Update(ctx, character.ID, &models.Character{
			ImageGenerationStatus: models.CharacterImageGenerationStatusFailed,
			ImageGenerationError:  &errMsg,
		})
		return nil, err
	}

	return character, nil
}

func (p *ApplyZoneSeedDraftProcessor) createQuestFromDraft(
	ctx context.Context,
	zone *models.Zone,
	poi *models.PointOfInterest,
	character *models.Character,
	draft models.ZoneSeedQuestDraft,
) error {
	acceptanceDialogue := models.StringArray(draft.AcceptanceDialogue)
	if acceptanceDialogue == nil {
		acceptanceDialogue = models.StringArray{}
	}

	now := time.Now()
	quest := &models.Quest{
		ID:                    uuid.New(),
		CreatedAt:             now,
		UpdatedAt:             now,
		Name:                  draft.Name,
		Description:           draft.Description,
		AcceptanceDialogue:    acceptanceDialogue,
		ZoneID:                &zone.ID,
		QuestGiverCharacterID: &character.ID,
		Gold:                  draft.Gold,
	}

	if imageURL, err := p.generateQuestImage(ctx, draft.Name, draft.Description); err == nil {
		quest.ImageURL = imageURL
	}

	if err := p.dbClient.Quest().Create(ctx, quest); err != nil {
		return err
	}

	if err := p.ensureQuestActionForCharacter(ctx, quest.ID, character.ID); err != nil {
		log.Printf("Failed to create quest action for character: %v", err)
	}

	node := &models.QuestNode{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		QuestID:           quest.ID,
		OrderIndex:        0,
		PointOfInterestID: nil,
		SubmissionType:    models.DefaultQuestNodeSubmissionType(),
	}
	if poi != nil {
		node.PointOfInterestID = &poi.ID
	}
	if err := p.dbClient.QuestNode().Create(ctx, node); err != nil {
		return err
	}

	challenge := &models.QuestNodeChallenge{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		QuestNodeID:    node.ID,
		Tier:           0,
		Question:       draft.ChallengeQuestion,
		Reward:         0,
		SubmissionType: models.DefaultQuestNodeSubmissionType(),
		Difficulty:     0,
	}
	if err := p.dbClient.QuestNodeChallenge().Create(ctx, challenge); err != nil {
		return err
	}

	if poi != nil {
		if err := p.dbClient.PointOfInterest().UpdateLastUsedInQuest(ctx, poi.ID); err != nil {
			log.Printf("Failed to update POI last_used_in_quest_at: %v", err)
		}
	}

	return nil
}

func (p *ApplyZoneSeedDraftProcessor) ensureQuestActionForCharacter(ctx context.Context, questID uuid.UUID, characterID uuid.UUID) error {
	actions, err := p.dbClient.CharacterAction().FindByCharacterID(ctx, characterID)
	if err != nil {
		return err
	}
	questIDStr := questID.String()
	for _, action := range actions {
		if action.ActionType != models.ActionTypeGiveQuest {
			continue
		}
		if action.Metadata == nil {
			continue
		}
		if value, ok := action.Metadata["questId"]; ok && fmt.Sprint(value) == questIDStr {
			return nil
		}
	}

	action := &models.CharacterAction{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CharacterID: characterID,
		ActionType:  models.ActionTypeGiveQuest,
		Dialogue:    []models.DialogueMessage{},
		Metadata:    map[string]interface{}{"questId": questIDStr},
	}
	return p.dbClient.CharacterAction().Create(ctx, action)
}

const questImagePromptTemplate = `
You are a video game designer tasked with creating visual assets for quests in a fantasy role playing game.

The quest is this:

%s

Please describe what an iconic moment from this quest would look like to an outside observer.

Please format your response as a JSON object with the following fields:
{
  "description": "string"
}
`

type questImagePrompt struct {
	Description string `json:"description"`
}

func (p *ApplyZoneSeedDraftProcessor) generateQuestImage(ctx context.Context, name, description string) (string, error) {
	if strings.TrimSpace(name) == "" && strings.TrimSpace(description) == "" {
		return "", fmt.Errorf("quest copy was empty")
	}

	prompt := fmt.Sprintf(questImagePromptTemplate, fmt.Sprintf("%s\n\n%s", name, description))
	answer, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return "", err
	}

	var imagePrompt questImagePrompt
	if err := json.Unmarshal([]byte(extractJSON(answer.Answer)), &imagePrompt); err != nil {
		imagePrompt.Description = strings.TrimSpace(answer.Answer)
	}
	imagePrompt.Description = strings.TrimSpace(imagePrompt.Description)
	if imagePrompt.Description == "" {
		return "", fmt.Errorf("quest image prompt was empty")
	}

	request := deep_priest.GenerateImageRequest{
		Prompt: imagePrompt.Description,
	}
	deep_priest.ApplyGenerateImageDefaults(&request)
	base64Image, err := p.deepPriest.GenerateImage(request)
	if err != nil {
		return "", err
	}

	return p.uploadImage(ctx, base64Image)
}

func (p *ApplyZoneSeedDraftProcessor) uploadImage(ctx context.Context, base64Image string) (string, error) {
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
	imageName := timestamp + "-" + uuid.New().String() + "." + imageExtension

	imageURL, err := p.awsClient.UploadImageToS3("crew-profile-icons", imageName, imageBytes)
	if err != nil {
		return "", err
	}

	return imageURL, nil
}

func parsePointOfInterestCoords(poi *models.PointOfInterest) (float64, float64, error) {
	if poi == nil {
		return 0, 0, fmt.Errorf("poi is nil")
	}
	if strings.TrimSpace(poi.Lat) == "" || strings.TrimSpace(poi.Lng) == "" {
		return 0, 0, fmt.Errorf("poi has no coordinates")
	}
	lat, err := strconv.ParseFloat(poi.Lat, 64)
	if err != nil {
		return 0, 0, err
	}
	lng, err := strconv.ParseFloat(poi.Lng, 64)
	if err != nil {
		return 0, 0, err
	}
	return lat, lng, nil
}
