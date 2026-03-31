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

const questArchetypeSuggestionPromptTemplate = `
You are designing reusable quest archetype bundles for Unclaimed Streets, an urban fantasy MMORPG.

Generate exactly %d quest archetype bundles.

Requested direction:
- theme prompt: %s
- family tags: %s
- character tags to bias toward: %s
- internal tags to bias toward: %s
- required location archetypes that must appear in each draft: %s
- required location metadata tags to use when appropriate: %s

Recent quest archetypes to avoid echoing:
%s

Allowed location archetypes:
%s

Allowed monster templates:
%s

Return JSON only:
{
  "drafts": [
    {
      "name": "string",
      "hook": "string",
      "description": "string",
      "acceptanceDialogue": ["line 1", "line 2", "line 3"],
      "characterTags": ["tag_a", "tag_b"],
      "internalTags": ["tag_a", "tag_b"],
      "difficultyMode": "scale|fixed",
      "difficulty": 1,
      "monsterEncounterTargetLevel": 1,
      "whyThisScales": "string",
      "challengeTemplateSeeds": ["seed one", "seed two"],
      "scenarioTemplateSeeds": ["seed one", "seed two"],
      "monsterTemplateSeeds": ["seed one", "seed two"],
      "steps": [
        {
          "source": "location|proximity",
          "content": "challenge|scenario|monster",
          "locationConcept": "string",
          "locationArchetypeName": "name from allowed location archetypes or empty string for proximity steps",
          "locationMetadataTags": ["market", "storefront"],
          "distanceMeters": 120,
          "templateConcept": "string",
          "potentialContent": ["idea one", "idea two", "idea three"],
          "challengeQuestion": "required for challenge steps",
          "challengeDescription": "required for challenge steps",
          "challengeSubmissionType": "photo|text|video",
          "challengeProficiency": "optional short proficiency",
          "challengeStatTags": ["strength", "wisdom"],
          "scenarioPrompt": "required for scenario steps",
          "scenarioOpenEnded": true,
          "scenarioBeats": ["beat one", "beat two"],
          "monsterTemplateNames": ["names chosen from allowed monster templates"],
          "encounterTone": ["urban", "scrappy"]
        }
      ]
    }
  ]
}

Rules:
- Output exactly %d drafts.
- Output JSON only. No markdown.
- Keep the tone urban fantasy, street-level, and reusable.
- Prefer 2-3 node routes.
- At least one third of drafts should be mostly non-combat.
- At least one third of drafts should end in combat.
- Challenge steps must use source "location".
- Proximity steps may only use content "scenario" or "monster".
- Use lowercase snake_case tags for characterTags, internalTags, and locationMetadataTags.
- Every step must include 2-5 locationMetadataTags.
- locationArchetypeName must be selected from the allowed list exactly when source is "location".
- monsterTemplateNames must be selected from the allowed list exactly for monster steps.
- Every draft must include each required location archetype at least once as a location step when a required list is provided.
- Challenge steps must be concrete, enjoyable real-world tasks the player can actually complete at the location right now.
- A challenge must be gradable from the player's submission alone.
- Good challenge patterns: photograph a specific detail, spot and record a pattern, identify something visible, compare two visible features, describe ambience or signage actually present on site.
- Never make a challenge depend on fictional missing objects, hidden clues, NPC cooperation, interviewing strangers, asking around, or facts that may not exist at the real location.
- If the content is about how the player would help, investigate, negotiate, persuade, intervene, solve a problem, or respond to a roleplaying situation, that is a scenario step instead of a challenge step.
- challengeQuestion should be an imperative action, not a mystery question.
- Make challengeQuestion and challengeDescription explicit and production-usable.
- Make scenarioPrompt explicit and production-usable.
- Make the content materially distinct across the batch.
`

type questArchetypeSuggestionResponse struct {
	Drafts []questArchetypeSuggestionDraftPayload `json:"drafts"`
}

