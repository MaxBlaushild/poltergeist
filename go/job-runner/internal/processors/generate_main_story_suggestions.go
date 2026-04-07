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
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
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
          "requiredLocationArchetypeNames": ["existing or new reusable location archetype name"],
          "preferredContentMix": ["challenge", "scenario"],
          "questGiverCharacterKey": "tag_a",
          "name": "quest archetype name",
          "hook": "one line hook",
          "description": "quest archetype description",
          "acceptanceDialogue": ["line 1", "line 2", "line 3"],
          "requiredStoryFlags": ["optional_flag"],
          "setStoryFlags": ["optional_flag"],
          "clearStoryFlags": ["optional_flag"],
          "questGiverRelationshipEffects": {
            "trust": 1,
            "respect": 0,
            "fear": 0,
            "debt": -1
          },
          "worldChanges": [
            {
              "type": "show_poi_text",
              "targetKey": "quest_giver_poi",
              "description": "optional one-paragraph change to the surrounding place after this beat",
              "clue": "optional short clue update"
            },
            {
              "type": "move_character",
              "targetKey": "quest_giver",
              "destinationHint": "brief destination vibe like chapel refuge or hidden back room"
            }
          ],
          "unlockedScenarios": [
            {
              "name": "string",
              "prompt": "open-ended scenario prompt unlocked after this beat",
              "pointOfInterestHint": "optional nearby place vibe",
              "internalTags": ["tag_a", "tag_b"],
              "difficulty": 2
            }
          ],
          "unlockedChallenges": [
            {
              "question": "concrete enjoyable task prompt",
              "description": "what the player actually does on site",
              "pointOfInterestHint": "optional nearby place vibe",
              "submissionType": "photo|text|video",
              "proficiency": "optional short proficiency",
              "statTags": ["strength", "wisdom"],
              "difficulty": 2
            }
          ],
          "unlockedMonsterEncounters": [
            {
              "name": "string",
              "description": "what threat appears in the district after this beat",
              "pointOfInterestHint": "optional nearby place vibe",
              "encounterType": "monster|boss",
              "monsterCount": 2,
              "encounterTone": ["urban", "predatory"],
              "monsterTemplateHints": ["feral choir brute", "ash-fed scavenger"],
              "targetLevel": 8
            }
          ],
          "questGiverAfterDescription": "optional one-paragraph ambient change for the quest giver after this beat is complete",
          "questGiverAfterDialogue": ["line 1", "line 2"],
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
              "locationArchetypeName": "existing or new reusable location archetype name, or empty string for proximity steps",
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
- Each beat should feel like a chapter with its own beginning, middle, and turn, not a single paragraph wrapped in one open-ended prompt.
- Every beat should usually break into 2-3 quest steps/subbeats. A single-step beat is only acceptable if the beat is intentionally very short and still feels complete.
- The beat-level chapterSummary, purpose, whatChanges, hook, and description should clearly establish who is involved, what is happening, where it happens, and why this beat matters now.
- Each step must represent a distinct subbeat with a clear purpose, not just a restatement of the overall beat summary.
- Vary step content inside a beat when it makes sense. A good beat often mixes discovery, action, and consequence instead of repeating the same interaction type.
- Challenge steps must be concrete, enjoyable real-world tasks the player can actually complete at the location right now.
- Investigation, negotiation, helping someone, or “what do you do?” situations should become scenarios, not challenges.
- Scenario steps must be specific dramatic situations with a concrete local context, not generic prompts that simply restate the beat summary.
- Monster steps should represent an actual escalation, ambush, guardian, or consequence tied to the beat, not filler combat.
- Every campaign must feel district-specific and grounded in urban fantasy.
- Use lowercase snake_case tags/keys for themeTags, internalTags, factionKeys, characterKeys, revealKeys, requiredZoneTags, characterTags, and preferredContentMix.
- Use lowercase snake_case for requiredStoryFlags, setStoryFlags, and clearStoryFlags when you include them.
- questGiverRelationshipEffects should use sparse, small integers between -3 and 3 and only when the beat meaningfully changes how that NPC feels about the player.
- worldChanges should be sparse and legible. Use only move_character or show_poi_text in this format.
- show_poi_text should usually target quest_giver_poi and describe how the place changes after the beat.
- move_character should usually target quest_giver and use destinationHint to describe the kind of place they relocate to.
- unlockedScenarios should be sparse, open-ended, and feel like optional story fallout the player can meaningfully engage with after the beat.
- unlockedChallenges should be concrete, fun, gradeable real-world tasks that make sense at the hinted location.
- unlockedMonsterEncounters should feel like consequences or escalating threats caused by the beat, not random filler combat.
- Every beat must include 1-3 requiredZoneTags.
- Every beat must include at least one requiredLocationArchetypeName.
- requiredLocationArchetypeNames and step locationArchetypeName values may reference an existing allowed archetype or propose a new reusable archetype if the story needs one.
- When proposing a new archetype, keep the name practical, atmospheric, and reusable across many districts.
- Across the full campaign, include each required location archetype at least once when a required list is provided.
- Vary the content mix across beats. Do not make every beat combat-heavy.
- questGiverAfterDialogue should feel like reactive NPC follow-up dialogue that acknowledges what changed after this beat.
`

const mainStoryBeatExpansionPromptTemplate = `
You are upgrading one main story beat for Unclaimed Streets so it becomes a richer, more varied quest.

Campaign context:
- campaign premise: %s
- district fit: %s
- tone: %s
- theme tags: %s
- character keys in play: %s
- reveal keys in play: %s

Beat context:
- order index: %d
- act: %d
- story role: %s
- chapter title: %s
- chapter summary: %s
- purpose: %s
- what changes: %s
- hook: %s
- description: %s
- preferred content mix: %s
- character tags: %s
- required zone tags: %s
- required location archetype names: %s
- challenge template seeds: %s
- scenario template seeds: %s
- monster template seeds: %s
- current steps json: %s

Allowed location archetypes:
%s

Allowed monster templates:
%s

