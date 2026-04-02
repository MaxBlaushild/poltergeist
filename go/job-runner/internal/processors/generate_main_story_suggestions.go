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

const mainStorySuggestionPromptTemplate = `
You are designing reusable district-scale main story campaigns for Unclaimed Streets, an urban fantasy MMORPG.

Generate exactly %d main story campaign drafts.
Each draft should represent a complete, coherent main story with exactly %d quests/beats.

Requested direction:
- theme prompt: %s
- district fit: %s
- tone: %s
- family tags: %s
- character tags to bias toward: %s
- internal tags to bias toward: %s
- required location archetypes that must appear across the campaign: %s
- required location metadata tags to use when appropriate: %s

Recent main story templates to avoid echoing:
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
      "premise": "string",
      "districtFit": "string",
      "tone": "string",
      "themeTags": ["tag_a", "tag_b"],
      "internalTags": ["tag_a", "tag_b"],
      "factionKeys": ["tag_a", "tag_b"],
      "characterKeys": ["tag_a", "tag_b"],
      "revealKeys": ["tag_a", "tag_b"],
      "climaxSummary": "string",
      "resolutionSummary": "string",
      "whyItWorks": "string",
      "beats": [
        {
          "orderIndex": 1,
          "act": 1,
          "storyRole": "inciting_incident",
          "chapterTitle": "string",
          "chapterSummary": "string",
          "purpose": "string",
          "whatChanges": "string",
          "introducedCharacterKeys": ["tag_a"],
          "requiredCharacterKeys": ["tag_b"],
          "introducedRevealKeys": ["tag_a"],
          "requiredRevealKeys": ["tag_b"],
          "requiredZoneTags": ["market", "nightlife"],
          "requiredLocationArchetypeNames": ["name from allowed location archetypes"],
          "preferredContentMix": ["challenge", "scenario"],
          "questGiverCharacterKey": "tag_a",
          "name": "quest archetype name",
          "hook": "one line hook",
          "description": "quest archetype description",
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
  ]
}

Rules:
- Output exactly %d drafts.
- Each draft must contain exactly %d beats.
- Output JSON only. No markdown.
- The campaign should read like a complete book or game main story, not a loose anthology.
- Use a three-act feeling across the beats.
- Every beat must advance the same central conflict.
- Early beats introduce people, factions, and mysteries. Mid beats escalate and complicate. Late beats reveal, climax, and resolve.
- Reuse recurring characters and reveals consistently with the introduced/required keys.
- Each beat must include questGiverCharacterKey, chosen from that beat's introducedCharacterKeys or requiredCharacterKeys whenever possible.
- Each beat must be a valid quest archetype package that can stand on its own while still serving the larger story.
- Challenge steps must be concrete, enjoyable real-world tasks the player can actually complete at the location right now.
- Investigation, negotiation, helping someone, or “what do you do?” situations should become scenarios, not challenges.
- Every campaign must feel district-specific and grounded in urban fantasy.
- Use lowercase snake_case tags/keys for themeTags, internalTags, factionKeys, characterKeys, revealKeys, requiredZoneTags, characterTags, and preferredContentMix.
- Every beat must include 1-3 requiredZoneTags.
- Every beat must include at least one requiredLocationArchetypeName.
- Across the full campaign, include each required location archetype at least once when a required list is provided.
- Vary the content mix across beats. Do not make every beat combat-heavy.
`

type mainStorySuggestionResponse struct {
	Drafts []mainStorySuggestionDraftPayload `json:"drafts"`
}