type questArchetypeSuggestionDraftPayload struct {
	Name                        string                                `json:"name"`
	Hook                        string                                `json:"hook"`
	Description                 string                                `json:"description"`
	AcceptanceDialogue          []string                              `json:"acceptanceDialogue"`
	CharacterTags               []string                              `json:"characterTags"`
	InternalTags                []string                              `json:"internalTags"`
	DifficultyMode              string                                `json:"difficultyMode"`
	Difficulty                  int                                   `json:"difficulty"`
	MonsterEncounterTargetLevel int                                   `json:"monsterEncounterTargetLevel"`
	WhyThisScales               string                                `json:"whyThisScales"`
	ChallengeTemplateSeeds      []string                              `json:"challengeTemplateSeeds"`
	ScenarioTemplateSeeds       []string                              `json:"scenarioTemplateSeeds"`
	MonsterTemplateSeeds        []string                              `json:"monsterTemplateSeeds"`
	Steps                       []questArchetypeSuggestionStepPayload `json:"steps"`
}

type questArchetypeSuggestionStepPayload struct {
	Source                  string   `json:"source"`
	Content                 string   `json:"content"`
	LocationConcept         string   `json:"locationConcept"`
	LocationArchetypeName   string   `json:"locationArchetypeName"`
	LocationMetadataTags    []string `json:"locationMetadataTags"`
	DistanceMeters          *int     `json:"distanceMeters"`
	TemplateConcept         string   `json:"templateConcept"`
	PotentialContent        []string `json:"potentialContent"`
	ChallengeQuestion       string   `json:"challengeQuestion"`
	ChallengeDescription    string   `json:"challengeDescription"`
	ChallengeSubmissionType string   `json:"challengeSubmissionType"`
	ChallengeProficiency    string   `json:"challengeProficiency"`
	ChallengeStatTags       []string `json:"challengeStatTags"`
	ScenarioPrompt          string   `json:"scenarioPrompt"`
	ScenarioOpenEnded       bool     `json:"scenarioOpenEnded"`
	ScenarioBeats           []string `json:"scenarioBeats"`
	MonsterTemplateNames    []string `json:"monsterTemplateNames"`
	EncounterTone           []string `json:"encounterTone"`
}

type GenerateQuestArchetypeSuggestionsProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
}

func NewGenerateQuestArchetypeSuggestionsProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
) GenerateQuestArchetypeSuggestionsProcessor {
	log.Println("Initializing GenerateQuestArchetypeSuggestionsProcessor")
	return GenerateQuestArchetypeSuggestionsProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
	}
}

func (p *GenerateQuestArchetypeSuggestionsProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing quest archetype suggestion task: %s", task.Type())

	var payload jobs.GenerateQuestArchetypeSuggestionsTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := p.dbClient.QuestArchetypeSuggestionJob().FindByID(ctx, payload.JobID)
	if err != nil {
		return err
	}
	if job == nil {
		log.Printf("Quest archetype suggestion job %s not found", payload.JobID)
		return nil
	}

	job.Status = models.QuestArchetypeSuggestionJobStatusInProgress
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.QuestArchetypeSuggestionJob().Update(ctx, job); err != nil {
		return err
	}

	if err := p.generateDrafts(ctx, job); err != nil {
		return p.failJob(ctx, job, err)
	}

	return nil
}