Return JSON only:
{
  "chapterSummary": "optional improved beat summary with stronger who/what/where/why",
  "hook": "optional improved hook",
  "description": "optional improved quest archetype description",
  "challengeTemplateSeeds": ["seed one", "seed two"],
  "scenarioTemplateSeeds": ["seed one", "seed two"],
  "monsterTemplateSeeds": ["seed one", "seed two"],
  "steps": [
    {
      "source": "location|proximity",
      "content": "challenge|scenario|monster",
      "locationConcept": "specific subbeat location vibe",
      "locationArchetypeName": "existing or new reusable location archetype name, or empty string for proximity steps",
      "locationMetadataTags": ["market", "nightlife"],
      "distanceMeters": 120,
      "templateConcept": "what the player is doing in this subbeat",
      "potentialContent": ["idea one", "idea two"],
      "challengeQuestion": "required for challenge steps",
      "challengeDescription": "required for challenge steps",
      "challengeSubmissionType": "photo|text|video",
      "challengeProficiency": "optional short proficiency",
      "challengeStatTags": ["strength", "wisdom"],
      "scenarioPrompt": "required for scenario steps",
      "scenarioOpenEnded": true,
      "scenarioBeats": ["beat one", "beat two"],
      "monsterTemplateNames": ["names chosen from allowed monster templates when possible"],
      "encounterTone": ["urban", "predatory"]
    }
  ]
}

Rules:
- Output 2-3 steps unless there is a very strong reason for only 1.
- The first step must not use proximity.
- Each step must feel like a distinct subbeat with its own local purpose.
- Do not restate the same scene three times. Escalate or pivot across the steps.
- Use challenges only for concrete enjoyable real-world tasks that can be completed right now at the location.
- Use scenarios for negotiations, discoveries, moral choices, roleplay, or response-driven situations.
- Use monster encounters when the beat genuinely calls for danger, pursuit, a guardian, or a violent consequence.
- If the beat involves a specific monster concept and no perfect template exists, use monsterTemplateSeeds and/or monsterTemplateNames to point toward the needed monster type.
- If the beat needs a more specific place than the provided list, you may propose a new reusable location archetype name.
- Keep everything grounded in the same chapter title, purpose, and what-changes arc.
- Make the step prompts specific and flavorful, not generic.
- Output JSON only.
`

const mainStoryMissingLocationArchetypePromptTemplate = `
You are defining reusable location archetypes for an urban fantasy MMORPG.

The story generator proposed location archetype names that do not already exist. For each missing name, map it to sensible Google place types so it can become a real reusable archetype.

Story context:
- theme prompt: %s
- district fit: %s
- tone: %s

Existing archetype names to avoid duplicating semantically:
%s

Allowed Google place types:
%s

Missing location archetype names that need definitions:
%s

Return JSON only:
{
  "archetypes": [
    {
      "name": "must exactly match one of the missing names",
      "includedTypes": ["1-6 exact place type values from the allowed list"],
      "excludedTypes": ["0-6 exact place type values from the allowed list"]
    }
  ]
}

Rules:
- Output one entry for each missing name when possible.
- The name must exactly match one of the missing names.
- includedTypes and excludedTypes must use exact allowed Google place types.
- Favor broad, reusable mappings over narrow one-off concepts.
- Keep excludedTypes sparse.
- Do not invent extra names not in the missing list.
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
	OrderIndex                     int                                     `json:"orderIndex"`
	Act                            int                                     `json:"act"`
	StoryRole                      string                                  `json:"storyRole"`
	ChapterTitle                   string                                  `json:"chapterTitle"`
	ChapterSummary                 string                                  `json:"chapterSummary"`
	Purpose                        string                                  `json:"purpose"`
	WhatChanges                    string                                  `json:"whatChanges"`
	IntroducedCharacterKeys        []string                                `json:"introducedCharacterKeys"`
	RequiredCharacterKeys          []string                                `json:"requiredCharacterKeys"`
	IntroducedRevealKeys           []string                                `json:"introducedRevealKeys"`
	RequiredRevealKeys             []string                                `json:"requiredRevealKeys"`
	RequiredZoneTags               []string                                `json:"requiredZoneTags"`
	RequiredLocationArchetypeNames []string                                `json:"requiredLocationArchetypeNames"`
	PreferredContentMix            []string                                `json:"preferredContentMix"`
	QuestGiverCharacterKey         string                                  `json:"questGiverCharacterKey"`
	Name                           string                                  `json:"name"`
	Hook                           string                                  `json:"hook"`
	Description                    string                                  `json:"description"`
	AcceptanceDialogue             []string                                `json:"acceptanceDialogue"`
	RequiredStoryFlags             []string                                `json:"requiredStoryFlags"`
	SetStoryFlags                  []string                                `json:"setStoryFlags"`
	ClearStoryFlags                []string                                `json:"clearStoryFlags"`
	QuestGiverRelationshipEffects  models.CharacterRelationshipState       `json:"questGiverRelationshipEffects"`
	WorldChanges                   []mainStorySuggestionWorldChangePayload `json:"worldChanges"`
	UnlockedScenarios              []mainStorySuggestionScenarioPayload    `json:"unlockedScenarios"`
	UnlockedChallenges             []mainStorySuggestionChallengePayload   `json:"unlockedChallenges"`
	UnlockedMonsterEncounters      []mainStorySuggestionEncounterPayload   `json:"unlockedMonsterEncounters"`
	QuestGiverAfterDescription     string                                  `json:"questGiverAfterDescription"`
	QuestGiverAfterDialogue        []string                                `json:"questGiverAfterDialogue"`
	CharacterTags                  []string                                `json:"characterTags"`
	InternalTags                   []string                                `json:"internalTags"`
	DifficultyMode                 string                                  `json:"difficultyMode"`
	Difficulty                     int                                     `json:"difficulty"`
	MonsterEncounterTargetLevel    int                                     `json:"monsterEncounterTargetLevel"`
	WhyThisScales                  string                                  `json:"whyThisScales"`
	ChallengeTemplateSeeds         []string                                `json:"challengeTemplateSeeds"`
	ScenarioTemplateSeeds          []string                                `json:"scenarioTemplateSeeds"`
	MonsterTemplateSeeds           []string                                `json:"monsterTemplateSeeds"`
	Steps                          []questArchetypeSuggestionStepPayload   `json:"steps"`
}

