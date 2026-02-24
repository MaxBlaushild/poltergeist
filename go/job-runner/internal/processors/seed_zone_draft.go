package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type SeedZoneDraftProcessor struct {
	dbClient         db.DbClient
	googlemapsClient googlemaps.Client
	deepPriest       deep_priest.DeepPriest
}

func NewSeedZoneDraftProcessor(
	dbClient db.DbClient,
	googlemapsClient googlemaps.Client,
	deepPriest deep_priest.DeepPriest,
) SeedZoneDraftProcessor {
	log.Println("Initializing SeedZoneDraftProcessor")
	return SeedZoneDraftProcessor{
		dbClient:         dbClient,
		googlemapsClient: googlemapsClient,
		deepPriest:       deepPriest,
	}
}

func (p *SeedZoneDraftProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing seed zone draft task: %v", task.Type())

	var payload jobs.SeedZoneDraftTaskPayload
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

	job.Status = models.ZoneSeedStatusInProgress
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ZoneSeedJob().Update(ctx, job); err != nil {
		log.Printf("Failed to update zone seed job status: %v", err)
		return err
	}

	zone, err := p.dbClient.Zone().FindByID(ctx, job.ZoneID)
	if err != nil {
		return p.failZoneSeedJob(ctx, job, fmt.Errorf("failed to find zone: %w", err))
	}
	if zone == nil {
		return p.failZoneSeedJob(ctx, job, fmt.Errorf("zone not found"))
	}

	placeCount := job.PlaceCount
	if placeCount <= 0 {
		placeCount = 6
	}

	places, err := p.findTopPlacesInZone(ctx, *zone, placeCount)
	if err != nil {
		return p.failZoneSeedJob(ctx, job, fmt.Errorf("failed to find top places: %w", err))
	}
	if len(places) == 0 {
		return p.failZoneSeedJob(ctx, job, fmt.Errorf("no places found in zone"))
	}

	branding, err := p.generateZoneBranding(ctx, *zone, places)
	if err != nil {
		return p.failZoneSeedJob(ctx, job, fmt.Errorf("failed to generate zone branding: %w", err))
	}

	characterCount := job.CharacterCount
	if characterCount <= 0 {
		characterCount = 4
	}

	characters, err := p.generateCharacters(ctx, *zone, branding, places, characterCount)
	if err != nil {
		return p.failZoneSeedJob(ctx, job, fmt.Errorf("failed to generate characters: %w", err))
	}

	questCount := job.QuestCount
	if questCount <= 0 {
		questCount = 4
	}

	quests, err := p.generateQuests(ctx, *zone, branding, places, characters, questCount)
	if err != nil {
		return p.failZoneSeedJob(ctx, job, fmt.Errorf("failed to generate quests: %w", err))
	}

	poiDrafts := make([]models.ZoneSeedPointOfInterestDraft, 0, len(places))
	for _, place := range places {
		poiDrafts = append(poiDrafts, models.ZoneSeedPointOfInterestDraft{
			DraftID:          uuid.New(),
			PlaceID:          place.ID,
			Name:             place.DisplayName.Text,
			Address:          place.FormattedAddress,
			Types:            place.Types,
			Latitude:         place.Location.Latitude,
			Longitude:        place.Location.Longitude,
			Rating:           place.Rating,
			UserRatingCount:  valueOrZero(place.UserRatingCount),
			EditorialSummary: place.EditorialSummary.Text,
		})
	}

	job.Draft = models.ZoneSeedDraft{
		FantasyName:      branding.FantasyName,
		ZoneDescription:  branding.ZoneDescription,
		PointsOfInterest: poiDrafts,
		Characters:       characters,
		Quests:           quests,
	}
	job.Status = models.ZoneSeedStatusAwaitingApproval
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ZoneSeedJob().Update(ctx, job); err != nil {
		log.Printf("Failed to update zone seed job with draft: %v", err)
		return err
	}

	log.Printf("Zone seed draft job %v completed", job.ID)
	return nil
}

