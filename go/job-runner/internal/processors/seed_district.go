package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/locationseeder"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type SeedDistrictProcessor struct {
	dbClient       db.DbClient
	deepPriest     deep_priest.DeepPriest
	dungeonmaster  dungeonmaster.Client
	locationSeeder locationseeder.Client
	asyncClient    *asynq.Client
}

type districtSeedCharacterResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

const districtSeedCharacterPromptTemplate = `
You are creating a fantasy RPG quest giver for a district seeding job.

Zone: %s
Zone description: %s
Quest template name: %s
Quest template description: %s
Quest giver tags: %s
Anchor point of interest: %s

Create one quest giver who feels native to the zone and clearly fits the quest giver tags.
Keep the character grounded in the zone and point of interest.

Respond ONLY as JSON:
{
  "name": "Character name",
  "description": "1-3 sentences"
}
`

func NewSeedDistrictProcessor(
	dbClient db.DbClient,
	deepPriest deep_priest.DeepPriest,
	dungeonmaster dungeonmaster.Client,
	locationSeeder locationseeder.Client,
	asyncClient *asynq.Client,
) SeedDistrictProcessor {
	log.Println("Initializing SeedDistrictProcessor")
	return SeedDistrictProcessor{
		dbClient:       dbClient,
		deepPriest:     deepPriest,
		dungeonmaster:  dungeonmaster,
		locationSeeder: locationSeeder,
		asyncClient:    asyncClient,
	}
}

func (p *SeedDistrictProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing district seed task: %v", task.Type())

	var payload jobs.SeedDistrictTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := p.dbClient.DistrictSeedJob().FindByID(ctx, payload.JobID)
	if err != nil {
		return err
	}
	if job == nil {
		return nil
	}
	if job.Status == models.DistrictSeedJobStatusCompleted {
		return nil
	}

	job.Status = models.DistrictSeedJobStatusInProgress
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.DistrictSeedJob().Update(ctx, job); err != nil {
		return err
	}

	district, err := p.dbClient.District().FindByID(ctx, job.DistrictID)
	if err != nil {
		return p.failDistrictSeedJob(ctx, job, fmt.Errorf("failed to load district: %w", err))
	}
	if district == nil {
		return p.failDistrictSeedJob(ctx, job, fmt.Errorf("district not found"))
	}
	if len(district.Zones) == 0 {
		return p.failDistrictSeedJob(ctx, job, fmt.Errorf("district has no zones"))
	}

	if job.ZoneSeedSettings.HasContent() && len(job.ZoneSeedJobIDs) == 0 {
		zoneSeedJobIDs, err := p.queueZoneSeedJobsForDistrict(ctx, district.Zones, job.ZoneSeedSettings)
		if err != nil {
			return p.failDistrictSeedJob(ctx, job, fmt.Errorf("failed to queue district zone seed jobs: %w", err))
		}
		job.ZoneSeedJobIDs = models.StringArray(zoneSeedJobIDs)
		job.UpdatedAt = time.Now()
		if err := p.dbClient.DistrictSeedJob().Update(ctx, job); err != nil {
			return err
		}
	}

	results := job.Results
	if len(results) == 0 && len(job.QuestArchetypeIDs) > 0 {
		results = make(models.DistrictSeedResults, 0, len(job.QuestArchetypeIDs))
		for _, questArchetypeID := range job.QuestArchetypeIDs {
			results = append(results, models.DistrictSeedResult{
				QuestArchetypeID: questArchetypeID,
				Status:           models.DistrictSeedResultStatusQueued,
			})
		}
	}

	failedCount := 0
	for idx := range results {
		if results[idx].Status == models.DistrictSeedResultStatusCompleted && results[idx].QuestID != nil {
			continue
		}

		updatedResult, resultErr := p.processDistrictSeedResult(ctx, district, results[idx])
		results[idx] = updatedResult
		job.Results = results
		job.UpdatedAt = time.Now()
		if updateErr := p.dbClient.DistrictSeedJob().Update(ctx, job); updateErr != nil {
			return updateErr
		}
		if resultErr != nil {
			failedCount++
		}
	}

	job.Results = results
	job.UpdatedAt = time.Now()
	finalizeDistrictSeedJob(job, failedCount, len(results))
	return p.dbClient.DistrictSeedJob().Update(ctx, job)
}

