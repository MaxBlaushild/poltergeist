package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/locationseeder"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type ApplyZoneSeedDraftProcessor struct {
	dbClient       db.DbClient
	locationSeeder locationseeder.Client
	deepPriest     deep_priest.DeepPriest
	asyncClient    *asynq.Client
}

const (
	zoneSeedHealingFountainDefaultName        = "Healing Fountain"
	zoneSeedHealingFountainDefaultDescription = "A mythic spring that restores travelers. Discover it to unlock its blessing."
	zoneSeedHealingFountainDefaultThumbnail   = "https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/poi-undiscovered.png"
)

type questRewardStatBonuses struct {
	Strength     int
	Dexterity    int
	Constitution int
	Intelligence int
	Wisdom       int
	Charisma     int
}

func (b questRewardStatBonuses) total() int {
	return b.Strength + b.Dexterity + b.Constitution + b.Intelligence + b.Wisdom + b.Charisma
}

func NewApplyZoneSeedDraftProcessor(
	dbClient db.DbClient,
	locationSeeder locationseeder.Client,
	deepPriest deep_priest.DeepPriest,
	asyncClient *asynq.Client,
) ApplyZoneSeedDraftProcessor {
	log.Println("Initializing ApplyZoneSeedDraftProcessor")
	return ApplyZoneSeedDraftProcessor{
		dbClient:       dbClient,
		locationSeeder: locationSeeder,
		deepPriest:     deepPriest,
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

	nextDescription := zone.Description
	if strings.TrimSpace(job.Draft.ZoneDescription) != "" {
		nextDescription = job.Draft.ZoneDescription
	}
	nextKind := zone.Kind
	if strings.TrimSpace(job.ZoneKind) != "" {
		nextKind = models.NormalizeZoneKind(job.ZoneKind)
	}
	if nextDescription != zone.Description || nextKind != zone.Kind {
		updatedZone, err := p.dbClient.Zone().UpdateMetadata(
			ctx,
			zone.ID,
			zone.Name,
			nextDescription,
			nextKind,
			zone.InternalTags,
		)
		if err != nil {
			return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("failed to update zone metadata: %w", err))
		}
		if updatedZone != nil {
			zone = updatedZone
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
		if err := p.ensureTalkActionForCharacterDialogue(ctx, character, draftCharacter.Dialogue, poi); err != nil {
			return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("failed to create talk action: %w", err))
		}
		if err := p.ensureShopActionForCharacterTags(ctx, character, draftCharacter.ShopItemTags); err != nil {
			return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("failed to create shop action: %w", err))
		}
		characterByDraftID[draftCharacter.DraftID] = character
	}

	if err := p.seedStandaloneChallengesForPOIs(ctx, zone, job, characterByDraftID); err != nil {
		return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("failed to seed standalone challenges: %w", err))
	}

	if err := p.seedMonsterEncountersForZone(ctx, zone, job); err != nil {
		return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("failed to seed monster encounters: %w", err))
	}

	scenarioFallbackLocations := zoneSeedScenarioLocations(job.Draft.PointsOfInterest)
	inputQueued, inputFailed := p.enqueueScenarioGenerationJobs(
		ctx,
		zone,
		job.ZoneID,
		true,
		job.InputEncounterCount,
		scenarioFallbackLocations,
		true,
	)
	optionQueued, optionFailed := p.enqueueScenarioGenerationJobs(
		ctx,
		zone,
		job.ZoneID,
		false,
		job.OptionEncounterCount,
		scenarioFallbackLocations,
		true,
	)
	if inputFailed > 0 || optionFailed > 0 {
		log.Printf(
			"Zone seed job %v seeded scenario generation jobs with failures (input queued=%d failed=%d, option queued=%d failed=%d)",
			job.ID,
			inputQueued,
			inputFailed,
			optionQueued,
			optionFailed,
		)
	}

	if err := p.seedTreasureChestsForZone(ctx, zone, job); err != nil {
		return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("failed to seed treasure chests: %w", err))
	}
	if err := p.seedHealingFountainsForZone(ctx, zone, job); err != nil {
		return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("failed to seed healing fountains: %w", err))
	}
	if err := p.seedResourcesForZone(ctx, zone, job); err != nil {
		return p.failApplyZoneSeedJob(ctx, job, fmt.Errorf("failed to seed resources: %w", err))
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
		if zone != nil {
			if err := p.dbClient.Zone().AddPointOfInterestToZone(ctx, zone.ID, existing.ID); err != nil {
				return nil, err
			}
		}
		return existing, nil
	}

	poi, err := p.locationSeeder.ImportPlace(ctx, placeID, *zone, nil)
	if err != nil {
		return nil, err
	}
	p.enqueueThumbnailTask(poi.ID, poi.ImageUrl)
	return poi, nil
}

func (p *ApplyZoneSeedDraftProcessor) enqueueThumbnailTask(poiID uuid.UUID, imageURL string) {
	if p.asyncClient == nil || strings.TrimSpace(imageURL) == "" {
		return
	}
	payload := jobs.GenerateImageThumbnailTaskPayload{
		EntityType: jobs.ThumbnailEntityPointOfInterest,
		EntityID:   poiID,
		SourceUrl:  imageURL,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal thumbnail task payload: %v", err)
		return
	}
	if _, err := p.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateImageThumbnailTaskType, payloadBytes)); err != nil {
		log.Printf("Failed to enqueue thumbnail task: %v", err)
	}
}

type zoneSeedScenarioLocation struct {
	Latitude  float64
	Longitude float64
}

func zoneSeedScenarioLocations(pois []models.ZoneSeedPointOfInterestDraft) []zoneSeedScenarioLocation {
	locations := make([]zoneSeedScenarioLocation, 0, len(pois))
	for _, poi := range pois {
		if poi.Latitude < -90 || poi.Latitude > 90 || poi.Longitude < -180 || poi.Longitude > 180 {
			continue
		}
		locations = append(locations, zoneSeedScenarioLocation{
			Latitude:  poi.Latitude,
			Longitude: poi.Longitude,
		})
	}
	return locations
}

func (p *ApplyZoneSeedDraftProcessor) seedMonsterEncountersForZone(
	ctx context.Context,
	zone *models.Zone,
	job *models.ZoneSeedJob,
) error {
	totalEncounterCount := job.MonsterCount + job.BossEncounterCount + job.RaidEncounterCount
	if totalEncounterCount <= 0 {
		return nil
	}

	templates, err := p.dbClient.MonsterTemplate().FindAllActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to load monster templates: %w", err)
	}
	if len(templates) == 0 {
		return fmt.Errorf("no monster templates available")
	}

	locations := zoneSeedScenarioLocations(job.Draft.PointsOfInterest)
	type zoneSeedEncounterBatch struct {
		count         int
		encounterType models.MonsterEncounterType
	}
	batches := []zoneSeedEncounterBatch{
		{count: job.MonsterCount, encounterType: models.MonsterEncounterTypeMonster},
		{count: job.BossEncounterCount, encounterType: models.MonsterEncounterTypeBoss},
		{count: job.RaidEncounterCount, encounterType: models.MonsterEncounterTypeRaid},
	}
	encounterOrdinal := 0
	for _, batch := range batches {
		templatePool := preferredZoneSeedTemplatesForEncounterType(templates, batch.encounterType)
		for i := 0; i < batch.count; i++ {
			encounterOrdinal++
			if err := p.createZoneSeedEncounter(ctx, zone, locations, templatePool, encounterOrdinal, totalEncounterCount, batch.encounterType); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *ApplyZoneSeedDraftProcessor) createZoneSeedEncounter(
	ctx context.Context,
	zone *models.Zone,
	locations []zoneSeedScenarioLocation,
	templates []models.MonsterTemplate,
	encounterOrdinal int,
	totalEncounterCount int,
	encounterType models.MonsterEncounterType,
) error {
	location := p.randomLocationForZone(zone, locations)
	leadTemplate := templates[rand.Intn(len(templates))]
	memberCount := zoneSeedEncounterMemberCount(encounterType)
	now := time.Now()
	recurringID, recurrenceFrequency, nextRecurrenceAt, err := zoneSeedDefaultRecurrence(now)
	if err != nil {
		return fmt.Errorf("failed to calculate monster encounter recurrence: %w", err)
	}

	encounterName := zoneSeedEncounterName(strings.TrimSpace(leadTemplate.Name), memberCount, encounterType)
	encounterDescription := zoneSeedEncounterDescription(strings.TrimSpace(leadTemplate.Description), memberCount, encounterType)
	encounterImageURL := strings.TrimSpace(leadTemplate.ImageURL)
	encounterThumbnailURL := strings.TrimSpace(leadTemplate.ThumbnailURL)
	if encounterThumbnailURL == "" {
		encounterThumbnailURL = encounterImageURL
	}

	encounter := &models.MonsterEncounter{
		ID:                          uuid.New(),
		CreatedAt:                   now,
		UpdatedAt:                   now,
		Name:                        encounterName,
		Description:                 encounterDescription,
		ImageURL:                    encounterImageURL,
		ThumbnailURL:                encounterThumbnailURL,
		EncounterType:               encounterType,
		ScaleWithUserLevel:          true,
		RecurringMonsterEncounterID: &recurringID,
		RecurrenceFrequency:         &recurrenceFrequency,
		NextRecurrenceAt:            &nextRecurrenceAt,
		ZoneID:                      zone.ID,
		Latitude:                    location.Latitude,
		Longitude:                   location.Longitude,
	}
	if err := p.dbClient.MonsterEncounter().Create(ctx, encounter); err != nil {
		return fmt.Errorf(
			"failed to create %s encounter %d/%d: %w",
			string(encounterType),
			encounterOrdinal,
			totalEncounterCount,
			err,
		)
	}

	members := make([]models.MonsterEncounterMember, 0, memberCount)
	for slot := 0; slot < memberCount; slot++ {
		template := leadTemplate
		if slot > 0 {
			template = templates[rand.Intn(len(templates))]
		}
		templateID := template.ID

		imageURL := strings.TrimSpace(template.ImageURL)
		thumbnailURL := strings.TrimSpace(template.ThumbnailURL)
		if thumbnailURL == "" {
			thumbnailURL = imageURL
		}

		monster := &models.Monster{
			ID:               uuid.New(),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			Name:             strings.TrimSpace(template.Name),
			Description:      strings.TrimSpace(template.Description),
			ImageURL:         imageURL,
			ThumbnailURL:     thumbnailURL,
			ZoneID:           zone.ID,
			Latitude:         location.Latitude,
			Longitude:        location.Longitude,
			TemplateID:       &templateID,
			Level:            5 + rand.Intn(26),
			RewardMode:       models.RewardModeRandom,
			RandomRewardSize: models.RandomRewardSizeSmall,
		}
		if imageURL != "" {
			monster.ImageGenerationStatus = models.MonsterImageGenerationStatusComplete
			emptyError := ""
			monster.ImageGenerationError = &emptyError
		} else {
			monster.ImageGenerationStatus = models.MonsterImageGenerationStatusNone
		}

		if err := p.dbClient.Monster().Create(ctx, monster); err != nil {
			return fmt.Errorf(
				"failed to create %s encounter member %d/%d: %w",
				string(encounterType),
				slot+1,
				memberCount,
				err,
			)
		}
		members = append(members, models.MonsterEncounterMember{
			MonsterID: monster.ID,
			Slot:      slot + 1,
		})
	}

	if err := p.dbClient.MonsterEncounter().ReplaceMembers(ctx, encounter.ID, members); err != nil {
		return fmt.Errorf("failed to attach %s encounter members: %w", string(encounterType), err)
	}

	return nil
}

func preferredZoneSeedTemplatesForEncounterType(
	templates []models.MonsterTemplate,
	encounterType models.MonsterEncounterType,
) []models.MonsterTemplate {
	preferredTypes := []models.MonsterTemplateType{models.MonsterTemplateTypeMonster}
	switch encounterType {
	case models.MonsterEncounterTypeBoss:
		preferredTypes = []models.MonsterTemplateType{
			models.MonsterTemplateTypeBoss,
			models.MonsterTemplateTypeMonster,
			models.MonsterTemplateTypeRaid,
		}
	case models.MonsterEncounterTypeRaid:
		preferredTypes = []models.MonsterTemplateType{
			models.MonsterTemplateTypeRaid,
			models.MonsterTemplateTypeBoss,
			models.MonsterTemplateTypeMonster,
		}
	}

	for _, preferredType := range preferredTypes {
		matches := make([]models.MonsterTemplate, 0, len(templates))
		for _, template := range templates {
			if models.NormalizeMonsterTemplateType(string(template.MonsterType)) != preferredType {
				continue
			}
			matches = append(matches, template)
		}
		if len(matches) > 0 {
			return matches
		}
	}

	return templates
}

func (p *ApplyZoneSeedDraftProcessor) enqueueScenarioGenerationJobs(
	ctx context.Context,
	zone *models.Zone,
	zoneID uuid.UUID,
	openEnded bool,
	count int,
	fallbackLocations []zoneSeedScenarioLocation,
	scaleWithUserLevel bool,
) (queued int, failed int) {
	if count <= 0 {
		return 0, 0
	}

	for i := 0; i < count; i++ {
		var latitude *float64
		var longitude *float64
		loc := p.randomLocationForZone(zone, fallbackLocations)
		lat := loc.Latitude
		lng := loc.Longitude
		latitude = &lat
		longitude = &lng
		now := time.Now()
		recurringID, recurrenceFrequency, nextRecurrenceAt, err := zoneSeedDefaultRecurrence(now)
		if err != nil {
			log.Printf("Failed to calculate scenario recurrence (openEnded=%t): %v", openEnded, err)
			failed++
			continue
		}

		job := &models.ScenarioGenerationJob{
			ID:                  uuid.New(),
			CreatedAt:           now,
			UpdatedAt:           now,
			ZoneID:              zoneID,
			Status:              models.ScenarioGenerationStatusQueued,
			OpenEnded:           openEnded,
			ScaleWithUserLevel:  scaleWithUserLevel,
			Latitude:            latitude,
			Longitude:           longitude,
			RecurringScenarioID: &recurringID,
			RecurrenceFrequency: &recurrenceFrequency,
			NextRecurrenceAt:    &nextRecurrenceAt,
		}
		if err := p.dbClient.ScenarioGenerationJob().Create(ctx, job); err != nil {
			log.Printf("Failed to create scenario generation job (openEnded=%t): %v", openEnded, err)
			failed++
			continue
		}

		payload, err := json.Marshal(jobs.GenerateScenarioTaskPayload{JobID: job.ID})
		if err != nil {
			errMsg := err.Error()
			job.Status = models.ScenarioGenerationStatusFailed
			job.ErrorMessage = &errMsg
			job.UpdatedAt = time.Now()
			_ = p.dbClient.ScenarioGenerationJob().Update(ctx, job)
			log.Printf("Failed to marshal scenario generation payload for job %s: %v", job.ID, err)
			failed++
			continue
		}

		if p.asyncClient == nil {
			errMsg := "async client unavailable"
			job.Status = models.ScenarioGenerationStatusFailed
			job.ErrorMessage = &errMsg
			job.UpdatedAt = time.Now()
			_ = p.dbClient.ScenarioGenerationJob().Update(ctx, job)
			log.Printf("Failed to enqueue scenario generation task for job %s: %s", job.ID, errMsg)
			failed++
			continue
		}

		if _, err := p.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateScenarioTaskType, payload)); err != nil {
			errMsg := err.Error()
			job.Status = models.ScenarioGenerationStatusFailed
			job.ErrorMessage = &errMsg
			job.UpdatedAt = time.Now()
			_ = p.dbClient.ScenarioGenerationJob().Update(ctx, job)
			log.Printf("Failed to enqueue scenario generation task for job %s: %v", job.ID, err)
			failed++
			continue
		}

		queued++
	}

	return queued, failed
}

