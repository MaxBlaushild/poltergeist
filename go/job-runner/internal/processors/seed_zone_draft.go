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
	"github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster/poilocals"
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
	if zone.Boundary != "" {
		log.Printf("SeedZoneDraft: zone %s (%s) boundary length=%d", zone.ID, zone.Name, len(zone.Boundary))
	} else {
		log.Printf("SeedZoneDraft: zone %s (%s) boundary is empty", zone.ID, zone.Name)
	}
	log.Printf("SeedZoneDraft: zone %s (%s) boundary points=%d", zone.ID, zone.Name, len(zone.Points))
	polygon := zone.GetPolygon()
	if polygon == nil {
		log.Printf("SeedZoneDraft: zone %s (%s) polygon is nil", zone.ID, zone.Name)
	} else {
		bounds := polygon.Bound()
		log.Printf(
			"SeedZoneDraft: zone %s (%s) polygon bounds min=(%f,%f) max=(%f,%f)",
			zone.ID,
			zone.Name,
			bounds.Min.Y(),
			bounds.Min.X(),
			bounds.Max.Y(),
			bounds.Max.X(),
		)
	}

	placeCount := job.PlaceCount
	if placeCount < 0 {
		placeCount = 0
	}

	requiredTags := normalizeRequiredPlaceTags(job.RequiredPlaceTags)
	places, err := p.findTopPlacesInZone(ctx, *zone, placeCount, requiredTags)
	if err != nil {
		return p.failZoneSeedJob(ctx, job, fmt.Errorf("failed to find top places: %w", err))
	}
	if placeCount > 0 && len(places) == 0 {
		return p.failZoneSeedJob(ctx, job, fmt.Errorf("no places found in zone"))
	}

	branding := fallbackZoneBranding(*zone)
	if placeCount > 0 {
		branding, err = p.generateZoneBranding(ctx, *zone, places)
		if err != nil {
			return p.failZoneSeedJob(ctx, job, fmt.Errorf("failed to generate zone branding: %w", err))
		}
	}

	characters := []models.ZoneSeedCharacterDraft{}
	if len(places) > 0 {
		characters = p.generateCharacters(*zone, branding, places)
	}
	shopkeeperItemTags := normalizeZoneSeedShopkeeperItemTags(job.ShopkeeperItemTags)
	if len(shopkeeperItemTags) > 0 {
		characters = append(characters, generateZoneSeedShopkeepers(*zone, shopkeeperItemTags)...)
	}

	quests := []models.ZoneSeedQuestDraft{}
	mainQuests := []models.ZoneSeedMainQuestDraft{}

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
		MainQuests:       mainQuests,
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

type mainQuestNodeDraftResponse struct {
	Title               string `json:"title"`
	Story               string `json:"story"`
	PlaceID             string `json:"placeId"`
	ChallengeQuestion   string `json:"challengeQuestion"`
	ChallengeDifficulty *int   `json:"challengeDifficulty,omitempty"`
}

type mainQuestGenerationResponse struct {
	MainQuests []struct {
		Name               string   `json:"name"`
		Description        string   `json:"description"`
		AcceptanceDialogue []string `json:"acceptanceDialogue"`
		QuestGiverDraftID  string   `json:"questGiverDraftId"`
		RewardItem         *struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			RarityTier  string `json:"rarityTier"`
		} `json:"rewardItem,omitempty"`
		Nodes []mainQuestNodeDraftResponse `json:"nodes"`
	} `json:"mainQuests"`
}

type enjoyablePlaceFilterResponse struct {
	EnjoyablePlaceIDs []string `json:"enjoyablePlaceIds"`
}

const zoneBrandingPromptTemplate = `
You are a fantasy RPG worldbuilder tasked with rebranding a real-world neighborhood.

Neighborhood name: %s
Existing description (if any): %s

Top points of interest in this neighborhood:
%s

World naming context:
%s

Create a fantasy district name and a vivid 1-2 paragraph description that captures the unique flavor of the neighborhood.
Keep the tone whimsical yet grounded in the POI list. Do not mention real-world brand names directly.
Do not start the new name with an overused opening word from the world naming context, and avoid repetitive adjective-led naming patterns.

Respond ONLY as JSON:
{
  "fantasyName": "string",
  "zoneDescription": "string"
}
`

const mainQuestGenerationPromptTemplate = `
You are a narrative RPG quest designer creating multi-step "main quests."

Fantasy district: %s
District description: %s

Characters (use questGiverDraftId and prefer the quest giver's placeId for quest locations):
%s

Points of interest (use only these placeIds):
%s

Create %d main quests. Each main quest must:
- Be a longer story arc with a 2-3 paragraph description
- Use a quest giver from the character list (by questGiverDraftId)
- Include 4-8 short acceptance dialogue lines
- Include exactly 3 nodes (acts), in order, ideally at distinct POIs
- Each node must include a short title, 1-2 sentence story beat, a placeId, and a challengeQuestion
- Challenge questions must be real-world, single-player activities at the POI
  - Ignore fantasy flavor; base the challenge only on the real-world POI type
  - Safe, legal, respectful, and no restricted areas or staff interaction
  - Single-input only: EITHER a photo proof OR a short text response (1-2 sentences), never both
  - Require meaningful participation in the POI's core activity (not just approaching it)
  - Avoid knowledge-based or hard-to-verify prompts; prefer proof-of-participation tied to the main activity at the POI
  - Do NOT use signage-only prompts (storefront sign, menu board, entrance, marquee, poster, or facade) as the main proof
    (bookstore: pick a book and photograph it; comedy club: photograph the stage/lineup during a set; cafe: photograph a drink or menu choice)
  - If the POI is food/drink-focused, the challenge should involve getting a drink/food item and photographing the selected item
  - Enjoyable on-site activity; answerable without external research
- Include a challengeDifficulty integer between 25 and 50 (inclusive) for each node
- Include a rewardItem with a short name, 1-2 sentence description, and rarityTier (Common, Uncommon, Epic, Mythic)

Respond ONLY as JSON:
{
  "mainQuests": [
    {
      "name": "string",
      "description": "string",
      "acceptanceDialogue": ["string"],
      "questGiverDraftId": "string",
      "rewardItem": {
        "name": "string",
        "description": "string",
        "rarityTier": "Common"
      },
      "nodes": [
        {
          "title": "string",
          "story": "string",
          "placeId": "string",
          "challengeQuestion": "string",
          "challengeDifficulty": 35
        }
      ]
    }
  ]
}
`

const characterGenerationPromptTemplate = `
You are a fantasy RPG character designer.

Fantasy district: %s
District description: %s

Points of interest (use only these placeIds):
%s

Create %d characters who belong in this district. Each character must be associated with one POI from the list.
Ensure each character description includes distinct fantasy styling (attire, archetype, or magical motif) that reflects the district's fantasy branding.
Avoid modern streetwear, real-world brands, or purely contemporary descriptions.
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

const enjoyablePlaceFilterPromptTemplate = `
You are curating places that are enjoyable to stumble upon in a neighborhood.

Select places that people would enjoy visiting casually (cafes, parks, boutiques, bookstores, markets, museums, galleries, scenic spots, etc).
Exclude utilitarian/errand services (dentist, doctor, locksmith, hardware store, auto repair, banks, offices, government, storage, schools, gas, parking, etc).