func (p *SeedDistrictProcessor) queueZoneSeedJobsForDistrict(
	ctx context.Context,
	zones []models.Zone,
	settings models.DistrictZoneSeedSettings,
) ([]string, error) {
	if !settings.HasContent() || len(zones) == 0 {
		return []string{}, nil
	}

	jobIDs := make([]string, 0, len(zones))
	for _, zone := range zones {
		job := &models.ZoneSeedJob{
			ID:                   uuid.New(),
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
			ZoneID:               zone.ID,
			ZoneKind:             settings.ZoneKind,
			Status:               models.ZoneSeedStatusQueued,
			SeedMode:             models.ZoneSeedModeManual,
			CountMode:            models.ZoneSeedCountModeAbsolute,
			PlaceCount:           settings.PlaceCount,
			CharacterCount:       0,
			QuestCount:           0,
			MainQuestCount:       0,
			MonsterCount:         settings.MonsterCount,
			BossEncounterCount:   settings.BossEncounterCount,
			RaidEncounterCount:   settings.RaidEncounterCount,
			InputEncounterCount:  settings.InputEncounterCount,
			OptionEncounterCount: settings.OptionEncounterCount,
			TreasureChestCount:   settings.TreasureChestCount,
			HealingFountainCount: settings.HealingFountainCount,
			RequiredPlaceTags:    settings.RequiredPlaceTags,
			ShopkeeperItemTags:   settings.ShopkeeperItemTags,
		}
		if err := p.dbClient.ZoneSeedJob().Create(ctx, job); err != nil {
			return nil, err
		}

		payloadBytes, err := json.Marshal(jobs.SeedZoneDraftTaskPayload{JobID: job.ID})
		if err != nil {
			return nil, err
		}
		if _, err := p.asyncClient.Enqueue(asynq.NewTask(jobs.SeedZoneDraftTaskType, payloadBytes)); err != nil {
			return nil, err
		}

		jobIDs = append(jobIDs, job.ID.String())
	}

	return jobIDs, nil
}

func finalizeDistrictSeedJob(job *models.DistrictSeedJob, failedCount int, total int) {
	if job == nil {
		return
	}
	if failedCount > 0 {
		log.Printf(
			"District seed job %s completed with %d failed quest template(s) out of %d",
			job.ID,
			failedCount,
			total,
		)
	}
	job.Status = models.DistrictSeedJobStatusCompleted
	job.ErrorMessage = nil
}

func (p *SeedDistrictProcessor) failDistrictSeedJob(ctx context.Context, job *models.DistrictSeedJob, err error) error {
	msg := err.Error()
	job.Status = models.DistrictSeedJobStatusFailed
	job.ErrorMessage = &msg
	job.UpdatedAt = time.Now()
	if updateErr := p.dbClient.DistrictSeedJob().Update(ctx, job); updateErr != nil {
		log.Printf("Failed to update district seed job failure state: %v", updateErr)
	}
	return err
}

