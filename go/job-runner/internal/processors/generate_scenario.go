package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	scenarioPlaceholderImageURL = "https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/scenario-undiscovered.png"
)

var scenarioGenerationValidStatTags = map[string]struct{}{
	"strength":     {},
	"dexterity":    {},
	"constitution": {},
	"intelligence": {},
	"wisdom":       {},
	"charisma":     {},
}

const openEndedScenarioGenerationPromptTemplate = `
You are designing one fantasy RPG map scenario for a location-based game.

Zone:
- name: %s
- description: %s

Location context:
- latitude: %.6f
- longitude: %.6f

Variance directives (treat these as hard constraints for diversity):
%s

Recent scenarios in this zone to avoid echoing:
%s

Create an OPEN-ENDED scenario (free-text response from player).

Return JSON only:
{
  "zoneKind": "forest",
  "prompt": "2-4 vivid sentences",
  "difficulty": 0-40,
  "rewardExperience": 0-120,
  "rewardGold": 0-120
}

Rules:
- zoneKind must be one of the allowed slugs exactly as written.
- Choose the single best-fit zone kind for this generated scenario.
- Prompt must be specific to this zone and location, with a clear conflict/opportunity.
- The scenario must feel materially different from the recent scenarios listed above.
- Keep tone adventurous and grounded in physical surroundings.
`

const choiceScenarioGenerationPromptTemplate = `
You are designing one fantasy RPG map scenario for a location-based game.

Zone:
- name: %s
- description: %s

Location context:
- latitude: %.6f
- longitude: %.6f

Variance directives (treat these as hard constraints for diversity):
%s

Recent scenarios in this zone to avoid echoing:
%s

Create a CHOICE-BASED scenario with 3 options.

Return JSON only:
{
  "zoneKind": "forest",
  "prompt": "2-4 vivid sentences",
  "difficulty": 0-40,
  "options": [
    {
      "optionText": "player action text",
      "successText": "one or two sentences",
      "failureText": "one or two sentences",
      "statTag": "strength|dexterity|constitution|intelligence|wisdom|charisma",
      "proficiencies": ["0-3 short proficiencies"],
      "difficulty": null or 0-40,
      "rewardExperience": 0-80,
      "rewardGold": 0-80
    }
  ]
}

Rules:
- zoneKind must be one of the allowed slugs exactly as written.
- Choose the single best-fit zone kind for this generated scenario.
- Prompt must be specific to this zone and location, with a clear conflict/opportunity.
- The scenario must feel materially different from the recent scenarios listed above.
- options must contain exactly 3 entries and each option should feel distinct.
- proficiencies should be practical, short labels.
`

var scenarioGenerationEncounterAnchors = []string{
	"a rooftop weather-vane platform exposed to strong winds",
	"a cellar beneath a public gathering place",
	"a narrow bridge crossing dark water",
	"an overgrown courtyard behind shuttered storefronts",
	"a market stall district at closing time",
	"a watchtower stairwell filled with echoing footsteps",
	"a shrine-side lantern path with failing light",
	"a freight dock where crates are being moved in haste",
}

var scenarioGenerationComplications = []string{
	"a time-sensitive hazard that escalates every minute",
	"an innocent bystander with conflicting goals",
	"a fragile object that must not be damaged",
	"a noisy crowd that can panic if mishandled",
	"limited visibility due to fog, smoke, or dust",
	"a deceptive clue that points to the wrong threat",
	"a rival faction attempting to intervene",
	"an unstable structure that might collapse",
}

var scenarioGenerationStakes = []string{
	"protecting civilians from immediate harm",
	"preventing the loss of a rare local resource",
	"stopping an emerging chain reaction",
	"preserving trust between local groups",
	"keeping a key route open for the district",
	"recovering evidence before it disappears",
	"preventing sabotage of local infrastructure",
	"containing a supernatural spillover before dawn",
}

var scenarioGenerationToneStyles = []string{
	"tense and tactical",
	"mysterious and uncanny",
	"social and negotiation-heavy",
	"investigative with subtle clues",
	"urgent and cinematic",
	"grim but hopeful",
	"wry and streetwise",
	"ominous with restrained humor",
}