type mainStorySuggestionDraftPayload struct {
	Name              string                           `json:"name"`
	Premise           string                           `json:"premise"`
	DistrictFit       string                           `json:"districtFit"`
	Tone              string                           `json:"tone"`
	ThemeTags         []string                         `json:"themeTags"`
	InternalTags      []string                         `json:"internalTags"`
	FactionKeys       []string                         `json:"factionKeys"`
	CharacterKeys     []string                         `json:"characterKeys"`
	RevealKeys        []string                         `json:"revealKeys"`
	ClimaxSummary     string                           `json:"climaxSummary"`
	ResolutionSummary string                           `json:"resolutionSummary"`
	WhyItWorks        string                           `json:"whyItWorks"`
	Beats             []mainStorySuggestionBeatPayload `json:"beats"`
}

type mainStorySuggestionBeatPayload struct {
	OrderIndex                     int                                   `json:"orderIndex"`
	Act                            int                                   `json:"act"`
	StoryRole                      string                                `json:"storyRole"`
	ChapterTitle                   string                                `json:"chapterTitle"`
	ChapterSummary                 string                                `json:"chapterSummary"`
	Purpose                        string                                `json:"purpose"`
	WhatChanges                    string                                `json:"whatChanges"`
	IntroducedCharacterKeys        []string                              `json:"introducedCharacterKeys"`
	RequiredCharacterKeys          []string                              `json:"requiredCharacterKeys"`
	IntroducedRevealKeys           []string                              `json:"introducedRevealKeys"`
	RequiredRevealKeys             []string                              `json:"requiredRevealKeys"`
	RequiredZoneTags               []string                              `json:"requiredZoneTags"`
	RequiredLocationArchetypeNames []string                              `json:"requiredLocationArchetypeNames"`
	PreferredContentMix            []string                              `json:"preferredContentMix"`
	QuestGiverCharacterKey         string                                `json:"questGiverCharacterKey"`
	Name                           string                                `json:"name"`
	Hook                           string                                `json:"hook"`
	Description                    string                                `json:"description"`
	AcceptanceDialogue             []string                              `json:"acceptanceDialogue"`
	CharacterTags                  []string                              `json:"characterTags"`
	InternalTags                   []string                              `json:"internalTags"`
	DifficultyMode                 string                                `json:"difficultyMode"`
	Difficulty                     int                                   `json:"difficulty"`
	MonsterEncounterTargetLevel    int                                   `json:"monsterEncounterTargetLevel"`
	WhyThisScales                  string                                `json:"whyThisScales"`
	ChallengeTemplateSeeds         []string                              `json:"challengeTemplateSeeds"`
	ScenarioTemplateSeeds          []string                              `json:"scenarioTemplateSeeds"`
	MonsterTemplateSeeds           []string                              `json:"monsterTemplateSeeds"`
	Steps                          []questArchetypeSuggestionStepPayload `json:"steps"`
}

type GenerateMainStorySuggestionsProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
}

func NewGenerateMainStorySuggestionsProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
) GenerateMainStorySuggestionsProcessor {
	log.Println("Initializing GenerateMainStorySuggestionsProcessor")
	return GenerateMainStorySuggestionsProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
	}
}

func (p *GenerateMainStorySuggestionsProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing main story suggestion task: %s", task.Type())

	var payload jobs.GenerateMainStorySuggestionsTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := p.dbClient.MainStorySuggestionJob().FindByID(ctx, payload.JobID)
	if err != nil {
		return err
	}
	if job == nil {
		return nil
	}

	job.Status = models.MainStorySuggestionJobStatusInProgress
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.MainStorySuggestionJob().Update(ctx, job); err != nil {
		return err
	}

	if err := p.generateDrafts(ctx, job); err != nil {
		return p.failJob(ctx, job, err)
	}
	return nil
}