func (p *SeedZoneDraftProcessor) failZoneSeedJob(ctx context.Context, job *models.ZoneSeedJob, err error) error {
	msg := err.Error()
	job.Status = models.ZoneSeedStatusFailed
	job.ErrorMessage = &msg
	job.UpdatedAt = time.Now()
	if updateErr := p.dbClient.ZoneSeedJob().Update(ctx, job); updateErr != nil {
		log.Printf("Failed to mark zone seed job failed: %v", updateErr)
	}
	return err
}

type zoneBrandingResponse struct {
	FantasyName     string `json:"fantasyName"`
	ZoneDescription string `json:"zoneDescription"`
}

type characterGenerationResponse struct {
	Characters []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		PlaceID     string `json:"placeId"`
	} `json:"characters"`
}

type questGenerationResponse struct {
	Quests []struct {
		Name                string   `json:"name"`
		Description         string   `json:"description"`
		AcceptanceDialogue  []string `json:"acceptanceDialogue"`
		QuestGiverDraftID   string   `json:"questGiverDraftId"`
		PlaceID             string   `json:"placeId"`
		ChallengeQuestion   string   `json:"challengeQuestion"`
		ChallengeDifficulty *int     `json:"challengeDifficulty,omitempty"`
		RewardItem          *struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			RarityTier  string `json:"rarityTier"`
		} `json:"rewardItem,omitempty"`
	} `json:"quests"`
}

const zoneBrandingPromptTemplate = `
You are a fantasy RPG worldbuilder tasked with rebranding a real-world neighborhood.

Neighborhood name: %s
Existing description (if any): %s

Top points of interest in this neighborhood:
%s

Create a fantasy district name and a vivid 1-2 paragraph description that captures the unique flavor of the neighborhood.
Keep the tone whimsical yet grounded in the POI list. Do not mention real-world brand names directly.

Respond ONLY as JSON:
{
  "fantasyName": "string",
  "zoneDescription": "string"
}
`

const characterGenerationPromptTemplate = `
You are a fantasy RPG character designer.

Fantasy district: %s
District description: %s

Points of interest (use only these placeIds):
%s

Create %d characters who belong in this district. Each character must be associated with one POI from the list.
Respond ONLY as JSON:
{
  "characters": [
    {
      "name": "string",
      "description": "string",
      "placeId": "string"
    }
  ]
}
`

const questGenerationPromptTemplate = `
You are a fantasy RPG quest designer.

Fantasy district: %s
District description: %s

Characters (use questGiverDraftId and prefer the quest giver's placeId for quest locations):
%s

Points of interest (use only these placeIds):
%s

Create %d quests that fit the district flavor. Each quest must:
- Use a quest giver from the character list (by questGiverDraftId)
- Reference a placeId from the POI list
- Include 3-6 short acceptance dialogue lines
- Include a short challengeQuestion for the player that can be completed by a single person on-site at the POI
  - Ignore fantasy flavor; base the challenge only on the real-world POI type
  - Safe, legal, respectful, and no purchase required; no restricted areas or staff interaction
  - Make it an enjoyable on-site activity (coffee shop: write a 4-line poem or sketch the mug; park: sketch a tree or count benches; museum: note a color motif; bookstore: find a title with a specific word)
  - Answerable on-site without external research
- Include a challengeDifficulty integer between 25 and 50 (inclusive)
- Include a rewardItem with a short name, 1-2 sentence description, and rarityTier (Common, Uncommon, Epic, Mythic)

Respond ONLY as JSON:
{
  "quests": [
    {
      "name": "string",
      "description": "string",
      "acceptanceDialogue": ["string"],
      "questGiverDraftId": "string",
      "placeId": "string",
      "challengeQuestion": "string",
      "challengeDifficulty": 35,
      "rewardItem": {
        "name": "string",
        "description": "string",
        "rarityTier": "Common"
      }
    }
  ]
}
`