func (p *ApplyZoneSeedDraftProcessor) seedStandaloneChallengesForPOIs(
	ctx context.Context,
	zone *models.Zone,
	job *models.ZoneSeedJob,
	characterByDraftID map[uuid.UUID]*models.Character,
) error {
	if len(job.Draft.PointsOfInterest) == 0 {
		return nil
	}

	narrator := pickZoneSeedNarrator(characterByDraftID)
	for _, draftPOI := range job.Draft.PointsOfInterest {
		poi, err := p.ensurePointOfInterest(ctx, zone, draftPOI.PlaceID)
		if err != nil {
			return err
		}
		if poi == nil {
			continue
		}

		lat, lng, err := challengeCoordinatesFromPOI(poi, draftPOI)
		if err != nil {
			return err
		}

		questDraft := models.ZoneSeedQuestDraft{
			Name:        fmt.Sprintf("Field Challenge at %s", strings.TrimSpace(poi.Name)),
			Description: strings.TrimSpace(zone.Description),
		}
		question, difficulty := p.generateQuestChallenge(ctx, zone, poi, &draftPOI, narrator, questDraft)
		question, submissionType := normalizeAppliedChallengeQuestion(question, poi, &draftPOI)
		description := p.generateStandalonePOIChallengeDescription(ctx, zone, poi, &draftPOI, question)
		questDraft.ChallengeQuestion = question
		statTags := p.classifyQuestStatTags(ctx, questDraft)
		if statTags == nil {
			statTags = models.StringArray{}
		}

		now := time.Now()
		recurringID, recurrenceFrequency, nextRecurrenceAt, err := zoneSeedDefaultRecurrence(now)
		if err != nil {
			return fmt.Errorf("failed to calculate challenge recurrence: %w", err)
		}
		challenge := &models.Challenge{
			ID:                   uuid.New(),
			CreatedAt:            now,
			UpdatedAt:            now,
			ZoneID:               zone.ID,
			PointOfInterestID:    &poi.ID,
			Latitude:             lat,
			Longitude:            lng,
			Question:             strings.TrimSpace(question),
			Description:          strings.TrimSpace(description),
			SubmissionType:       submissionType,
			Difficulty:           clampQuestDifficulty(difficulty),
			ScaleWithUserLevel:   true,
			RecurringChallengeID: &recurringID,
			RecurrenceFrequency:  &recurrenceFrequency,
			NextRecurrenceAt:     &nextRecurrenceAt,
			RewardMode:           models.RewardModeRandom,
			RandomRewardSize:     models.RandomRewardSizeSmall,
			StatTags:             statTags,
		}
		if err := p.dbClient.Challenge().Create(ctx, challenge); err != nil {
			return err
		}
		p.enqueueChallengeImageTask(challenge.ID)
	}

	return nil
}

func (p *ApplyZoneSeedDraftProcessor) generateStandalonePOIChallengeDescription(
	ctx context.Context,
	zone *models.Zone,
	poi *models.PointOfInterest,
	poiDraft *models.ZoneSeedPointOfInterestDraft,
	question string,
) string {
	zoneName := ""
	zoneDescription := ""
	if zone != nil {
		zoneName = strings.TrimSpace(zone.Name)
		zoneDescription = strings.TrimSpace(zone.Description)
	}
	poiDetails := formatZoneSeedPOIForPrompt(poi, poiDraft)
	prompt := fmt.Sprintf(
		`You are writing fantasy flavor text for a location-based RPG challenge.
Keep the real-world action intact, but dress it with atmospheric fantasy style.

Zone: %s
Zone description: %s
Point of Interest:
%s

Challenge question:
%s

Write 2-4 sentences of evocative flavor text (40-120 words) suitable for challenge image generation.
No markdown. No bullet points. Output JSON only:
{"description":"string"}`,
		truncate(zoneName, 120),
		truncate(zoneDescription, 400),
		poiDetails,
		truncate(strings.TrimSpace(question), 320),
	)

	answer, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return fallbackStandalonePOIChallengeDescription(zoneName, poi, poiDraft, question)
	}
	var parsed struct {
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(extractJSON(answer.Answer)), &parsed); err != nil {
		return fallbackStandalonePOIChallengeDescription(zoneName, poi, poiDraft, question)
	}
	description := strings.TrimSpace(parsed.Description)
	if description == "" {
		return fallbackStandalonePOIChallengeDescription(zoneName, poi, poiDraft, question)
	}
	return description
}

func (p *ApplyZoneSeedDraftProcessor) enqueueChallengeImageTask(challengeID uuid.UUID) {
	if p.asyncClient == nil {
		return
	}
	payload, err := json.Marshal(jobs.GenerateChallengeImageTaskPayload{ChallengeID: challengeID})
	if err != nil {
		log.Printf("Failed to marshal challenge image payload: %v", err)
		return
	}
	if _, err := p.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateChallengeImageTaskType, payload)); err != nil {
		log.Printf("Failed to enqueue challenge image task: %v", err)
	}
}

func (p *ApplyZoneSeedDraftProcessor) seedTreasureChestsForZone(
	ctx context.Context,
	zone *models.Zone,
	job *models.ZoneSeedJob,
) error {
	chestCount := job.TreasureChestCount
	if chestCount <= 0 {
		return nil
	}

	fallbackLocations := zoneSeedScenarioLocations(job.Draft.PointsOfInterest)
	for i := 0; i < chestCount; i++ {
		location := p.randomLocationForZone(zone, fallbackLocations)
		targetLevel, size := zoneSeedTreasureChestRewardProfile()
		unlockTier := targetLevel
		chest := &models.TreasureChest{
			Latitude:         location.Latitude,
			Longitude:        location.Longitude,
			ZoneID:           zone.ID,
			RewardMode:       models.RewardModeRandom,
			RandomRewardSize: size,
			RewardExperience: 0,
			Gold:             nil,
			UnlockTier:       &unlockTier,
			Invalidated:      false,
		}
		if err := p.dbClient.TreasureChest().Create(ctx, chest); err != nil {
			return fmt.Errorf("failed to create treasure chest %d/%d: %w", i+1, chestCount, err)
		}
	}

	return nil
}

func (p *ApplyZoneSeedDraftProcessor) seedHealingFountainsForZone(
	ctx context.Context,
	zone *models.Zone,
	job *models.ZoneSeedJob,
) error {
	fountainCount := job.HealingFountainCount
	if fountainCount <= 0 {
		return nil
	}

	fallbackLocations := zoneSeedScenarioLocations(job.Draft.PointsOfInterest)
	for i := 0; i < fountainCount; i++ {
		location := p.randomLocationForZone(zone, fallbackLocations)
		fountain := &models.HealingFountain{
			Name:         zoneSeedHealingFountainDefaultName,
			Description:  zoneSeedHealingFountainDefaultDescription,
			ThumbnailURL: zoneSeedHealingFountainDefaultThumbnail,
			ZoneID:       zone.ID,
			Latitude:     location.Latitude,
			Longitude:    location.Longitude,
			Invalidated:  false,
		}
		if err := p.dbClient.HealingFountain().Create(ctx, fountain); err != nil {
			return fmt.Errorf("failed to create healing fountain %d/%d: %w", i+1, fountainCount, err)
		}
	}

	return nil
}

type zoneSeedResourcePool struct {
	resourceType   models.ResourceType
	inventoryItems []models.InventoryItem
}

const (
	zoneSeedHerbalismResourceTypeSlug = "herbalism"
	zoneSeedMiningResourceTypeSlug    = "mining"
)

func zoneSeedNormalizeResourceTypeSlug(resourceType models.ResourceType) string {
	slug := strings.TrimSpace(resourceType.Slug)
	if slug == "" {
		slug = strings.TrimSpace(resourceType.Name)
	}
	return models.NormalizeZoneKind(slug)
}

func zoneSeedBuildResourcePools(
	resourceTypes []models.ResourceType,
	inventoryItems []models.InventoryItem,
) []zoneSeedResourcePool {
	resourceTypeByID := make(map[uuid.UUID]models.ResourceType, len(resourceTypes))
	for _, resourceType := range resourceTypes {
		if resourceType.ID == uuid.Nil {
			continue
		}
		resourceTypeByID[resourceType.ID] = resourceType
	}

	itemsByTypeID := make(map[uuid.UUID][]models.InventoryItem)
	for _, item := range inventoryItems {
		if item.ResourceTypeID == nil || *item.ResourceTypeID == uuid.Nil {
			continue
		}
		if _, ok := resourceTypeByID[*item.ResourceTypeID]; !ok {
			continue
		}
		itemsByTypeID[*item.ResourceTypeID] = append(itemsByTypeID[*item.ResourceTypeID], item)
	}

	pools := make([]zoneSeedResourcePool, 0, len(resourceTypes))
	for _, resourceType := range resourceTypes {
		items := itemsByTypeID[resourceType.ID]
		if len(items) == 0 {
			continue
		}
		pools = append(pools, zoneSeedResourcePool{
			resourceType:   resourceType,
			inventoryItems: items,
		})
	}
	return pools
}

func zoneSeedBuildResourcePoolsBySlug(
	resourceTypes []models.ResourceType,
	inventoryItems []models.InventoryItem,
) map[string][]zoneSeedResourcePool {
	poolsBySlug := map[string][]zoneSeedResourcePool{}
	for _, pool := range zoneSeedBuildResourcePools(resourceTypes, inventoryItems) {
		slug := zoneSeedNormalizeResourceTypeSlug(pool.resourceType)
		if slug == "" {
			continue
		}
		poolsBySlug[slug] = append(poolsBySlug[slug], pool)
	}
	return poolsBySlug
}

func (p *ApplyZoneSeedDraftProcessor) seedResourceNodesForZone(
	ctx context.Context,
	zone *models.Zone,
	resourcePools []zoneSeedResourcePool,
	resourceCount int,
	resourceLabel string,
	fallbackLocations []zoneSeedScenarioLocation,
) error {
	if resourceCount <= 0 {
		return nil
	}
	if len(resourcePools) == 0 {
		return fmt.Errorf("no eligible %s resource types with active inventory items are available", resourceLabel)
	}

	poolOrder := []int{}
	for i := 0; i < resourceCount; i++ {
		if len(poolOrder) == 0 {
			poolOrder = rand.Perm(len(resourcePools))
		}
		poolIndex := poolOrder[0]
		poolOrder = poolOrder[1:]
		pool := resourcePools[poolIndex]
		location := p.randomLocationForZone(zone, fallbackLocations)

		resource := &models.Resource{
			ZoneID:         zone.ID,
			ResourceTypeID: pool.resourceType.ID,
			Quantity:       1,
			Latitude:       location.Latitude,
			Longitude:      location.Longitude,
			Invalidated:    false,
		}
		if err := p.dbClient.Resource().Create(ctx, resource); err != nil {
			return fmt.Errorf(
				"failed to create %s resource %d/%d for type %s: %w",
				resourceLabel,
				i+1,
				resourceCount,
				strings.TrimSpace(pool.resourceType.Name),
				err,
			)
		}
	}

	return nil
}