func (p *SeedDistrictProcessor) processDistrictSeedResult(
	ctx context.Context,
	district *models.District,
	result models.DistrictSeedResult,
) (models.DistrictSeedResult, error) {
	result.Status = models.DistrictSeedResultStatusQueued
	result.ErrorMessage = nil

	questArchetypeID, err := uuid.Parse(strings.TrimSpace(result.QuestArchetypeID))
	if err != nil {
		msg := "invalid quest archetype ID"
		result.Status = models.DistrictSeedResultStatusFailed
		result.ErrorMessage = &msg
		return result, fmt.Errorf("%s", msg)
	}

	questArchetype, err := p.dbClient.QuestArchetype().FindByID(ctx, questArchetypeID)
	if err != nil {
		result.Status = models.DistrictSeedResultStatusFailed
		msg := fmt.Sprintf("failed to load quest archetype: %v", err)
		result.ErrorMessage = &msg
		return result, err
	}
	if questArchetype == nil {
		msg := "quest archetype not found"
		result.Status = models.DistrictSeedResultStatusFailed
		result.ErrorMessage = &msg
		return result, fmt.Errorf("%s", msg)
	}
	result.QuestArchetypeName = questArchetype.Name

	zone, matchCount, err := selectBestDistrictSeedZone(district.Zones, questArchetype.InternalTags)
	if err != nil {
		result.Status = models.DistrictSeedResultStatusFailed
		msg := err.Error()
		result.ErrorMessage = &msg
		return result, err
	}
	rankedZones, err := rankDistrictSeedZones(district.Zones, questArchetype.InternalTags)
	if err != nil {
		result.Status = models.DistrictSeedResultStatusFailed
		msg := err.Error()
		result.ErrorMessage = &msg
		return result, err
	}

	orderedZones := make([]districtSeedZoneCandidate, 0, len(rankedZones))
	orderedZones = append(orderedZones, districtSeedZoneCandidate{
		zone:       zone,
		matchCount: matchCount,
	})
	for _, candidate := range rankedZones {
		if candidate.zone == nil || candidate.zone.ID == zone.ID {
			continue
		}
		orderedZones = append(orderedZones, candidate)
	}

	var lastErr error
	for zoneIndex, candidate := range orderedZones {
		if candidate.zone == nil {
			continue
		}
		applyDistrictSeedZoneToResult(&result, candidate.zone, candidate.matchCount)

		questGiverCharacterID, questGiverName, generatedCharacterID, generatedCharacterName, err := p.ensureQuestGiverCharacter(
			ctx,
			candidate.zone,
			questArchetype,
		)
		if err != nil {
			result.Status = models.DistrictSeedResultStatusFailed
			msg := fmt.Sprintf("failed to ensure quest giver: %v", err)
			result.ErrorMessage = &msg
			return result, err
		}
		applyDistrictSeedQuestGiverToResult(
			&result,
			questGiverCharacterID,
			questGiverName,
			generatedCharacterID,
			generatedCharacterName,
		)

		quest, err := p.dungeonmaster.GenerateQuest(
			ctx,
			candidate.zone,
			questArchetype.ID,
			questGiverCharacterID,
		)
		if err == nil {
			if quest == nil {
				lastErr = fmt.Errorf("quest generation returned no quest")
				break
			}
			questID := quest.ID.String()
			result.QuestID = &questID
			result.Status = models.DistrictSeedResultStatusCompleted
			result.ErrorMessage = nil
			return result, nil
		}

		lastErr = err
		if shouldRetryDistrictSeedQuestInAnotherZone(err) && zoneIndex+1 < len(orderedZones) {
			log.Printf(
				"District seed retrying quest archetype %s in another zone after compatibility failure in zone %s: %v",
				questArchetype.ID,
				candidate.zone.ID,
				err,
			)
			continue
		}
		break
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("quest generation returned no quest")
	}
	result.Status = models.DistrictSeedResultStatusFailed
	msg := fmt.Sprintf("failed to generate quest: %v", lastErr)
	result.ErrorMessage = &msg
	return result, lastErr
}

type districtSeedZoneCandidate struct {
	zone       *models.Zone
	matchCount int
}

func applyDistrictSeedZoneToResult(
	result *models.DistrictSeedResult,
	zone *models.Zone,
	matchCount int,
) {
	if result == nil {
		return
	}
	result.ZoneID = nil
	result.ZoneName = ""
	result.MatchCount = matchCount
	if zone == nil {
		return
	}
	zoneID := zone.ID.String()
	result.ZoneID = &zoneID
	result.ZoneName = zone.Name
}