var scenarioGenerationTwists = []string{
	"the apparent victim is secretly orchestrating events",
	"the obvious threat is a distraction for a quieter risk",
	"the conflict can be de-escalated through unexpected empathy",
	"a trusted authority figure is unintentionally worsening things",
	"the key tool works once and then breaks",
	"an environmental detail becomes the central leverage point",
	"the safe solution costs immediate personal reputation",
	"success requires cooperation with a former rival",
}

type scenarioGenerationRewardPayload struct {
	InventoryItemID int `json:"inventoryItemId"`
	Quantity        int `json:"quantity"`
}

type openEndedScenarioGenerationResponse struct {
	ZoneKind         string                            `json:"zoneKind"`
	Prompt           string                            `json:"prompt"`
	Difficulty       *int                              `json:"difficulty"`
	RewardExperience int                               `json:"rewardExperience"`
	RewardGold       int                               `json:"rewardGold"`
	ItemRewards      []scenarioGenerationRewardPayload `json:"itemRewards"`
}

type choiceScenarioGenerationOptionPayload struct {
	OptionText       string                            `json:"optionText"`
	SuccessText      string                            `json:"successText"`
	FailureText      string                            `json:"failureText"`
	StatTag          string                            `json:"statTag"`
	Proficiencies    []string                          `json:"proficiencies"`
	Difficulty       *int                              `json:"difficulty"`
	RewardExperience int                               `json:"rewardExperience"`
	RewardGold       int                               `json:"rewardGold"`
	ItemRewards      []scenarioGenerationRewardPayload `json:"itemRewards"`
}

type choiceScenarioGenerationResponse struct {
	ZoneKind   string                                  `json:"zoneKind"`
	Prompt     string                                  `json:"prompt"`
	Difficulty *int                                    `json:"difficulty"`
	Options    []choiceScenarioGenerationOptionPayload `json:"options"`
}

type GenerateScenarioProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	asyncClient      *asynq.Client
}

func NewGenerateScenarioProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
	asyncClient *asynq.Client,
) GenerateScenarioProcessor {
	log.Println("Initializing GenerateScenarioProcessor")
	return GenerateScenarioProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		asyncClient:      asyncClient,
	}
}

func (p *GenerateScenarioProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate scenario task: %v", task.Type())

	var payload jobs.GenerateScenarioTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := p.dbClient.ScenarioGenerationJob().FindByID(ctx, payload.JobID)
	if err != nil {
		return err
	}
	if job == nil {
		log.Printf("Scenario generation job %s not found", payload.JobID)
		return nil
	}

	job.Status = models.ScenarioGenerationStatusInProgress
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ScenarioGenerationJob().Update(ctx, job); err != nil {
		return err
	}

	if err := p.generateScenario(ctx, job); err != nil {
		return p.failScenarioGenerationJob(ctx, job, err)
	}

	return nil
}