func (p *GenerateMainStorySuggestionsProcessor) generateDrafts(
	ctx context.Context,
	job *models.MainStorySuggestionJob,
) error {
	locationArchetypes, err := p.dbClient.LocationArchetype().FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load location archetypes: %w", err)
	}
	monsterTemplates, err := p.dbClient.MonsterTemplate().FindAllActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to load monster templates: %w", err)
	}
	recentTemplates, err := p.dbClient.MainStoryTemplate().FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load main story templates: %w", err)
	}

	prompt := fmt.Sprintf(
		mainStorySuggestionPromptTemplate,
		maxInt(1, job.Count),
		maxInt(3, job.QuestCount),
		quotedOrNone(job.ThemePrompt),
		quotedOrNone(job.DistrictFit),
		quotedOrNone(job.Tone),
		renderTagList(job.FamilyTags),
		renderTagList(job.CharacterTags),
		renderTagList(job.InternalTags),
		buildRequiredLocationArchetypesPrompt(job.RequiredLocationArchetypeIDs, locationArchetypes),
		renderTagList(job.RequiredLocationMetadataTags),
		buildMainStorySuggestionAvoidance(recentTemplates, 12),
		buildAllowedLocationArchetypesPrompt(locationArchetypes),
		buildAllowedMonsterTemplatesPrompt(monsterTemplates),
		maxInt(1, job.Count),
		maxInt(3, job.QuestCount),
	)

	answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return fmt.Errorf("failed to generate main story suggestions: %w", err)
	}

	generated := &mainStorySuggestionResponse{}
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), generated); err != nil {
		return fmt.Errorf("failed to parse main story suggestion payload: %w", err)
	}
	if len(generated.Drafts) == 0 {
		return fmt.Errorf("main story suggestion payload did not include any drafts")
	}

	locationIndex := buildLocationArchetypeIndex(locationArchetypes)
	monsterIndex := buildMonsterTemplateIndex(monsterTemplates)
	requiredLocationArchetypes := resolveRequiredLocationArchetypes(
		job.RequiredLocationArchetypeIDs,
		locationArchetypes,
	)

	createdCount := 0
	for _, spec := range generated.Drafts {
		draft := sanitizeMainStorySuggestionDraft(spec, locationIndex, monsterIndex, requiredLocationArchetypes, job)
		draft.JobID = job.ID
		draft.Status = models.MainStorySuggestionDraftStatusSuggested
		if err := p.dbClient.MainStorySuggestionDraft().Create(ctx, draft); err != nil {
			job.CreatedCount = createdCount
			return fmt.Errorf("failed to create main story suggestion draft: %w", err)
		}
		createdCount++
	}

	job.CreatedCount = createdCount
	job.Status = models.MainStorySuggestionJobStatusCompleted
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.MainStorySuggestionJob().Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update main story suggestion job: %w", err)
	}
	return nil
}

func (p *GenerateMainStorySuggestionsProcessor) failJob(
	ctx context.Context,
	job *models.MainStorySuggestionJob,
	err error,
) error {
	msg := err.Error()
	job.Status = models.MainStorySuggestionJobStatusFailed
	job.ErrorMessage = &msg
	job.UpdatedAt = time.Now()
	if updateErr := p.dbClient.MainStorySuggestionJob().Update(ctx, job); updateErr != nil {
		log.Printf("Failed to mark main story suggestion job %s as failed: %v", job.ID, updateErr)
	}
	return err
}

func buildMainStorySuggestionAvoidance(recent []models.MainStoryTemplate, limit int) string {
	if len(recent) == 0 {
		return "- none"
	}
	items := append([]models.MainStoryTemplate{}, recent...)
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = "Unnamed main story"
		}
		premise := strings.TrimSpace(item.Premise)
		if len(premise) > 120 {
			premise = strings.TrimSpace(premise[:120]) + "..."
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", name, premise))
	}
	if len(lines) == 0 {
		return "- none"
	}
	return strings.Join(lines, "\n")
}