func (p *ApplyZoneSeedDraftProcessor) seedResourcesForZone(
	ctx context.Context,
	zone *models.Zone,
	job *models.ZoneSeedJob,
) error {
	herbalismResourceCount := job.EffectiveHerbalismResourceCount()
	miningResourceCount := job.EffectiveMiningResourceCount()
	if herbalismResourceCount <= 0 && miningResourceCount <= 0 {
		return nil
	}

	resourceTypes, err := p.dbClient.ResourceType().FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load resource types: %w", err)
	}
	activeInventoryItems, err := p.dbClient.InventoryItem().FindAllActiveInventoryItems(ctx)
	if err != nil {
		return fmt.Errorf("failed to load active inventory items: %w", err)
	}

	fallbackLocations := zoneSeedScenarioLocations(job.Draft.PointsOfInterest)
	resourcePoolsBySlug := zoneSeedBuildResourcePoolsBySlug(resourceTypes, activeInventoryItems)
	if err := p.seedResourceNodesForZone(
		ctx,
		zone,
		resourcePoolsBySlug[zoneSeedHerbalismResourceTypeSlug],
		herbalismResourceCount,
		zoneSeedHerbalismResourceTypeSlug,
		fallbackLocations,
	); err != nil {
		return err
	}
	if err := p.seedResourceNodesForZone(
		ctx,
		zone,
		resourcePoolsBySlug[zoneSeedMiningResourceTypeSlug],
		miningResourceCount,
		zoneSeedMiningResourceTypeSlug,
		fallbackLocations,
	); err != nil {
		return err
	}

	return nil
}

func (p *ApplyZoneSeedDraftProcessor) randomLocationForZone(
	zone *models.Zone,
	fallbackLocations []zoneSeedScenarioLocation,
) zoneSeedScenarioLocation {
	if zone != nil {
		point := zone.GetRandomPoint()
		lat := point.Y()
		lng := point.X()
		if lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180 {
			// Zone.GetRandomPoint can return (0,0) when zone geometry is missing.
			// Prefer fallback/zone center unless the zone is actually centered at (0,0).
			if (lat != 0 || lng != 0) || (zone.Latitude == 0 && zone.Longitude == 0) {
				return zoneSeedScenarioLocation{Latitude: lat, Longitude: lng}
			}
		}
	}
	if len(fallbackLocations) > 0 {
		return fallbackLocations[rand.Intn(len(fallbackLocations))]
	}
	if zone != nil {
		if zone.Latitude >= -90 && zone.Latitude <= 90 && zone.Longitude >= -180 && zone.Longitude <= 180 {
			return zoneSeedScenarioLocation{Latitude: zone.Latitude, Longitude: zone.Longitude}
		}
	}
	return zoneSeedScenarioLocation{}
}

func pickZoneSeedNarrator(characterByDraftID map[uuid.UUID]*models.Character) *models.Character {
	for _, character := range characterByDraftID {
		if character != nil {
			return character
		}
	}
	return &models.Character{
		Name:        "Guild Chronicler",
		Description: "A field chronicler who frames local tasks in mythic language.",
	}
}

func challengeCoordinatesFromPOI(
	poi *models.PointOfInterest,
	draftPOI models.ZoneSeedPointOfInterestDraft,
) (float64, float64, error) {
	if poi != nil {
		lat, latErr := strconv.ParseFloat(strings.TrimSpace(poi.Lat), 64)
		lng, lngErr := strconv.ParseFloat(strings.TrimSpace(poi.Lng), 64)
		if latErr == nil && lngErr == nil {
			return lat, lng, nil
		}
	}
	if draftPOI.Latitude < -90 || draftPOI.Latitude > 90 || draftPOI.Longitude < -180 || draftPOI.Longitude > 180 {
		return 0, 0, fmt.Errorf("invalid point of interest coordinates for challenge")
	}
	return draftPOI.Latitude, draftPOI.Longitude, nil
}

func fallbackStandalonePOIChallengeDescription(
	zoneName string,
	poi *models.PointOfInterest,
	poiDraft *models.ZoneSeedPointOfInterestDraft,
	question string,
) string {
	name := poiNameForPrompt(poi, poiDraft)
	if strings.TrimSpace(zoneName) == "" {
		return fmt.Sprintf("Rumors coil around %s like drifting embers. Step into the scene, play your role, and prove it by completing this action: %s", name, strings.TrimSpace(question))
	}
	return fmt.Sprintf("In %s, %s feels like a threshold between the ordinary and the uncanny. Play your part in the unfolding tale, then prove it by completing this action: %s", strings.TrimSpace(zoneName), name, strings.TrimSpace(question))
}

func zoneSeedEncounterName(baseName string, memberCount int, encounterType models.MonsterEncounterType) string {
	trimmed := strings.TrimSpace(baseName)
	if trimmed == "" {
		trimmed = "Host"
	}
	switch encounterType {
	case models.MonsterEncounterTypeBoss:
		return fmt.Sprintf("%s Boss", trimmed)
	case models.MonsterEncounterTypeRaid:
		return fmt.Sprintf("%s Raid", trimmed)
	}
	if memberCount <= 1 {
		return fmt.Sprintf("Wandering %s", trimmed)
	}
	return fmt.Sprintf("%s Warband", trimmed)
}

func zoneSeedEncounterDescription(baseDescription string, memberCount int, encounterType models.MonsterEncounterType) string {
	trimmed := strings.TrimSpace(baseDescription)
	switch encounterType {
	case models.MonsterEncounterTypeBoss:
		if trimmed == "" {
			return "An elite hostile presence dominates this part of the zone."
		}
		return fmt.Sprintf("%s This elite encounter is tuned above the player level curve.", trimmed)
	case models.MonsterEncounterTypeRaid:
		if trimmed == "" {
			return "A large-scale hostile force has gathered here, demanding a full party response."
		}
		return fmt.Sprintf("%s This raid encounter is tuned for a full party of adventurers.", trimmed)
	}
	if trimmed == "" {
		if memberCount <= 1 {
			return "A lone hostile presence prowls this part of the zone."
		}
		return "A coordinated pack of hostiles has been spotted in this part of the zone."
	}
	if memberCount <= 1 {
		return trimmed
	}
	return fmt.Sprintf("%s This encounter has multiple hostile monsters acting together.", trimmed)
}

func zoneSeedEncounterMemberCount(encounterType models.MonsterEncounterType) int {
	switch encounterType {
	case models.MonsterEncounterTypeBoss:
		return 1
	case models.MonsterEncounterTypeRaid:
		return 5
	}
	roll := rand.Intn(100)
	switch {
	case roll < 45:
		return 1
	case roll < 90:
		return 2
	default:
		// Rare 3-monster encounter.
		return 3
	}
}

func zoneSeedTreasureChestRewardProfile() (int, models.RandomRewardSize) {
	roll := rand.Intn(100)
	switch {
	case roll < 45:
		return 10, models.RandomRewardSizeSmall
	case roll < 78:
		return 25, models.RandomRewardSizeSmall
	case roll < 95:
		return 50, models.RandomRewardSizeMedium
	default:
		return 70, models.RandomRewardSizeLarge
	}
}