func (p *GenerateScenarioProcessor) generateScenario(ctx context.Context, job *models.ScenarioGenerationJob) error {
	zone, err := p.dbClient.Zone().FindByID(ctx, job.ZoneID)
	if err != nil {
		return fmt.Errorf("failed to load zone: %w", err)
	}
	if zone == nil {
		return fmt.Errorf("zone not found")
	}
	genre, err := loadScenarioGenre(ctx, p.dbClient, job.GenreID, job.Genre)
	if err != nil {
		return fmt.Errorf("failed to load scenario genre: %w", err)
	}
	zoneKinds, err := p.dbClient.ZoneKind().FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load zone kinds for scenario classification: %w", err)
	}
	parentZoneKind := findZoneKindBySlug(zoneKinds, zone.Kind)

	lat, lng := scenarioGenerationLocation(*zone, job.Latitude, job.Longitude)

	scenario := &models.Scenario{
		ZoneID:              job.ZoneID,
		GenreID:             genre.ID,
		Genre:               genre,
		Latitude:            lat,
		Longitude:           lng,
		ImageURL:            scenarioPlaceholderImageURL,
		ThumbnailURL:        scenarioPlaceholderImageURL,
		ScaleWithUserLevel:  job.ScaleWithUserLevel,
		RecurringScenarioID: job.RecurringScenarioID,
		RecurrenceFrequency: job.RecurrenceFrequency,
		NextRecurrenceAt:    job.NextRecurrenceAt,
		RewardMode:          models.RewardModeExplicit,
		RandomRewardSize:    models.RandomRewardSizeSmall,
		Difficulty:          24,
		RewardExperience:    0,
		RewardGold:          0,
		OpenEnded:           job.OpenEnded,
	}
	options := make([]models.ScenarioOption, 0)
	rewards := make([]models.ScenarioItemReward, 0)

	zoneName := strings.TrimSpace(zone.Name)
	if zoneName == "" {
		zoneName = "Unknown Zone"
	}
	zoneDescription := strings.TrimSpace(zone.Description)
	if zoneDescription == "" {
		zoneDescription = "No description available."
	}
	varianceSalt := buildScenarioVarianceSalt(job, zoneName)
	recentScenarioAvoidance := p.buildRecentScenarioAvoidance(ctx, job, genre, 6)

	if job.OpenEnded {
		prompt := buildOpenEndedScenarioGenerationPrompt(
			zoneName,
			zoneDescription,
			lat,
			lng,
			varianceSalt,
			recentScenarioAvoidance,
			genre,
			zoneKinds,
			parentZoneKind,
		)
		answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
		if err != nil {
			return fmt.Errorf("failed to generate open-ended scenario: %w", err)
		}
		generated := &openEndedScenarioGenerationResponse{}
		if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), generated); err != nil {
			return fmt.Errorf("failed to parse open-ended scenario payload: %w", err)
		}

		scenario.Prompt = sanitizeScenarioPrompt(generated.Prompt)
		scenario.Difficulty = sanitizeScenarioDifficulty(generated.Difficulty, 24)
		scenario.RewardExperience = clampInt(generated.RewardExperience, 0, 120)
		scenario.RewardGold = clampInt(generated.RewardGold, 0, 120)
		scenario.ZoneKind = normalizeScenarioGeneratedZoneKind(
			generated.ZoneKind,
			zoneKinds,
			deriveScenarioZoneKindHeuristically(
				zoneKinds,
				zone.Kind,
				zoneName,
				zoneDescription,
				scenario.Prompt,
				scenarioGenrePromptLabel(genre),
			),
		)
	} else {
		prompt := buildChoiceScenarioGenerationPrompt(
			zoneName,
			zoneDescription,
			lat,
			lng,
			varianceSalt,
			recentScenarioAvoidance,
			genre,
			zoneKinds,
			parentZoneKind,
		)
		answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
		if err != nil {
			return fmt.Errorf("failed to generate choice scenario: %w", err)
		}
		generated := &choiceScenarioGenerationResponse{}
		if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), generated); err != nil {
			return fmt.Errorf("failed to parse choice scenario payload: %w", err)
		}

		scenario.Prompt = sanitizeScenarioPrompt(generated.Prompt)
		scenario.Difficulty = sanitizeScenarioDifficulty(generated.Difficulty, 24)
		options = sanitizeScenarioOptions(generated.Options, nil)
		if len(options) == 0 {
			options = append(options, fallbackScenarioOption())
		}
		scenario.ZoneKind = normalizeScenarioGeneratedZoneKind(
			generated.ZoneKind,
			zoneKinds,
			deriveScenarioZoneKindHeuristically(
				zoneKinds,
				zone.Kind,
				zoneName,
				zoneDescription,
				scenario.Prompt,
				scenarioGenrePromptLabel(genre),
			),
		)
	}

	if err := p.dbClient.Scenario().Create(ctx, scenario); err != nil {
		return fmt.Errorf("failed to create scenario: %w", err)
	}
	if err := p.dbClient.Scenario().ReplaceOptions(ctx, scenario.ID, options); err != nil {
		return fmt.Errorf("failed to create scenario options: %w", err)
	}
	if err := p.dbClient.Scenario().ReplaceItemRewards(ctx, scenario.ID, rewards); err != nil {
		return fmt.Errorf("failed to create scenario rewards: %w", err)
	}

	job.Status = models.ScenarioGenerationStatusCompleted
	job.GeneratedScenarioID = &scenario.ID
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ScenarioGenerationJob().Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update scenario generation job: %w", err)
	}

	if p.asyncClient != nil {
		imagePayload, err := json.Marshal(jobs.GenerateScenarioImageTaskPayload{
			ScenarioID: scenario.ID,
		})
		if err == nil {
			if _, enqueueErr := p.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateScenarioImageTaskType, imagePayload)); enqueueErr != nil {
				log.Printf("Failed to enqueue scenario image generation for scenario %s: %v", scenario.ID, enqueueErr)
			}
		}
	}

	return nil
}