func (p *GenerateQuestArchetypeSuggestionsProcessor) generateDrafts(
	ctx context.Context,
	job *models.QuestArchetypeSuggestionJob,
) error {
	locationArchetypes, err := p.dbClient.LocationArchetype().FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load location archetypes: %w", err)
	}
	monsterTemplates, err := p.dbClient.MonsterTemplate().FindAllActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to load monster templates: %w", err)
	}
	recentArchetypes, err := p.dbClient.QuestArchetype().FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load quest archetypes: %w", err)
	}

	prompt := fmt.Sprintf(
		questArchetypeSuggestionPromptTemplate,
		maxInt(1, job.Count),
		quotedOrNone(job.ThemePrompt),
		renderTagList(job.FamilyTags),
		renderTagList(job.CharacterTags),
		renderTagList(job.InternalTags),
		buildRequiredLocationArchetypesPrompt(job.RequiredLocationArchetypeIDs, locationArchetypes),
		renderTagList(job.RequiredLocationMetadataTags),
		buildQuestArchetypeSuggestionAvoidance(recentArchetypes, 18),
		buildAllowedLocationArchetypesPrompt(locationArchetypes),
		buildAllowedMonsterTemplatesPrompt(monsterTemplates),
		maxInt(1, job.Count),
	)

	answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return fmt.Errorf("failed to generate quest archetype suggestions: %w", err)
	}

	generated := &questArchetypeSuggestionResponse{}
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), generated); err != nil {
		return fmt.Errorf("failed to parse quest archetype suggestion payload: %w", err)
	}
	if len(generated.Drafts) == 0 {
		return fmt.Errorf("quest archetype suggestion payload did not include any drafts")
	}

	locationIndex := buildLocationArchetypeIndex(locationArchetypes)
	monsterIndex := buildMonsterTemplateIndex(monsterTemplates)
	requiredLocationArchetypes := resolveRequiredLocationArchetypes(
		job.RequiredLocationArchetypeIDs,
		locationArchetypes,
	)
	createdCount := 0
	for _, spec := range generated.Drafts {
		draft := sanitizeQuestArchetypeSuggestionDraft(
			spec,
			locationIndex,
			monsterIndex,
			requiredLocationArchetypes,
		)
		draft.JobID = job.ID
		draft.Status = models.QuestArchetypeSuggestionDraftStatusSuggested
		if err := p.dbClient.QuestArchetypeSuggestionDraft().Create(ctx, draft); err != nil {
			job.CreatedCount = createdCount
			return fmt.Errorf("failed to create quest archetype suggestion draft: %w", err)
		}
		createdCount++
	}

	job.CreatedCount = createdCount
	job.Status = models.QuestArchetypeSuggestionJobStatusCompleted
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.QuestArchetypeSuggestionJob().Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update quest archetype suggestion job: %w", err)
	}

	return nil
}

func (p *GenerateQuestArchetypeSuggestionsProcessor) failJob(
	ctx context.Context,
	job *models.QuestArchetypeSuggestionJob,
	err error,
) error {
	msg := err.Error()
	job.Status = models.QuestArchetypeSuggestionJobStatusFailed
	job.ErrorMessage = &msg
	job.UpdatedAt = time.Now()
	if updateErr := p.dbClient.QuestArchetypeSuggestionJob().Update(ctx, job); updateErr != nil {
		log.Printf("Failed to mark quest archetype suggestion job %s as failed: %v", job.ID, updateErr)
	}
	return err
}