func (p *ApplyZoneSeedDraftProcessor) createCharacterFromDraft(
	ctx context.Context,
	zone *models.Zone,
	poi *models.PointOfInterest,
	draft models.ZoneSeedCharacterDraft,
) (*models.Character, error) {
	startingLat := 0.0
	startingLng := 0.0
	hasStartingLocation := false
	switch {
	case draft.Latitude != nil && draft.Longitude != nil:
		startingLat = *draft.Latitude
		startingLng = *draft.Longitude
		hasStartingLocation = true
	case poi != nil:
		lat, latErr := strconv.ParseFloat(strings.TrimSpace(poi.Lat), 64)
		lng, lngErr := strconv.ParseFloat(strings.TrimSpace(poi.Lng), 64)
		if latErr == nil && lngErr == nil {
			startingLat = lat
			startingLng = lng
			hasStartingLocation = true
		}
	}
	if !hasStartingLocation || startingLat < -90 || startingLat > 90 || startingLng < -180 || startingLng > 180 {
		fallback := p.randomLocationForZone(zone, nil)
		startingLat = fallback.Latitude
		startingLng = fallback.Longitude
	}

	character := &models.Character{
		Name:                  draft.Name,
		Description:           draft.Description,
		PointOfInterestID:     nil,
		InternalTags:          zoneSeedCharacterInternalTags(draft),
		ImageGenerationStatus: models.CharacterImageGenerationStatusQueued,
	}
	if poi != nil {
		character.PointOfInterestID = &poi.ID
	}

	if err := p.dbClient.Character().Create(ctx, character); err != nil {
		return nil, err
	}
	if poi == nil {
		if err := p.dbClient.CharacterLocation().ReplaceForCharacter(ctx, character.ID, []models.CharacterLocation{
			{
				Latitude:  startingLat,
				Longitude: startingLng,
			},
		}); err != nil {
			return nil, err
		}
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

func zoneSeedCharacterInternalTags(draft models.ZoneSeedCharacterDraft) models.StringArray {
	if strings.TrimSpace(draft.PlaceID) == "" {
		return models.StringArray{}
	}
	if len(normalizeShopActionTags(draft.ShopItemTags)) > 0 {
		return models.StringArray{}
	}
	return models.StringArray{models.CharacterInternalTagGeneratedPOILocal}
}

func normalizeShopActionTags(tags models.StringArray) []string {
	normalized := make([]string, 0, len(tags))
	seen := map[string]struct{}{}
	for _, rawTag := range []string(tags) {
		tag := strings.ToLower(strings.TrimSpace(rawTag))
		if tag == "" {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}
	sort.Strings(normalized)
	return normalized
}

func (p *ApplyZoneSeedDraftProcessor) ensureShopActionForCharacterTags(
	ctx context.Context,
	character *models.Character,
	tags models.StringArray,
) error {
	if character == nil {
		return fmt.Errorf("character is nil")
	}

	normalizedTags := normalizeShopActionTags(tags)
	if len(normalizedTags) == 0 {
		return nil
	}

	actions, err := p.dbClient.CharacterAction().FindByCharacterID(ctx, character.ID)
	if err != nil {
		return err
	}
	for _, action := range actions {
		if action == nil || action.ActionType != models.ActionTypeShop || action.Metadata == nil {
			continue
		}
		mode := strings.ToLower(strings.TrimSpace(fmt.Sprint(action.Metadata["shopMode"])))
		if mode != "tags" {
			continue
		}
		rawTags, ok := action.Metadata["shopItemTags"].([]interface{})
		if !ok {
			continue
		}
		existing := make([]string, 0, len(rawTags))
		for _, raw := range rawTags {
			tag := strings.ToLower(strings.TrimSpace(fmt.Sprint(raw)))
			if tag != "" {
				existing = append(existing, tag)
			}
		}
		sort.Strings(existing)
		if len(existing) != len(normalizedTags) {
			continue
		}
		matched := true
		for idx := range existing {
			if existing[idx] != normalizedTags[idx] {
				matched = false
				break
			}
		}
		if matched {
			return nil
		}
	}

	shopTags := make([]interface{}, 0, len(normalizedTags))
	for _, tag := range normalizedTags {
		shopTags = append(shopTags, tag)
	}
	action := &models.CharacterAction{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CharacterID: character.ID,
		ActionType:  models.ActionTypeShop,
		Dialogue:    p.generateQuestGiverAmbientDialogue(ctx, character),
		Metadata: map[string]interface{}{
			"shopMode":     "tags",
			"shopItemTags": shopTags,
			"inventory":    []interface{}{},
		},
	}
	return p.dbClient.CharacterAction().Create(ctx, action)
}

func (p *ApplyZoneSeedDraftProcessor) ensureTalkActionForCharacterDialogue(
	ctx context.Context,
	character *models.Character,
	dialogueLines []string,
	poi *models.PointOfInterest,
) error {
	if character == nil {
		return fmt.Errorf("character is nil")
	}

	lines := sanitizeZoneSeedTalkDialogueLines(dialogueLines)
	if len(lines) == 0 {
		return nil
	}

	actions, err := p.dbClient.CharacterAction().FindByCharacterID(ctx, character.ID)
	if err != nil {
		return err
	}
	for _, action := range actions {
		if action == nil || action.ActionType != models.ActionTypeTalk || action.Metadata == nil {
			continue
		}
		if strings.ToLower(strings.TrimSpace(fmt.Sprint(action.Metadata["source"]))) == "poilocal" {
			return nil
		}
	}

	metadata := map[string]interface{}{"source": "poiLocal"}
	if poi != nil {
		metadata["pointOfInterestId"] = poi.ID.String()
	}
	return p.dbClient.CharacterAction().Create(ctx, &models.CharacterAction{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CharacterID: character.ID,
		ActionType:  models.ActionTypeTalk,
		Dialogue:    models.DialogueSequenceFromStringLines(lines),
		Metadata:    metadata,
	})
}

func sanitizeZoneSeedTalkDialogueLines(lines []string) []string {
	seen := map[string]struct{}{}
	sanitized := make([]string, 0, 3)
	for _, raw := range lines {
		line := strings.TrimSpace(strings.ReplaceAll(raw, "\n", " "))
		if line == "" {
			continue
		}
		if len(line) > 180 {
			line = strings.TrimSpace(line[:180])
		}
		key := strings.ToLower(line)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		sanitized = append(sanitized, line)
		if len(sanitized) >= 3 {
			break
		}
	}
	return sanitized
}

func (p *ApplyZoneSeedDraftProcessor) createQuestFromDraft(
	ctx context.Context,
	zone *models.Zone,
	poi *models.PointOfInterest,
	poiDraft *models.ZoneSeedPointOfInterestDraft,
	character *models.Character,
	draft models.ZoneSeedQuestDraft,
) error {
	acceptanceDialogue := models.DialogueSequenceFromStringLines(draft.AcceptanceDialogue)

	challengeQuestion := strings.TrimSpace(draft.ChallengeQuestion)
	challengeDifficulty := draft.ChallengeDifficulty
	if challengeQuestion == "" || challengeDifficulty <= 0 {
		challengeQuestion, challengeDifficulty = p.generateQuestChallenge(ctx, zone, poi, poiDraft, character, draft)
	} else {
		challengeDifficulty = clampQuestDifficulty(challengeDifficulty)
	}
	challengeQuestion, submissionType := normalizeAppliedChallengeQuestion(challengeQuestion, poi, poiDraft)
	draftForTags := draft
	draftForTags.ChallengeQuestion = challengeQuestion
	statTags := p.classifyQuestStatTags(ctx, draftForTags)

	gold := draft.Gold
	if gold <= 0 {
		gold = 200 + rand.Intn(601)
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
		RewardMode:            models.RewardModeExplicit,
		RandomRewardSize:      models.RandomRewardSizeSmall,
		Gold:                  gold,
	}

	if err := p.dbClient.Quest().Create(ctx, quest); err != nil {
		return err
	}

	rewardItem, err := p.generateQuestRewardItem(ctx, zone, poi, poiDraft, character, draft, challengeQuestion, draft.RewardItem, false)
	if err != nil {
		return err
	}
	if rewardItem != nil {
		rewards := []models.QuestItemReward{
			{
				ID:              uuid.New(),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				QuestID:         quest.ID,
				InventoryItemID: rewardItem.ID,
				Quantity:        1,
			},
		}
		if err := p.dbClient.QuestItemReward().ReplaceForQuest(ctx, quest.ID, rewards); err != nil {
			return err
		}
	}

	if err := p.ensureQuestActionForCharacter(ctx, quest.ID, character); err != nil {
		log.Printf("Failed to create quest action for character: %v", err)
	}

	node := &models.QuestNode{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		QuestID:        quest.ID,
		OrderIndex:     0,
		SubmissionType: submissionType,
	}
	locationChallenge, err := p.createQuestNodeLocationChallenge(
		zone.ID,
		poi,
		challengeQuestion,
		draft.Description,
		submissionType,
		challengeDifficulty,
		statTags,
	)
	if err != nil {
		return err
	}
	if err := p.dbClient.Challenge().Create(ctx, locationChallenge); err != nil {
		return err
	}
	node.ChallengeID = &locationChallenge.ID
	if err := p.dbClient.QuestNode().Create(ctx, node); err != nil {
		return err
	}

	if poi != nil {
		if err := p.dbClient.PointOfInterest().UpdateLastUsedInQuest(ctx, poi.ID); err != nil {
			log.Printf("Failed to update POI last_used_in_quest_at: %v", err)
		}
	}

	return nil
}

func (p *ApplyZoneSeedDraftProcessor) createMainQuestFromDraft(
	ctx context.Context,
	zone *models.Zone,
	character *models.Character,
	draft models.ZoneSeedMainQuestDraft,
) error {
	acceptanceDialogue := models.DialogueSequenceFromStringLines(draft.AcceptanceDialogue)

	gold := draft.Gold
	if gold <= 0 {
		gold = 50 + rand.Intn(451)
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
		RewardMode:            models.RewardModeExplicit,
		RandomRewardSize:      models.RandomRewardSizeSmall,
		Gold:                  gold,
	}

	if err := p.dbClient.Quest().Create(ctx, quest); err != nil {
		return err
	}

	if err := p.ensureQuestActionForCharacter(ctx, quest.ID, character); err != nil {
		log.Printf("Failed to create quest action for character: %v", err)
	}

	if len(draft.Nodes) == 0 {
		return fmt.Errorf("main quest has no nodes")
	}

	nodes := append([]models.ZoneSeedMainQuestNodeDraft(nil), draft.Nodes...)
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].OrderIndex < nodes[j].OrderIndex
	})

	var rewardPOI *models.PointOfInterest
	if len(nodes) > 0 {
		lastNode := nodes[len(nodes)-1]
		poi, err := p.ensurePointOfInterest(ctx, zone, lastNode.PlaceID)
		if err != nil {
			return err
		}
		rewardPOI = poi
	}

	rewardDraft := models.ZoneSeedQuestDraft{
		Name:               draft.Name,
		Description:        draft.Description,
		AcceptanceDialogue: draft.AcceptanceDialogue,
	}
	rewardQuestion := ""
	if len(nodes) > 0 {
		rewardQuestion = nodes[len(nodes)-1].ChallengeQuestion
	}
	rewardItem, err := p.generateQuestRewardItem(ctx, zone, rewardPOI, nil, character, rewardDraft, rewardQuestion, draft.RewardItem, true)
	if err != nil {
		return err
	}
	if rewardItem != nil {
		rewards := []models.QuestItemReward{
			{
				ID:              uuid.New(),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				QuestID:         quest.ID,
				InventoryItemID: rewardItem.ID,
				Quantity:        1,
			},
		}
		if err := p.dbClient.QuestItemReward().ReplaceForQuest(ctx, quest.ID, rewards); err != nil {
			return err
		}
	}

	for idx, nodeDraft := range nodes {
		poi, err := p.ensurePointOfInterest(ctx, zone, nodeDraft.PlaceID)
		if err != nil {
			return err
		}
		challengeQuestion := strings.TrimSpace(nodeDraft.ChallengeQuestion)
		if challengeQuestion == "" {
			challengeQuestion = fallbackQuestChallengeQuestion(poi, nil)
		}
		challengeQuestion, submissionType := normalizeAppliedChallengeQuestion(challengeQuestion, poi, nil)

		node := &models.QuestNode{
			ID:             uuid.New(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			QuestID:        quest.ID,
			OrderIndex:     idx,
			SubmissionType: submissionType,
		}
		difficulty := nodeDraft.ChallengeDifficulty
		if difficulty <= 0 {
			difficulty = randomQuestDifficulty()
		}
		difficulty = clampQuestDifficulty(difficulty)
		statDraft := models.ZoneSeedQuestDraft{
			Name:               draft.Name,
			Description:        draft.Description,
			AcceptanceDialogue: draft.AcceptanceDialogue,
			ChallengeQuestion:  challengeQuestion,
		}
		statTags := p.classifyQuestStatTags(ctx, statDraft)
		locationChallenge, err := p.createQuestNodeLocationChallenge(
			zone.ID,
			poi,
			challengeQuestion,
			draft.Description,
			submissionType,
			difficulty,
			statTags,
		)
		if err != nil {
			return err
		}
		if err := p.dbClient.Challenge().Create(ctx, locationChallenge); err != nil {
			return err
		}
		node.ChallengeID = &locationChallenge.ID
		if err := p.dbClient.QuestNode().Create(ctx, node); err != nil {
			return err
		}

		if poi != nil {
			if err := p.dbClient.PointOfInterest().UpdateLastUsedInQuest(ctx, poi.ID); err != nil {
				log.Printf("Failed to update POI last_used_in_quest_at: %v", err)
			}
		}
	}

	return nil
}

func (p *ApplyZoneSeedDraftProcessor) createQuestNodeLocationChallenge(
	zoneID uuid.UUID,
	poi *models.PointOfInterest,
	question string,
	description string,
	submissionType models.QuestNodeSubmissionType,
	difficulty int,
	statTags models.StringArray,
) (*models.Challenge, error) {
	if poi == nil {
		return nil, fmt.Errorf("quest node point of interest is required")
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(poi.Lat), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid point of interest latitude: %w", err)
	}
	lng, err := strconv.ParseFloat(strings.TrimSpace(poi.Lng), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid point of interest longitude: %w", err)
	}
	challengeQuestion := strings.TrimSpace(question)
	if challengeQuestion == "" {
		challengeQuestion = fallbackQuestChallengeQuestion(poi, nil)
	}
	challengeDescription := strings.TrimSpace(description)
	if challengeDescription == "" {
		challengeDescription = strings.TrimSpace(poi.Description)
	}
	now := time.Now()
	if statTags == nil {
		statTags = models.StringArray{}
	}
	return &models.Challenge{
		ID:             uuid.New(),
		CreatedAt:      now,
		UpdatedAt:      now,
		ZoneID:         zoneID,
		Latitude:       lat,
		Longitude:      lng,
		Question:       challengeQuestion,
		Description:    challengeDescription,
		SubmissionType: submissionType,
		Reward:         0,
		Difficulty:     difficulty,
		StatTags:       statTags,
	}, nil
}

func (p *ApplyZoneSeedDraftProcessor) ensureQuestActionForCharacter(ctx context.Context, questID uuid.UUID, character *models.Character) error {
	if character == nil {
		return fmt.Errorf("character is nil")
	}

	actions, err := p.dbClient.CharacterAction().FindByCharacterID(ctx, character.ID)
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
		CharacterID: character.ID,
		ActionType:  models.ActionTypeGiveQuest,
		Dialogue:    p.generateQuestGiverAmbientDialogue(ctx, character),
		Metadata:    map[string]interface{}{"questId": questIDStr},
	}
	return p.dbClient.CharacterAction().Create(ctx, action)
}

type questGiverAmbientDialogueResponse struct {
	Lines []string `json:"lines"`
}

const questGiverAmbientDialoguePromptTemplate = `
You are writing incidental in-character dialogue for an RPG quest giver.

Character name: %s
Character description: %s

Write 1-2 short lines this character might casually say in conversation.
Constraints:
- Keep lines in character.
- Do NOT mention quests, missions, tasks, objectives, rewards, or adventuring.
- Do NOT reference the specific quest content.
- Keep each line to one short sentence.

Respond ONLY as JSON:
{
  "lines": ["string"]
}
`

func (p *ApplyZoneSeedDraftProcessor) generateQuestGiverAmbientDialogue(
	ctx context.Context,
	character *models.Character,
) []models.DialogueMessage {
	lines := []string{}
	if character != nil {
		prompt := fmt.Sprintf(
			questGiverAmbientDialoguePromptTemplate,
			truncate(strings.TrimSpace(character.Name), 80),
			truncate(strings.TrimSpace(character.Description), 280),
		)

		answer, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
		if err == nil {
			var response questGiverAmbientDialogueResponse
			if parseErr := json.Unmarshal([]byte(extractJSON(answer.Answer)), &response); parseErr == nil {
				lines = sanitizeQuestGiverAmbientDialogueLines(response.Lines)
			}
		}
	}

	if len(lines) == 0 {
		lines = fallbackQuestGiverAmbientDialogueLines(character)
	}
	if len(lines) == 0 {
		lines = []string{"I like this corner of the city when it gets quiet."}
	}

	dialogue := make([]models.DialogueMessage, 0, len(lines))
	for i, line := range lines {
		dialogue = append(dialogue, models.DialogueMessage{
			Speaker: "character",
			Text:    line,
			Order:   i,
		})
	}
	return dialogue
}

func sanitizeQuestGiverAmbientDialogueLines(lines []string) []string {
	blocked := []string{
		"quest", "mission", "task", "objective", "reward", "contract", "job", "adventure", "adventurer",
	}
	seen := map[string]struct{}{}
	sanitized := make([]string, 0, 2)
	for _, raw := range lines {
		line := strings.TrimSpace(strings.ReplaceAll(raw, "\n", " "))
		if line == "" {
			continue
		}
		if len(line) > 180 {
			line = strings.TrimSpace(line[:180])
		}
		lower := strings.ToLower(line)
		skip := false
		for _, token := range blocked {
			if strings.Contains(lower, token) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		if _, ok := seen[lower]; ok {
			continue
		}
		seen[lower] = struct{}{}
		sanitized = append(sanitized, line)
		if len(sanitized) >= 2 {
			break
		}
	}
	return sanitized
}

func fallbackQuestGiverAmbientDialogueLines(character *models.Character) []string {
	pool := []string{
		"I keep a little notebook of things most people walk right past.",
		"Some corners of the city sound better if you stand still for a minute.",
		"I trust good timing more than good luck.",
		"I notice shoes first. They tell you where someone has really been.",
		"Morning light is honest. Evening light is forgiving.",
	}

	if character != nil {
		desc := strings.ToLower(character.Description)
		switch {
		case strings.Contains(desc, "smith"), strings.Contains(desc, "forge"), strings.Contains(desc, "blacksmith"), strings.Contains(desc, "hammer"):
			pool = append(pool, "A clean strike and a steady breath solve most problems.")
		case strings.Contains(desc, "bard"), strings.Contains(desc, "music"), strings.Contains(desc, "singer"), strings.Contains(desc, "violin"), strings.Contains(desc, "jazz"):
			pool = append(pool, "I collect melodies the way other people collect favors.")
		case strings.Contains(desc, "alchemist"), strings.Contains(desc, "apothecary"), strings.Contains(desc, "herb"), strings.Contains(desc, "botanist"):
			pool = append(pool, "The air always changes a moment before the weather does.")
		case strings.Contains(desc, "guard"), strings.Contains(desc, "knight"), strings.Contains(desc, "watch"), strings.Contains(desc, "sentinel"):
			pool = append(pool, "I still wake before dawn, even on days with nowhere urgent to be.")
		case strings.Contains(desc, "merchant"), strings.Contains(desc, "shop"), strings.Contains(desc, "trader"), strings.Contains(desc, "vendor"):
			pool = append(pool, "I remember people by what they linger over, not what they buy.")
		case strings.Contains(desc, "scholar"), strings.Contains(desc, "scribe"), strings.Contains(desc, "librarian"), strings.Contains(desc, "professor"):
			pool = append(pool, "Half my best ideas start as arguments with my own notes.")
		case strings.Contains(desc, "cook"), strings.Contains(desc, "chef"), strings.Contains(desc, "baker"), strings.Contains(desc, "barista"), strings.Contains(desc, "cafe"):
			pool = append(pool, "Strong coffee and warm bread can rescue a rough morning.")
		case strings.Contains(desc, "artist"), strings.Contains(desc, "painter"), strings.Contains(desc, "poet"), strings.Contains(desc, "actor"):
			pool = append(pool, "I'm still chasing one perfect detail I once saw for three seconds.")
		}
	}

	count := 1 + rand.Intn(2)
	perm := rand.Perm(len(pool))
	lines := make([]string, 0, count)
	for _, idx := range perm {
		lines = append(lines, pool[idx])
		if len(lines) >= count {
			break
		}
	}
	return sanitizeQuestGiverAmbientDialogueLines(lines)
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

type questChallengeResponse struct {
	Question   string `json:"question"`
	Difficulty int    `json:"difficulty"`
}

const questChallengePromptTemplate = `
You are a quest designer creating a single-player, real-world challenge.

Ignore any fantasy flavor in the quest/zone. Base the challenge only on the real-world POI category.

Zone: %s
Quest name: %s
Quest description: %s
Quest giver: %s
Quest giver description: %s

Point of Interest:
%s

Create one challenge that can be completed by a single person while physically at the POI.
Constraints:
- Safe, legal, and respectful. Do not require entering restricted areas or interacting with staff.
- Single-input only: EITHER a photo proof OR a short text response (1-2 sentences), never both.
- Require meaningful participation in the POI's core activity (not just approaching it).
- Avoid knowledge-based or hard-to-verify prompts; prefer proof-of-participation in the activity itself.
- Do NOT use signage-only prompts (storefront sign, menu board, entrance, marquee, poster, or facade) as the main proof.
- If the POI is food/drink-focused, the challenge should involve getting a drink/food item (for example, "get a coffee"), and proof should show the selected item.
- Make it an enjoyable on-site activity that fits the POI type.
- Answerable on-site without external research.
- 1-2 short sentences.
- Provide a difficulty score from 25 to 50 (inclusive).

Respond ONLY as JSON:
{
  "question": "string",
  "difficulty": 32
}
`

func (p *ApplyZoneSeedDraftProcessor) generateQuestChallenge(
	ctx context.Context,
	zone *models.Zone,
	poi *models.PointOfInterest,
	poiDraft *models.ZoneSeedPointOfInterestDraft,
	character *models.Character,
	draft models.ZoneSeedQuestDraft,
) (string, int) {
	poiDetails := formatZoneSeedPOIForPrompt(poi, poiDraft)
	zoneName := ""
	if zone != nil {
		zoneName = zone.Name
	}
	prompt := fmt.Sprintf(
		questChallengePromptTemplate,
		truncate(zoneName, 120),
		truncate(draft.Name, 120),
		truncate(draft.Description, 400),
		truncate(character.Name, 80),
		truncate(character.Description, 200),
		poiDetails,
	)

	answer, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return fallbackQuestChallengeQuestion(poi, poiDraft), randomQuestDifficulty()
	}

	var response questChallengeResponse
	if err := json.Unmarshal([]byte(extractJSON(answer.Answer)), &response); err != nil {
		return fallbackQuestChallengeQuestion(poi, poiDraft), randomQuestDifficulty()
	}

	question := strings.TrimSpace(response.Question)
	if question == "" {
		question = fallbackQuestChallengeQuestion(poi, poiDraft)
	}

	difficulty := response.Difficulty
	if difficulty <= 0 {
		difficulty = randomQuestDifficulty()
	}
	difficulty = clampQuestDifficulty(difficulty)

	return question, difficulty
}

type questRewardItemResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	RarityTier  string `json:"rarityTier"`
}