func (p *SeedZoneDraftProcessor) generateZoneBranding(ctx context.Context, zone models.Zone, places []googlemaps.Place) (*zoneBrandingResponse, error) {
	prompt := fmt.Sprintf(
		zoneBrandingPromptTemplate,
		zone.Name,
		truncate(zone.Description, 300),
		formatPlacesForPrompt(places, 8),
	)

	answer, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return nil, err
	}

	var response zoneBrandingResponse
	if err := json.Unmarshal([]byte(extractJSON(answer.Answer)), &response); err != nil {
		return nil, err
	}

	response.FantasyName = strings.TrimSpace(response.FantasyName)
	response.ZoneDescription = strings.TrimSpace(response.ZoneDescription)
	if response.FantasyName == "" {
		response.FantasyName = zone.Name
	}
	if response.ZoneDescription == "" {
		return nil, fmt.Errorf("zone description was empty")
	}

	return &response, nil
}

func (p *SeedZoneDraftProcessor) generateCharacters(
	ctx context.Context,
	zone models.Zone,
	branding *zoneBrandingResponse,
	places []googlemaps.Place,
	count int,
) ([]models.ZoneSeedCharacterDraft, error) {
	prompt := fmt.Sprintf(
		characterGenerationPromptTemplate,
		branding.FantasyName,
		branding.ZoneDescription,
		formatPlacesForPrompt(places, 12),
		count,
	)

	answer, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return nil, err
	}

	var response characterGenerationResponse
	if err := json.Unmarshal([]byte(extractJSON(answer.Answer)), &response); err != nil {
		return nil, err
	}

	placeIDs := make(map[string]googlemaps.Place)
	for _, place := range places {
		placeIDs[place.ID] = place
	}

	characters := make([]models.ZoneSeedCharacterDraft, 0, len(response.Characters))
	for _, draft := range response.Characters {
		name := strings.TrimSpace(draft.Name)
		if name == "" {
			continue
		}
		description := strings.TrimSpace(draft.Description)
		placeID := strings.TrimSpace(draft.PlaceID)
		if _, ok := placeIDs[placeID]; !ok {
			placeID = pickFallbackPlaceID(places)
		}
		characters = append(characters, models.ZoneSeedCharacterDraft{
			DraftID:     uuid.New(),
			Name:        name,
			Description: description,
			PlaceID:     placeID,
		})
	}

	if len(characters) == 0 {
		return nil, fmt.Errorf("no valid characters generated")
	}

	return characters, nil
}

func (p *SeedZoneDraftProcessor) generateQuests(
	ctx context.Context,
	zone models.Zone,
	branding *zoneBrandingResponse,
	places []googlemaps.Place,
	characters []models.ZoneSeedCharacterDraft,
	count int,
) ([]models.ZoneSeedQuestDraft, error) {
	characterLines := make([]string, 0, len(characters))
	for _, character := range characters {
		characterLines = append(characterLines, fmt.Sprintf(
			"- %s | questGiverDraftId=%s | placeId=%s",
			character.Name,
			character.DraftID.String(),
			character.PlaceID,
		))
	}

	prompt := fmt.Sprintf(
		questGenerationPromptTemplate,
		branding.FantasyName,
		branding.ZoneDescription,
		strings.Join(characterLines, "\n"),
		formatPlacesForPrompt(places, 12),
		count,
	)

	answer, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return nil, err
	}

	var response questGenerationResponse
	if err := json.Unmarshal([]byte(extractJSON(answer.Answer)), &response); err != nil {
		return nil, err
	}

	placeIDs := make(map[string]googlemaps.Place)
	for _, place := range places {
		placeIDs[place.ID] = place
	}
	characterByID := make(map[uuid.UUID]models.ZoneSeedCharacterDraft)
	for _, character := range characters {
		characterByID[character.DraftID] = character
	}

	quests := make([]models.ZoneSeedQuestDraft, 0, len(response.Quests))
	for _, draft := range response.Quests {
		name := strings.TrimSpace(draft.Name)
		if name == "" {
			continue
		}
		description := strings.TrimSpace(draft.Description)

		questGiverID, err := uuid.Parse(strings.TrimSpace(draft.QuestGiverDraftID))
		if err != nil {
			questGiverID = pickFallbackCharacterID(characters)
		}
		questGiver, ok := characterByID[questGiverID]
		if !ok {
			questGiverID = pickFallbackCharacterID(characters)
			questGiver = characterByID[questGiverID]
		}

		placeID := strings.TrimSpace(draft.PlaceID)
		if _, ok := placeIDs[placeID]; !ok {
			placeID = questGiver.PlaceID
		}
		if questGiver.PlaceID != "" && placeID != questGiver.PlaceID {
			placeID = questGiver.PlaceID
		}
		if placeID == "" {
			placeID = pickFallbackPlaceID(places)
		}

		gold := 50 + rand.Intn(451)

		dialogue := make([]string, 0, len(draft.AcceptanceDialogue))
		for _, line := range draft.AcceptanceDialogue {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			dialogue = append(dialogue, trimmed)
		}
		if len(dialogue) > 6 {
			dialogue = dialogue[:6]
		}

		challengeQuestion := strings.TrimSpace(draft.ChallengeQuestion)
		if challengeQuestion == "" {
			challengeQuestion = fallbackSeedQuestChallengeQuestion(placeIDs[placeID])
		}
		challengeDifficulty := 0
		if draft.ChallengeDifficulty != nil {
			challengeDifficulty = *draft.ChallengeDifficulty
		}
		if challengeDifficulty <= 0 {
			challengeDifficulty = randomQuestDifficulty()
		}
		challengeDifficulty = clampQuestDifficulty(challengeDifficulty)

		rewardItem := buildQuestRewardItemDraft(draft.RewardItem, placeIDs[placeID], name)

		quests = append(quests, models.ZoneSeedQuestDraft{
			DraftID:             uuid.New(),
			Name:                name,
			Description:         description,
			AcceptanceDialogue:  dialogue,
			PlaceID:             placeID,
			QuestGiverDraftID:   questGiverID,
			ChallengeQuestion:   challengeQuestion,
			ChallengeDifficulty: challengeDifficulty,
			Gold:                gold,
			RewardItem:          rewardItem,
		})
	}

	if len(quests) == 0 {
		return nil, fmt.Errorf("no valid quests generated")
	}

	return quests, nil
}