func applyDistrictSeedQuestGiverToResult(
	result *models.DistrictSeedResult,
	questGiverCharacterID *uuid.UUID,
	questGiverName string,
	generatedCharacterID *uuid.UUID,
	generatedCharacterName string,
) {
	if result == nil {
		return
	}
	result.QuestGiverCharacterID = nil
	result.QuestGiverCharacterName = ""
	result.GeneratedCharacterID = nil
	result.GeneratedCharacterName = ""
	if questGiverCharacterID != nil {
		id := questGiverCharacterID.String()
		result.QuestGiverCharacterID = &id
		result.QuestGiverCharacterName = questGiverName
	}
	if generatedCharacterID != nil {
		id := generatedCharacterID.String()
		result.GeneratedCharacterID = &id
		result.GeneratedCharacterName = generatedCharacterName
	}
}

func shouldRetryDistrictSeedQuestInAnotherZone(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "no points of interest found for location archetype") ||
		strings.Contains(message, "no unused points of interest found for location archetype")
}

func selectBestDistrictSeedZone(zones []models.Zone, questInternalTags models.StringArray) (*models.Zone, int, error) {
	if len(zones) == 0 {
		return nil, 0, fmt.Errorf("district has no zones")
	}

	desiredTags := normalizeDistrictSeedTags(questInternalTags)
	bestScore := -1
	bestIndexes := make([]int, 0, len(zones))
	for idx := range zones {
		score := districtSeedMatchCount(zones[idx].InternalTags, desiredTags)
		if score > bestScore {
			bestScore = score
			bestIndexes = []int{idx}
			continue
		}
		if score == bestScore {
			bestIndexes = append(bestIndexes, idx)
		}
	}

	selectedIndex := bestIndexes[rand.Intn(len(bestIndexes))]
	return &zones[selectedIndex], bestScore, nil
}

func rankDistrictSeedZones(
	zones []models.Zone,
	questInternalTags models.StringArray,
) ([]districtSeedZoneCandidate, error) {
	if len(zones) == 0 {
		return nil, fmt.Errorf("district has no zones")
	}

	desiredTags := normalizeDistrictSeedTags(questInternalTags)
	ranked := make([]districtSeedZoneCandidate, 0, len(zones))
	for idx := range zones {
		ranked = append(ranked, districtSeedZoneCandidate{
			zone:       &zones[idx],
			matchCount: districtSeedMatchCount(zones[idx].InternalTags, desiredTags),
		})
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		return ranked[i].matchCount > ranked[j].matchCount
	})
	return ranked, nil
}