type mainStorySuggestionWorldChangePayload struct {
	Type                string   `json:"type"`
	TargetKey           string   `json:"targetKey"`
	CharacterKey        string   `json:"characterKey"`
	PointOfInterestHint string   `json:"pointOfInterestHint"`
	DestinationHint     string   `json:"destinationHint"`
	ZoneTags            []string `json:"zoneTags"`
	Description         string   `json:"description"`
	Clue                string   `json:"clue"`
}

type mainStorySuggestionScenarioPayload struct {
	Name                string   `json:"name"`
	Prompt              string   `json:"prompt"`
	PointOfInterestHint string   `json:"pointOfInterestHint"`
	InternalTags        []string `json:"internalTags"`
	Difficulty          int      `json:"difficulty"`
}

type mainStorySuggestionChallengePayload struct {
	Question            string   `json:"question"`
	Description         string   `json:"description"`
	PointOfInterestHint string   `json:"pointOfInterestHint"`
	SubmissionType      string   `json:"submissionType"`
	Proficiency         string   `json:"proficiency"`
	StatTags            []string `json:"statTags"`
	Difficulty          int      `json:"difficulty"`
}

type mainStorySuggestionEncounterPayload struct {
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	PointOfInterestHint  string   `json:"pointOfInterestHint"`
	EncounterType        string   `json:"encounterType"`
	MonsterCount         int      `json:"monsterCount"`
	EncounterTone        []string `json:"encounterTone"`
	MonsterTemplateHints []string `json:"monsterTemplateHints"`
	TargetLevel          int      `json:"targetLevel"`
}