type questRewardHandProfile struct {
	Category   string
	Handedness string
}

const questRewardItemPromptTemplate = `
You are a fantasy RPG item designer creating a quest reward.

Zone: %s
Quest name: %s
Quest description: %s
Challenge: %s
Quest giver: %s

Point of Interest:
%s

Create a reward item that fits the quest and location flavor.
Constraints:
- Tangible item a player could carry.
- Avoid real-world brand names.
- Name should be short (<= 6 words).
- Description 1-2 sentences.
- Rarity tier must be one of: Common, Uncommon, Epic, Mythic.
- Flavor intensity should match the rarity.

Respond ONLY as JSON:
{
  "name": "string",
  "description": "string",
  "rarityTier": "Common"
}
`

const questRewardEquipmentPromptTemplate = `
You are a fantasy RPG item designer creating an EQUIPPABLE quest reward.

Zone: %s
Quest name: %s
Quest description: %s
Challenge: %s
Quest giver: %s

Point of Interest:
%s

Create a reward item that fits the quest and location flavor.
Equipment slot: %s
Hand equipment constraints:
%s
Constraints:
- Must be wearable/usable equipment for the given slot.
- Tangible item a player could carry.
- Avoid real-world brand names.
- Name should be short (<= 6 words).
- Description 1-2 sentences.
- Rarity tier must be one of: Common, Uncommon, Epic, Mythic.
- Flavor intensity should match the rarity.

Respond ONLY as JSON:
{
  "name": "string",
  "description": "string",
  "rarityTier": "Common"
}
`

func (p *ApplyZoneSeedDraftProcessor) generateQuestRewardItem(
	ctx context.Context,
	zone *models.Zone,
	poi *models.PointOfInterest,
	poiDraft *models.ZoneSeedPointOfInterestDraft,
	character *models.Character,
	draft models.ZoneSeedQuestDraft,
	challengeQuestion string,
	draftReward *models.ZoneSeedQuestRewardItemDraft,
	forceEquipment bool,
) (*models.InventoryItem, error) {
	poiDetails := formatZoneSeedPOIForPrompt(poi, poiDraft)
	zoneName := ""
	if zone != nil {
		zoneName = zone.Name
	}
	equipSlot := pickQuestRewardEquipSlot(forceEquipment)
	handProfile := pickQuestRewardHandProfile(equipSlot)
	fallback := fallbackQuestRewardItem(poi, poiDraft, draft)
	if equipSlot != nil {
		fallback = fallbackQuestRewardEquipment(*equipSlot, handProfile, poi, poiDraft, draft)
	}
	if draftReward != nil && strings.TrimSpace(draftReward.Name) != "" {
		name := strings.TrimSpace(draftReward.Name)
		description := strings.TrimSpace(draftReward.Description)
		if description == "" {
			description = fallback.Description
		}
		rarity := normalizeRarityTier(draftReward.RarityTier)
		if rarity == "" {
			rarity = fallback.RarityTier
		}
		bonuses := questRewardStatBonuses{}
		if equipSlot != nil {
			name = ensureQuestRewardSlotName(name, *equipSlot, handProfile)
			bonuses = rollQuestRewardStatBonuses(rarity)
			name, description = applyQuestRewardIntensity(name, description, rarity, bonuses)
		}
		return p.createInventoryItemReward(ctx, name, description, rarity, equipSlot, &bonuses, handProfile)
	}

	prompt := ""
	if equipSlot != nil {
		prompt = fmt.Sprintf(
			questRewardEquipmentPromptTemplate,
			truncate(zoneName, 120),
			truncate(draft.Name, 120),
			truncate(draft.Description, 400),
			truncate(challengeQuestion, 200),
			truncate(character.Name, 80),
			poiDetails,
			*equipSlot,
			formatQuestRewardHandProfileForPrompt(handProfile),
		)
	} else {
		prompt = fmt.Sprintf(
			questRewardItemPromptTemplate,
			truncate(zoneName, 120),
			truncate(draft.Name, 120),
			truncate(draft.Description, 400),
			truncate(challengeQuestion, 200),
			truncate(character.Name, 80),
			poiDetails,
		)
	}

	answer, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	var response questRewardItemResponse
	if err == nil {
		if err := json.Unmarshal([]byte(extractJSON(answer.Answer)), &response); err != nil {
			response = questRewardItemResponse{}
		}
	}

	name := strings.TrimSpace(response.Name)
	if name == "" {
		name = fallback.Name
	}
	if equipSlot != nil {
		name = ensureQuestRewardSlotName(name, *equipSlot, handProfile)
	}
	description := strings.TrimSpace(response.Description)
	if description == "" {
		description = fallback.Description
	}
	rarity := normalizeRarityTier(response.RarityTier)
	if rarity == "" {
		rarity = fallback.RarityTier
	}

	bonuses := questRewardStatBonuses{}
	if equipSlot != nil {
		bonuses = rollQuestRewardStatBonuses(rarity)
		name, description = applyQuestRewardIntensity(name, description, rarity, bonuses)
	}
	return p.createInventoryItemReward(ctx, name, description, rarity, equipSlot, &bonuses, handProfile)
}

func (p *ApplyZoneSeedDraftProcessor) createInventoryItemReward(
	ctx context.Context,
	name string,
	description string,
	rarity string,
	equipSlot *string,
	bonuses *questRewardStatBonuses,
	handProfile *questRewardHandProfile,
) (*models.InventoryItem, error) {
	var normalizedSlot *string
	if equipSlot != nil {
		slot := strings.TrimSpace(*equipSlot)
		if slot != "" && models.IsValidInventoryEquipSlot(slot) {
			normalizedSlot = &slot
		}
	}
	if normalizedSlot == nil {
		bonuses = nil
		handProfile = nil
	}
	handAttrs := models.HandEquipmentAttributes{}
	if normalizedSlot != nil && handProfile != nil {
		handAttrs = rollQuestRewardHandAttributes(*normalizedSlot, rarity, handProfile)
	}
	validatedHandAttrs, err := models.NormalizeAndValidateHandEquipment(normalizedSlot, handAttrs)
	if err != nil {
		// Fall back to a non-hand reward if generated hand attributes are invalid.
		normalizedSlot = nil
		bonuses = nil
		validatedHandAttrs = models.HandEquipmentAttributes{}
	}
	item := &models.InventoryItem{
		Name:                  name,
		FlavorText:            description,
		RarityTier:            rarity,
		IsCaptureType:         false,
		ItemLevel:             1,
		EquipSlot:             normalizedSlot,
		ImageGenerationStatus: models.InventoryImageGenerationStatusQueued,
	}
	if bonuses != nil {
		item.StrengthMod = bonuses.Strength
		item.DexterityMod = bonuses.Dexterity
		item.ConstitutionMod = bonuses.Constitution
		item.IntelligenceMod = bonuses.Intelligence
		item.WisdomMod = bonuses.Wisdom
		item.CharismaMod = bonuses.Charisma
	}
	item.HandItemCategory = validatedHandAttrs.HandItemCategory
	item.Handedness = validatedHandAttrs.Handedness
	item.DamageMin = validatedHandAttrs.DamageMin
	item.DamageMax = validatedHandAttrs.DamageMax
	item.DamageAffinity = validatedHandAttrs.DamageAffinity
	item.SwipesPerAttack = validatedHandAttrs.SwipesPerAttack
	item.BlockPercentage = validatedHandAttrs.BlockPercentage
	item.DamageBlocked = validatedHandAttrs.DamageBlocked
	item.SpellDamageBonusPercent = validatedHandAttrs.SpellDamageBonusPercent

	if err := p.dbClient.InventoryItem().CreateInventoryItem(ctx, item); err != nil {
		return nil, err
	}

	payload := jobs.GenerateInventoryItemImageTaskPayload{
		InventoryItemID: item.ID,
		Name:            item.Name,
		Description:     item.FlavorText,
		RarityTier:      item.RarityTier,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		errMsg := err.Error()
		_ = p.dbClient.InventoryItem().UpdateInventoryItem(ctx, item.ID, map[string]interface{}{
			"image_generation_status": models.InventoryImageGenerationStatusFailed,
			"image_generation_error":  errMsg,
		})
		return item, nil
	}

	if _, err := p.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateInventoryItemImageTaskType, payloadBytes)); err != nil {
		errMsg := err.Error()
		_ = p.dbClient.InventoryItem().UpdateInventoryItem(ctx, item.ID, map[string]interface{}{
			"image_generation_status": models.InventoryImageGenerationStatusFailed,
			"image_generation_error":  errMsg,
		})
	}

	return item, nil
}

func formatZoneSeedPOIForPrompt(poi *models.PointOfInterest, draft *models.ZoneSeedPointOfInterestDraft) string {
	lines := []string{}

	name := poiNameForPrompt(poi, draft)
	if name != "" {
		lines = append(lines, fmt.Sprintf("Name: %s", name))
	}
	if draft != nil {
		if address := strings.TrimSpace(draft.Address); address != "" {
			lines = append(lines, fmt.Sprintf("Address: %s", address))
		}
		if len(draft.Types) > 0 {
			lines = append(lines, fmt.Sprintf("Types: %s", strings.Join(draft.Types, ", ")))
		}
		if summary := strings.TrimSpace(draft.EditorialSummary); summary != "" {
			lines = append(lines, fmt.Sprintf("Summary: %s", truncate(summary, 200)))
		}
		if draft.Rating > 0 {
			lines = append(lines, fmt.Sprintf("Rating: %.1f (%d reviews)", draft.Rating, draft.UserRatingCount))
		}
	}
	if poi != nil {
		if alt := strings.TrimSpace(poi.OriginalName); alt != "" && alt != name {
			lines = append(lines, fmt.Sprintf("Original name: %s", alt))
		}
		if desc := strings.TrimSpace(poi.Description); desc != "" {
			lines = append(lines, fmt.Sprintf("Description: %s", truncate(desc, 200)))
		}
	}

	if len(lines) == 0 {
		return "No POI details available."
	}
	return strings.Join(lines, "\n")
}

func poiNameForPrompt(poi *models.PointOfInterest, draft *models.ZoneSeedPointOfInterestDraft) string {
	if draft != nil {
		if name := strings.TrimSpace(draft.Name); name != "" {
			return name
		}
	}
	if poi != nil {
		return strings.TrimSpace(poi.Name)
	}
	return ""
}

func fallbackQuestChallengeQuestion(poi *models.PointOfInterest, draft *models.ZoneSeedPointOfInterestDraft) string {
	name := poiNameForPrompt(poi, draft)
	types := []string{}
	if draft != nil && len(draft.Types) > 0 {
		types = draft.Types
	}
	if name == "" {
		return buildHeuristicChallengeQuestion("this location", types)
	}
	return buildHeuristicChallengeQuestion(name, types)
}

