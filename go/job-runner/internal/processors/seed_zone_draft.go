package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
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
		Name               string   `json:"name"`
		Description        string   `json:"description"`
		AcceptanceDialogue []string `json:"acceptanceDialogue"`
		QuestGiverDraftID  string   `json:"questGiverDraftId"`
		PlaceID            string   `json:"placeId"`
		ChallengeQuestion  string   `json:"challengeQuestion"`
		Gold               *int     `json:"gold,omitempty"`
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
- Include a short challengeQuestion for the player

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
      "gold": 10
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

		gold := 0
		if draft.Gold != nil {
			gold = *draft.Gold
		}

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

		quests = append(quests, models.ZoneSeedQuestDraft{
			DraftID:            uuid.New(),
			Name:               name,
			Description:        description,
			AcceptanceDialogue: dialogue,
			PlaceID:            placeID,
			QuestGiverDraftID:  questGiverID,
			ChallengeQuestion:  strings.TrimSpace(draft.ChallengeQuestion),
			Gold:               gold,
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

	radius := zone.Radius
	if radius < 500 {
		radius = 500
	}
	if radius > 5000 {
		radius = 5000
	}

	seen := make(map[string]googlemaps.Place)
	maxAttempts := 6
	for attempt := 0; attempt < maxAttempts && len(seen) < desired; attempt++ {
		point := zone.GetRandomPoint()
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

		for _, place := range places {
			if place.ID == "" {
				continue
			}
			if !zone.IsPointInBoundary(place.Location.Latitude, place.Location.Longitude) {
				continue
			}
			if _, ok := seen[place.ID]; ok {
				continue
			}
			seen[place.ID] = place
		}
	}

	results := make([]googlemaps.Place, 0, len(seen))
	for _, place := range seen {
		results = append(results, place)
	}

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