type mainStoryBeatExpansionResponse struct {
	ChapterSummary         string                                `json:"chapterSummary"`
	Hook                   string                                `json:"hook"`
	Description            string                                `json:"description"`
	ChallengeTemplateSeeds []string                              `json:"challengeTemplateSeeds"`
	ScenarioTemplateSeeds  []string                              `json:"scenarioTemplateSeeds"`
	MonsterTemplateSeeds   []string                              `json:"monsterTemplateSeeds"`
	Steps                  []questArchetypeSuggestionStepPayload `json:"steps"`
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

	targetDraftCount := maxInt(1, job.Count)
	targetQuestCount := maxInt(3, job.QuestCount)
	generatedDrafts := make([]mainStorySuggestionDraftPayload, 0, targetDraftCount)
	for len(generatedDrafts) < targetDraftCount {
		prompt := buildMainStorySuggestionPrompt(
			job,
			locationArchetypes,
			monsterTemplates,
			recentTemplates,
			generatedDrafts,
			1,
			targetQuestCount,
		)

		generated, err := p.requestMainStorySuggestionResponse(prompt)
		if err != nil {
			return err
		}
		if len(generated.Drafts) == 0 {
			return fmt.Errorf("main story suggestion payload did not include any drafts")
		}

		for _, draft := range generated.Drafts {
			if len(generatedDrafts) >= targetDraftCount {
				break
			}
			generatedDrafts = append(generatedDrafts, draft)
		}
	}

	locationArchetypes, err = p.ensureGeneratedMainStoryLocationArchetypes(
		ctx,
		job,
		locationArchetypes,
		generatedDrafts,
	)
	if err != nil {
		return fmt.Errorf("failed to create missing location archetypes: %w", err)
	}

	generatedDrafts = p.enrichMainStoryDraftBeats(
		job,
		generatedDrafts,
		locationArchetypes,
		monsterTemplates,
	)

	locationIndex := buildLocationArchetypeIndex(locationArchetypes)
	monsterIndex := buildMonsterTemplateIndex(monsterTemplates)
	requiredLocationArchetypes := resolveRequiredLocationArchetypes(
		job.RequiredLocationArchetypeIDs,
		locationArchetypes,
	)

	createdCount := 0
	for _, spec := range generatedDrafts {
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

func buildMainStorySuggestionPrompt(
	job *models.MainStorySuggestionJob,
	locationArchetypes []*models.LocationArchetype,
	monsterTemplates []models.MonsterTemplate,
	recentTemplates []models.MainStoryTemplate,
	generatedDrafts []mainStorySuggestionDraftPayload,
	draftCount int,
	questCount int,
) string {
	avoidance := buildMainStorySuggestionAvoidance(recentTemplates, 12)
	inBatchAvoidance := buildGeneratedMainStoryDraftAvoidance(generatedDrafts, 8)
	if inBatchAvoidance != "" {
		avoidance = strings.TrimSpace(avoidance) + "\nDrafts already generated in this batch to avoid echoing:\n" + inBatchAvoidance
	}
	return fmt.Sprintf(
		mainStorySuggestionPromptTemplate,
		maxInt(1, draftCount),
		maxInt(3, questCount),
		quotedOrNone(job.ThemePrompt),
		quotedOrNone(job.DistrictFit),
		quotedOrNone(job.Tone),
		renderTagList(job.FamilyTags),
		renderTagList(job.CharacterTags),
		renderTagList(job.InternalTags),
		buildRequiredLocationArchetypesPrompt(job.RequiredLocationArchetypeIDs, locationArchetypes),
		renderTagList(job.RequiredLocationMetadataTags),
		avoidance,
		buildAllowedLocationArchetypesPrompt(locationArchetypes),
		buildAllowedMonsterTemplatesPrompt(monsterTemplates),
		maxInt(1, draftCount),
		maxInt(3, questCount),
	)
}

func (p *GenerateMainStorySuggestionsProcessor) requestMainStorySuggestionResponse(
	basePrompt string,
) (*mainStorySuggestionResponse, error) {
	var lastErr error
	prompts := []string{
		basePrompt,
		basePrompt + "\n\nIMPORTANT: Your last response was malformed or incomplete. Return the full JSON object in one complete response with no markdown, no commentary, and no truncation.",
		basePrompt + "\n\nIMPORTANT: Return a smaller, cleaner response body if needed, but it must still be valid complete JSON matching the schema exactly. Do not stop early.",
	}
	for attempt, prompt := range prompts {
		answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
		if err != nil {
			lastErr = fmt.Errorf("failed to generate main story suggestions: %w", err)
			continue
		}

		generated := &mainStorySuggestionResponse{}
		rawJSON := extractGeneratedJSONObject(answer.Answer)
		if err := json.Unmarshal([]byte(rawJSON), generated); err != nil {
			lastErr = fmt.Errorf("failed to parse main story suggestion payload on attempt %d: %w", attempt+1, err)
			log.Printf(
				"main story suggestion payload parse failed on attempt %d: %v | payload preview=%q",
				attempt+1,
				err,
				truncateForMainStoryLog(rawJSON, 500),
			)
			continue
		}
		return generated, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("failed to generate main story suggestions")
	}
	return nil, lastErr
}

func truncateForMainStoryLog(raw string, limit int) string {
	trimmed := strings.TrimSpace(raw)
	if limit <= 0 || len(trimmed) <= limit {
		return trimmed
	}
	return trimmed[:limit] + "..."
}

func collectMissingMainStoryLocationArchetypeNames(
	locationIndex map[string]locationArchetypeIndexEntry,
	drafts []mainStorySuggestionDraftPayload,
) []string {
	missing := make([]string, 0)
	seen := map[string]struct{}{}
	addIfMissing := func(raw string) {
		name := collapseWhitespace(raw)
		if name == "" {
			return
		}
		if _, ok := resolveLocationArchetypeByName(name, locationIndex); ok {
			return
		}
		key := normalizeLocationArchetypeNameKey(name)
		if key == "" {
			return
		}
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		missing = append(missing, name)
	}
	for _, draft := range drafts {
		for _, beat := range draft.Beats {
			for _, name := range beat.RequiredLocationArchetypeNames {
				addIfMissing(name)
			}
			for _, step := range beat.Steps {
				addIfMissing(step.LocationArchetypeName)
			}
		}
	}
	sort.Strings(missing)
	return missing
}

func buildMissingMainStoryLocationArchetypePrompt(
	job *models.MainStorySuggestionJob,
	existing []*models.LocationArchetype,
	missing []string,
) string {
	allowedPlaceTypes := googlemaps.GetAllPlaceTypes()
	allowedNames := make([]string, 0, len(allowedPlaceTypes))
	for _, placeType := range allowedPlaceTypes {
		allowedNames = append(allowedNames, string(placeType))
	}
	existingNames := make([]string, 0, len(existing))
	for _, archetype := range existing {
		if archetype == nil {
			continue
		}
		name := collapseWhitespace(archetype.Name)
		if name == "" {
			continue
		}
		existingNames = append(existingNames, name)
	}
	sort.Strings(existingNames)
	return fmt.Sprintf(
		mainStoryMissingLocationArchetypePromptTemplate,
		quotedOrNone(job.ThemePrompt),
		quotedOrNone(job.DistrictFit),
		quotedOrNone(job.Tone),
		joinLocationArchetypeAvoidanceNames(existingNames, 250),
		strings.Join(allowedNames, ", "),
		strings.Join(missing, "; "),
	)
}

func (p *GenerateMainStorySuggestionsProcessor) ensureGeneratedMainStoryLocationArchetypes(
	ctx context.Context,
	job *models.MainStorySuggestionJob,
	existing []*models.LocationArchetype,
	drafts []mainStorySuggestionDraftPayload,
) ([]*models.LocationArchetype, error) {
	locationIndex := buildLocationArchetypeIndex(existing)
	missingNames := collectMissingMainStoryLocationArchetypeNames(locationIndex, drafts)
	if len(missingNames) == 0 {
		return existing, nil
	}
	if p.deepPriestClient == nil {
		return existing, nil
	}

	prompt := buildMissingMainStoryLocationArchetypePrompt(job, existing, missingNames)
	answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return existing, err
	}

	var generated generatedLocationArchetypesResponse
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), &generated); err != nil {
		return existing, err
	}

	allowedTypeIndex := buildLocationArchetypePlaceTypeIndex(googlemaps.GetAllPlaceTypes())
	sanitized := sanitizeGeneratedLocationArchetypes(generated.Archetypes, allowedTypeIndex, len(missingNames))
	missingByKey := make(map[string]string, len(missingNames))
	for _, name := range missingNames {
		key := normalizeLocationArchetypeNameKey(name)
		if key == "" {
			continue
		}
		missingByKey[key] = name
	}

	createdAny := false
	for _, spec := range sanitized {
		key := normalizeLocationArchetypeNameKey(spec.Name)
		expectedName, ok := missingByKey[key]
		if !ok {
			continue
		}
		if _, exists := resolveLocationArchetypeByName(expectedName, locationIndex); exists {
			continue
		}
		archetype := &models.LocationArchetype{
			ID:            uuid.New(),
			Name:          expectedName,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			IncludedTypes: spec.IncludedTypes,
			ExcludedTypes: spec.ExcludedTypes,
			Challenges:    models.LocationArchetypeChallenges{},
		}
		if err := p.dbClient.LocationArchetype().Create(ctx, archetype); err != nil {
			return existing, err
		}
		existing = append(existing, archetype)
		locationIndex[strings.ToLower(expectedName)] = locationArchetypeIndexEntry{
			ID:   archetype.ID,
			Name: expectedName,
		}
		createdAny = true
	}

	if !createdAny {
		return existing, nil
	}
	return existing, nil
}