func sanitizeMainStorySuggestionDraft(
	payload mainStorySuggestionDraftPayload,
	locationIndex map[string]locationArchetypeIndexEntry,
	monsterIndex map[string]monsterTemplateIndexEntry,
	requiredLocationArchetypes []locationArchetypeIndexEntry,
	job *models.MainStorySuggestionJob,
) *models.MainStorySuggestionDraft {
	now := time.Now()
	warnings := models.StringArray{}
	beats := make(models.MainStoryBeatDrafts, 0, len(payload.Beats))
	usedRequiredLocations := map[uuid.UUID]struct{}{}

	for index, rawBeat := range payload.Beats {
		beat := sanitizeMainStorySuggestionBeat(rawBeat, locationIndex, monsterIndex, index+1)
		for _, id := range beat.RequiredLocationArchetypeIDs {
			parsed, err := uuid.Parse(strings.TrimSpace(id))
			if err == nil && parsed != uuid.Nil {
				usedRequiredLocations[parsed] = struct{}{}
			}
		}
		for _, step := range beat.Steps {
			if step.LocationArchetypeID != nil {
				usedRequiredLocations[*step.LocationArchetypeID] = struct{}{}
			}
		}
		for _, warning := range beat.Warnings {
			warnings = append(warnings, fmt.Sprintf("beat %d: %s", beat.OrderIndex, warning))
		}
		beats = append(beats, beat)
	}

	for _, required := range requiredLocationArchetypes {
		if _, ok := usedRequiredLocations[required.ID]; ok {
			continue
		}
		warnings = append(warnings, fmt.Sprintf("required location archetype %q was not used in this draft", required.Name))
	}
	if len(beats) != maxInt(1, job.QuestCount) {
		warnings = append(warnings, fmt.Sprintf("expected %d beats but generated %d", maxInt(1, job.QuestCount), len(beats)))
	}

	name := strings.TrimSpace(payload.Name)
	if name == "" {
		name = "Generated Main Story Draft"
		warnings = append(warnings, "name was empty and replaced with a fallback")
	}
	premise := strings.TrimSpace(payload.Premise)
	if premise == "" {
		premise = "A district-scale urban fantasy mystery unfolds."
		warnings = append(warnings, "premise was empty and replaced with a fallback")
	}

	return &models.MainStorySuggestionDraft{
		ID:                uuid.New(),
		CreatedAt:         now,
		UpdatedAt:         now,
		Status:            models.MainStorySuggestionDraftStatusSuggested,
		Name:              name,
		Premise:           premise,
		DistrictFit:       strings.TrimSpace(payload.DistrictFit),
		Tone:              strings.TrimSpace(payload.Tone),
		ThemeTags:         normalizeSuggestionTags(payload.ThemeTags),
		InternalTags:      normalizeSuggestionTags(payload.InternalTags),
		FactionKeys:       normalizeSuggestionTags(payload.FactionKeys),
		CharacterKeys:     normalizeSuggestionTags(payload.CharacterKeys),
		RevealKeys:        normalizeSuggestionTags(payload.RevealKeys),
		ClimaxSummary:     strings.TrimSpace(payload.ClimaxSummary),
		ResolutionSummary: strings.TrimSpace(payload.ResolutionSummary),
		WhyItWorks:        strings.TrimSpace(payload.WhyItWorks),
		Beats:             beats,
		Warnings:          normalizeSuggestionLines(warnings),
	}
}