Return ONLY JSON:
{
  "enjoyablePlaceIds": ["string"]
}

Places:
%s
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
  - Safe, legal, respectful, and no restricted areas or staff interaction
  - Single-input only: EITHER a photo proof OR a short text response (1-2 sentences), never both
  - Require meaningful participation in the POI's core activity (not just approaching it)
  - Avoid knowledge-based or hard-to-verify prompts; prefer proof-of-participation tied to the main activity at the POI
  - Do NOT use signage-only prompts (storefront sign, menu board, entrance, marquee, poster, or facade) as the main proof
    (bookstore: pick a book and photograph it; comedy club: photograph the stage/lineup during a set; cafe: photograph a drink or menu choice)
  - If the POI is food/drink-focused, the challenge should involve getting a drink/food item and photographing the selected item
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
	diversityContext := buildZoneNameDiversityContext(ctx, p.dbClient, zone.ID)
	basePrompt := fmt.Sprintf(
		zoneBrandingPromptTemplate,
		zone.Name,
		truncate(zone.Description, 300),
		formatPlacesForPrompt(places, 8),
		diversityContext.Guidance,
	)

	var (
		response zoneBrandingResponse
		lastErr  error
	)
	attempts := 1
	if len(diversityContext.ForbiddenLeadingRoots) > 0 {
		attempts = 3
	}

	for attempt := 0; attempt < attempts; attempt++ {
		attemptPrompt := basePrompt
		if attempt > 0 && len(diversityContext.ForbiddenLeadingRoots) > 0 {
			attemptPrompt = fmt.Sprintf(
				"%s\nAdditional correction:\n- The previous candidate still used a forbidden repeated opening root.\n- Absolutely do not begin the fantasy name with any variant of: %s.\n- Choose a different opening word and a distinct naming rhythm.\n",
				basePrompt,
				strings.Join(diversityContext.ForbiddenLeadingRoots, ", "),
			)
		}

		answer, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: attemptPrompt})
		if err != nil {
			lastErr = err
			continue
		}

		if err := json.Unmarshal([]byte(extractJSON(answer.Answer)), &response); err != nil {
			lastErr = err
			continue
		}

		response.FantasyName = strings.TrimSpace(response.FantasyName)
		response.ZoneDescription = strings.TrimSpace(response.ZoneDescription)
		if response.FantasyName == "" {
			response.FantasyName = zone.Name
		}
		if response.ZoneDescription == "" {
			lastErr = fmt.Errorf("zone description was empty")
			continue
		}
		if zoneNameUsesForbiddenLeadingRoot(response.FantasyName, diversityContext.ForbiddenLeadingRoots) {
			lastErr = fmt.Errorf("generated zone name %q used an overused leading root", response.FantasyName)
			continue
		}

		lastErr = nil
		break
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return &response, nil
}

func fallbackZoneBranding(zone models.Zone) *zoneBrandingResponse {
	fantasyName := strings.TrimSpace(zone.Name)
	if fantasyName == "" {
		fantasyName = "Untamed District"
	}
	description := strings.TrimSpace(zone.Description)
	if description == "" {
		description = fmt.Sprintf("%s is a shifting quarter where new stories and wandering merchants appear each day.", fantasyName)
	}
	return &zoneBrandingResponse{
		FantasyName:     fantasyName,
		ZoneDescription: description,
	}
}

func (p *SeedZoneDraftProcessor) generateCharacters(
	zone models.Zone,
	branding *zoneBrandingResponse,
	places []googlemaps.Place,
) []models.ZoneSeedCharacterDraft {
	placeContexts := make([]poilocals.PlaceContext, 0, len(places))
	for _, place := range places {
		placeContexts = append(placeContexts, poilocals.PlaceContext{
			ID:               place.ID,
			Name:             strings.TrimSpace(place.DisplayName.Text),
			OriginalName:     strings.TrimSpace(place.Name),
			Description:      strings.TrimSpace(place.EditorialSummary.Text),
			Address:          strings.TrimSpace(place.FormattedAddress),
			EditorialSummary: strings.TrimSpace(place.EditorialSummary.Text),
			Types:            append([]string{}, place.Types...),
		})
	}

	generated := poilocals.GenerateDrafts(
		p.deepPriest,
		poilocals.ZoneContext{
			Name:        strings.TrimSpace(branding.FantasyName),
			Description: strings.TrimSpace(branding.ZoneDescription),
		},
		placeContexts,
	)

	characters := make([]models.ZoneSeedCharacterDraft, 0, len(generated))
	for _, draft := range generated {
		placeID := strings.TrimSpace(draft.PlaceID)
		if placeID == "" {
			placeID = pickFallbackPlaceID(places)
		}
		characters = append(characters, models.ZoneSeedCharacterDraft{
			DraftID:     uuid.New(),
			Name:        strings.TrimSpace(draft.Name),
			Description: strings.TrimSpace(draft.Description),
			PlaceID:     placeID,
			Dialogue:    append([]string{}, draft.Dialogue...),
		})
	}

	return characters
}

func normalizeZoneSeedShopkeeperItemTags(input models.StringArray) []string {
	normalized := make([]string, 0, len(input))
	seen := map[string]struct{}{}
	for _, raw := range []string(input) {
		tag := strings.ToLower(strings.TrimSpace(raw))
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

func generateZoneSeedShopkeepers(zone models.Zone, tags []string) []models.ZoneSeedCharacterDraft {
	shopkeepers := make([]models.ZoneSeedCharacterDraft, 0, len(tags))
	usedNames := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		lat := zone.Latitude
		lng := zone.Longitude
		point := zone.GetRandomPoint()
		if point.Y() >= -90 && point.Y() <= 90 && point.X() >= -180 && point.X() <= 180 {
			lat = point.Y()
			lng = point.X()
		}

		tagLabel := humanizeShopkeeperTag(tag)
		description := fmt.Sprintf(
			"A traveling quartermaster who curates %s wares and barters with adventurers passing through %s.",
			tagLabel,
			strings.TrimSpace(zone.Name),
		)
		if strings.TrimSpace(zone.Name) == "" {
			description = fmt.Sprintf(
				"A traveling quartermaster who curates %s wares and barters with adventurers.",
				tagLabel,
			)
		}

		shopkeeperName := generateZoneSeedShopkeeperName(tag, zone.Name, usedNames)
		shopkeepers = append(shopkeepers, models.ZoneSeedCharacterDraft{
			DraftID:      uuid.New(),
			Name:         shopkeeperName,
			Description:  description,
			PlaceID:      "",
			Latitude:     float64Ptr(lat),
			Longitude:    float64Ptr(lng),
			ShopItemTags: models.StringArray{tag},
		})
	}
	return shopkeepers
}

func float64Ptr(v float64) *float64 {
	return &v
}

func humanizeShopkeeperTag(tag string) string {
	cleaned := strings.ReplaceAll(strings.TrimSpace(tag), "_", " ")
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	if cleaned == "" {
		return "specialized"
	}
	return strings.ToLower(cleaned)
}

func generateZoneSeedShopkeeperName(tag string, zoneName string, used map[string]struct{}) string {
	firstNames := []string{
		"Aria", "Bren", "Cassian", "Dahlia", "Elric", "Fiora", "Galen", "Helena",
		"Iris", "Jasper", "Kael", "Liora", "Marek", "Nadia", "Orin", "Petra",
		"Quinn", "Rowan", "Selene", "Theron", "Ulric", "Vera", "Wren", "Yara", "Zane",
	}
	lastNames := []string{
		"Amberfall", "Blackwood", "Cinder", "Dawnmere", "Emberlane", "Fairwind", "Graves",
		"Hawthorne", "Ironvale", "Jade", "Kingsley", "Larkspur", "Morrow", "North", "Oakhart",
		"Pryce", "Quill", "Ravencrest", "Storm", "Thorne", "Umber", "Vale", "West", "York", "Zephyr",
	}

	key := strings.ToLower(strings.TrimSpace(tag + "|" + zoneName))
	if key == "" {
		key = "shopkeeper"
	}
	hash := 0
	for _, r := range key {
		hash = (hash*31 + int(r)) % 1_000_000
	}

	for offset := 0; offset < len(firstNames)*len(lastNames); offset++ {
		first := firstNames[(hash+offset)%len(firstNames)]
		last := lastNames[(hash/len(firstNames)+offset*7)%len(lastNames)]
		name := first + " " + last
		if _, exists := used[name]; exists {
			continue
		}
		used[name] = struct{}{}
		return name
	}

	fallback := "Avery Merchant"
	used[fallback] = struct{}{}
	return fallback
}

func (p *SeedZoneDraftProcessor) requestCharacterDrafts(
	ctx context.Context,
	prompt string,
	attempts int,
) (characterGenerationResponse, error) {
	var response characterGenerationResponse
	requestPrompt := prompt
	for attempt := 1; attempt <= attempts; attempt++ {
		answer, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: requestPrompt})
		if err != nil {
			return response, err
		}
		parsed, parseErr := parseCharacterGenerationResponse(answer.Answer)
		if parseErr == nil {
			return parsed, nil
		}
		if attempt == attempts {
			return response, parseErr
		}
		requestPrompt = prompt + "\n\nReturn ONLY valid JSON with all braces/quotes closed. No markdown, no commentary."
	}
	return response, fmt.Errorf("failed to parse character response")
}