func (p *GenerateMainStorySuggestionsProcessor) enrichMainStoryDraftBeats(
	job *models.MainStorySuggestionJob,
	drafts []mainStorySuggestionDraftPayload,
	locationArchetypes []*models.LocationArchetype,
	monsterTemplates []models.MonsterTemplate,
) []mainStorySuggestionDraftPayload {
	if len(drafts) == 0 {
		return drafts
	}

	allowedLocationPrompt := buildAllowedLocationArchetypesPrompt(locationArchetypes)
	allowedMonsterPrompt := buildAllowedMonsterTemplatesPrompt(monsterTemplates)

	enriched := make([]mainStorySuggestionDraftPayload, 0, len(drafts))
	for _, draft := range drafts {
		nextDraft := draft
		nextDraft.Beats = make([]mainStorySuggestionBeatPayload, 0, len(draft.Beats))
		for _, beat := range draft.Beats {
			expanded := beat
			if mainStoryBeatNeedsExpansion(beat) {
				if response, err := p.requestMainStoryBeatExpansion(job, draft, beat, allowedLocationPrompt, allowedMonsterPrompt); err == nil && response != nil {
					expanded = mergeMainStoryBeatExpansion(expanded, *response)
				} else if err != nil {
					log.Printf(
						"main story beat expansion fallback for draft=%q beat=%q: %v",
						strings.TrimSpace(draft.Name),
						strings.TrimSpace(beat.ChapterTitle),
						err,
					)
				}
			}
			expanded = applyMainStoryBeatFallbackArc(expanded)
			nextDraft.Beats = append(nextDraft.Beats, expanded)
		}
		enriched = append(enriched, nextDraft)
	}
	return enriched
}

func mainStoryBeatNeedsExpansion(beat mainStorySuggestionBeatPayload) bool {
	if len(beat.Steps) < 2 {
		return true
	}
	contentCounts := map[string]int{}
	openEndedScenarioCount := 0
	for _, step := range beat.Steps {
		content := normalizeSuggestionContent(step.Content)
		contentCounts[content]++
		if content == "scenario" && step.ScenarioOpenEnded && len(normalizeSuggestionLines(step.ScenarioBeats)) <= 1 {
			openEndedScenarioCount++
		}
	}
	if len(contentCounts) == 1 && openEndedScenarioCount == len(beat.Steps) {
		return true
	}
	return false
}

func (p *GenerateMainStorySuggestionsProcessor) requestMainStoryBeatExpansion(
	job *models.MainStorySuggestionJob,
	draft mainStorySuggestionDraftPayload,
	beat mainStorySuggestionBeatPayload,
	allowedLocationPrompt string,
	allowedMonsterPrompt string,
) (*mainStoryBeatExpansionResponse, error) {
	if p.deepPriestClient == nil {
		return nil, fmt.Errorf("deep priest client unavailable")
	}
	currentStepsJSON, err := json.Marshal(beat.Steps)
	if err != nil {
		return nil, err
	}
	prompt := fmt.Sprintf(
		mainStoryBeatExpansionPromptTemplate,
		quotedOrNone(draft.Premise),
		quotedOrNone(draft.DistrictFit),
		quotedOrNone(draft.Tone),
		renderTagList(draft.ThemeTags),
		renderTagList(draft.CharacterKeys),
		renderTagList(draft.RevealKeys),
		maxInt(1, beat.OrderIndex),
		maxInt(1, beat.Act),
		quotedOrNone(beat.StoryRole),
		quotedOrNone(beat.ChapterTitle),
		quotedOrNone(beat.ChapterSummary),
		quotedOrNone(beat.Purpose),
		quotedOrNone(beat.WhatChanges),
		quotedOrNone(beat.Hook),
		quotedOrNone(beat.Description),
		renderTagList(beat.PreferredContentMix),
		renderTagList(beat.CharacterTags),
		renderTagList(beat.RequiredZoneTags),
		renderTagList(beat.RequiredLocationArchetypeNames),
		renderTagList(beat.ChallengeTemplateSeeds),
		renderTagList(beat.ScenarioTemplateSeeds),
		renderTagList(beat.MonsterTemplateSeeds),
		string(currentStepsJSON),
		allowedLocationPrompt,
		allowedMonsterPrompt,
	)

	answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return nil, err
	}

	var response mainStoryBeatExpansionResponse
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func mergeMainStoryBeatExpansion(
	beat mainStorySuggestionBeatPayload,
	expansion mainStoryBeatExpansionResponse,
) mainStorySuggestionBeatPayload {
	if summary := strings.TrimSpace(expansion.ChapterSummary); summary != "" {
		beat.ChapterSummary = summary
	}
	if hook := strings.TrimSpace(expansion.Hook); hook != "" {
		beat.Hook = hook
	}
	if description := strings.TrimSpace(expansion.Description); description != "" {
		beat.Description = description
	}
	if len(expansion.ChallengeTemplateSeeds) > 0 {
		beat.ChallengeTemplateSeeds = expansion.ChallengeTemplateSeeds
	}
	if len(expansion.ScenarioTemplateSeeds) > 0 {
		beat.ScenarioTemplateSeeds = expansion.ScenarioTemplateSeeds
	}
	if len(expansion.MonsterTemplateSeeds) > 0 {
		beat.MonsterTemplateSeeds = expansion.MonsterTemplateSeeds
	}
	if len(expansion.Steps) > 0 {
		beat.Steps = expansion.Steps
	}
	return beat
}