func (p *SeedZoneDraftProcessor) findTopPlacesInZone(ctx context.Context, zone models.Zone, count int) ([]googlemaps.Place, error) {
	if count <= 0 {
		return []googlemaps.Place{}, nil
	}

	desired := count
	if desired < 3 {
		desired = 3
	}

	radius := placeSearchRadius(zone)
	if radius <= 0 {
		return []googlemaps.Place{}, nil
	}

	seen := make(map[string]googlemaps.Place)
	seenFallback := make(map[string]googlemaps.Place)
	maxAttempts := 6
	for attempt := 0; attempt < maxAttempts && len(seen) < desired; attempt++ {
		point := zone.GetRandomPoint()
		if point.X() == 0 && point.Y() == 0 {
			log.Printf("FindTopPlaces: zone %s (%s) returned empty random point on attempt %d", zone.ID, zone.Name, attempt+1)
		}
		places, err := p.googlemapsClient.FindPlaces(googlemaps.PlaceQuery{
			Lat:            point.Y(),
			Long:           point.X(),
			Radius:         radius,
			MaxResultCount: 20,
			RankPreference: googlemaps.RankPreferencePopularity,
		})
		if err != nil {
			return nil, err
		}

		inBoundaryCount := 0
		enjoyableCount := 0
		for _, place := range places {
			if place.ID == "" {
				continue
			}
			if !zone.IsPointInBoundary(place.Location.Latitude, place.Location.Longitude) {
				continue
			}
			inBoundaryCount++
			if _, ok := seen[place.ID]; ok {
				continue
			}
			if _, ok := seenFallback[place.ID]; !ok {
				seenFallback[place.ID] = place
			}
			if !isEnjoyablePlace(place) {
				continue
			}
			enjoyableCount++
			seen[place.ID] = place
		}
		log.Printf(
			"FindTopPlaces: zone %s (%s) attempt %d found %d places (%d in boundary, %d enjoyable). totals: enjoyable=%d fallback=%d",
			zone.ID,
			zone.Name,
			attempt+1,
			len(places),
			inBoundaryCount,
			enjoyableCount,
			len(seen),
			len(seenFallback),
		)
	}

	results := make([]googlemaps.Place, 0, len(seen))
	for _, place := range seen {
		results = append(results, place)
	}
	if len(results) == 0 && len(seenFallback) > 0 {
		log.Printf(
			"FindTopPlaces: zone %s (%s) had no enjoyable places, falling back to %d in-boundary places",
			zone.ID,
			zone.Name,
			len(seenFallback),
		)
		for _, place := range seenFallback {
			results = append(results, place)
		}
	}
	log.Printf(
		"FindTopPlaces: zone %s (%s) returning %d places (desired=%d, radius=%.0f)",
		zone.ID,
		zone.Name,
		len(results),
		count,
		radius,
	)

	sort.Slice(results, func(i, j int) bool {
		return scorePlace(results[i]) > scorePlace(results[j])
	})

	if len(results) > count {
		results = results[:count]
	}

	return results, nil
}