func districtSeedMatchCount(tags models.StringArray, desired map[string]struct{}) int {
	if len(desired) == 0 {
		return 0
	}
	count := 0
	seen := map[string]struct{}{}
	for _, rawTag := range []string(tags) {
		tag := strings.ToLower(strings.TrimSpace(rawTag))
		if tag == "" {
			continue
		}
		if _, exists := desired[tag]; !exists {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		count++
	}
	return count
}

func normalizeDistrictSeedTags(tags models.StringArray) map[string]struct{} {
	normalized := make(map[string]struct{}, len(tags))
	for _, rawTag := range []string(tags) {
		tag := strings.ToLower(strings.TrimSpace(rawTag))
		if tag == "" {
			continue
		}
		normalized[tag] = struct{}{}
	}
	return normalized
}

func (p *SeedDistrictProcessor) ensureQuestGiverCharacter(
	ctx context.Context,
	zone *models.Zone,
	questArchetype *models.QuestArchetype,
) (*uuid.UUID, string, *uuid.UUID, string, error) {
	desiredTags := normalizeDistrictSeedTags(questArchetype.CharacterTags)
	if len(desiredTags) == 0 {
		return nil, "", nil, "", nil
	}

	pointsOfInterest, err := p.dbClient.PointOfInterest().FindAllForZone(ctx, zone.ID)
	if err != nil {
		return nil, "", nil, "", err
	}
	pointOfInterestIDs := make(map[uuid.UUID]struct{}, len(pointsOfInterest))
	for _, poi := range pointsOfInterest {
		pointOfInterestIDs[poi.ID] = struct{}{}
	}

	characters, err := p.dbClient.Character().FindAll(ctx)
	if err != nil {
		return nil, "", nil, "", err
	}

	type candidate struct {
		character *models.Character
		score     int
		priority  int
	}
	bestScore := 0
	bestPriority := 2
	candidates := make([]candidate, 0)
	for _, character := range characters {
		if character == nil {
			continue
		}
		score := districtSeedMatchCount(character.InternalTags, desiredTags)
		if score == 0 {
			continue
		}

		priority := 2
		if character.PointOfInterestID != nil {
			if _, ok := pointOfInterestIDs[*character.PointOfInterestID]; ok {
				priority = 0
			}
		}
		if priority == 2 && districtSeedCharacterInZone(zone, character) {
			priority = 1
		}
		if priority == 2 {
			continue
		}

		if score > bestScore || (score == bestScore && priority < bestPriority) {
			bestScore = score
			bestPriority = priority
			candidates = []candidate{{character: character, score: score, priority: priority}}
			continue
		}
		if score == bestScore && priority == bestPriority {
			candidates = append(candidates, candidate{character: character, score: score, priority: priority})
		}
	}

	if len(candidates) > 0 {
		selected := candidates[rand.Intn(len(candidates))].character
		return &selected.ID, selected.Name, nil, "", nil
	}

	anchorPOI, err := p.ensureDistrictSeedAnchorPOI(ctx, zone, questArchetype)
	if err != nil {
		return nil, "", nil, "", err
	}

	generated, err := p.generateDistrictSeedCharacter(ctx, zone, questArchetype, anchorPOI)
	if err != nil {
		return nil, "", nil, "", err
	}

	character, err := p.createDistrictSeedCharacter(ctx, zone, anchorPOI, generated, questArchetype.CharacterTags)
	if err != nil {
		return nil, "", nil, "", err
	}

	return &character.ID, character.Name, &character.ID, character.Name, nil
}

func districtSeedCharacterInZone(zone *models.Zone, character *models.Character) bool {
	if zone == nil || character == nil {
		return false
	}
	for _, location := range character.Locations {
		if zone.IsPointInBoundary(location.Latitude, location.Longitude) {
			return true
		}
	}
	return false
}

func (p *SeedDistrictProcessor) ensureDistrictSeedAnchorPOI(
	ctx context.Context,
	zone *models.Zone,
	questArchetype *models.QuestArchetype,
) (*models.PointOfInterest, error) {
	if zone == nil {
		return nil, nil
	}

	if questArchetype != nil &&
		questArchetype.Root.LocationArchetypeID != nil &&
		*questArchetype.Root.LocationArchetypeID != uuid.Nil &&
		questArchetype.Root.LocationArchetype != nil {
		pois, err := p.locationSeeder.SeedPointsOfInterest(
			ctx,
			*zone,
			questArchetype.Root.LocationArchetype.IncludedTypes,
			questArchetype.Root.LocationArchetype.ExcludedTypes,
			1,
			nil,
		)
		if err == nil && len(pois) > 0 {
			return pois[0], nil
		}
		if err != nil {
			log.Printf("District seed failed to pre-seed POI for quest giver: %v", err)
		}
	}

	pointsOfInterest, err := p.dbClient.PointOfInterest().FindAllForZone(ctx, zone.ID)
	if err != nil {
		return nil, err
	}
	if len(pointsOfInterest) == 0 {
		return nil, nil
	}
	selected := pointsOfInterest[rand.Intn(len(pointsOfInterest))]
	return &selected, nil
}

func (p *SeedDistrictProcessor) generateDistrictSeedCharacter(
	ctx context.Context,
	zone *models.Zone,
	questArchetype *models.QuestArchetype,
	poi *models.PointOfInterest,
) (districtSeedCharacterResponse, error) {
	anchorName := "None"
	if poi != nil && strings.TrimSpace(poi.Name) != "" {
		anchorName = poi.Name
	}

	prompt := fmt.Sprintf(
		districtSeedCharacterPromptTemplate,
		zoneNameForPrompt(zone),
		zoneDescriptionForPrompt(zone),
		strings.TrimSpace(questArchetype.Name),
		strings.TrimSpace(questArchetype.Description),
		strings.Join([]string(questArchetype.CharacterTags), ", "),
		anchorName,
	)

	answer, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return fallbackDistrictSeedCharacter(zone, questArchetype, poi), nil
	}

	var response districtSeedCharacterResponse
	if err := json.Unmarshal([]byte(extractJSON(answer.Answer)), &response); err != nil {
		return fallbackDistrictSeedCharacter(zone, questArchetype, poi), nil
	}

	response.Name = strings.TrimSpace(response.Name)
	response.Description = strings.TrimSpace(response.Description)
	if response.Name == "" || response.Description == "" {
		return fallbackDistrictSeedCharacter(zone, questArchetype, poi), nil
	}
	return response, nil
}