func sanitizeMainStorySuggestionBeat(
	payload mainStorySuggestionBeatPayload,
	locationIndex map[string]locationArchetypeIndexEntry,
	monsterIndex map[string]monsterTemplateIndexEntry,
	fallbackOrder int,
) models.MainStoryBeatDraft {
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

	requiredLocationIDs := make(models.StringArray, 0, len(payload.RequiredLocationArchetypeNames))
	for _, rawName := range payload.RequiredLocationArchetypeNames {
		entry, ok := resolveLocationArchetypeByName(rawName, locationIndex)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("required location archetype %q could not be resolved", strings.TrimSpace(rawName)))
			continue
		}
		requiredLocationIDs = append(requiredLocationIDs, entry.ID.String())
	}

	difficultyMode := models.NormalizeQuestDifficultyMode(payload.DifficultyMode)
	difficulty := models.NormalizeQuestDifficulty(payload.Difficulty)
	monsterLevel := models.NormalizeMonsterEncounterTargetLevel(payload.MonsterEncounterTargetLevel)
	orderIndex := payload.OrderIndex
	if orderIndex <= 0 {
		orderIndex = fallbackOrder
		warnings = append(warnings, "orderIndex was missing and replaced with a fallback")
	}
	act := payload.Act
	if act <= 0 {
		switch {
		case orderIndex >= 11:
			act = 3
		case orderIndex >= 6:
			act = 2
		default:
			act = 1
		}
	}

	name := strings.TrimSpace(payload.Name)
	if name == "" {
		name = fmt.Sprintf("Generated Story Beat %d", orderIndex)
		warnings = append(warnings, "quest name was empty and replaced with a fallback")
	}

	return models.MainStoryBeatDraft{
		OrderIndex:                   orderIndex,
		Act:                          act,
		StoryRole:                    strings.TrimSpace(strings.ToLower(payload.StoryRole)),
		ChapterTitle:                 strings.TrimSpace(payload.ChapterTitle),
		ChapterSummary:               strings.TrimSpace(payload.ChapterSummary),
		Purpose:                      strings.TrimSpace(payload.Purpose),
		WhatChanges:                  strings.TrimSpace(payload.WhatChanges),
		IntroducedCharacterKeys:      normalizeSuggestionTags(payload.IntroducedCharacterKeys),
		RequiredCharacterKeys:        normalizeSuggestionTags(payload.RequiredCharacterKeys),
		IntroducedRevealKeys:         normalizeSuggestionTags(payload.IntroducedRevealKeys),
		RequiredRevealKeys:           normalizeSuggestionTags(payload.RequiredRevealKeys),
		RequiredZoneTags:             normalizeSuggestionTags(payload.RequiredZoneTags),
		RequiredLocationArchetypeIDs: requiredLocationIDs,
		PreferredContentMix:          normalizeSuggestionTags(payload.PreferredContentMix),
		QuestGiverCharacterKey:       sanitizeMainStoryQuestGiverCharacterKey(payload),
		Name:                         name,
		Hook:                         strings.TrimSpace(payload.Hook),
		Description:                  strings.TrimSpace(payload.Description),
		AcceptanceDialogue:           normalizeSuggestionLines(payload.AcceptanceDialogue),
		CharacterTags:                normalizeSuggestionTags(payload.CharacterTags),
		InternalTags:                 normalizeSuggestionTags(payload.InternalTags),
		DifficultyMode:               difficultyMode,
		Difficulty:                   difficulty,
		MonsterEncounterTargetLevel:  monsterLevel,
		WhyThisScales:                strings.TrimSpace(payload.WhyThisScales),
		Steps:                        steps,
		ChallengeTemplateSeeds:       normalizeSuggestionLines(payload.ChallengeTemplateSeeds),
		ScenarioTemplateSeeds:        normalizeSuggestionLines(payload.ScenarioTemplateSeeds),
		MonsterTemplateSeeds:         normalizeSuggestionLines(payload.MonsterTemplateSeeds),
		Warnings:                     normalizeSuggestionLines(warnings),
	}
}

func sanitizeMainStoryQuestGiverCharacterKey(
	payload mainStorySuggestionBeatPayload,
) string {
	candidate := strings.TrimSpace(strings.ToLower(payload.QuestGiverCharacterKey))
	available := normalizeSuggestionTags(
		append(
			append([]string{}, payload.RequiredCharacterKeys...),
			payload.IntroducedCharacterKeys...,
		),
	)
	if candidate != "" {
		for _, key := range available {
			if key == candidate {
				return candidate
			}
		}
	}
	if len(available) > 0 {
		return available[0]
	}
	return candidate
}