func applyMainStoryBeatFallbackArc(
	beat mainStorySuggestionBeatPayload,
) mainStorySuggestionBeatPayload {
	if len(beat.Steps) >= 2 {
		return beat
	}

	requiredLocations := normalizeSuggestionLines(beat.RequiredLocationArchetypeNames)
	primaryLocation := "street-level site"
	if len(requiredLocations) > 0 {
		primaryLocation = strings.TrimSpace(requiredLocations[0])
	}
	secondaryLocation := primaryLocation
	if len(requiredLocations) > 1 {
		secondaryLocation = strings.TrimSpace(requiredLocations[1])
	}

	preferredMix := normalizeSuggestionTags(beat.PreferredContentMix)
	firstContent := pickFallbackBeatContent(preferredMix, []string{"challenge", "scenario", "monster"}, "")
	secondContent := pickFallbackBeatContent(preferredMix, []string{"scenario", "challenge", "monster"}, firstContent)
	if secondContent == "" {
		secondContent = "scenario"
	}
	thirdContent := ""
	if shouldAddThirdFallbackBeatStep(beat, preferredMix, firstContent, secondContent) {
		thirdContent = pickFallbackBeatContent(preferredMix, []string{"monster", "scenario", "challenge"}, secondContent, firstContent)
	}

	steps := make([]questArchetypeSuggestionStepPayload, 0, 3)
	steps = append(steps, buildFallbackBeatStep(beat, firstContent, primaryLocation, false, 0))
	steps = append(steps, buildFallbackBeatStep(beat, secondContent, secondaryLocation, false, 0))
	if thirdContent != "" {
		steps = append(steps, buildFallbackBeatStep(beat, thirdContent, secondaryLocation, thirdContent == "monster", 120))
	}

	beat.Steps = steps
	if strings.TrimSpace(beat.Description) == "" {
		beat.Description = strings.TrimSpace(strings.Join([]string{
			beat.ChapterSummary,
			beat.Purpose,
			beat.WhatChanges,
		}, " "))
	}
	if len(beat.ChallengeTemplateSeeds) == 0 {
		beat.ChallengeTemplateSeeds = normalizeSuggestionLines([]string{
			strings.TrimSpace(beat.ChapterTitle),
			strings.TrimSpace(beat.WhatChanges),
		})
	}
	if len(beat.ScenarioTemplateSeeds) == 0 {
		beat.ScenarioTemplateSeeds = normalizeSuggestionLines([]string{
			strings.TrimSpace(beat.Hook),
			strings.TrimSpace(beat.ChapterSummary),
		})
	}
	if len(beat.MonsterTemplateSeeds) == 0 {
		beat.MonsterTemplateSeeds = normalizeSuggestionLines([]string{
			strings.TrimSpace(beat.ChapterTitle),
			strings.TrimSpace(beat.WhatChanges),
			strings.TrimSpace(beat.Description),
		})
	}
	return beat
}

func pickFallbackBeatContent(preferred []string, defaults []string, excluded ...string) string {
	excludedSet := map[string]struct{}{}
	for _, item := range excluded {
		normalized := normalizeSuggestionContent(item)
		if normalized != "" {
			excludedSet[normalized] = struct{}{}
		}
	}
	for _, item := range preferred {
		normalized := normalizeSuggestionContent(item)
		if normalized == "" {
			continue
		}
		if _, blocked := excludedSet[normalized]; blocked {
			continue
		}
		return normalized
	}
	for _, item := range defaults {
		normalized := normalizeSuggestionContent(item)
		if normalized == "" {
			continue
		}
		if _, blocked := excludedSet[normalized]; blocked {
			continue
		}
		return normalized
	}
	return ""
}

func shouldAddThirdFallbackBeatStep(
	beat mainStorySuggestionBeatPayload,
	preferred []string,
	firstContent string,
	secondContent string,
) bool {
	role := strings.ToLower(strings.TrimSpace(beat.StoryRole))
	if strings.Contains(role, "climax") || strings.Contains(role, "reveal") || strings.Contains(role, "betrayal") {
		return true
	}
	for _, item := range preferred {
		if normalizeSuggestionContent(item) == "monster" && firstContent != "monster" && secondContent != "monster" {
			return true
		}
	}
	return false
}

func buildFallbackBeatStep(
	beat mainStorySuggestionBeatPayload,
	content string,
	locationArchetypeName string,
	useProximity bool,
	distanceMeters int,
) questArchetypeSuggestionStepPayload {
	locationConcept := collapseWhitespace(strings.TrimSpace(beat.ChapterTitle))
	if locationConcept == "" {
		locationConcept = "district site"
	}
	templateConcept := collapseWhitespace(strings.TrimSpace(beat.WhatChanges))
	if templateConcept == "" {
		templateConcept = collapseWhitespace(strings.TrimSpace(beat.ChapterSummary))
	}
	if templateConcept == "" {
		templateConcept = "follow the next lead in the story"
	}
	locationMetadata := normalizeSuggestionTags(append([]string{}, beat.RequiredZoneTags...))
	if len(locationMetadata) == 0 {
		locationMetadata = models.StringArray{"street_level"}
	}
	step := questArchetypeSuggestionStepPayload{
		Source:                "location",
		Content:               content,
		LocationConcept:       locationConcept,
		LocationArchetypeName: locationArchetypeName,
		LocationMetadataTags:  []string(locationMetadata),
		TemplateConcept:       templateConcept,
		PotentialContent: []string{
			strings.TrimSpace(beat.ChapterSummary),
			strings.TrimSpace(beat.WhatChanges),
		},
		ScenarioOpenEnded: true,
		EncounterTone:     []string(normalizeSuggestionTags(append([]string{}, beat.InternalTags...))),
	}
	if useProximity {
		step.Source = "proximity"
		if distanceMeters > 0 {
			distance := distanceMeters
			step.DistanceMeters = &distance
		}
		step.LocationArchetypeName = ""
	}

	switch content {
	case "challenge":
		step.ChallengeQuestion = fmt.Sprintf(
			"Capture or describe a concrete detail at the %s that advances %s.",
			locationConcept,
			lowercaseFirst(strings.TrimSpace(beat.ChapterTitle)),
		)
		step.ChallengeDescription = fmt.Sprintf(
			"Use the place itself to gather proof, spot a telling detail, or complete a small real-world task that pushes this lead forward.",
		)
		step.ChallengeSubmissionType = "text"
		step.ChallengeStatTags = []string{"wisdom", "intelligence"}
	case "monster":
		step.MonsterTemplateNames = normalizeSuggestionLines(append([]string{}, beat.MonsterTemplateSeeds...))
		if len(step.MonsterTemplateNames) == 0 {
			step.MonsterTemplateNames = []string{strings.TrimSpace(beat.ChapterTitle)}
		}
		if len(step.EncounterTone) == 0 {
			step.EncounterTone = []string{"urban", "escalating"}
		}
	default:
		step.Content = "scenario"
		step.ScenarioPrompt = fmt.Sprintf(
			"%s At the %s, the situation sharpens around %s. What do you do?",
			ensureSentenceWithTerminal(strings.TrimSpace(beat.ChapterSummary)),
			locationConcept,
			lowercaseFirst(strings.TrimSpace(beat.WhatChanges)),
		)
		step.ScenarioBeats = normalizeSuggestionLines([]string{
			strings.TrimSpace(beat.Purpose),
			strings.TrimSpace(beat.WhatChanges),
		})
	}

	return step
}