func parseCharacterGenerationResponse(raw string) (characterGenerationResponse, error) {
	var response characterGenerationResponse
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return response, fmt.Errorf("empty character response")
	}
	candidate := strings.TrimSpace(extractJSON(trimmed))
	if candidate == "" {
		candidate = trimmed
	}
	if err := json.Unmarshal([]byte(candidate), &response); err == nil {
		return response, nil
	} else {
		snippet := truncate(candidate, 400)
		if snippet == "" {
			snippet = truncate(trimmed, 400)
		}
		log.Printf("Character generation JSON parse failed: %v. Snippet: %s", err, snippet)
		return response, fmt.Errorf("character generation JSON parse failed: %w (snippet: %s)", err, snippet)
	}
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

		challenge := buildZoneSeedChallengeMetadata(
			placeIDs[placeID],
			draft.ChallengeQuestion,
			draft.ChallengeDifficulty,
		)

		rewardItem := buildQuestRewardItemDraft(draft.RewardItem, placeIDs[placeID], name)

		quests = append(quests, models.ZoneSeedQuestDraft{
			DraftID:             uuid.New(),
			Name:                name,
			Description:         description,
			AcceptanceDialogue:  dialogue,
			PlaceID:             placeID,
			QuestGiverDraftID:   questGiverID,
			ChallengeQuestion:   challenge.Question,
			ChallengeDifficulty: challenge.Difficulty,
			Gold:                gold,
			RewardItem:          rewardItem,
		})
	}

	if len(quests) == 0 {
		return nil, fmt.Errorf("no valid quests generated")
	}

	return quests, nil
}

func (p *SeedZoneDraftProcessor) generateMainQuests(
	ctx context.Context,
	zone models.Zone,
	branding *zoneBrandingResponse,
	places []googlemaps.Place,
	characters []models.ZoneSeedCharacterDraft,
	count int,
) ([]models.ZoneSeedMainQuestDraft, error) {
	if count <= 0 {
		return []models.ZoneSeedMainQuestDraft{}, nil
	}
	if countUniquePlaceIDs(places) < 3 {
		return nil, fmt.Errorf("not enough distinct places to build main quest nodes")
	}

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
		mainQuestGenerationPromptTemplate,
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

	var response mainQuestGenerationResponse
	if err := json.Unmarshal([]byte(extractJSON(answer.Answer)), &response); err != nil {
		return nil, err
	}

	placeByID := make(map[string]googlemaps.Place)
	for _, place := range places {
		placeByID[place.ID] = place
	}
	characterByID := make(map[uuid.UUID]models.ZoneSeedCharacterDraft)
	for _, character := range characters {
		characterByID[character.DraftID] = character
	}

	mainQuests := make([]models.ZoneSeedMainQuestDraft, 0, len(response.MainQuests))
	for _, draft := range response.MainQuests {
		name := strings.TrimSpace(draft.Name)
		if name == "" {
			continue
		}

		description := strings.TrimSpace(draft.Description)
		if description == "" {
			description = fmt.Sprintf("A three-part journey unfolds across %s.", branding.FantasyName)
		}

		questGiverID, err := uuid.Parse(strings.TrimSpace(draft.QuestGiverDraftID))
		if err != nil {
			questGiverID = pickFallbackCharacterID(characters)
		}
		if _, ok := characterByID[questGiverID]; !ok {
			questGiverID = pickFallbackCharacterID(characters)
		}

		nodes := normalizeMainQuestNodes(draft.Nodes, places, placeByID)
		if len(nodes) == 0 {
			nodes = buildFallbackMainQuestNodes(places, placeByID, 3)
		}
		if len(nodes) == 0 {
			continue
		}

		description = buildMainQuestDescription(description, nodes)

		dialogue := make([]string, 0, len(draft.AcceptanceDialogue))
		for _, line := range draft.AcceptanceDialogue {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			dialogue = append(dialogue, trimmed)
		}
		if len(dialogue) > 8 {
			dialogue = dialogue[:8]
		}

		gold := 200 + rand.Intn(601)

		rewardPlace := placeByID[nodes[len(nodes)-1].PlaceID]
		rewardItem := buildQuestRewardItemDraft(draft.RewardItem, rewardPlace, name)

		mainQuests = append(mainQuests, models.ZoneSeedMainQuestDraft{
			DraftID:            uuid.New(),
			Name:               name,
			Description:        description,
			AcceptanceDialogue: dialogue,
			QuestGiverDraftID:  questGiverID,
			Nodes:              nodes,
			Gold:               gold,
			RewardItem:         rewardItem,
		})
	}

	if len(mainQuests) == 0 {
		return nil, fmt.Errorf("no valid main quests generated")
	}

	return mainQuests, nil
}