func scorePlace(place googlemaps.Place) float64 {
	count := float64(valueOrZero(place.UserRatingCount))
	return place.Rating * math.Log10(count+1)
}

func valueOrZero(value *int32) int32 {
	if value == nil {
		return 0
	}
	return *value
}

func formatPlacesForPrompt(places []googlemaps.Place, limit int) string {
	lines := make([]string, 0, len(places))
	for i, place := range places {
		if limit > 0 && i >= limit {
			break
		}
		summary := strings.TrimSpace(place.EditorialSummary.Text)
		if summary == "" {
			summary = place.PrimaryTypeDisplayName.Text
		}
		if summary == "" {
			summary = strings.Join(place.Types, ", ")
		}
		if summary == "" {
			summary = "local landmark"
		}
		line := fmt.Sprintf("- %s | placeId=%s | %s", place.DisplayName.Text, place.ID, truncate(summary, 120))
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func truncate(value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if max <= 0 || len(trimmed) <= max {
		return trimmed
	}
	return strings.TrimSpace(trimmed[:max]) + "..."
}

func extractJSON(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
	}
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end >= start {
		return trimmed[start : end+1]
	}
	return trimmed
}

func pickFallbackPlaceID(places []googlemaps.Place) string {
	if len(places) == 0 {
		return ""
	}
	return places[0].ID
}

func pickFallbackCharacterID(characters []models.ZoneSeedCharacterDraft) uuid.UUID {
	if len(characters) == 0 {
		return uuid.Nil
	}
	return characters[0].DraftID
}

func isEnjoyablePlace(place googlemaps.Place) bool {
	lowerTypes := normalizePlaceTypes(place)
	if len(lowerTypes) == 0 {
		return false
	}

	blocked := []string{
		"accounting", "insurance_agency", "real_estate_agency", "lawyer",
		"bank", "atm", "police", "fire_station", "post_office", "courthouse",
		"city_hall", "local_government_office",
		"doctor", "dentist", "hospital", "health", "pharmacy", "drugstore", "veterinary_care",
		"school", "primary_school", "secondary_school", "university",
		"gas_station", "car_repair", "car_wash", "parking", "storage",
		"train_station", "subway_station", "bus_station", "transit_station", "airport",
		"lodging",
	}
	if hasAnyType(lowerTypes, blocked) {
		return false
	}

	allowed := []string{
		"cafe", "coffee_shop", "bakery", "restaurant", "bar", "meal_takeaway", "meal_delivery",
		"ice_cream_shop", "dessert",
		"park", "garden", "playground", "trail", "hiking_area", "campground", "natural_feature",
		"beach", "marina", "harbor", "water",
		"museum", "art_gallery", "gallery", "tourist_attraction", "point_of_interest", "landmark",
		"library", "book_store", "movie_theater", "theater", "night_club", "music_venue",
		"stadium", "sports_complex", "gym",
		"zoo", "aquarium", "amusement_park", "bowling_alley",
		"plaza", "square", "bridge",
		"store", "shopping_mall", "market", "supermarket", "convenience_store",
		"clothing_store", "department_store", "electronics_store", "furniture_store", "home_goods_store",
		"pet_store", "florist", "spa", "beauty_salon", "barber", "hair_care",
	}

	if hasAnyType(lowerTypes, allowed) {
		return true
	}

	display := strings.ToLower(strings.TrimSpace(place.PrimaryTypeDisplayName.Text))
	if display != "" {
		keywords := []string{
			"cafe", "coffee", "bakery", "restaurant", "bar", "park", "garden", "museum",
			"gallery", "library", "book", "theater", "cinema", "market", "shop", "store",
			"plaza", "square", "trail", "beach", "zoo", "aquarium", "amusement",
		}
		for _, keyword := range keywords {
			if strings.Contains(display, keyword) {
				return true
			}
		}
	}

	return false
}

func normalizePlaceTypes(place googlemaps.Place) []string {
	result := make([]string, 0, len(place.Types)+1)
	for _, t := range place.Types {
		trimmed := strings.ToLower(strings.TrimSpace(t))
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if primary := strings.ToLower(strings.TrimSpace(place.PrimaryType)); primary != "" {
		result = append(result, primary)
	}
	return result
}

func hasAnyType(types []string, targets []string) bool {
	for _, t := range types {
		for _, target := range targets {
			if t == target || strings.Contains(t, target) {
				return true
			}
		}
	}
	return false
}

func placeSearchRadius(zone models.Zone) float64 {
	polygon := zone.GetPolygon()
	if polygon == nil {
		return 0
	}
	bounds := polygon.Bound()
	if bounds.IsEmpty() {
		return 0
	}

	centerLng := (bounds.Min.X() + bounds.Max.X()) / 2
	centerLat := (bounds.Min.Y() + bounds.Max.Y()) / 2

	corners := [][2]float64{
		{bounds.Min.Y(), bounds.Min.X()},
		{bounds.Min.Y(), bounds.Max.X()},
		{bounds.Max.Y(), bounds.Min.X()},
		{bounds.Max.Y(), bounds.Max.X()},
	}

	maxDistance := 0.0
	for _, corner := range corners {
		distance := util.HaversineDistance(centerLat, centerLng, corner[0], corner[1])
		if distance > maxDistance {
			maxDistance = distance
		}
	}

	if maxDistance < 500 {
		return 500
	}
	if maxDistance > 5000 {
		return 5000
	}
	return maxDistance
}

func fallbackSeedQuestChallengeQuestion(place googlemaps.Place) string {
	name := strings.TrimSpace(place.DisplayName.Text)
	if name == "" {
		return buildSeedHeuristicChallengeQuestion("this location", place.Types)
	}
	return buildSeedHeuristicChallengeQuestion(name, place.Types)
}

func buildQuestRewardItemDraft(
	reward *struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		RarityTier  string `json:"rarityTier"`
	},
	place googlemaps.Place,
	questName string,
) *models.ZoneSeedQuestRewardItemDraft {
	if reward != nil {
		name := strings.TrimSpace(reward.Name)
		description := strings.TrimSpace(reward.Description)
		rarity := normalizeRarityTier(reward.RarityTier)
		if name != "" {
			if description == "" {
				description = fallbackRewardItemDescription(place, questName)
			}
			if rarity == "" {
				rarity = randomRarityTier()
			}
			return &models.ZoneSeedQuestRewardItemDraft{
				Name:        name,
				Description: description,
				RarityTier:  rarity,
			}
		}
	}

	fallbackName := fallbackRewardItemName(questName)
	return &models.ZoneSeedQuestRewardItemDraft{
		Name:        fallbackName,
		Description: fallbackRewardItemDescription(place, questName),
		RarityTier:  randomRarityTier(),
	}
}

func fallbackRewardItemName(questName string) string {
	name := strings.TrimSpace(questName)
	if name == "" {
		return "Traveler's Token"
	}
	if len(name) > 40 {
		name = name[:40]
	}
	return fmt.Sprintf("%s Token", strings.TrimSpace(name))
}

func fallbackRewardItemDescription(place googlemaps.Place, questName string) string {
	poiName := strings.TrimSpace(place.DisplayName.Text)
	if poiName != "" {
		return fmt.Sprintf("A keepsake earned near %s, warm with the memory of the place.", poiName)
	}
	if questName != "" {
		return fmt.Sprintf("A keepsake earned after completing %s.", strings.TrimSpace(questName))
	}
	return "A keepsake earned by completing a local favor."
}

func buildSeedHeuristicChallengeQuestion(name string, types []string) string {
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