func normalizeAppliedChallengeQuestion(
	question string,
	poi *models.PointOfInterest,
	draft *models.ZoneSeedPointOfInterestDraft,
) (string, models.QuestNodeSubmissionType) {
	trimmed := strings.TrimSpace(question)
	name := poiNameForPrompt(poi, draft)
	if name == "" {
		name = "this location"
	}
	types := []string{}
	if draft != nil && len(draft.Types) > 0 {
		types = draft.Types
	}

	if trimmed == "" {
		return buildAppliedPhotoProofQuestion(name, types), models.QuestNodeSubmissionTypePhoto
	}

	lower := strings.ToLower(trimmed)
	hasPhoto := hasAnyKeywordApplied(lower, "photo", "picture", "snapshot", "selfie", "photograph")
	hasText := hasAnyKeywordApplied(lower, "write", "describe", "list", "note", "count", "identify", "observe", "explain", "summarize", "record", "tell")

	if requiresProofOnlyApplied(lower) || (hasPhoto && hasText) {
		return buildAppliedPhotoProofQuestion(name, types), models.QuestNodeSubmissionTypePhoto
	}

	if hasPhoto || hasAnyKeywordApplied(lower, "sketch", "draw") {
		if hasAnyKeywordApplied(lower, "sketch", "draw") && !hasPhoto {
			return fmt.Sprintf("Sketch something you notice at %s and take a photo of your sketch.", name), models.QuestNodeSubmissionTypePhoto
		}
		if strings.HasPrefix(lower, "take") || strings.HasPrefix(lower, "photograph") || strings.HasPrefix(lower, "photo") {
			if needsParticipationOverrideApplied(lower) {
				return buildAppliedPhotoProofQuestion(name, types), models.QuestNodeSubmissionTypePhoto
			}
			return trimmed, models.QuestNodeSubmissionTypePhoto
		}
		return buildAppliedPhotoProofQuestion(name, types), models.QuestNodeSubmissionTypePhoto
	}

	if hasText {
		if needsParticipationOverrideApplied(lower) {
			return buildAppliedPhotoProofQuestion(name, types), models.QuestNodeSubmissionTypePhoto
		}
		return ensureTextOnlyQuestionApplied(trimmed), models.QuestNodeSubmissionTypeText
	}

	return buildAppliedPhotoProofQuestion(name, types), models.QuestNodeSubmissionTypePhoto
}

func buildAppliedPhotoProofQuestion(name string, types []string) string {
	lowerName := strings.ToLower(strings.TrimSpace(name))
	lowerTypes := make([]string, 0, len(types))
	for _, t := range types {
		trimmed := strings.ToLower(strings.TrimSpace(t))
		if trimmed != "" {
			lowerTypes = append(lowerTypes, trimmed)
		}
	}

	hasType := func(targets ...string) bool {
		for _, t := range lowerTypes {
			for _, target := range targets {
				if t == target || strings.Contains(t, target) {
					return true
				}
			}
		}
		return false
	}

	switch {
	case hasType("cafe", "coffee", "coffee_shop", "bakery", "restaurant", "bar", "meal_takeaway", "meal_delivery") || strings.Contains(lowerName, "coffee") || strings.Contains(lowerName, "cafe"):
		return fmt.Sprintf("Get a coffee, drink, or food item at %s and photograph the item you chose.", name)
	case hasType("park", "garden", "playground", "trail", "campground", "natural_feature") || strings.Contains(lowerName, "park") || strings.Contains(lowerName, "garden"):
		return fmt.Sprintf("Spend a moment at %s, then photograph a specific spot where you paused (bench, path, or tree).", name)
	case hasType("museum", "art_gallery", "gallery", "tourist_attraction", "landmark") || strings.Contains(lowerName, "museum") || strings.Contains(lowerName, "gallery"):
		return fmt.Sprintf("Pick an exhibit or artwork at %s and photograph its title card or placard.", name)
	case hasType("movie_theater", "cinema", "theater", "performing_arts_theater", "night_club", "concert_hall") || strings.Contains(lowerName, "theater") || strings.Contains(lowerName, "cinema"):
		return fmt.Sprintf("Choose a show or performance at %s and photograph the listing or ticket detail for what you picked.", name)
	case hasType("store", "shopping_mall", "supermarket", "market", "book_store", "clothing_store") || strings.Contains(lowerName, "market") || strings.Contains(lowerName, "shop"):
		return fmt.Sprintf("Find something at %s that caught your attention and photograph the item or display.", name)
	case hasType("plaza", "square", "bridge", "beach", "marina", "harbor") || strings.Contains(lowerName, "plaza") || strings.Contains(lowerName, "square"):
		return fmt.Sprintf("Explore %s and photograph a specific viewpoint or feature where you spent time.", name)
	default:
		return fmt.Sprintf("Take a photo that clearly shows you participating in the main activity at %s.", name)
	}
}

func needsParticipationOverrideApplied(text string) bool {
	if text == "" {
		return false
	}
	lower := strings.ToLower(text)
	if !hasAnyKeywordApplied(lower,
		"sign", "storefront", "entrance", "exterior", "outside", "front",
		"marquee", "poster", "window", "logo", "facade", "building",
	) {
		return false
	}
	if hasAnyKeywordApplied(lower,
		"drink", "meal", "food", "book", "read", "page", "shelf",
		"stage", "show", "performance", "set", "lineup", "ticket", "program",
		"exhibit", "gallery", "trail", "play", "game", "class", "workout",
		"ride", "viewing", "display", "bench", "picnic", "market", "shop",
		"browse", "order", "sip", "eat", "watch", "listen",
	) {
		return false
	}
	return true
}

func hasAnyKeywordApplied(text string, keywords ...string) bool {
	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func requiresProofOnlyApplied(text string) bool {
	return hasAnyKeywordApplied(text,
		"oldest", "first", "main ingredient", "ingredient", "recipe", "menu", "price", "cost",
		"listen", "instrument", "song", "piece", "performance", "band", "taste", "flavor",
		"review", "best", "favorite", "count", "number of",
	)
}

func ensureTextOnlyQuestionApplied(question string) string {
	trimmed := strings.TrimSpace(question)
	if trimmed == "" {
		return "Write one sentence about something you notice here."
	}
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "write") || strings.HasPrefix(lower, "describe") || strings.HasPrefix(lower, "list") || strings.HasPrefix(lower, "note") {
		return trimmed
	}
	return fmt.Sprintf("Write 1-2 sentences: %s", trimmed)
}

func randomQuestDifficulty() int {
	return 25 + rand.Intn(26)
}

func clampQuestDifficulty(value int) int {
	if value < 25 {
		return 25
	}
	if value > 50 {
		return 50
	}
	return value
}

func normalizeRarityTier(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "common":
		return "Common"
	case "uncommon":
		return "Uncommon"
	case "epic":
		return "Epic"
	case "mythic":
		return "Mythic"
	case "legendary":
		return "Mythic"
	case "rare":
		return "Uncommon"
	case "very rare":
		return "Epic"
	default:
		return ""
	}
}

func randomRarityTier() string {
	roll := rand.Intn(100)
	switch {
	case roll < 50:
		return "Common"
	case roll < 80:
		return "Uncommon"
	case roll < 95:
		return "Epic"
	default:
		return "Mythic"
	}
}

func pickQuestRewardEquipSlot(force bool) *string {
	if !force {
		if rand.Float64() > 0.35 {
			return nil
		}
	}
	if len(models.InventoryEquipSlots) == 0 {
		return nil
	}
	slot := string(models.InventoryEquipSlots[rand.Intn(len(models.InventoryEquipSlots))])
	return &slot
}

func pickQuestRewardHandProfile(equipSlot *string) *questRewardHandProfile {
	if equipSlot == nil {
		return nil
	}
	slot := strings.TrimSpace(*equipSlot)
	switch slot {
	case string(models.EquipmentSlotDominantHand):
		if rand.Float64() < 0.25 {
			return &questRewardHandProfile{
				Category:   string(models.HandItemCategoryStaff),
				Handedness: string(models.HandednessTwoHanded),
			}
		}
		handedness := string(models.HandednessOneHanded)
		if rand.Float64() < 0.35 {
			handedness = string(models.HandednessTwoHanded)
		}
		return &questRewardHandProfile{
			Category:   string(models.HandItemCategoryWeapon),
			Handedness: handedness,
		}
	case string(models.EquipmentSlotOffHand):
		category := string(models.HandItemCategoryShield)
		if rand.Float64() < 0.45 {
			category = string(models.HandItemCategoryOrb)
		}
		return &questRewardHandProfile{
			Category:   category,
			Handedness: string(models.HandednessOneHanded),
		}
	default:
		return nil
	}
}

func formatQuestRewardHandProfileForPrompt(handProfile *questRewardHandProfile) string {
	if handProfile == nil {
		return "- Not hand equipment."
	}
	switch handProfile.Category {
	case string(models.HandItemCategoryWeapon):
		return fmt.Sprintf("- category: weapon\n- handedness: %s\n- must include a damage range and swipe cadence", handProfile.Handedness)
	case string(models.HandItemCategoryStaff):
		return "- category: staff\n- handedness: two_handed\n- must imply both physical damage and a spell power boost"
	case string(models.HandItemCategoryShield):
		return "- category: shield\n- handedness: one_handed\n- should imply defensive blocking capability"
	case string(models.HandItemCategoryOrb):
		return "- category: orb\n- handedness: one_handed\n- should imply spell amplification"
	default:
		return "- Follow slot rules."
	}
}

func rollQuestRewardHandAttributes(slot string, rarity string, handProfile *questRewardHandProfile) models.HandEquipmentAttributes {
	if handProfile == nil {
		return models.HandEquipmentAttributes{}
	}
	attrs := models.HandEquipmentAttributes{
		HandItemCategory: stringPtr(handProfile.Category),
		Handedness:       stringPtr(handProfile.Handedness),
	}
	switch slot {
	case string(models.EquipmentSlotDominantHand):
		damageMin, damageMax := questRewardDamageRangeByRarity(rarity)
		if handProfile.Handedness == string(models.HandednessTwoHanded) {
			damageMin = int(float64(damageMin) * 1.35)
			damageMax = int(float64(damageMax) * 1.35)
		}
		if handProfile.Category == string(models.HandItemCategoryStaff) {
			damageMin = int(float64(damageMin) * 0.9)
			damageMax = int(float64(damageMax) * 0.9)
			bonusMin, bonusMax := questRewardSpellBonusRangeByRarity(rarity)
			attrs.SpellDamageBonusPercent = intPtr(rollRange(bonusMin, bonusMax))
		}
		attrs.DamageMin = intPtr(maxInt(1, damageMin))
		attrs.DamageMax = intPtr(maxInt(*attrs.DamageMin, damageMax))
		attrs.SwipesPerAttack = intPtr(questRewardSwipesPerAttack(handProfile))
	case string(models.EquipmentSlotOffHand):
		if handProfile.Category == string(models.HandItemCategoryShield) {
			blockPctMin, blockPctMax := questRewardBlockPercentRangeByRarity(rarity)
			blockMin, blockMax := questRewardBlockedDamageRangeByRarity(rarity)
			attrs.BlockPercentage = intPtr(rollRange(blockPctMin, blockPctMax))
			attrs.DamageBlocked = intPtr(rollRange(blockMin, blockMax))
		}
		if handProfile.Category == string(models.HandItemCategoryOrb) {
			bonusMin, bonusMax := questRewardSpellBonusRangeByRarity(rarity)
			attrs.SpellDamageBonusPercent = intPtr(rollRange(bonusMin, bonusMax))
		}
	}
	return attrs
}

func questRewardDamageRangeByRarity(rarity string) (int, int) {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "common":
		return 3, 6
	case "uncommon":
		return 5, 10
	case "epic":
		return 9, 16
	case "mythic":
		return 14, 24
	default:
		return 3, 6
	}
}

func questRewardSwipesPerAttack(handProfile *questRewardHandProfile) int {
	if handProfile == nil {
		return 1
	}
	if handProfile.Category == string(models.HandItemCategoryStaff) {
		return rollRange(1, 2)
	}
	if handProfile.Handedness == string(models.HandednessTwoHanded) {
		return rollRange(1, 2)
	}
	return rollRange(2, 4)
}

func questRewardBlockPercentRangeByRarity(rarity string) (int, int) {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "common":
		return 10, 18
	case "uncommon":
		return 18, 28
	case "epic":
		return 28, 42
	case "mythic":
		return 40, 58
	default:
		return 10, 18
	}
}

func questRewardBlockedDamageRangeByRarity(rarity string) (int, int) {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "common":
		return 2, 5
	case "uncommon":
		return 4, 8
	case "epic":
		return 8, 14
	case "mythic":
		return 13, 20
	default:
		return 2, 5
	}
}

func questRewardSpellBonusRangeByRarity(rarity string) (int, int) {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "common":
		return 8, 14
	case "uncommon":
		return 14, 22
	case "epic":
		return 22, 34
	case "mythic":
		return 34, 50
	default:
		return 8, 14
	}
}

func rollRange(minValue int, maxValue int) int {
	if maxValue < minValue {
		maxValue = minValue
	}
	return minValue + rand.Intn(maxValue-minValue+1)
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func intPtr(value int) *int {
	v := value
	return &v
}

func stringPtr(value string) *string {
	v := value
	return &v
}

func rollQuestRewardStatBonuses(rarity string) questRewardStatBonuses {
	minBonus, maxBonus := questRewardBonusRange(rarity)
	if minBonus <= 0 || maxBonus <= 0 {
		return questRewardStatBonuses{}
	}
	if maxBonus < minBonus {
		maxBonus = minBonus
	}
	total := minBonus + rand.Intn(maxBonus-minBonus+1)
	if total <= 0 {
		return questRewardStatBonuses{}
	}
	stats := []string{"strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma"}
	rand.Shuffle(len(stats), func(i, j int) { stats[i], stats[j] = stats[j], stats[i] })
	numStats := 1
	if total >= 3 {
		numStats = 2
	}
	first := total
	second := 0
	if numStats == 2 {
		first = 1 + rand.Intn(total-1)
		second = total - first
	}
	bonuses := questRewardStatBonuses{}
	assignBonus := func(stat string, bonus int) {
		switch stat {
		case "strength":
			bonuses.Strength = bonus
		case "dexterity":
			bonuses.Dexterity = bonus
		case "constitution":
			bonuses.Constitution = bonus
		case "intelligence":
			bonuses.Intelligence = bonus
		case "wisdom":
			bonuses.Wisdom = bonus
		case "charisma":
			bonuses.Charisma = bonus
		}
	}
	assignBonus(stats[0], first)
	if numStats == 2 {
		assignBonus(stats[1], second)
	}
	return bonuses
}

func questRewardBonusRange(rarity string) (int, int) {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "common":
		return 1, 1
	case "uncommon":
		return 1, 2
	case "epic":
		return 2, 4
	case "mythic":
		return 3, 5
	default:
		return 1, 1
	}
}