func (p *SeedZoneDraftProcessor) findTopPlacesInZone(ctx context.Context, zone models.Zone, count int, requiredTags []string) ([]googlemaps.Place, error) {
	if count <= 0 {
		return []googlemaps.Place{}, nil
	}

	desired := count
	if desired < 3 {
		desired = 3
	}

	requiredTags = normalizeRequiredPlaceTags(requiredTags)

	radius := placeSearchRadius(zone)
	if radius <= 0 {
		return []googlemaps.Place{}, nil
	}

	seen := make(map[string]googlemaps.Place)
	seenFallback := make(map[string]googlemaps.Place)
	maxAttempts := 6
	if len(requiredTags) > 0 {
		maxAttempts = 12
	}
	missingTags := requiredTags
	for attempt := 0; attempt < maxAttempts; attempt++ {
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
		inBoundaryPlaces := make([]googlemaps.Place, 0, len(places))
		for _, place := range places {
			if place.ID == "" {
				continue
			}
			if !zone.IsPointInBoundary(place.Location.Latitude, place.Location.Longitude) {
				continue
			}
			inBoundaryCount++
			inBoundaryPlaces = append(inBoundaryPlaces, place)
		}

		llmEnjoyable := map[string]struct{}{}
		llmUsed := false
		if len(inBoundaryPlaces) > 0 && len(seen) < desired {
			ids, err := p.filterEnjoyablePlacesLLM(inBoundaryPlaces)
			if err != nil {
				log.Printf("FindTopPlaces: zone %s (%s) LLM filter failed: %v", zone.ID, zone.Name, err)
			} else if len(ids) > 0 {
				llmEnjoyable = ids
				llmUsed = true
				log.Printf(
					"FindTopPlaces: zone %s (%s) LLM filter selected %d of %d in-boundary places",
					zone.ID,
					zone.Name,
					len(llmEnjoyable),
					len(inBoundaryPlaces),
				)
			} else {
				log.Printf("FindTopPlaces: zone %s (%s) LLM filter returned 0 places; falling back to heuristic", zone.ID, zone.Name)
			}
		}

		for _, place := range inBoundaryPlaces {
			if _, ok := seen[place.ID]; ok {
				continue
			}
			if _, ok := seenFallback[place.ID]; !ok {
				seenFallback[place.ID] = place
			}
			matchesRequired := placeMatchesAnyTag(place, requiredTags)
			if llmUsed {
				if _, ok := llmEnjoyable[place.ID]; !ok && !matchesRequired {
					continue
				}
			} else {
				if !isEnjoyablePlace(place) && !matchesRequired {
					continue
				}
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
		if len(requiredTags) > 0 {
			missingTags = missingRequiredPlaceTags(requiredTags, seenFallback)
			if len(missingTags) > 0 {
				log.Printf("FindTopPlaces: zone %s (%s) missing required tags=%v after attempt %d", zone.ID, zone.Name, missingTags, attempt+1)
			}
		}

		if len(seen) >= desired && len(missingTags) == 0 {
			break
		}
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
	results = preferTopSpots(results)
	results = diversifyPlaces(results, count)

	requiredSelections := map[string]googlemaps.Place{}
	if len(requiredTags) > 0 {
		var missing []string
		requiredSelections, missing = p.searchForRequiredTags(ctx, zone, requiredTags, radius, seen, seenFallback)
		if len(missing) > 0 {
			log.Printf("TopSpots: missing required tags after targeted search=%v", missing)
		}
	}

	candidatePool := make([]googlemaps.Place, 0, len(seenFallback))
	for _, place := range seenFallback {
		candidatePool = append(candidatePool, place)
	}
	if len(candidatePool) == 0 {
		candidatePool = results
	}

	if len(requiredTags) > 0 {
		results = ensureRequiredSelections(requiredTags, requiredSelections, results, candidatePool, count)
	}

	if len(results) > count {
		results = results[:count]
	}

	return results, nil
}

func normalizeMainQuestNodes(
	nodes []mainQuestNodeDraftResponse,
	places []googlemaps.Place,
	placeByID map[string]googlemaps.Place,
) []models.ZoneSeedMainQuestNodeDraft {
	if len(nodes) == 0 || len(places) == 0 {
		return nil
	}

	used := make(map[string]struct{})
	results := make([]models.ZoneSeedMainQuestNodeDraft, 0, 3)

	for _, node := range nodes {
		if len(results) >= 3 {
			break
		}
		placeID := strings.TrimSpace(node.PlaceID)
		if placeID == "" || placeByID[placeID].ID == "" {
			placeID = ""
		}
		if placeID != "" {
			if _, ok := used[placeID]; ok {
				placeID = ""
			}
		}
		if placeID == "" {
			placeID = pickUnusedPlaceID(places, used)
		}
		if placeID == "" {
			placeID = pickFallbackPlaceID(places)
		}
		if placeID == "" {
			continue
		}
		if _, ok := used[placeID]; ok {
			continue
		}
		used[placeID] = struct{}{}

		challenge := buildZoneSeedChallengeMetadata(
			placeByID[placeID],
			node.ChallengeQuestion,
			node.ChallengeDifficulty,
		)

		results = append(results, models.ZoneSeedMainQuestNodeDraft{
			DraftID:             uuid.New(),
			OrderIndex:          len(results),
			Title:               strings.TrimSpace(node.Title),
			Story:               withFallbackStory(strings.TrimSpace(node.Story), len(results)),
			PlaceID:             placeID,
			ChallengeQuestion:   challenge.Question,
			ChallengeDifficulty: challenge.Difficulty,
		})
	}

	for len(results) < 3 {
		placeID := pickUnusedPlaceID(places, used)
		if placeID == "" {
			placeID = pickFallbackPlaceID(places)
		}
		if placeID == "" {
			break
		}
		used[placeID] = struct{}{}
		title := fmt.Sprintf("Chapter %d", len(results)+1)
		fallbackChallenge := regenerateZoneSeedChallengeMetadata(placeByID[placeID])
		results = append(results, models.ZoneSeedMainQuestNodeDraft{
			DraftID:             uuid.New(),
			OrderIndex:          len(results),
			Title:               title,
			Story:               withFallbackStory("", len(results)),
			PlaceID:             placeID,
			ChallengeQuestion:   fallbackChallenge.Question,
			ChallengeDifficulty: fallbackChallenge.Difficulty,
		})
	}

	if len(results) < 3 {
		return nil
	}

	return results
}

func buildFallbackMainQuestNodes(
	places []googlemaps.Place,
	placeByID map[string]googlemaps.Place,
	count int,
) []models.ZoneSeedMainQuestNodeDraft {
	if len(places) == 0 || count <= 0 {
		return nil
	}
	used := make(map[string]struct{})
	nodes := make([]models.ZoneSeedMainQuestNodeDraft, 0, count)
	for i := 0; i < count; i++ {
		placeID := pickUnusedPlaceID(places, used)
		if placeID == "" {
			placeID = pickFallbackPlaceID(places)
		}
		if placeID == "" {
			break
		}
		used[placeID] = struct{}{}
		challenge := regenerateZoneSeedChallengeMetadata(placeByID[placeID])
		nodes = append(nodes, models.ZoneSeedMainQuestNodeDraft{
			DraftID:             uuid.New(),
			OrderIndex:          i,
			Title:               fmt.Sprintf("Chapter %d", i+1),
			Story:               withFallbackStory("", i),
			PlaceID:             placeID,
			ChallengeQuestion:   challenge.Question,
			ChallengeDifficulty: challenge.Difficulty,
		})
	}
	if len(nodes) < count {
		return nil
	}
	return nodes
}

func pickUnusedPlaceID(places []googlemaps.Place, used map[string]struct{}) string {
	for _, place := range places {
		if place.ID == "" {
			continue
		}
		if _, ok := used[place.ID]; ok {
			continue
		}
		return place.ID
	}
	return ""
}

func countUniquePlaceIDs(places []googlemaps.Place) int {
	seen := make(map[string]struct{}, len(places))
	for _, place := range places {
		if place.ID == "" {
			continue
		}
		seen[place.ID] = struct{}{}
	}
	return len(seen)
}

func buildMainQuestDescription(base string, nodes []models.ZoneSeedMainQuestNodeDraft) string {
	trimmed := strings.TrimSpace(base)
	beats := make([]string, 0, len(nodes))
	for idx, node := range nodes {
		story := strings.TrimSpace(node.Story)
		if story == "" {
			continue
		}
		title := strings.TrimSpace(node.Title)
		if title == "" {
			title = fmt.Sprintf("Act %d", idx+1)
		}
		beats = append(beats, fmt.Sprintf("%s: %s", title, story))
	}
	if len(beats) == 0 {
		return trimmed
	}
	if trimmed == "" {
		return strings.Join(beats, "\n")
	}
	return trimmed + "\n\n" + strings.Join(beats, "\n")
}

func withFallbackStory(story string, index int) string {
	trimmed := strings.TrimSpace(story)
	if trimmed != "" {
		return trimmed
	}
	return fallbackMainQuestStory(index)
}

func fallbackMainQuestStory(index int) string {
	switch index {
	case 0:
		return "The first clue is uncovered, hinting at a larger mystery."
	case 1:
		return "A second lead deepens the trail and raises the stakes."
	default:
		return "The final piece clicks into place, bringing the tale to its close."
	}
}

func preferTopSpots(places []googlemaps.Place) []googlemaps.Place {
	if len(places) == 0 {
		return places
	}

	highReviews, midReviews, lowReviews := computeReviewThresholds(places)
	log.Printf("TopSpots: review thresholds high=%d mid=%d low=%d", highReviews, midReviews, lowReviews)
	filters := []struct {
		minRating  float64
		minReviews int32
		label      string
	}{
		{minRating: 4.5, minReviews: highReviews, label: fmt.Sprintf("4.5+ and %d+", highReviews)},
		{minRating: 4.3, minReviews: midReviews, label: fmt.Sprintf("4.3+ and %d+", midReviews)},
		{minRating: 4.0, minReviews: lowReviews, label: fmt.Sprintf("4.0+ and %d+", lowReviews)},
	}

	for _, filter := range filters {
		preferred := make([]googlemaps.Place, 0, len(places))
		fallback := make([]googlemaps.Place, 0, len(places))
		for _, place := range places {
			if place.Rating >= filter.minRating && valueOrZero(place.UserRatingCount) >= filter.minReviews {
				preferred = append(preferred, place)
			} else {
				fallback = append(fallback, place)
			}
		}
		if len(preferred) > 0 {
			log.Printf("TopSpots: using rating filter %s; preferred=%d fallback=%d", filter.label, len(preferred), len(fallback))
			return append(preferred, fallback...)
		}
	}

	return places
}

func normalizeRequiredPlaceTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.ToLower(strings.TrimSpace(tag))
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func applyRequiredPlaceTags(
	requiredTags []string,
	candidatePool []googlemaps.Place,
	base []googlemaps.Place,
	desired int,
) ([]googlemaps.Place, error) {
	normalized := normalizeRequiredPlaceTags(requiredTags)
	if len(normalized) == 0 {
		return base, nil
	}

	used := make(map[string]struct{})
	required := make([]googlemaps.Place, 0, len(normalized))
	missing := make([]string, 0)

	for _, tag := range normalized {
		best, ok := bestPlaceForTag(tag, candidatePool, used)
		if !ok {
			missing = append(missing, tag)
			continue
		}
		required = append(required, best)
		used[best.ID] = struct{}{}
	}

	if len(missing) > 0 {
		log.Printf("TopSpots: missing required place tags=%v; continuing without them", missing)
	}

	result := make([]googlemaps.Place, 0, minInt(desired, len(candidatePool)))
	result = append(result, required...)

	for _, place := range base {
		if len(result) >= desired {
			break
		}
		if place.ID == "" {
			continue
		}
		if _, ok := used[place.ID]; ok {
			continue
		}
		used[place.ID] = struct{}{}
		result = append(result, place)
	}

	if len(result) < desired {
		sort.Slice(candidatePool, func(i, j int) bool {
			return scorePlace(candidatePool[i]) > scorePlace(candidatePool[j])
		})
		for _, place := range candidatePool {
			if len(result) >= desired {
				break
			}
			if place.ID == "" {
				continue
			}
			if _, ok := used[place.ID]; ok {
				continue
			}
			used[place.ID] = struct{}{}
			result = append(result, place)
		}
	}

	log.Printf("TopSpots: enforcing required tags=%v", normalized)
	return result, nil
}

func ensureRequiredSelections(
	requiredTags []string,
	selections map[string]googlemaps.Place,
	base []googlemaps.Place,
	candidatePool []googlemaps.Place,
	desired int,
) []googlemaps.Place {
	if desired <= 0 {
		return base
	}

	result := make([]googlemaps.Place, 0, minInt(desired, len(candidatePool)))
	used := make(map[string]struct{})

	for _, tag := range requiredTags {
		place, ok := selections[tag]
		if !ok {
			continue
		}
		if place.ID == "" {
			continue
		}
		if _, ok := used[place.ID]; ok {
			continue
		}
		used[place.ID] = struct{}{}
		result = append(result, place)
	}

	for _, place := range base {
		if len(result) >= desired {
			break
		}
		if place.ID == "" {
			continue
		}
		if _, ok := used[place.ID]; ok {
			continue
		}
		used[place.ID] = struct{}{}
		result = append(result, place)
	}

	if len(result) < desired {
		sort.Slice(candidatePool, func(i, j int) bool {
			return scorePlace(candidatePool[i]) > scorePlace(candidatePool[j])
		})
		for _, place := range candidatePool {
			if len(result) >= desired {
				break
			}
			if place.ID == "" {
				continue
			}
			if _, ok := used[place.ID]; ok {
				continue
			}
			used[place.ID] = struct{}{}
			result = append(result, place)
		}
	}

	return result
}

func missingRequiredPlaceTags(requiredTags []string, pool map[string]googlemaps.Place) []string {
	if len(requiredTags) == 0 {
		return nil
	}
	if len(pool) == 0 {
		return requiredTags
	}

	places := make([]googlemaps.Place, 0, len(pool))
	for _, place := range pool {
		places = append(places, place)
	}

	missing := make([]string, 0)
	for _, tag := range requiredTags {
		found := false
		for _, place := range places {
			if placeMatchesTag(place, tag) {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, tag)
		}
	}
	return missing
}

func (p *SeedZoneDraftProcessor) searchForRequiredTags(
	ctx context.Context,
	zone models.Zone,
	tags []string,
	radius float64,
	seen map[string]googlemaps.Place,
	seenFallback map[string]googlemaps.Place,
) (map[string]googlemaps.Place, []string) {
	selections := make(map[string]googlemaps.Place)
	missing := make([]string, 0)

	for _, tag := range tags {
		types := requiredTagPlaceTypes(tag)
		if len(types) == 0 {
			missing = append(missing, tag)
			continue
		}

		bestScore := -1.0
		found := false
		var best googlemaps.Place

		for attempt := 0; attempt < 3; attempt++ {
			point := zone.GetRandomPoint()
			if point.X() == 0 && point.Y() == 0 {
				continue
			}
			places, err := p.googlemapsClient.FindPlaces(googlemaps.PlaceQuery{
				Lat:            point.Y(),
				Long:           point.X(),
				Radius:         radius,
				MaxResultCount: 20,
				IncludedTypes:  types,
				RankPreference: googlemaps.RankPreferencePopularity,
			})
			if err != nil {
				log.Printf("FindTopPlaces: zone %s (%s) required tag %s search failed: %v", zone.ID, zone.Name, tag, err)
				continue
			}
			for _, place := range places {
				if place.ID == "" {
					continue
				}
				if !zone.IsPointInBoundary(place.Location.Latitude, place.Location.Longitude) {
					continue
				}
				if _, ok := seenFallback[place.ID]; !ok {
					seenFallback[place.ID] = place
				}
				score := scorePlace(place)
				if !placeMatchesTag(place, tag) {
					continue
				}
				if !found || score > bestScore {
					best = place
					bestScore = score
					found = true
				}
			}
			if found {
				break
			}
		}

		if !found {
			missing = append(missing, tag)
			continue
		}

		selections[tag] = best
		if _, ok := seenFallback[best.ID]; !ok {
			seenFallback[best.ID] = best
		}
		if _, ok := seen[best.ID]; !ok {
			seen[best.ID] = best
		}
		log.Printf("FindTopPlaces: zone %s (%s) required tag %s selected %s", zone.ID, zone.Name, tag, best.DisplayName.Text)
	}

	return selections, missing
}

func requiredTagPlaceTypes(tag string) []googlemaps.PlaceType {
	switch strings.ToLower(strings.TrimSpace(tag)) {
	case "coffee_shop", "cafe", "coffee":
		return []googlemaps.PlaceType{
			googlemaps.TypeCoffeeShop,
			googlemaps.TypeCafe,
			googlemaps.TypeTeaHouse,
			googlemaps.TypeBakery,
		}
	case "park", "garden", "playground", "trail", "hiking", "natural_feature":
		return []googlemaps.PlaceType{
			googlemaps.TypePark,
			googlemaps.TypeGarden,
			googlemaps.TypePlayground,
			googlemaps.TypeDogPark,
			googlemaps.TypeHikingArea,
			googlemaps.TypePicnicGround,
			googlemaps.TypeBotanicalGarden,
		}
	case "museum", "gallery", "art_gallery":
		return []googlemaps.PlaceType{
			googlemaps.TypeMuseum,
			googlemaps.TypeArtGallery,
			googlemaps.TypeTouristAttraction,
		}
	case "library", "book_store", "bookstore":
		return []googlemaps.PlaceType{
			googlemaps.TypeLibrary,
			googlemaps.TypeBookStore,
		}
	case "market", "shopping", "store", "shopping_mall":
		return []googlemaps.PlaceType{
			googlemaps.TypeMarket,
			googlemaps.TypeShoppingMall,
			googlemaps.TypeStore,
			googlemaps.TypeSupermarket,
			googlemaps.TypeClothingStore,
		}
	case "scenic", "plaza", "square", "beach", "marina":
		return []googlemaps.PlaceType{
			googlemaps.TypePlaza,
			googlemaps.TypeBeach,
			googlemaps.TypeMarina,
		}
	case "theater", "cinema", "movie_theater":
		return []googlemaps.PlaceType{
			googlemaps.TypeMovieTheater,
			googlemaps.TypePerformingArtsTheater,
		}
	case "entertainment", "music_venue":
		return []googlemaps.PlaceType{
			googlemaps.TypeConcertHall,
			googlemaps.TypeAmphitheatre,
			googlemaps.TypeNightClub,
			googlemaps.TypeStadium,
			googlemaps.TypeAmusementPark,
			googlemaps.TypeZoo,
			googlemaps.TypeAquarium,
		}
	default:
		return nil
	}
}

func bestPlaceForTag(tag string, places []googlemaps.Place, used map[string]struct{}) (googlemaps.Place, bool) {
	var best googlemaps.Place
	found := false
	bestScore := -1.0
	for _, place := range places {
		if place.ID == "" {
			continue
		}
		if _, ok := used[place.ID]; ok {
			continue
		}
		if !placeMatchesTag(place, tag) {
			continue
		}
		score := scorePlace(place)
		if !found || score > bestScore {
			best = place
			bestScore = score
			found = true
		}
	}
	return best, found
}

func placeMatchesTag(place googlemaps.Place, tag string) bool {
	needle := strings.ToLower(strings.TrimSpace(tag))
	if needle == "" {
		return false
	}

	expanded := expandRequiredTagAliases(needle)

	types := normalizePlaceTypes(place)
	for _, t := range types {
		for _, alias := range expanded {
			if t == alias || strings.Contains(t, alias) {
				return true
			}
		}
	}

	display := strings.ToLower(strings.TrimSpace(place.PrimaryTypeDisplayName.Text))
	if display != "" {
		for _, alias := range expanded {
			if strings.Contains(display, alias) {
				return true
			}
		}
	}

	name := strings.ToLower(strings.TrimSpace(place.DisplayName.Text))
	if name != "" {
		for _, alias := range expanded {
			if strings.Contains(name, alias) {
				return true
			}
		}
	}

	return false
}

func placeMatchesAnyTag(place googlemaps.Place, tags []string) bool {
	if len(tags) == 0 {
		return false
	}
	for _, tag := range tags {
		if placeMatchesTag(place, tag) {
			return true
		}
	}
	return false
}

func expandRequiredTagAliases(tag string) []string {
	aliases := map[string][]string{
		"coffee_shop":   {"coffee_shop", "cafe", "coffee", "bakery", "espresso", "tea"},
		"cafe":          {"cafe", "coffee_shop", "coffee", "bakery", "espresso", "tea"},
		"park":          {"park", "garden", "playground", "trail", "hiking", "natural_feature"},
		"museum":        {"museum", "gallery", "art_gallery", "exhibit"},
		"gallery":       {"gallery", "art_gallery", "museum"},
		"library":       {"library", "book_store", "bookstore", "book"},
		"book_store":    {"book_store", "bookstore", "library", "book"},
		"market":        {"market", "shopping_mall", "store", "shopping", "supermarket"},
		"shopping":      {"shopping_mall", "store", "market", "shopping", "clothing_store"},
		"scenic":        {"plaza", "square", "bridge", "beach", "marina", "harbor", "view"},
		"theater":       {"theater", "movie_theater", "cinema"},
		"entertainment": {"theater", "movie_theater", "cinema", "music", "stadium", "amusement", "zoo", "aquarium"},
	}

	if expanded, ok := aliases[tag]; ok {
		return expanded
	}

	return []string{tag}
}

func diversifyPlaces(places []googlemaps.Place, desired int) []googlemaps.Place {
	if len(places) == 0 || desired <= 0 {
		return places
	}

	maxPerCategory := int(math.Ceil(float64(desired) * 0.4))
	if maxPerCategory < 1 {
		maxPerCategory = 1
	}

	buckets := map[string][]googlemaps.Place{}
	order := make([]string, 0)
	seenCategory := map[string]bool{}
	for _, place := range places {
		category := placeCategory(place)
		if !seenCategory[category] {
			seenCategory[category] = true
			order = append(order, category)
		}
		buckets[category] = append(buckets[category], place)
	}

	result := make([]googlemaps.Place, 0, minInt(desired, len(places)))
	categoryCounts := map[string]int{}
	relaxed := false

	for len(result) < desired && len(result) < len(places) {
		added := false
		for _, category := range order {
			if len(result) >= desired {
				break
			}
			if len(buckets[category]) == 0 {
				continue
			}
			if !relaxed && categoryCounts[category] >= maxPerCategory {
				continue
			}
			next := buckets[category][0]
			buckets[category] = buckets[category][1:]
			result = append(result, next)
			categoryCounts[category]++
			added = true
		}
		if !added {
			if relaxed {
				break
			}
			relaxed = true
		}
	}

	if len(result) == 0 {
		return places
	}

	log.Printf("TopSpots: diversified mix with maxPerCategory=%d; categories=%v", maxPerCategory, categoryCounts)
	return result
}

func placeCategory(place googlemaps.Place) string {
	types := normalizePlaceTypes(place)
	if len(types) == 0 {
		return "other"
	}

	if hasAnyType(types, []string{
		"restaurant", "cafe", "coffee_shop", "bakery", "bar", "meal_takeaway", "meal_delivery",
		"ice_cream_shop", "dessert",
	}) {
		return "food"
	}
	if hasAnyType(types, []string{
		"park", "garden", "playground", "trail", "hiking_area", "campground", "natural_feature",
	}) {
		return "park"
	}
	if hasAnyType(types, []string{
		"museum", "art_gallery", "gallery", "tourist_attraction", "landmark", "library",
	}) {
		return "culture"
	}
	if hasAnyType(types, []string{
		"movie_theater", "theater", "night_club", "music_venue", "stadium", "sports_complex",
		"gym", "bowling_alley", "amusement_park", "zoo", "aquarium",
	}) {
		return "entertainment"
	}
	if hasAnyType(types, []string{
		"store", "shopping_mall", "market", "supermarket", "convenience_store", "book_store",
		"clothing_store", "department_store", "electronics_store", "furniture_store", "home_goods_store",
		"pet_store", "florist",
	}) {
		return "shopping"
	}
	if hasAnyType(types, []string{
		"plaza", "square", "bridge", "beach", "marina", "harbor", "water",
	}) {
		return "scenic"
	}

	display := strings.ToLower(strings.TrimSpace(place.PrimaryTypeDisplayName.Text))
	if display != "" {
		switch {
		case strings.Contains(display, "cafe") || strings.Contains(display, "coffee") || strings.Contains(display, "restaurant") || strings.Contains(display, "bar"):
			return "food"
		case strings.Contains(display, "park") || strings.Contains(display, "garden") || strings.Contains(display, "trail"):
			return "park"
		case strings.Contains(display, "museum") || strings.Contains(display, "gallery") || strings.Contains(display, "library"):
			return "culture"
		case strings.Contains(display, "theater") || strings.Contains(display, "cinema") || strings.Contains(display, "music") || strings.Contains(display, "stadium"):
			return "entertainment"
		case strings.Contains(display, "market") || strings.Contains(display, "shop") || strings.Contains(display, "store"):
			return "shopping"
		case strings.Contains(display, "plaza") || strings.Contains(display, "square") || strings.Contains(display, "bridge") || strings.Contains(display, "beach"):
			return "scenic"
		}
	}

	return "other"
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func computeReviewThresholds(places []googlemaps.Place) (int32, int32, int32) {
	counts := make([]int, 0, len(places))
	for _, place := range places {
		if place.UserRatingCount == nil {
			continue
		}
		if *place.UserRatingCount <= 0 {
			continue
		}
		counts = append(counts, int(*place.UserRatingCount))
	}

	if len(counts) == 0 {
		return 0, 0, 0
	}

	sort.Ints(counts)
	maxCount := int32(counts[len(counts)-1])
	percentile := func(p float64) int32 {
		if len(counts) == 1 {
			return int32(counts[0])
		}
		idx := int(math.Round(p * float64(len(counts)-1)))
		if idx < 0 {
			idx = 0
		} else if idx >= len(counts) {
			idx = len(counts) - 1
		}
		return int32(counts[idx])
	}

	high := percentile(0.75)
	mid := percentile(0.50)
	low := percentile(0.30)

	baseFloor := int32(10)
	if len(counts) < 10 {
		baseFloor = 5
	}
	if maxCount < baseFloor {
		baseFloor = maxCount
	}

	if high < baseFloor {
		high = baseFloor
	}
	if mid < baseFloor/2 {
		mid = baseFloor / 2
	}
	if low < 3 {
		low = 3
	}

	if high < mid {
		high = mid
	}
	if mid < low {
		mid = low
	}
	if high > maxCount {
		high = maxCount
	}
	if mid > maxCount {
		mid = maxCount
	}
	if low > maxCount {
		low = maxCount
	}

	return high, mid, low
}

func (p *SeedZoneDraftProcessor) filterEnjoyablePlacesLLM(
	places []googlemaps.Place,
) (map[string]struct{}, error) {
	if len(places) == 0 {
		return map[string]struct{}{}, nil
	}

	prompt := fmt.Sprintf(enjoyablePlaceFilterPromptTemplate, formatPlacesForEnjoyableFilter(places, 20))
	answer, err := p.deepPriest.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return nil, err
	}

	var response enjoyablePlaceFilterResponse
	if err := json.Unmarshal([]byte(extractJSON(answer.Answer)), &response); err != nil {
		return nil, err
	}

	allowed := make(map[string]struct{}, len(places))
	for _, place := range places {
		if place.ID != "" {
			allowed[place.ID] = struct{}{}
		}
	}

	selected := make(map[string]struct{}, len(response.EnjoyablePlaceIDs))
	for _, id := range response.EnjoyablePlaceIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, ok := allowed[trimmed]; ok {
			selected[trimmed] = struct{}{}
		}
	}

	return selected, nil
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

func formatPlacesForEnjoyableFilter(places []googlemaps.Place, limit int) string {
	lines := make([]string, 0, len(places))
	for i, place := range places {
		if limit > 0 && i >= limit {
			break
		}
		name := strings.TrimSpace(place.DisplayName.Text)
		if name == "" {
			name = "Unknown place"
		}
		summary := strings.TrimSpace(place.EditorialSummary.Text)
		if summary == "" {
			summary = place.PrimaryTypeDisplayName.Text
		}
		if summary == "" {
			summary = strings.Join(place.Types, ", ")
		}
		address := strings.TrimSpace(place.FormattedAddress)
		if address != "" {
			address = " | " + truncate(address, 80)
		}
		line := fmt.Sprintf(
			"- %s | placeId=%s | primaryType=%s | types=%s | %s%s",
			name,
			place.ID,
			place.PrimaryType,
			strings.Join(place.Types, ","),
			truncate(summary, 120),
			address,
		)
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
		"locksmith", "hardware_store", "home_improvement_store",
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

func normalizeSeedChallengeQuestion(question string, place googlemaps.Place) string {
	trimmed := strings.TrimSpace(question)
	name := strings.TrimSpace(place.DisplayName.Text)
	if name == "" {
		name = "this location"
	}
	types := normalizePlaceTypes(place)

	if trimmed == "" {
		return buildSeedPhotoProofQuestion(name, types)
	}

	lower := strings.ToLower(trimmed)
	hasPhoto := hasAnyKeyword(lower, "photo", "picture", "snapshot", "selfie", "photograph")
	hasText := hasAnyKeyword(lower, "write", "describe", "list", "note", "count", "identify", "observe", "explain", "summarize", "record", "tell")

	if requiresProofOnly(lower) || (hasPhoto && hasText) {
		return buildSeedPhotoProofQuestion(name, types)
	}

	if hasPhoto || hasAnyKeyword(lower, "sketch", "draw") {
		if hasAnyKeyword(lower, "sketch", "draw") && !hasPhoto {
			return fmt.Sprintf("Sketch something you notice at %s and take a photo of your sketch.", name)
		}
		if strings.HasPrefix(lower, "take") || strings.HasPrefix(lower, "photograph") || strings.HasPrefix(lower, "photo") {
			if needsParticipationOverride(lower) {
				return buildSeedParticipationPhotoQuestion(name, types)
			}
			return trimmed
		}
		return buildSeedPhotoProofQuestion(name, types)
	}

	if hasText {
		if needsParticipationOverride(lower) {
			return buildSeedParticipationPhotoQuestion(name, types)
		}
		return ensureTextOnlyQuestion(trimmed)
	}

	return buildSeedPhotoProofQuestion(name, types)
}

func buildSeedPhotoProofQuestion(name string, types []string) string {
	return buildSeedParticipationPhotoQuestion(name, types)
}

func buildSeedParticipationPhotoQuestion(name string, types []string) string {
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
	case hasType("book_store", "bookstore") || strings.Contains(lowerName, "book"):
		return fmt.Sprintf("Pick a book at %s, open to a page that grabs you, and photograph the open book.", name)
	case hasType("library") || strings.Contains(lowerName, "library"):
		return fmt.Sprintf("Find a book or display at %s that interests you and photograph the cover or display.", name)
	case hasType("comedy_club") || strings.Contains(lowerName, "comedy"):
		return fmt.Sprintf("Photograph the stage or lineup board for a comedy set at %s.", name)
	case hasType("movie_theater", "cinema") || strings.Contains(lowerName, "cinema") || strings.Contains(lowerName, "theater"):
		return fmt.Sprintf("Photograph the showtime board or poster for the film you're seeing at %s.", name)
	case hasType("performing_arts_theater", "theater") || strings.Contains(lowerName, "theatre"):
		return fmt.Sprintf("Photograph the program, poster, or stage entrance for the performance at %s.", name)
	case hasType("live_music_venue", "music_venue", "concert_hall", "night_club") || strings.Contains(lowerName, "music"):
		return fmt.Sprintf("Photograph the stage or lineup board for the performance at %s.", name)
	case hasType("museum", "art_gallery", "gallery") || strings.Contains(lowerName, "museum") || strings.Contains(lowerName, "gallery"):
		return fmt.Sprintf("Find an exhibit label or artwork title at %s and photograph it.", name)
	case hasType("park", "garden", "playground", "trail", "campground", "natural_feature") || strings.Contains(lowerName, "park") || strings.Contains(lowerName, "garden"):
		return fmt.Sprintf("Photograph a spot you spent time at %s (bench, trail marker, or play area).", name)
	case hasType("cafe", "coffee", "coffee_shop", "bakery") || strings.Contains(lowerName, "coffee") || strings.Contains(lowerName, "cafe"):
		return fmt.Sprintf("Get a coffee, drink, or pastry at %s and photograph the item you chose.", name)
	case hasType("restaurant", "bar", "brewery", "meal_takeaway", "meal_delivery") || strings.Contains(lowerName, "restaurant") || strings.Contains(lowerName, "bar"):
		return fmt.Sprintf("Get a meal or drink at %s and photograph the item you chose.", name)
	case hasType("ice_cream_shop", "dessert") || strings.Contains(lowerName, "ice cream") || strings.Contains(lowerName, "gelato"):
		return fmt.Sprintf("Photograph the dessert or flavor board you picked at %s.", name)
	case hasType("market", "store", "shopping_mall", "supermarket", "clothing_store", "shoe_store", "department_store") || strings.Contains(lowerName, "market") || strings.Contains(lowerName, "shop"):
		return fmt.Sprintf("Pick an item or display at %s that caught your eye and photograph it.", name)
	case hasType("amusement_park") || strings.Contains(lowerName, "amusement"):
		return fmt.Sprintf("Photograph the ride or attraction sign you visited at %s.", name)
	case hasType("aquarium", "zoo") || strings.Contains(lowerName, "aquarium") || strings.Contains(lowerName, "zoo"):
		return fmt.Sprintf("Photograph an exhibit sign or viewing area you visited at %s.", name)
	case hasType("gym", "fitness") || strings.Contains(lowerName, "gym") || strings.Contains(lowerName, "fitness"):
		return fmt.Sprintf("Photograph the workout station or class board you used at %s.", name)
	case hasType("tourist_attraction", "landmark", "monument", "historic_site") || strings.Contains(lowerName, "monument"):
		return fmt.Sprintf("Photograph the main feature or plaque you stopped to read at %s.", name)
	case hasType("beach", "lake", "river", "water", "harbor", "marina") || strings.Contains(lowerName, "beach") || strings.Contains(lowerName, "lake"):
		return fmt.Sprintf("Photograph the shoreline or view where you spent time at %s.", name)
	default:
		return fmt.Sprintf("Take a photo that shows you participating in the main activity at %s.", name)
	}
}

func hasAnyKeyword(text string, keywords ...string) bool {
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

func requiresProofOnly(text string) bool {
	return hasAnyKeyword(text,
		"oldest", "first", "main ingredient", "ingredient", "recipe", "menu", "price", "cost",
		"listen", "instrument", "song", "piece", "performance", "band", "taste", "flavor",
		"review", "best", "favorite", "count", "number of",
	)
}

func ensureTextOnlyQuestion(question string) string {
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

func buildSeedHeuristicChallengeQuestion(name string, types []string) string {
	return buildSeedParticipationPhotoQuestion(name, types)
}

func needsParticipationOverride(text string) bool {
	if text == "" {
		return false
	}
	lower := strings.ToLower(text)
	if !hasAnyKeyword(lower,
		"sign", "storefront", "entrance", "exterior", "outside", "front",
		"marquee", "poster", "window", "logo", "facade", "building",
	) {
		return false
	}
	if hasAnyKeyword(lower,
		"menu", "drink", "meal", "food", "book", "read", "page", "shelf",
		"stage", "show", "performance", "set", "lineup", "ticket", "program",
		"exhibit", "gallery", "trail", "play", "game", "class", "workout",
		"ride", "viewing", "display", "bench", "picnic", "market", "shop",
		"browse", "order", "sip", "eat", "watch", "listen",
	) {
		return false
	}
	return true
}