func buildQuestArchetypeSuggestionAvoidance(
	recent []*models.QuestArchetype,
	limit int,
) string {
	if len(recent) == 0 {
		return "- none"
	}
	items := make([]*models.QuestArchetype, 0, len(recent))
	for _, item := range recent {
		if item != nil {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	lines := make([]string, 0, len(items))
	for _, archetype := range items {
		name := strings.TrimSpace(archetype.Name)
		if name == "" {
			name = "Unnamed archetype"
		}
		description := strings.TrimSpace(archetype.Description)
		if len(description) > 120 {
			description = strings.TrimSpace(description[:120]) + "..."
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", name, description))
	}
	if len(lines) == 0 {
		return "- none"
	}
	return strings.Join(lines, "\n")
}

func buildAllowedLocationArchetypesPrompt(items []*models.LocationArchetype) string {
	if len(items) == 0 {
		return "- none"
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		intent := make([]string, 0, len(item.IncludedTypes))
		for _, placeType := range item.IncludedTypes {
			intent = append(intent, strings.TrimSpace(string(placeType)))
		}
		if len(intent) > 4 {
			intent = intent[:4]
		}
		if len(intent) == 0 {
			lines = append(lines, fmt.Sprintf("- %s", name))
			continue
		}
		lines = append(lines, fmt.Sprintf("- %s (intent: %s)", name, strings.Join(intent, ", ")))
	}
	if len(lines) == 0 {
		return "- none"
	}
	return strings.Join(lines, "\n")
}

func buildRequiredLocationArchetypesPrompt(
	requiredIDs []string,
	items []*models.LocationArchetype,
) string {
	required := resolveRequiredLocationArchetypes(requiredIDs, items)
	if len(required) == 0 {
		return "- none"
	}
	lines := make([]string, 0, len(required))
	for _, item := range required {
		lines = append(lines, fmt.Sprintf("- %s", item.Name))
	}
	return strings.Join(lines, "\n")
}

func buildAllowedMonsterTemplatesPrompt(items []models.MonsterTemplate) string {
	if len(items) == 0 {
		return "- none"
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		desc := strings.TrimSpace(item.Description)
		if len(desc) > 80 {
			desc = strings.TrimSpace(desc[:80]) + "..."
		}
		monsterType := strings.TrimSpace(string(item.MonsterType))
		if monsterType == "" {
			monsterType = "monster"
		}
		lines = append(lines, fmt.Sprintf("- %s [%s]: %s", name, monsterType, desc))
	}
	if len(lines) == 0 {
		return "- none"
	}
	return strings.Join(lines, "\n")
}

type locationArchetypeIndexEntry struct {
	ID   uuid.UUID
	Name string
}

func buildLocationArchetypeIndex(items []*models.LocationArchetype) map[string]locationArchetypeIndexEntry {
	index := map[string]locationArchetypeIndexEntry{}
	for _, item := range items {
		if item == nil {
			continue
		}
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		index[strings.ToLower(name)] = locationArchetypeIndexEntry{ID: item.ID, Name: name}
	}
	return index
}

func resolveRequiredLocationArchetypes(
	requiredIDs []string,
	items []*models.LocationArchetype,
) []locationArchetypeIndexEntry {
	if len(requiredIDs) == 0 || len(items) == 0 {
		return nil
	}
	byID := make(map[uuid.UUID]locationArchetypeIndexEntry, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		name := strings.TrimSpace(item.Name)
		if item.ID == uuid.Nil || name == "" {
			continue
		}
		byID[item.ID] = locationArchetypeIndexEntry{ID: item.ID, Name: name}
	}
	required := make([]locationArchetypeIndexEntry, 0, len(requiredIDs))
	seen := make(map[uuid.UUID]struct{}, len(requiredIDs))
	for _, rawID := range requiredIDs {
		parsedID, err := uuid.Parse(strings.TrimSpace(rawID))
		if err != nil || parsedID == uuid.Nil {
			continue
		}
		entry, exists := byID[parsedID]
		if !exists {
			continue
		}
		if _, duplicate := seen[parsedID]; duplicate {
			continue
		}
		seen[parsedID] = struct{}{}
		required = append(required, entry)
	}
	sort.Slice(required, func(i, j int) bool {
		return required[i].Name < required[j].Name
	})
	return required
}

type monsterTemplateIndexEntry struct {
	ID   uuid.UUID
	Name string
}

func buildMonsterTemplateIndex(items []models.MonsterTemplate) map[string]monsterTemplateIndexEntry {
	index := map[string]monsterTemplateIndexEntry{}
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		index[strings.ToLower(name)] = monsterTemplateIndexEntry{ID: item.ID, Name: name}
	}
	return index
}

func sanitizeQuestArchetypeSuggestionDraft(
	payload questArchetypeSuggestionDraftPayload,
	locationIndex map[string]locationArchetypeIndexEntry,
	monsterIndex map[string]monsterTemplateIndexEntry,
	requiredLocationArchetypes []locationArchetypeIndexEntry,
) *models.QuestArchetypeSuggestionDraft {
	now := time.Now()
	warnings := models.StringArray{}
	steps := make(models.QuestArchetypeSuggestionSteps, 0, len(payload.Steps))
	for index, rawStep := range payload.Steps {
		step, stepWarnings := sanitizeQuestArchetypeSuggestionStep(rawStep, locationIndex, monsterIndex)
		for _, warning := range stepWarnings {
			warnings = append(warnings, fmt.Sprintf("step %d: %s", index+1, warning))
		}
		steps = append(steps, step)
	}
	if len(steps) == 0 {
		warnings = append(warnings, "no usable steps were generated")
	}
	for _, missing := range missingRequiredLocationArchetypes(
		steps,
		requiredLocationArchetypes,
	) {
		warnings = append(
			warnings,
			fmt.Sprintf("required location archetype %q was not used in this draft", missing),
		)
	}

	difficultyMode := models.NormalizeQuestDifficultyMode(payload.DifficultyMode)
	difficulty := models.NormalizeQuestDifficulty(payload.Difficulty)
	monsterLevel := models.NormalizeMonsterEncounterTargetLevel(payload.MonsterEncounterTargetLevel)

	name := strings.TrimSpace(payload.Name)
	if name == "" {
		name = "Generated Quest Archetype Draft"
		warnings = append(warnings, "name was empty and replaced with a fallback")
	}
	description := strings.TrimSpace(payload.Description)
	if description == "" {
		description = "Generated quest archetype draft."
		warnings = append(warnings, "description was empty and replaced with a fallback")
	}
	hook := strings.TrimSpace(payload.Hook)
	whyThisScales := strings.TrimSpace(payload.WhyThisScales)

	return &models.QuestArchetypeSuggestionDraft{
		ID:                          uuid.New(),
		CreatedAt:                   now,
		UpdatedAt:                   now,
		Status:                      models.QuestArchetypeSuggestionDraftStatusSuggested,
		Name:                        name,
		Hook:                        hook,
		Description:                 description,
		AcceptanceDialogue:          normalizeSuggestionLines(payload.AcceptanceDialogue),
		CharacterTags:               normalizeSuggestionTags(payload.CharacterTags),
		InternalTags:                normalizeSuggestionTags(payload.InternalTags),
		DifficultyMode:              difficultyMode,
		Difficulty:                  difficulty,
		MonsterEncounterTargetLevel: monsterLevel,
		WhyThisScales:               whyThisScales,
		Steps:                       steps,
		ChallengeTemplateSeeds:      normalizeSuggestionLines(payload.ChallengeTemplateSeeds),
		ScenarioTemplateSeeds:       normalizeSuggestionLines(payload.ScenarioTemplateSeeds),
		MonsterTemplateSeeds:        normalizeSuggestionLines(payload.MonsterTemplateSeeds),
		Warnings:                    normalizeSuggestionLines(warnings),
	}
}

func missingRequiredLocationArchetypes(
	steps models.QuestArchetypeSuggestionSteps,
	required []locationArchetypeIndexEntry,
) []string {
	if len(required) == 0 {
		return nil
	}
	used := make(map[uuid.UUID]struct{}, len(steps))
	for _, step := range steps {
		if step.LocationArchetypeID == nil || *step.LocationArchetypeID == uuid.Nil {
			continue
		}
		used[*step.LocationArchetypeID] = struct{}{}
	}
	missing := make([]string, 0, len(required))
	for _, entry := range required {
		if _, exists := used[entry.ID]; exists {
			continue
		}
		missing = append(missing, entry.Name)
	}
	return missing
}

func sanitizeQuestArchetypeSuggestionStep(
	payload questArchetypeSuggestionStepPayload,
	locationIndex map[string]locationArchetypeIndexEntry,
	monsterIndex map[string]monsterTemplateIndexEntry,
) (models.QuestArchetypeSuggestionStep, []string) {
	warnings := []string{}
	source := normalizeSuggestionSource(payload.Source)
	content := normalizeSuggestionContent(payload.Content)

	step := models.QuestArchetypeSuggestionStep{
		Source:                source,
		Content:               content,
		LocationConcept:       strings.TrimSpace(payload.LocationConcept),
		LocationArchetypeName: strings.TrimSpace(payload.LocationArchetypeName),
		LocationMetadataTags:  []string(normalizeSuggestionTags(payload.LocationMetadataTags)),
		TemplateConcept:       strings.TrimSpace(payload.TemplateConcept),
		PotentialContent:      []string(normalizeSuggestionLines(payload.PotentialContent)),
		ScenarioOpenEnded:     payload.ScenarioOpenEnded,
		ScenarioBeats:         []string(normalizeSuggestionLines(payload.ScenarioBeats)),
		EncounterTone:         []string(normalizeSuggestionTags(payload.EncounterTone)),
	}

	if step.LocationConcept == "" {
		step.LocationConcept = "urban site"
	}
	if len(step.LocationMetadataTags) == 0 {
		step.LocationMetadataTags = []string{"street_level"}
		warnings = append(warnings, "location metadata tags were empty")
	}
	if source == "location" {
		if entry, ok := resolveLocationArchetypeByName(step.LocationArchetypeName, locationIndex); ok {
			locationID := entry.ID
			step.LocationArchetypeID = &locationID
			step.LocationArchetypeName = entry.Name
		} else {
			warnings = append(warnings, fmt.Sprintf("location archetype %q could not be resolved", step.LocationArchetypeName))
		}
	} else {
		step.LocationArchetypeName = ""
		step.LocationArchetypeID = nil
		if payload.DistanceMeters != nil && *payload.DistanceMeters >= 0 {
			distance := *payload.DistanceMeters
			step.DistanceMeters = &distance
		} else {
			defaultDistance := 100
			step.DistanceMeters = &defaultDistance
			warnings = append(warnings, "proximity distance was missing and defaulted to 100m")
		}
	}

	switch content {
	case "challenge":
		if source != "location" {
			warnings = append(warnings, "challenge step must use a location source")
		}
		step.ChallengeQuestion = strings.TrimSpace(payload.ChallengeQuestion)
		step.ChallengeDescription = strings.TrimSpace(payload.ChallengeDescription)
		step.ChallengeSubmissionType = normalizeSuggestionSubmissionType(payload.ChallengeSubmissionType)
		if step.ChallengeSubmissionType == "" {
			step.ChallengeSubmissionType = models.DefaultQuestNodeSubmissionType()
		}
		if proficiency := strings.TrimSpace(payload.ChallengeProficiency); proficiency != "" {
			step.ChallengeProficiency = &proficiency
		}
		step.ChallengeStatTags = []string(normalizeSuggestionTags(payload.ChallengeStatTags))
		if step.ChallengeQuestion == "" {
			step.ChallengeQuestion = strings.TrimSpace(step.TemplateConcept)
			warnings = append(warnings, "challenge question was empty and fell back to template concept")
		}
		if step.ChallengeDescription == "" {
			step.ChallengeDescription = "Generated challenge template."
			warnings = append(warnings, "challenge description was empty")
		}
		if shouldConvertSuggestionChallengeToScenario(step.ChallengeQuestion, step.ChallengeDescription) {
			prompt := buildScenarioPromptFromSuggestionChallenge(step)
			step.Content = "scenario"
			step.ScenarioPrompt = prompt
			step.ScenarioOpenEnded = true
			step.ScenarioBeats = nil
			step.ChallengeQuestion = ""
			step.ChallengeDescription = ""
			step.ChallengeSubmissionType = ""
			step.ChallengeProficiency = nil
			step.ChallengeStatTags = nil
			warnings = append(warnings, "challenge read like a roleplaying or investigation scenario and was converted to an open-ended scenario")
		}
	case "scenario":
		step.ScenarioPrompt = strings.TrimSpace(payload.ScenarioPrompt)
		if step.ScenarioPrompt == "" {
			step.ScenarioPrompt = strings.TrimSpace(step.TemplateConcept)
			warnings = append(warnings, "scenario prompt was empty and fell back to template concept")
		}
	case "monster":
		resolvedNames := []string{}
		resolvedIDs := []string{}
		for _, name := range payload.MonsterTemplateNames {
			entry, ok := resolveMonsterTemplateByName(name, monsterIndex)
			if !ok {
				warnings = append(warnings, fmt.Sprintf("monster template %q could not be resolved", strings.TrimSpace(name)))
				continue
			}
			resolvedNames = append(resolvedNames, entry.Name)
			resolvedIDs = append(resolvedIDs, entry.ID.String())
		}
		if len(resolvedIDs) == 0 {
			warnings = append(warnings, "no monster templates could be resolved")
		}
		step.MonsterTemplateNames = resolvedNames
		step.MonsterTemplateIDs = resolvedIDs
	default:
		warnings = append(warnings, "step content defaulted to challenge")
	}

	return step, warnings
}

var suggestionChallengeScenarioLikePhrases = []string{
	"interview locals",
	"ask locals",
	"ask around",
	"search the area",
	"search for clues",
	"find clues",
	"look for clues",
	"track down",
	"find out",
	"figure out",
	"convince",
	"persuade",
	"negotiate",
	"mediate",
	"resolve the dispute",
	"solve the problem",
	"help settle",
	"what do you do",
	"how would you",
	"how do you",
	"decide how",
	"respond to",
	"intervene",
	"missing sketchbook",
	"missing journal",
	"missing item",
	"lost sketchbook",
	"lost journal",
}

func shouldConvertSuggestionChallengeToScenario(question string, description string) bool {
	trimmedQuestion := strings.TrimSpace(strings.ToLower(question))
	if trimmedQuestion == "" {
		return false
	}

	if strings.HasSuffix(trimmedQuestion, "?") {
		return true
	}

	for _, prefix := range []string{"where ", "who ", "why ", "how ", "what ", "when ", "which "} {
		if strings.HasPrefix(trimmedQuestion, prefix) {
			return true
		}
	}

	combined := strings.ToLower(strings.TrimSpace(question + " " + description))
	for _, phrase := range suggestionChallengeScenarioLikePhrases {
		if strings.Contains(combined, phrase) {
			return true
		}
	}

	return false
}

func buildScenarioPromptFromSuggestionChallenge(step models.QuestArchetypeSuggestionStep) string {
	location := strings.TrimSpace(step.LocationConcept)
	if location == "" {
		location = "location"
	}

	problem := strings.TrimSpace(step.ChallengeDescription)
	if problem == "" {
		problem = strings.TrimSpace(step.ChallengeQuestion)
	}
	problem = ensureSuggestionSentence(problem)
	if problem == "" {
		return fmt.Sprintf("At the %s, a complication unfolds that needs your response. What do you do?", location)
	}
	return fmt.Sprintf("At the %s, this complication unfolds: %s What do you do?", location, problem)
}

func ensureSuggestionSentence(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	last := trimmed[len(trimmed)-1]
	if last == '.' || last == '!' || last == '?' {
		return trimmed
	}
	return trimmed + "."
}

func normalizeSuggestionSource(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "proximity":
		return "proximity"
	default:
		return "location"
	}
}

func normalizeSuggestionContent(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "scenario":
		return "scenario"
	case "monster", "monster_encounter":
		return "monster"
	default:
		return "challenge"
	}
}