func ensureSentenceWithTerminal(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	last := trimmed[len(trimmed)-1]
	if last == '.' || last == '!' || last == '?' {
		return trimmed
	}
	return trimmed + "."
}

func lowercaseFirst(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "the situation"
	}
	if len(trimmed) == 1 {
		return strings.ToLower(trimmed)
	}
	return strings.ToLower(trimmed[:1]) + trimmed[1:]
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

func buildGeneratedMainStoryDraftAvoidance(
	drafts []mainStorySuggestionDraftPayload,
	limit int,
) string {
	if len(drafts) == 0 || limit <= 0 {
		return ""
	}
	lines := make([]string, 0, minInt(len(drafts), limit))
	for index, draft := range drafts {
		if index >= limit {
			break
		}
		name := collapseWhitespace(draft.Name)
		if name == "" {
			name = "Unnamed generated draft"
		}
		premise := collapseWhitespace(draft.Premise)
		if len(premise) > 120 {
			premise = strings.TrimSpace(premise[:120]) + "..."
		}
		if premise != "" {
			lines = append(lines, fmt.Sprintf("- %s: %s", name, premise))
			continue
		}
		lines = append(lines, fmt.Sprintf("- %s", name))
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
	draftID := uuid.New()
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
	flagPrefix := buildMainStoryFlagPrefix(payload.Name, draftID)
	beats = applyMainStoryBeatAutoFlags(beats, flagPrefix)

	return &models.MainStorySuggestionDraft{
		ID:                draftID,
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
		OrderIndex:                    orderIndex,
		Act:                           act,
		StoryRole:                     strings.TrimSpace(strings.ToLower(payload.StoryRole)),
		ChapterTitle:                  strings.TrimSpace(payload.ChapterTitle),
		ChapterSummary:                strings.TrimSpace(payload.ChapterSummary),
		Purpose:                       strings.TrimSpace(payload.Purpose),
		WhatChanges:                   strings.TrimSpace(payload.WhatChanges),
		IntroducedCharacterKeys:       normalizeSuggestionTags(payload.IntroducedCharacterKeys),
		RequiredCharacterKeys:         normalizeSuggestionTags(payload.RequiredCharacterKeys),
		IntroducedRevealKeys:          normalizeSuggestionTags(payload.IntroducedRevealKeys),
		RequiredRevealKeys:            normalizeSuggestionTags(payload.RequiredRevealKeys),
		RequiredZoneTags:              normalizeSuggestionTags(payload.RequiredZoneTags),
		RequiredLocationArchetypeIDs:  requiredLocationIDs,
		PreferredContentMix:           normalizeSuggestionTags(payload.PreferredContentMix),
		QuestGiverCharacterKey:        sanitizeMainStoryQuestGiverCharacterKey(payload),
		Name:                          name,
		Hook:                          strings.TrimSpace(payload.Hook),
		Description:                   strings.TrimSpace(payload.Description),
		AcceptanceDialogue:            normalizeSuggestionLines(payload.AcceptanceDialogue),
		RequiredStoryFlags:            normalizeSuggestionTags(payload.RequiredStoryFlags),
		SetStoryFlags:                 normalizeSuggestionTags(payload.SetStoryFlags),
		ClearStoryFlags:               normalizeSuggestionTags(payload.ClearStoryFlags),
		QuestGiverRelationshipEffects: normalizeCharacterRelationshipState(payload.QuestGiverRelationshipEffects),
		WorldChanges:                  sanitizeMainStoryWorldChanges(payload.WorldChanges),
		UnlockedScenarios:             sanitizeMainStoryUnlockedScenarios(payload.UnlockedScenarios),
		UnlockedChallenges:            sanitizeMainStoryUnlockedChallenges(payload.UnlockedChallenges),
		UnlockedMonsterEncounters:     sanitizeMainStoryUnlockedEncounters(payload.UnlockedMonsterEncounters),
		QuestGiverAfterDescription:    strings.TrimSpace(payload.QuestGiverAfterDescription),
		QuestGiverAfterDialogue:       normalizeSuggestionLines(payload.QuestGiverAfterDialogue),
		CharacterTags:                 normalizeSuggestionTags(payload.CharacterTags),
		InternalTags:                  normalizeSuggestionTags(payload.InternalTags),
		DifficultyMode:                difficultyMode,
		Difficulty:                    difficulty,
		MonsterEncounterTargetLevel:   monsterLevel,
		WhyThisScales:                 strings.TrimSpace(payload.WhyThisScales),
		Steps:                         steps,
		ChallengeTemplateSeeds:        normalizeSuggestionLines(payload.ChallengeTemplateSeeds),
		ScenarioTemplateSeeds:         normalizeSuggestionLines(payload.ScenarioTemplateSeeds),
		MonsterTemplateSeeds:          normalizeSuggestionLines(payload.MonsterTemplateSeeds),
		Warnings:                      normalizeSuggestionLines(warnings),
	}
}

func sanitizeMainStoryWorldChanges(
	payloads []mainStorySuggestionWorldChangePayload,
) []models.MainStoryWorldChange {
	changes := make([]models.MainStoryWorldChange, 0, len(payloads))
	for _, payload := range payloads {
		changeType := models.NormalizeStoryWorldChangeType(
			strings.TrimSpace(strings.ToLower(payload.Type)),
		)
		if changeType == "" {
			continue
		}
		targetKey := strings.TrimSpace(strings.ToLower(payload.TargetKey))
		if targetKey == "" {
			switch changeType {
			case models.StoryWorldChangeTypeMoveCharacter:
				targetKey = "quest_giver"
			case models.StoryWorldChangeTypeShowPOIText:
				targetKey = "quest_giver_poi"
			}
		}
		changes = append(changes, models.MainStoryWorldChange{
			Type:                changeType,
			TargetKey:           targetKey,
			CharacterKey:        strings.TrimSpace(strings.ToLower(payload.CharacterKey)),
			PointOfInterestHint: strings.TrimSpace(payload.PointOfInterestHint),
			DestinationHint:     strings.TrimSpace(payload.DestinationHint),
			ZoneTags:            normalizeSuggestionTags(payload.ZoneTags),
			Description:         strings.TrimSpace(payload.Description),
			Clue:                strings.TrimSpace(payload.Clue),
		})
	}
	return changes
}

func sanitizeMainStoryUnlockedScenarios(
	payloads []mainStorySuggestionScenarioPayload,
) []models.MainStoryUnlockedScenario {
	out := make([]models.MainStoryUnlockedScenario, 0, len(payloads))
	for _, payload := range payloads {
		prompt := strings.TrimSpace(payload.Prompt)
		if prompt == "" {
			continue
		}
		out = append(out, models.MainStoryUnlockedScenario{
			Name:                strings.TrimSpace(payload.Name),
			Prompt:              prompt,
			PointOfInterestHint: strings.TrimSpace(payload.PointOfInterestHint),
			InternalTags:        normalizeSuggestionTags(payload.InternalTags),
			Difficulty:          models.NormalizeQuestDifficulty(payload.Difficulty),
		})
	}
	return out
}

func sanitizeMainStoryUnlockedChallenges(
	payloads []mainStorySuggestionChallengePayload,
) []models.MainStoryUnlockedChallenge {
	out := make([]models.MainStoryUnlockedChallenge, 0, len(payloads))
	for _, payload := range payloads {
		question := strings.TrimSpace(payload.Question)
		description := strings.TrimSpace(payload.Description)
		if question == "" || description == "" {
			continue
		}
		submissionType := models.QuestNodeSubmissionType(strings.TrimSpace(strings.ToLower(payload.SubmissionType)))
		if !submissionType.IsValid() {
			submissionType = models.DefaultQuestNodeSubmissionType()
		}
		var proficiency *string
		if trimmed := strings.TrimSpace(payload.Proficiency); trimmed != "" {
			proficiency = &trimmed
		}
		out = append(out, models.MainStoryUnlockedChallenge{
			Question:            question,
			Description:         description,
			PointOfInterestHint: strings.TrimSpace(payload.PointOfInterestHint),
			SubmissionType:      submissionType,
			Proficiency:         proficiency,
			StatTags:            normalizeSuggestionTags(payload.StatTags),
			Difficulty:          models.NormalizeQuestDifficulty(payload.Difficulty),
		})
	}
	return out
}

func sanitizeMainStoryUnlockedEncounters(
	payloads []mainStorySuggestionEncounterPayload,
) []models.MainStoryUnlockedEncounter {
	out := make([]models.MainStoryUnlockedEncounter, 0, len(payloads))
	for _, payload := range payloads {
		name := strings.TrimSpace(payload.Name)
		description := strings.TrimSpace(payload.Description)
		if name == "" || description == "" {
			continue
		}
		out = append(out, models.MainStoryUnlockedEncounter{
			Name:                 name,
			Description:          description,
			PointOfInterestHint:  strings.TrimSpace(payload.PointOfInterestHint),
			EncounterType:        models.NormalizeMonsterEncounterType(payload.EncounterType),
			MonsterCount:         maxInt(1, minInt(4, payload.MonsterCount)),
			EncounterTone:        normalizeSuggestionTags(payload.EncounterTone),
			MonsterTemplateHints: normalizeSuggestionLines(payload.MonsterTemplateHints),
			TargetLevel:          models.NormalizeMonsterEncounterTargetLevel(payload.TargetLevel),
		})
	}
	return out
}

func buildMainStoryFlagPrefix(name string, draftID uuid.UUID) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = strings.ReplaceAll(slug, "'", "")
	replacer := strings.NewReplacer(
		" ", "_",
		"-", "_",
		"/", "_",
		".", "_",
		",", "_",
		":", "_",
		";", "_",
	)
	slug = replacer.Replace(slug)
	for strings.Contains(slug, "__") {
		slug = strings.ReplaceAll(slug, "__", "_")
	}
	slug = strings.Trim(slug, "_")
	if slug == "" {
		slug = "campaign"
	}
	return fmt.Sprintf("main_story_%s_%s", draftID.String()[:8], slug)
}