func applyQuestRewardIntensity(name string, description string, rarity string, bonuses questRewardStatBonuses) (string, string) {
	trimmedName := strings.TrimSpace(name)
	trimmedDesc := strings.TrimSpace(description)
	rarityKey := strings.ToLower(strings.TrimSpace(rarity))
	totalBonus := bonuses.total()

	prefix := ""
	switch rarityKey {
	case "common":
		if totalBonus >= 2 {
			prefix = "Sturdy"
		}
	case "uncommon":
		prefix = "Refined"
	case "epic":
		prefix = "Heroic"
	case "mythic":
		prefix = "Mythic"
	default:
		if totalBonus >= 3 {
			prefix = "Empowered"
		}
	}
	if totalBonus >= 4 && rarityKey != "mythic" {
		prefix = "Fabled"
	}

	statDescriptor := questRewardStatDescriptor(bonuses)
	if trimmedName != "" {
		candidate := trimmedName
		if statDescriptor != "" {
			candidate = fmt.Sprintf("%s %s", statDescriptor, candidate)
		}
		if prefix != "" && countWords(candidate) <= 5 && !strings.Contains(strings.ToLower(candidate), strings.ToLower(prefix)) {
			candidate = fmt.Sprintf("%s %s", prefix, candidate)
		}
		if countWords(candidate) <= 6 {
			trimmedName = candidate
		}
	}

	if trimmedDesc == "" {
		return trimmedName, trimmedDesc
	}

	intensity := ""
	switch rarityKey {
	case "common":
		intensity = "It carries a modest but dependable strength."
	case "uncommon":
		intensity = "It hums with a practiced edge."
	case "epic":
		intensity = "Power coils through it with heroic force."
	case "mythic":
		intensity = "Its power is overwhelming, the stuff of legends."
	default:
		intensity = "A steady power lingers within."
	}
	if totalBonus >= 4 {
		intensity = "Its power flares with unmistakable might."
	}

	statFlavor := questRewardStatFlavor(bonuses, rarityKey, totalBonus)
	if statFlavor == "" {
		statFlavor = intensity
	} else if intensity != "" && !strings.Contains(statFlavor, "power") {
		statFlavor = strings.TrimSpace(fmt.Sprintf("%s %s", statFlavor, intensity))
	}

	if strings.HasSuffix(trimmedDesc, ".") || strings.HasSuffix(trimmedDesc, "!") || strings.HasSuffix(trimmedDesc, "?") {
		trimmedDesc = fmt.Sprintf("%s %s", trimmedDesc, statFlavor)
	} else {
		trimmedDesc = fmt.Sprintf("%s. %s", trimmedDesc, statFlavor)
	}

	return trimmedName, strings.TrimSpace(trimmedDesc)
}

func countWords(value string) int {
	parts := strings.Fields(value)
	return len(parts)
}

func questRewardStatDescriptor(bonuses questRewardStatBonuses) string {
	stats := questRewardStatKeys(bonuses)
	if len(stats) == 0 {
		return ""
	}
	if len(stats) == 1 {
		return questRewardSingleStatDescriptor(stats[0])
	}
	pairKey := questRewardStatPairKey(stats[0], stats[1])
	if descriptor, ok := questRewardPairDescriptors[pairKey]; ok {
		return descriptor
	}
	return questRewardSingleStatDescriptor(stats[0])
}

func questRewardStatFlavor(bonuses questRewardStatBonuses, rarityKey string, totalBonus int) string {
	stats := questRewardStatKeys(bonuses)
	if len(stats) == 0 {
		return ""
	}
	list := questRewardStatList(stats)
	if list == "" {
		return ""
	}
	verb := "bolsters"
	switch rarityKey {
	case "epic", "mythic":
		verb = "surges with"
	}
	if totalBonus >= 4 {
		verb = "blazes with"
	}
	return fmt.Sprintf("It %s %s.", verb, list)
}

func questRewardStatKeys(bonuses questRewardStatBonuses) []string {
	type statEntry struct {
		key   string
		value int
	}
	entries := []statEntry{
		{key: "strength", value: bonuses.Strength},
		{key: "dexterity", value: bonuses.Dexterity},
		{key: "constitution", value: bonuses.Constitution},
		{key: "intelligence", value: bonuses.Intelligence},
		{key: "wisdom", value: bonuses.Wisdom},
		{key: "charisma", value: bonuses.Charisma},
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].value == entries[j].value {
			return entries[i].key < entries[j].key
		}
		return entries[i].value > entries[j].value
	})
	keys := []string{}
	for _, entry := range entries {
		if entry.value > 0 {
			keys = append(keys, entry.key)
		}
		if len(keys) == 2 {
			break
		}
	}
	return keys
}

func questRewardStatList(stats []string) string {
	if len(stats) == 0 {
		return ""
	}
	if len(stats) == 1 {
		return stats[0]
	}
	return fmt.Sprintf("%s and %s", stats[0], stats[1])
}

func questRewardStatPairKey(a string, b string) string {
	if a < b {
		return a + "+" + b
	}
	return b + "+" + a
}

func questRewardSingleStatDescriptor(stat string) string {
	switch stat {
	case "strength":
		return "Mighty"
	case "dexterity":
		return "Swift"
	case "constitution":
		return "Stalwart"
	case "intelligence":
		return "Keen"
	case "wisdom":
		return "Sage"
	case "charisma":
		return "Radiant"
	default:
		return ""
	}
}

var questRewardPairDescriptors = map[string]string{
	"strength+constitution":     "Stout",
	"strength+dexterity":        "Forceful",
	"strength+intelligence":     "Calculated",
	"strength+wisdom":           "Steadfast",
	"strength+charisma":         "Commanding",
	"dexterity+constitution":    "Agile",
	"dexterity+intelligence":    "Keen",
	"dexterity+wisdom":          "Wary",
	"dexterity+charisma":        "Dashing",
	"constitution+intelligence": "Resolute",
	"constitution+wisdom":       "Stalwart",
	"constitution+charisma":     "Dauntless",
	"intelligence+wisdom":       "Sage",
	"intelligence+charisma":     "Silvered",
	"wisdom+charisma":           "Radiant",
}

func ensureQuestRewardSlotName(name string, slot string, handProfile *questRewardHandProfile) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return name
	}
	slotKey := strings.ToLower(strings.TrimSpace(slot))
	keywords, slotNoun := questRewardSlotKeywords(slotKey, handProfile)
	if slotNoun == "" {
		return trimmed
	}
	lower := strings.ToLower(trimmed)
	for _, keyword := range keywords {
		if keyword != "" && strings.Contains(lower, keyword) {
			return trimmed
		}
	}
	words := strings.Fields(trimmed)
	if len(words) == 0 {
		return slotNoun
	}
	if len(words) >= 6 {
		words[len(words)-1] = slotNoun
		return strings.Join(words, " ")
	}
	return fmt.Sprintf("%s %s", trimmed, slotNoun)
}

func questRewardSlotKeywords(slot string, handProfile *questRewardHandProfile) ([]string, string) {
	switch slot {
	case string(models.EquipmentSlotHat):
		return []string{"hat", "helm", "cap", "hood", "cowl", "circlet"}, "Helm"
	case string(models.EquipmentSlotNecklace):
		return []string{"amulet", "pendant", "necklace", "torc", "choker"}, "Amulet"
	case string(models.EquipmentSlotChest):
		return []string{"chest", "armor", "vest", "cuirass", "coat", "tunic", "jerkin"}, "Cuirass"
	case string(models.EquipmentSlotLegs):
		return []string{"legs", "greaves", "leggings", "pants", "trousers", "kilt"}, "Greaves"
	case string(models.EquipmentSlotShoes):
		return []string{"boots", "shoes", "sandals", "sabaton"}, "Boots"
	case string(models.EquipmentSlotGloves):
		return []string{"gloves", "gauntlets", "bracers", "mitts"}, "Gloves"
	case string(models.EquipmentSlotDominantHand):
		if handProfile != nil && handProfile.Category == string(models.HandItemCategoryStaff) {
			return []string{"staff", "rod", "stave", "cane"}, "Staff"
		}
		if handProfile != nil && handProfile.Handedness == string(models.HandednessTwoHanded) {
			return []string{"greatsword", "axe", "hammer", "polearm", "blade"}, "Greatblade"
		}
		return []string{"blade", "sword", "dagger", "mace", "axe", "hammer", "tool"}, "Blade"
	case string(models.EquipmentSlotOffHand):
		if handProfile != nil && handProfile.Category == string(models.HandItemCategoryOrb) {
			return []string{"orb", "focus", "globe", "sphere"}, "Orb"
		}
		return []string{"shield", "buckler", "tome", "focus", "orb"}, "Shield"
	case string(models.EquipmentSlotRing), string(models.EquipmentSlotRingLeft), string(models.EquipmentSlotRingRight):
		return []string{"ring", "band", "signet"}, "Ring"
	default:
		return nil, ""
	}
}

func fallbackQuestRewardItem(poi *models.PointOfInterest, draft *models.ZoneSeedPointOfInterestDraft, quest models.ZoneSeedQuestDraft) questRewardItemResponse {
	name := strings.TrimSpace(quest.Name)
	if name == "" {
		name = "Traveler's Token"
	} else {
		name = fmt.Sprintf("%s Token", name)
	}

	poiName := poiNameForPrompt(poi, draft)
	description := "A keepsake earned by completing a local favor."
	if poiName != "" {
		description = fmt.Sprintf("A keepsake earned near %s, warm with the memory of the place.", poiName)
	}

	return questRewardItemResponse{
		Name:        name,
		Description: description,
		RarityTier:  randomRarityTier(),
	}
}

func fallbackQuestRewardEquipment(
	slot string,
	handProfile *questRewardHandProfile,
	poi *models.PointOfInterest,
	draft *models.ZoneSeedPointOfInterestDraft,
	quest models.ZoneSeedQuestDraft,
) questRewardItemResponse {
	slot = strings.TrimSpace(strings.ToLower(slot))
	base := strings.TrimSpace(quest.Name)
	if base == "" {
		base = "Wayfarer's"
	}

	slotNoun := "Gear"
	switch slot {
	case string(models.EquipmentSlotHat):
		slotNoun = "Cap"
	case string(models.EquipmentSlotNecklace):
		slotNoun = "Pendant"
	case string(models.EquipmentSlotChest):
		slotNoun = "Vest"
	case string(models.EquipmentSlotLegs):
		slotNoun = "Greaves"
	case string(models.EquipmentSlotShoes):
		slotNoun = "Boots"
	case string(models.EquipmentSlotGloves):
		slotNoun = "Gloves"
	case string(models.EquipmentSlotDominantHand):
		if handProfile != nil && handProfile.Category == string(models.HandItemCategoryStaff) {
			slotNoun = "Staff"
		} else if handProfile != nil && handProfile.Handedness == string(models.HandednessTwoHanded) {
			slotNoun = "Greatblade"
		} else {
			slotNoun = "Blade"
		}
	case string(models.EquipmentSlotOffHand):
		if handProfile != nil && handProfile.Category == string(models.HandItemCategoryOrb) {
			slotNoun = "Orb"
		} else {
			slotNoun = "Buckler"
		}
	case string(models.EquipmentSlotRing), string(models.EquipmentSlotRingLeft), string(models.EquipmentSlotRingRight):
		slotNoun = "Ring"
	}

	name := fmt.Sprintf("%s %s", base, slotNoun)
	poiName := poiNameForPrompt(poi, draft)
	description := "A well-used piece of gear earned for helping out nearby."
	if poiName != "" {
		description = fmt.Sprintf("A well-used piece of gear earned near %s, still warm with local pride.", poiName)
	}

	return questRewardItemResponse{
		Name:        name,
		Description: description,
		RarityTier:  randomRarityTier(),
	}
}