func normalizeSuggestionTags(input []string) models.StringArray {
	out := models.StringArray{}
	seen := map[string]struct{}{}
	for _, raw := range input {
		value := strings.TrimSpace(strings.ToLower(raw))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func normalizeSuggestionLines(input []string) models.StringArray {
	out := models.StringArray{}
	seen := map[string]struct{}{}
	for _, raw := range input {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func quotedOrNone(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "none"
	}
	return trimmed
}

func renderTagList(tags []string) string {
	normalized := normalizeSuggestionTags(tags)
	if len(normalized) == 0 {
		return "none"
	}
	return strings.Join(normalized, ", ")
}

func resolveLocationArchetypeByName(
	raw string,
	index map[string]locationArchetypeIndexEntry,
) (locationArchetypeIndexEntry, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return locationArchetypeIndexEntry{}, false
	}
	if entry, ok := index[normalized]; ok {
		return entry, true
	}
	bestScore := 0
	best := locationArchetypeIndexEntry{}
	queryTokens := tokenSet(normalized)
	for key, entry := range index {
		score := overlapScore(queryTokens, tokenSet(key))
		if score > bestScore {
			bestScore = score
			best = entry
		}
	}
	return best, bestScore > 0
}

func resolveMonsterTemplateByName(
	raw string,
	index map[string]monsterTemplateIndexEntry,
) (monsterTemplateIndexEntry, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return monsterTemplateIndexEntry{}, false
	}
	if entry, ok := index[normalized]; ok {
		return entry, true
	}
	bestScore := 0
	best := monsterTemplateIndexEntry{}
	queryTokens := tokenSet(normalized)
	for key, entry := range index {
		score := overlapScore(queryTokens, tokenSet(key))
		if score > bestScore {
			bestScore = score
			best = entry
		}
	}
	return best, bestScore > 0
}

func tokenSet(value string) map[string]struct{} {
	out := map[string]struct{}{}
	normalized := strings.ToLower(strings.TrimSpace(value))
	for _, part := range strings.FieldsFunc(normalized, func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	}) {
		if len(part) < 3 {
			continue
		}
		out[part] = struct{}{}
	}
	return out
}

func overlapScore(a map[string]struct{}, b map[string]struct{}) int {
	score := 0
	for token := range a {
		if _, ok := b[token]; ok {
			score++
		}
	}
	return score
}

func normalizeSuggestionSubmissionType(raw string) models.QuestNodeSubmissionType {
	value := models.QuestNodeSubmissionType(strings.TrimSpace(strings.ToLower(raw)))
	if value.IsValid() {
		return value
	}
	return models.DefaultQuestNodeSubmissionType()
}