func applyMainStoryBeatAutoFlags(
	beats models.MainStoryBeatDrafts,
	flagPrefix string,
) models.MainStoryBeatDrafts {
	if len(beats) == 0 {
		return beats
	}
	sorted := append(models.MainStoryBeatDrafts{}, beats...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].OrderIndex != sorted[j].OrderIndex {
			return sorted[i].OrderIndex < sorted[j].OrderIndex
		}
		return i < j
	})
	var previousCompletionFlag string
	for index := range sorted {
		completionFlag := fmt.Sprintf(
			"%s_beat_%02d_complete",
			flagPrefix,
			maxInt(1, sorted[index].OrderIndex),
		)
		phaseFlag := fmt.Sprintf(
			"%s_phase_%d_reached",
			flagPrefix,
			maxInt(1, sorted[index].Act),
		)
		required := append([]string{}, []string(sorted[index].RequiredStoryFlags)...)
		if previousCompletionFlag != "" {
			required = append(required, previousCompletionFlag)
		}
		sorted[index].RequiredStoryFlags = normalizeSuggestionTags(required)
		sorted[index].SetStoryFlags = normalizeSuggestionTags(
			append(
				append([]string{}, []string(sorted[index].SetStoryFlags)...),
				completionFlag,
				phaseFlag,
			),
		)
		sorted[index].ClearStoryFlags = normalizeSuggestionTags(
			[]string(sorted[index].ClearStoryFlags),
		)
		if len(sorted[index].QuestGiverAfterDialogue) == 0 {
			fallbackLine := strings.TrimSpace(sorted[index].WhatChanges)
			if fallbackLine == "" {
				fallbackLine = strings.TrimSpace(sorted[index].ChapterSummary)
			}
			if fallbackLine != "" {
				sorted[index].QuestGiverAfterDialogue = models.StringArray{fallbackLine}
			}
		}
		previousCompletionFlag = completionFlag
	}
	return sorted
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