func buildHeuristicChallengeQuestion(name string, types []string) string {
	lowerName := strings.ToLower(strings.TrimSpace(name))
	lowerTypes := make([]string, 0, len(types))
	for _, t := range types {
		trimmed := strings.ToLower(strings.TrimSpace(t))
		if trimmed != "" {
			lowerTypes = append(lowerTypes, trimmed)
		}
	}

	hasType := func(targets ...string) bool {
		for _, t := range lowerTypes {
			for _, target := range targets {
				if t == target || strings.Contains(t, target) {
					return true
				}
			}
		}
		return false
	}

	switch {
	case hasType("cafe", "coffee", "coffee_shop", "bakery") || strings.Contains(lowerName, "coffee") || strings.Contains(lowerName, "cafe"):
		return fmt.Sprintf("Write a 4-line poem inspired by the smells or sounds near %s, then photograph the page with the storefront in view.", name)
	case hasType("park", "garden", "playground", "trail", "campground", "natural_feature") || strings.Contains(lowerName, "park") || strings.Contains(lowerName, "garden"):
		return fmt.Sprintf("Sketch one plant or tree you can see at %s and take a photo of the sketch with the area in view.", name)
	case hasType("playground") || strings.Contains(lowerName, "playground"):
		return fmt.Sprintf("Count the number of swings or slides you can see at %s and photograph them.", name)
	case hasType("trail", "hiking_trail", "walking_trail") || strings.Contains(lowerName, "trail"):
		return fmt.Sprintf("Find a trail marker or distance sign at %s and photograph it.", name)
	case hasType("library") || strings.Contains(lowerName, "library"):
		return fmt.Sprintf("Find a book title on a display or shelf at %s that includes a color word, and photograph the title.", name)
	case hasType("book_store", "bookstore") || strings.Contains(lowerName, "book"):
		return fmt.Sprintf("Find a book cover at %s with a creature on it and photograph the cover (no purchase needed).", name)
	case hasType("museum", "art_gallery", "gallery") || strings.Contains(lowerName, "museum") || strings.Contains(lowerName, "gallery"):
		return fmt.Sprintf("From outside %s, note two materials used in the building facade and photograph the entrance.", name)
	case hasType("amusement_park") || strings.Contains(lowerName, "amusement"):
		return fmt.Sprintf("Find a ride name or attraction map near %s and photograph it.", name)
	case hasType("aquarium", "zoo") || strings.Contains(lowerName, "aquarium") || strings.Contains(lowerName, "zoo"):
		return fmt.Sprintf("From outside %s, find a sign or poster featuring an animal and photograph it.", name)
	case hasType("bowling_alley") || strings.Contains(lowerName, "bowling"):
		return fmt.Sprintf("Photograph the lane count or main sign at %s from the lobby or entrance.", name)
	case hasType("casino") || strings.Contains(lowerName, "casino"):
		return fmt.Sprintf("Find the main entrance sign at %s and photograph it from outside.", name)
	case hasType("tourist_attraction", "point_of_interest", "landmark", "monument", "historic_site") || strings.Contains(lowerName, "monument"):
		return fmt.Sprintf("Find a plaque, marker, or date on or near %s and photograph it.", name)
	case hasType("cemetery") || strings.Contains(lowerName, "cemetery"):
		return fmt.Sprintf("From a respectful distance at %s, photograph a sign or gate that shows the cemetery name.", name)
	case hasType("plaza", "square") || strings.Contains(lowerName, "plaza") || strings.Contains(lowerName, "square"):
		return fmt.Sprintf("Count the number of visible benches or seating areas at %s and photograph the space.", name)
	case hasType("bridge") || strings.Contains(lowerName, "bridge"):
		return fmt.Sprintf("Count the number of arches or spans you can see on %s and photograph the structure.", name)
	case hasType("church", "place_of_worship", "synagogue", "mosque") || strings.Contains(lowerName, "church"):
		return fmt.Sprintf("Count the visible arches or windows on the exterior of %s and photograph the building.", name)
	case hasType("school", "primary_school", "secondary_school", "university", "college") || strings.Contains(lowerName, "school") || strings.Contains(lowerName, "university"):
		return fmt.Sprintf("Find the school name or crest on a sign near %s and photograph it from outside.", name)
	case hasType("movie_theater", "cinema", "theater", "performing_arts_theater") || strings.Contains(lowerName, "theater") || strings.Contains(lowerName, "cinema"):
		return fmt.Sprintf("Find a poster or marquee at %s and photograph the title of a show or film.", name)
	case hasType("music_venue", "night_club") || strings.Contains(lowerName, "venue"):
		return fmt.Sprintf("From outside %s, photograph the main sign and describe the vibe in one sentence.", name)
	case hasType("barber", "hair_care", "beauty_salon", "spa") || strings.Contains(lowerName, "salon") || strings.Contains(lowerName, "spa"):
		return fmt.Sprintf("Photograph the main sign at %s and describe the storefront style in one sentence.", name)
	case hasType("gym", "fitness") || strings.Contains(lowerName, "gym") || strings.Contains(lowerName, "fitness"):
		return fmt.Sprintf("Count the number of different exercise stations or windows you can see at %s from outside and photograph the front.", name)
	case hasType("restaurant", "bar", "brewery", "meal_takeaway", "meal_delivery") || strings.Contains(lowerName, "restaurant") || strings.Contains(lowerName, "bar"):
		return fmt.Sprintf("Take a photo of the main sign at %s and write a two-sentence review of the vibe from outside.", name)
	case hasType("ice_cream_shop", "dessert") || strings.Contains(lowerName, "ice cream") || strings.Contains(lowerName, "gelato"):
		return fmt.Sprintf("Find a flavor board or dessert sign at %s and photograph it.", name)
	case hasType("store", "shopping_mall", "supermarket", "market") || strings.Contains(lowerName, "market") || strings.Contains(lowerName, "shop"):
		return fmt.Sprintf("Find a window display at %s and describe the dominant color theme in one sentence, then photograph it.", name)
	case hasType("clothing_store", "shoe_store", "department_store") || strings.Contains(lowerName, "boutique"):
		return fmt.Sprintf("Find a window display at %s featuring an outfit and photograph it.", name)
	case hasType("electronics_store") || strings.Contains(lowerName, "electronics"):
		return fmt.Sprintf("Photograph the main sign at %s and note one product category highlighted in the window.", name)
	case hasType("furniture_store", "home_goods_store") || strings.Contains(lowerName, "furniture"):
		return fmt.Sprintf("Photograph a window display at %s and describe the texture of the featured item.", name)
	case hasType("hardware_store") || strings.Contains(lowerName, "hardware"):
		return fmt.Sprintf("Find the tool or materials signage at %s and photograph it.", name)
	case hasType("pet_store", "veterinary_care") || strings.Contains(lowerName, "pet"):
		return fmt.Sprintf("Photograph the pet-related sign at %s and note one animal featured.", name)
	case hasType("florist") || strings.Contains(lowerName, "florist"):
		return fmt.Sprintf("Find a flower display at %s and photograph the most vibrant color you see.", name)
	case hasType("hotel", "lodging") || strings.Contains(lowerName, "hotel"):
		return fmt.Sprintf("Find the hotel name on the exterior of %s and photograph it from the sidewalk.", name)
	case hasType("pharmacy", "drugstore") || strings.Contains(lowerName, "pharmacy"):
		return fmt.Sprintf("Photograph the main pharmacy sign at %s and note one color used in the branding.", name)
	case hasType("bank", "atm") || strings.Contains(lowerName, "bank"):
		return fmt.Sprintf("Find the posted hours or services signage at %s and photograph it.", name)
	case hasType("courthouse", "city_hall", "town_hall") || strings.Contains(lowerName, "city hall"):
		return fmt.Sprintf("Photograph the main civic building sign at %s and note the year if displayed.", name)
	case hasType("police", "fire_station") || strings.Contains(lowerName, "police") || strings.Contains(lowerName, "fire"):
		return fmt.Sprintf("From outside %s, photograph the station sign or emblem.", name)
	case hasType("hospital", "doctor", "dentist", "health", "clinic") || strings.Contains(lowerName, "hospital") || strings.Contains(lowerName, "clinic"):
		return fmt.Sprintf("Find the clinic or hospital name at %s and photograph it from outside.", name)
	case hasType("post_office") || strings.Contains(lowerName, "post office"):
		return fmt.Sprintf("Photograph the postal logo or hours sign outside %s.", name)
	case hasType("gas_station") || strings.Contains(lowerName, "gas"):
		return fmt.Sprintf("From outside %s, count the number of fuel pumps you can see and photograph the station.", name)
	case hasType("car_wash") || strings.Contains(lowerName, "car wash"):
		return fmt.Sprintf("Find the car wash entrance sign at %s and photograph it.", name)
	case hasType("car_repair") || strings.Contains(lowerName, "auto") || strings.Contains(lowerName, "mechanic"):
		return fmt.Sprintf("Photograph the service sign at %s and note one service offered.", name)
	case hasType("laundry", "dry_cleaner") || strings.Contains(lowerName, "laundry") || strings.Contains(lowerName, "cleaner"):
		return fmt.Sprintf("Find the posted hours at %s and photograph them.", name)
	case hasType("parking") || strings.Contains(lowerName, "garage"):
		return fmt.Sprintf("Find the maximum height or rate sign at %s and photograph it.", name)
	case hasType("stadium", "sports_complex", "gym") || strings.Contains(lowerName, "stadium") || strings.Contains(lowerName, "gym"):
		return fmt.Sprintf("Count the visible entrances to %s and photograph the main gate.", name)
	case hasType("train_station", "subway_station", "bus_station", "transit_station") || strings.Contains(lowerName, "station"):
		return fmt.Sprintf("Find the line color or route number on signage at %s and photograph it.", name)
	case hasType("bus_stop") || strings.Contains(lowerName, "bus stop"):
		return fmt.Sprintf("Find the route number on the stop sign at %s and photograph it.", name)
	case hasType("airport") || strings.Contains(lowerName, "airport"):
		return fmt.Sprintf("Photograph the terminal or airport name sign at %s from outside.", name)
	case hasType("beach", "lake", "river", "water", "harbor", "marina") || strings.Contains(lowerName, "beach") || strings.Contains(lowerName, "lake"):
		return fmt.Sprintf("Collect three different textures you can see at %s (sand, stone, water, etc.) and photograph them together.", name)
	case hasType("viewpoint", "scenic_lookout", "observation_deck") || strings.Contains(lowerName, "lookout") || strings.Contains(lowerName, "overlook"):
		return fmt.Sprintf("Find the best viewpoint at %s and take a photo that shows the horizon.", name)
	case hasType("hiking_area", "national_park") || strings.Contains(lowerName, "national park"):
		return fmt.Sprintf("Photograph a trail map or rules sign at %s.", name)
	default:
		return fmt.Sprintf("Take a clear photo of the main sign or entrance at %s.", name)
	}
}

type questStatTagResponse struct {
	StatTags []string `json:"statTags"`
}

const questStatTagPromptTemplate = `
You are a game designer classifying which character stats are most relevant for a quest.

Allowed stat tags (use only these, lowercase):
- strength
- dexterity
- constitution
- intelligence
- wisdom
- charisma

Quest details:
Name: %s
Description: %s
Challenge: %s
Acceptance dialogue:
%s

Pick up to 2 stat tags that best fit the quest. If none apply, return an empty array.

Respond ONLY as JSON:
{
  "statTags": ["strength"]
}
`

func (p *ApplyZoneSeedDraftProcessor) classifyQuestStatTags(
	ctx context.Context,
	draft models.ZoneSeedQuestDraft,
) models.StringArray {
	dialogue := strings.TrimSpace(strings.Join(draft.AcceptanceDialogue, "\n"))
	prompt := fmt.Sprintf(
		questStatTagPromptTemplate,
		strings.TrimSpace(draft.Name),
		strings.TrimSpace(draft.Description),
		strings.TrimSpace(draft.ChallengeQuestion),
		dialogue,
	)

	answer, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return inferQuestStatTagsHeuristic(draft)
	}

	var response questStatTagResponse
	if err := json.Unmarshal([]byte(extractJSON(answer.Answer)), &response); err != nil {
		return inferQuestStatTagsHeuristic(draft)
	}

	valid := map[string]struct{}{
		"strength":     {},
		"dexterity":    {},
		"constitution": {},
		"intelligence": {},
		"wisdom":       {},
		"charisma":     {},
	}
	seen := map[string]struct{}{}
	tags := make([]string, 0, len(response.StatTags))
	for _, tag := range response.StatTags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		if _, ok := valid[normalized]; !ok {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		tags = append(tags, normalized)
		if len(tags) >= 2 {
			break
		}
	}

	if len(tags) == 0 {
		return inferQuestStatTagsHeuristic(draft)
	}

	return models.StringArray(tags)
}

func inferQuestStatTagsHeuristic(draft models.ZoneSeedQuestDraft) models.StringArray {
	textParts := []string{draft.Name, draft.Description, draft.ChallengeQuestion}
	textParts = append(textParts, draft.AcceptanceDialogue...)
	text := strings.ToLower(strings.TrimSpace(strings.Join(textParts, " ")))
	if text == "" {
		return models.StringArray{}
	}

	keywordMap := map[string][]string{
		"strength": {
			"lift", "carry", "haul", "forge", "smith", "battle", "fight", "brawl",
			"guard", "muscle", "heavy", "strike", "hammer", "anvil",
		},
		"dexterity": {
			"sneak", "stealth", "pick", "lock", "climb", "dodge", "agile", "balance",
			"aim", "archery", "bow", "dagger", "quick", "swift",
		},
		"constitution": {
			"endure", "survive", "stamina", "resist", "poison", "tough", "weather",
			"long haul", "marathon", "patrol", "fortitude",
		},
		"intelligence": {
			"research", "investigate", "decode", "puzzle", "riddle", "analyze", "study",
			"alchemist", "invention", "map", "library", "ledger",
		},
		"wisdom": {
			"herb", "heal", "ritual", "spirit", "listen", "observe", "insight",
			"track", "wild", "nature", "garden", "grove",
		},
		"charisma": {
			"convince", "negotiate", "charm", "perform", "sing", "dance", "crowd",
			"influence", "trade", "barter", "festival", "story",
		},
	}

	type scoredTag struct {
		tag   string
		score int
	}
	scores := make([]scoredTag, 0, len(keywordMap))
	for tag, keywords := range keywordMap {
		score := 0
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				score++
			}
		}
		if score > 0 {
			scores = append(scores, scoredTag{tag: tag, score: score})
		}
	}

	if len(scores) == 0 {
		return models.StringArray{}
	}

	sort.Slice(scores, func(i, j int) bool {
		if scores[i].score == scores[j].score {
			return scores[i].tag < scores[j].tag
		}
		return scores[i].score > scores[j].score
	})

	limit := 2
	if len(scores) < limit {
		limit = len(scores)
	}
	tags := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		tags = append(tags, scores[i].tag)
	}

	return models.StringArray(tags)
}