func (p *GenerateScenarioProcessor) failScenarioGenerationJob(ctx context.Context, job *models.ScenarioGenerationJob, err error) error {
	msg := err.Error()
	job.Status = models.ScenarioGenerationStatusFailed
	job.ErrorMessage = &msg
	job.UpdatedAt = time.Now()
	if updateErr := p.dbClient.ScenarioGenerationJob().Update(ctx, job); updateErr != nil {
		log.Printf("Failed to mark scenario generation job %s as failed: %v", job.ID, updateErr)
	}
	return err
}

func extractGeneratedJSONObject(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```JSON")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
	}
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		return trimmed[start : end+1]
	}
	return trimmed
}

func scenarioGenerationLocation(zone models.Zone, latitude *float64, longitude *float64) (float64, float64) {
	if latitude != nil && longitude != nil {
		return *latitude, *longitude
	}

	point := zone.GetRandomPoint()
	if point.X() != 0 || point.Y() != 0 {
		return point.Y(), point.X()
	}

	if zone.Latitude != 0 || zone.Longitude != 0 {
		return zone.Latitude, zone.Longitude
	}

	return 40.7589, -73.98513
}

func buildAllowedItemsPrompt(items []models.InventoryItem) string {
	if len(items) == 0 {
		return "- none"
	}

	const maxItems = 120
	lines := make([]string, 0, min(len(items), maxItems)+1)
	for i, item := range items {
		if i >= maxItems {
			lines = append(lines, "- ...")
			break
		}
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = "Unnamed Item"
		}
		lines = append(lines, fmt.Sprintf("- %d: %s", item.ID, name))
	}
	return strings.Join(lines, "\n")
}

func (p *GenerateScenarioProcessor) buildRecentScenarioAvoidance(
	ctx context.Context,
	job *models.ScenarioGenerationJob,
	genre *models.ZoneGenre,
	limit int,
) string {
	if job == nil || limit <= 0 {
		return "- none"
	}

	scenarios, err := p.dbClient.Scenario().FindByZoneID(ctx, job.ZoneID)
	if err != nil || len(scenarios) == 0 {
		return "- none"
	}

	sort.Slice(scenarios, func(i, j int) bool {
		return scenarios[i].CreatedAt.After(scenarios[j].CreatedAt)
	})

	lines := make([]string, 0, limit)
	for _, scenario := range scenarios {
		if genre != nil && genre.ID != uuid.Nil && scenario.GenreID != uuid.Nil && scenario.GenreID != genre.ID {
			continue
		}
		prompt := strings.TrimSpace(scenario.Prompt)
		if prompt == "" {
			continue
		}
		prompt = strings.ReplaceAll(prompt, "\n", " ")
		if len(prompt) > 220 {
			prompt = strings.TrimSpace(prompt[:220]) + "..."
		}
		lines = append(lines, "- "+prompt)
		if len(lines) >= limit {
			break
		}
	}
	if len(lines) == 0 {
		return "- none"
	}
	return strings.Join(lines, "\n")
}

func buildOpenEndedScenarioGenerationPrompt(
	zoneName string,
	zoneDescription string,
	latitude float64,
	longitude float64,
	varianceSalt string,
	recentScenarioAvoidance string,
	genre *models.ZoneGenre,
	zoneKinds []models.ZoneKind,
	parentZoneKind *models.ZoneKind,
) string {
	base := fmt.Sprintf(
		openEndedScenarioGenerationPromptTemplate,
		zoneName,
		zoneDescription,
		latitude,
		longitude,
		varianceSalt,
		recentScenarioAvoidance,
	)
	instructionBlocks := []string{}
	if zoneKindBlock := buildScenarioZoneKindInstructionBlock(
		zoneKinds,
		parentZoneKind,
		"parent zone kind",
	); zoneKindBlock != "" {
		instructionBlocks = append(instructionBlocks, zoneKindBlock)
	}
	if !isBaselineFantasyScenarioGenre(genre) {
		instructionBlocks = append(
			instructionBlocks,
			scenarioGenreInstructionBlock(genre),
		)
	}
	if len(instructionBlocks) == 0 {
		return base
	}
	return strings.TrimSpace(strings.Join(instructionBlocks, "\n\n") + "\n\n" + base)
}

func buildChoiceScenarioGenerationPrompt(
	zoneName string,
	zoneDescription string,
	latitude float64,
	longitude float64,
	varianceSalt string,
	recentScenarioAvoidance string,
	genre *models.ZoneGenre,
	zoneKinds []models.ZoneKind,
	parentZoneKind *models.ZoneKind,
) string {
	base := fmt.Sprintf(
		choiceScenarioGenerationPromptTemplate,
		zoneName,
		zoneDescription,
		latitude,
		longitude,
		varianceSalt,
		recentScenarioAvoidance,
	)
	instructionBlocks := []string{}
	if zoneKindBlock := buildScenarioZoneKindInstructionBlock(
		zoneKinds,
		parentZoneKind,
		"parent zone kind",
	); zoneKindBlock != "" {
		instructionBlocks = append(instructionBlocks, zoneKindBlock)
	}
	if !isBaselineFantasyScenarioGenre(genre) {
		instructionBlocks = append(
			instructionBlocks,
			scenarioGenreInstructionBlock(genre),
		)
	}
	if len(instructionBlocks) == 0 {
		return base
	}
	return strings.TrimSpace(strings.Join(instructionBlocks, "\n\n") + "\n\n" + base)
}

func buildScenarioVarianceSalt(job *models.ScenarioGenerationJob, zoneName string) string {
	if job == nil {
		return "- use a clearly different encounter setup than recent scenarios"
	}

	seed := fmt.Sprintf(
		"%s|%s|%t|%.6f|%.6f",
		job.ID.String(),
		strings.ToLower(strings.TrimSpace(zoneName)),
		job.OpenEnded,
		derefFloat64(job.Latitude),
		derefFloat64(job.Longitude),
	)

	encounter := scenarioGenerationEncounterAnchors[saltIndex(seed, 1, len(scenarioGenerationEncounterAnchors))]
	complication := scenarioGenerationComplications[saltIndex(seed, 2, len(scenarioGenerationComplications))]
	stakes := scenarioGenerationStakes[saltIndex(seed, 3, len(scenarioGenerationStakes))]
	tone := scenarioGenerationToneStyles[saltIndex(seed, 4, len(scenarioGenerationToneStyles))]
	twist := scenarioGenerationTwists[saltIndex(seed, 5, len(scenarioGenerationTwists))]

	return strings.Join([]string{
		"- primary setting anchor: " + encounter,
		"- main complication: " + complication,
		"- core stakes: " + stakes,
		"- tonal style: " + tone,
		"- structural twist: " + twist,
		"- do not reuse the same opening beat or central prop from recent scenarios",
	}, "\n")
}

func saltIndex(seed string, salt int, size int) int {
	if size <= 0 {
		return 0
	}

	hash := 17 + salt*131
	for i := 0; i < len(seed); i++ {
		hash = hash*31 + int(seed[i]) + salt*7
	}
	if hash < 0 {
		hash = -hash
	}
	return hash % size
}

func derefFloat64(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func sanitizeScenarioPrompt(prompt string) string {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return "A tense moment unfolds in the district as locals look to you for help."
	}
	if len(trimmed) > 700 {
		return strings.TrimSpace(trimmed[:700])
	}
	return trimmed
}

func sanitizeScenarioDifficulty(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return clampInt(*value, 0, 40)
}

func sanitizeScenarioRewards(rewards []scenarioGenerationRewardPayload, allowedItemIDs map[int]struct{}, maxCount int) []models.ScenarioItemReward {
	result := make([]models.ScenarioItemReward, 0, min(len(rewards), maxCount))
	for _, reward := range rewards {
		if len(result) >= maxCount {
			break
		}
		if reward.InventoryItemID <= 0 || reward.Quantity <= 0 {
			continue
		}
		if _, ok := allowedItemIDs[reward.InventoryItemID]; !ok {
			continue
		}
		result = append(result, models.ScenarioItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        clampInt(reward.Quantity, 1, 3),
		})
	}
	return result
}

func sanitizeScenarioOptionRewards(rewards []scenarioGenerationRewardPayload, allowedItemIDs map[int]struct{}, maxCount int) []models.ScenarioOptionItemReward {
	result := make([]models.ScenarioOptionItemReward, 0, min(len(rewards), maxCount))
	for _, reward := range rewards {
		if len(result) >= maxCount {
			break
		}
		if reward.InventoryItemID <= 0 || reward.Quantity <= 0 {
			continue
		}
		if _, ok := allowedItemIDs[reward.InventoryItemID]; !ok {
			continue
		}
		result = append(result, models.ScenarioOptionItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        clampInt(reward.Quantity, 1, 2),
		})
	}
	return result
}

func sanitizeScenarioOptions(options []choiceScenarioGenerationOptionPayload, allowedItemIDs map[int]struct{}) []models.ScenarioOption {
	if len(options) == 0 {
		return nil
	}

	const maxOptions = 4
	result := make([]models.ScenarioOption, 0, min(len(options), maxOptions))
	for _, option := range options {
		if len(result) >= maxOptions {
			break
		}
		optionText := strings.TrimSpace(option.OptionText)
		if optionText == "" {
			continue
		}
		successText := strings.TrimSpace(option.SuccessText)
		if successText == "" {
			successText = "Your approach works, and momentum turns in your favor."
		}
		failureText := strings.TrimSpace(option.FailureText)
		if failureText == "" {
			failureText = "The attempt falls short, and the moment slips away."
		}

		statTag := strings.ToLower(strings.TrimSpace(option.StatTag))
		if _, ok := scenarioGenerationValidStatTags[statTag]; !ok {
			statTag = "charisma"
		}

		proficiencies := sanitizeScenarioProficiencies(option.Proficiencies, 3)

		var difficulty *int
		if option.Difficulty != nil {
			d := clampInt(*option.Difficulty, 0, 40)
			difficulty = &d
		}

		result = append(result, models.ScenarioOption{
			OptionText:       optionText,
			SuccessText:      successText,
			FailureText:      failureText,
			StatTag:          statTag,
			Proficiencies:    models.StringArray(proficiencies),
			Difficulty:       difficulty,
			RewardExperience: clampInt(option.RewardExperience, 0, 80),
			RewardGold:       clampInt(option.RewardGold, 0, 80),
			ItemRewards:      sanitizeScenarioOptionRewards(option.ItemRewards, allowedItemIDs, 2),
		})
	}

	return result
}

func sanitizeScenarioProficiencies(values []string, maxCount int) []string {
	if len(values) == 0 {
		return []string{}
	}
	seen := map[string]struct{}{}
	result := make([]string, 0, min(len(values), maxCount))
	for _, value := range values {
		if len(result) >= maxCount {
			break
		}
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func fallbackScenarioOption() models.ScenarioOption {
	return models.ScenarioOption{
		OptionText:       "Rally nearby allies and confront the threat together.",
		SuccessText:      "Your coordinated push turns the tide and restores order.",
		FailureText:      "The plan falters under pressure, leaving the danger unresolved.",
		StatTag:          "charisma",
		Proficiencies:    models.StringArray{},
		RewardExperience: 10,
		RewardGold:       8,
		ItemRewards:      []models.ScenarioOptionItemReward{},
	}
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