func zoneNameForPrompt(zone *models.Zone) string {
	if zone == nil || strings.TrimSpace(zone.Name) == "" {
		return "Unnamed zone"
	}
	return strings.TrimSpace(zone.Name)
}

func zoneDescriptionForPrompt(zone *models.Zone) string {
	if zone == nil || strings.TrimSpace(zone.Description) == "" {
		return "No zone description provided."
	}
	return strings.TrimSpace(zone.Description)
}

func fallbackDistrictSeedCharacter(
	zone *models.Zone,
	questArchetype *models.QuestArchetype,
	poi *models.PointOfInterest,
) districtSeedCharacterResponse {
	tagLabel := "local guide"
	if len(questArchetype.CharacterTags) > 0 {
		tagLabel = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(questArchetype.CharacterTags[0])), "_", " ")
	}
	anchorName := "the district"
	if poi != nil && strings.TrimSpace(poi.Name) != "" {
		anchorName = strings.TrimSpace(poi.Name)
	}
	zoneName := zoneNameForPrompt(zone)

	return districtSeedCharacterResponse{
		Name: generateZoneSeedShopkeeperName(tagLabel, zoneName, map[string]struct{}{}),
		Description: fmt.Sprintf(
			"A %s who keeps watch near %s and quietly steers adventurers toward troubles stirring in %s.",
			tagLabel,
			anchorName,
			zoneName,
		),
	}
}

func (p *SeedDistrictProcessor) createDistrictSeedCharacter(
	ctx context.Context,
	zone *models.Zone,
	poi *models.PointOfInterest,
	generated districtSeedCharacterResponse,
	characterTags models.StringArray,
) (*models.Character, error) {
	character := &models.Character{
		Name:                  generated.Name,
		Description:           generated.Description,
		InternalTags:          normalizeCharacterInternalTagsForSeed(characterTags),
		ImageGenerationStatus: models.CharacterImageGenerationStatusQueued,
	}
	if poi != nil {
		character.PointOfInterestID = &poi.ID
	}

	if err := p.dbClient.Character().Create(ctx, character); err != nil {
		return nil, err
	}

	if poi == nil && zone != nil {
		point := zone.GetRandomPoint()
		if err := p.dbClient.CharacterLocation().ReplaceForCharacter(ctx, character.ID, []models.CharacterLocation{
			{
				Latitude:  point.Y(),
				Longitude: point.X(),
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

func normalizeCharacterInternalTagsForSeed(tags models.StringArray) models.StringArray {
	normalized := make(models.StringArray, 0, len(tags))
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
	return normalized
}
