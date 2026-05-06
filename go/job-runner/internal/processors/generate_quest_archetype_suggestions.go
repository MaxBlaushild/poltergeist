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
You are designing reusable quest archetype bundles for StreetSekai, an urban fantasy MMORPG.

Generate exactly %d quest archetype bundles.

Requested direction:
- theme prompt: %s
- preferred zone kind: %s
- family tags: %s
- requested family mix targets across the batch: %s
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
      "hook": "vivid 1-2 sentence quest teaser",
      "description": "2-4 sentence live situation synopsis with stakes",
      "acceptanceDialogue": ["specific in-world line 1", "specific in-world line 2", "specific in-world line 3"],
      "characterTags": ["tag_a", "tag_b"],
      "internalTags": ["tag_a", "tag_b"],
      "difficultyMode": "scale|fixed",
      "difficulty": 1,
      "monsterEncounterTargetLevel": 1,
      "whyThisScales": "concrete explanation of how the premise escalates or adapts across zones/levels",
      "challengeTemplateSeeds": ["seed one", "seed two"],
      "scenarioTemplateSeeds": ["seed one", "seed two"],
      "monsterTemplateSeeds": ["seed one", "seed two"],
      "nodes": [
        {
          "nodeKey": "unique_lowercase_key",
          "source": "location|proximity",
          "content": "challenge|scenario|monster|exposition",
          "locationConcept": "string",
          "locationArchetypeName": "name from allowed location archetypes or empty string for proximity steps",
          "locationMetadataTags": ["market", "storefront"],
          "distanceMeters": 120,
          "templateConcept": "concrete beat or encounter premise, not a vague label",
          "potentialContent": ["sceneable detail one", "pressure point two", "complication three"],
          "challengeQuestion": "required for challenge steps",
          "challengeDescription": "required for challenge steps",
          "challengeSubmissionType": "photo|text|video",
          "challengeProficiency": "optional short proficiency",
          "challengeStatTags": ["strength", "wisdom"],
          "scenarioPrompt": "required for scenario steps",
          "scenarioOpenEnded": true,
          "scenarioBeats": ["beat one", "beat two"],
          "expositionTitle": "required for exposition steps",
          "expositionDescription": "required for exposition steps",
          "expositionSpeakerName": "required for exposition steps",
          "expositionPortraitUrl": "optional portrait url for exposition steps",
          "expositionDialogue": ["line one", "line two"],
          "monsterTemplateNames": ["names chosen from allowed monster templates"],
          "encounterTone": ["urban", "scrappy"],
          "outcomes": [
            {
              "outcome": "success",
              "nextNodeKey": "later_node_key"
            },
            {
              "outcome": "failure",
              "nextNodeKey": "later_node_key"
            }
          ]
        }
      ]
    }
  ]
}

Rules:
- Output exactly %d drafts.
- Output JSON only. No markdown.
- Keep the tone fantasy and reusable.
- Default to street-level urban fantasy when no preferred zone kind is provided.
- If a preferred zone kind is provided, make the quest premise and node content feel naturally suited to it while remaining reusable.
- Let the zone kind flavor influence hooks, descriptions, scenario prompts, challenge descriptions, metadata tags, and encounter tone where appropriate.
- Write the top-level hook like a quest board teaser or urgent NPC pitch, not a bland summary.
- Write the top-level description like a dungeon master framing the situation in 2-4 sentences with a clear live problem, concrete context, and stakes.
- Write acceptanceDialogue like 2-4 short in-world lines from someone handing the job over, with urgency, specificity, and personality.
- Write whyThisScales as a concrete explanation of how the same premise can escalate, intensify, or stay fresh as the content scales.
- Prefer 3-6 graph nodes with 1-2 meaningful branch points when branching helps.
- At least one third of drafts should be mostly non-combat.
- At least one third of drafts should end in combat.
- Node keys must be unique lowercase snake_case identifiers.
- List nodes in topological order. nextNodeKey may only point to a later node in the same draft.
- Every node may have at most one "success" outcome and at most one "failure" outcome.
- Success/failure branches may reconverge on a later node.
- The root node must be a location-based node, not a proximity node.
- Challenge nodes must use source "location".
- Exposition nodes must use source "location".
- Proximity nodes may only use content "scenario" or "monster".
- Use lowercase snake_case tags for characterTags, internalTags, and locationMetadataTags.
- Every node must include 2-5 locationMetadataTags.
- locationArchetypeName must be selected from the allowed list exactly when source is "location".
- monsterTemplateNames must be selected from the allowed list exactly for monster nodes.
- Every draft must include each required location archetype at least once as a location node when a required list is provided.
- Challenge nodes must be concrete, enjoyable real-world tasks the player can actually complete at the location right now.
- A challenge must be gradable from the player's submission alone.
- Good challenge patterns: photograph a specific detail, spot and record a pattern, identify something visible, compare two visible features, describe ambience or signage actually present on site.
- Never make a challenge depend on fictional missing objects, hidden clues, NPC cooperation, interviewing strangers, asking around, or facts that may not exist at the real location.
- If the content is about how the player would help, investigate, negotiate, persuade, intervene, solve a problem, or respond to a roleplaying situation, that is a scenario step instead of a challenge step.
- challengeQuestion should be an imperative action, not a mystery question.
- Make challengeQuestion and challengeDescription explicit and production-usable.
- Make challengeDescription concrete enough that the player can picture what they are looking for at the real location and why it matters, without relying on fictional hidden evidence.
- When challenge flavor references a fantasy object or clue, bridge it back to observable reality with phrases like "what looks like", "could pass for", "resembles", or "a book/sign/detail of your choice" instead of implying the magical thing literally exists.
- Example: say "Photograph what looks like a hidden sigil within a book of your choice" rather than "Photograph the hidden sigil within the tome."
- Write scenarioPrompt like a dungeon master setting a live scene for players in 2-4 sentences.
- The first 1-3 sentences should paint a concrete, vivid situation with specific sensory or social context, clear actors or forces in motion, and immediate stakes.
- The final sentence should pose the player-facing problem or question they must solve within that scene.
- Avoid thin prompts like "The ritual is underway." or "A confrontation is brewing." without explaining what is happening, where, and why it matters.
- Make scenarioPrompt explicit, vivid, and production-usable.
- Use exposition nodes for lore reveals, witness accounts, ambient warnings, magical residue, discoverable testimony, briefings, and omen-reading beats that deepen the quest without demanding an open-ended player answer.
- expositionTitle should feel like a reusable 2-5 word template title, not a full sentence.
- expositionDescription should read like a dungeon master introducing a vivid discoverable moment in 2-4 sentences, with concrete imagery and why this reveal matters right now.
- expositionSpeakerName should be a reusable 2-4 word voice label like "Witness Echo", "Marginal Warning", or "Shrine Whisper", not a specifically named NPC.
- expositionPortraitUrl may be empty, but when present it must be a direct https URL, not a made-up social profile or unrelated web page.
- expositionDialogue should contain 2-4 short in-world lines that can be delivered as found writing, a witness echo, a magical imprint, or an unnamed local voice.
- Exposition content should stay reusable and should not depend on a specifically named NPC, business, or one-off landmark.
- Use failure branches for fail-forward consequences, setbacks, detours, or escalations instead of dead ends.
- Around half the drafts should include at least one failure branch when it fits naturally.
- When a draft clearly belongs to one primary family such as investigation, delivery, negotiation, pursuit, containment, omen_chasing, ritual_interruption, survival, rescue, or combat_finale, include that exact slug in internalTags.
- When family mix targets are provided, hit those requested family counts across the batch when possible.
- Make the content materially distinct across the batch.
- Vary the core conflict, route texture, and final payoff across the batch instead of submitting palette swaps of the same quest.
- Spread the batch across multiple quest families such as delivery, investigation, negotiation, pursuit, containment, omen-chasing, ritual interruption, survival, or rescue when they fit the requested direction.
- Avoid repeating the same location fantasy, monster fantasy, or final beat more than twice across the batch.
- templateConcept and potentialContent should be concrete, sceneable ideas that help later generation feel vivid instead of generic.
`

const (
	questArchetypeSuggestionDuplicateSimilarityThreshold = 0.72
	questArchetypeSuggestionRecentSimilarityPenaltyStart = 0.48
	questArchetypeSuggestionConsultTimeout               = 90 * time.Second
	questArchetypeSuggestionAvoidancePromptLimit         = 12
	questArchetypeSuggestionLocationPromptLimit          = 40
	questArchetypeSuggestionMonsterPromptLimit           = 72
	questArchetypeSuggestionSpeakerPortraitPlaceholderURL = "https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/character-undiscovered.png"
)

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
	Nodes                       []questArchetypeSuggestionNodePayload `json:"nodes"`
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
	ExpositionTitle         string   `json:"expositionTitle"`
	ExpositionDescription   string   `json:"expositionDescription"`
	ExpositionSpeakerName   string   `json:"expositionSpeakerName"`
	ExpositionPortraitURL   string   `json:"expositionPortraitUrl"`
	ExpositionDialogue      []string `json:"expositionDialogue"`
	MonsterTemplateNames    []string `json:"monsterTemplateNames"`
	EncounterTone           []string `json:"encounterTone"`
}

type questArchetypeSuggestionNodePayload struct {
	NodeKey string `json:"nodeKey"`
	questArchetypeSuggestionStepPayload
	Outcomes []questArchetypeSuggestionOutcomePayload `json:"outcomes"`
}

type questArchetypeSuggestionOutcomePayload struct {
	Outcome     string `json:"outcome"`
	NextNodeKey string `json:"nextNodeKey"`
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
	targetCount := maxInt(1, job.Count)
	candidateCount := questArchetypeSuggestionCandidateCount(targetCount)
	locationArchetypes, err := p.dbClient.LocationArchetype().FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load location archetypes: %w", err)
	}
	zoneKind, err := loadOptionalZoneKind(ctx, p.dbClient, job.ZoneKind)
	if err != nil {
		return fmt.Errorf("failed to load quest archetype suggestion zone kind: %w", err)
	}
	monsterTemplates, err := p.dbClient.MonsterTemplate().FindAllActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to load monster templates: %w", err)
	}
	recentArchetypes, err := p.dbClient.QuestArchetype().FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load quest archetypes: %w", err)
	}
	promptLocationArchetypes := selectQuestArchetypeSuggestionLocationArchetypesForPrompt(
		locationArchetypes,
		job.RequiredLocationArchetypeIDs,
		questArchetypeSuggestionLocationPromptLimit,
	)
	promptMonsterTemplates := selectQuestArchetypeSuggestionMonsterTemplatesForPrompt(
		monsterTemplates,
		job.ZoneKind,
		questArchetypeSuggestionMonsterPromptLimit,
	)

	prompt := fmt.Sprintf(
		questArchetypeSuggestionPromptTemplate,
		candidateCount,
		quotedOrNone(job.ThemePrompt),
		renderQuestArchetypeSuggestionZoneKind(zoneKind),
		renderTagList(job.FamilyTags),
		renderQuestArchetypeSuggestionFamilyMixTargets(job.FamilyMixTargets),
		renderTagList(job.CharacterTags),
		renderTagList(job.InternalTags),
		buildRequiredLocationArchetypesPrompt(job.RequiredLocationArchetypeIDs, locationArchetypes),
		renderTagList(job.RequiredLocationMetadataTags),
		buildQuestArchetypeSuggestionAvoidance(recentArchetypes, questArchetypeSuggestionAvoidancePromptLimit),
		buildAllowedLocationArchetypesPrompt(promptLocationArchetypes),
		buildAllowedMonsterTemplatesPrompt(promptMonsterTemplates),
		candidateCount,
	)
	if zoneKindBlock := zoneKindInstructionBlock(zoneKind); zoneKindBlock != "" {
		prompt = strings.TrimSpace(zoneKindBlock + "\n\n" + prompt)
	}

	answer, err := petitionForQuestArchetypeSuggestions(p.deepPriestClient, prompt)
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
	sanitizedDrafts := make([]*models.QuestArchetypeSuggestionDraft, 0, len(generated.Drafts))
	for _, spec := range generated.Drafts {
		draft := sanitizeQuestArchetypeSuggestionDraft(
			spec,
			job.ZoneKind,
			locationIndex,
			monsterIndex,
			requiredLocationArchetypes,
		)
		sanitizedDrafts = append(sanitizedDrafts, draft)
	}
	selectedDrafts := selectQuestArchetypeSuggestionDrafts(
		sanitizedDrafts,
		recentArchetypes,
		targetCount,
		job.FamilyMixTargets,
	)
	if len(selectedDrafts) == 0 {
		return fmt.Errorf("quest archetype suggestion payload did not include any usable drafts")
	}

	createdCount := 0
	for _, draft := range selectedDrafts {
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

type questArchetypeSuggestionTimedPetitioner interface {
	PetitionTheFountWithTimeout(
		question *deep_priest.Question,
		timeout time.Duration,
	) (*deep_priest.Answer, error)
}

func petitionForQuestArchetypeSuggestions(
	client deep_priest.DeepPriest,
	prompt string,
) (*deep_priest.Answer, error) {
	if timedClient, ok := client.(questArchetypeSuggestionTimedPetitioner); ok {
		return timedClient.PetitionTheFountWithTimeout(
			&deep_priest.Question{Question: prompt},
			questArchetypeSuggestionConsultTimeout,
		)
	}
	return client.PetitionTheFount(&deep_priest.Question{Question: prompt})
}

type questArchetypeSuggestionDraftEvaluation struct {
	draft                        *models.QuestArchetypeSuggestionDraft
	score                        int
	warningCount                 int
	missingRequiredLocationCount int
	placeholderCount             int
	contentTypeCount             int
	nodeCount                    int
	routeKey                     string
	tokenSet                     map[string]struct{}
	mostlyNonCombat              bool
	endsInCombat                 bool
	failureBranchCount           int
	branchPointCount             int
	familySet                    map[string]struct{}
	maxRecentSimilarity          float64
}

func questArchetypeSuggestionCandidateCount(targetCount int) int {
	targetCount = maxInt(1, targetCount)
	if targetCount > 20 {
		return targetCount
	}
	extra := maxInt(4, targetCount/2)
	if extra > 8 {
		extra = 8
	}
	return targetCount + extra
}

func selectQuestArchetypeSuggestionDrafts(
	candidates []*models.QuestArchetypeSuggestionDraft,
	recentArchetypes []*models.QuestArchetype,
	targetCount int,
	familyMixTargets models.QuestArchetypeSuggestionFamilyMixTargets,
) []*models.QuestArchetypeSuggestionDraft {
	targetCount = maxInt(1, targetCount)
	if len(candidates) == 0 {
		return nil
	}

	recentTokenSets := buildRecentQuestArchetypeTokenSets(recentArchetypes, 24)
	evaluations := make([]questArchetypeSuggestionDraftEvaluation, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		evaluations = append(
			evaluations,
			evaluateQuestArchetypeSuggestionDraft(candidate, recentTokenSets),
		)
	}
	if len(evaluations) == 0 {
		return nil
	}

	sort.SliceStable(evaluations, func(left, right int) bool {
		if evaluations[left].score != evaluations[right].score {
			return evaluations[left].score > evaluations[right].score
		}
		if evaluations[left].missingRequiredLocationCount != evaluations[right].missingRequiredLocationCount {
			return evaluations[left].missingRequiredLocationCount < evaluations[right].missingRequiredLocationCount
		}
		if evaluations[left].warningCount != evaluations[right].warningCount {
			return evaluations[left].warningCount < evaluations[right].warningCount
		}
		if evaluations[left].placeholderCount != evaluations[right].placeholderCount {
			return evaluations[left].placeholderCount < evaluations[right].placeholderCount
		}
		if evaluations[left].maxRecentSimilarity != evaluations[right].maxRecentSimilarity {
			return evaluations[left].maxRecentSimilarity < evaluations[right].maxRecentSimilarity
		}
		return strings.ToLower(strings.TrimSpace(evaluations[left].draft.Name)) <
			strings.ToLower(strings.TrimSpace(evaluations[right].draft.Name))
	})

	selected := make([]questArchetypeSuggestionDraftEvaluation, 0, minInt(targetCount, len(evaluations)))
	selectedIndexes := map[int]struct{}{}

	addMatches := func(
		needed int,
		filter func(questArchetypeSuggestionDraftEvaluation) bool,
		enforceDistinct bool,
	) {
		if needed <= 0 {
			return
		}
		for index, evaluation := range evaluations {
			if len(selected) >= targetCount || needed <= 0 {
				return
			}
			if _, exists := selectedIndexes[index]; exists {
				continue
			}
			if filter != nil && !filter(evaluation) {
				continue
			}
			if enforceDistinct && questArchetypeSuggestionTooSimilarToSelection(evaluation, selected) {
				continue
			}
			selectedIndexes[index] = struct{}{}
			selected = append(selected, evaluation)
			needed--
		}
	}

	minimumCategoryCount := 0
	for _, familyTarget := range orderedQuestArchetypeSuggestionFamilyMixTargets(familyMixTargets) {
		addMatches(
			minInt(targetCount-len(selected), familyTarget.count),
			func(evaluation questArchetypeSuggestionDraftEvaluation) bool {
				return evaluation.missingRequiredLocationCount == 0 &&
					questArchetypeSuggestionEvaluationMatchesFamily(evaluation, familyTarget.slug)
			},
			true,
		)
	}
	minimumCategoryCount = 0
	if targetCount >= 3 {
		minimumCategoryCount = maxInt(1, targetCount/3)
	}
	if minimumCategoryCount > 0 {
		addMatches(
			minimumCategoryCount,
			func(evaluation questArchetypeSuggestionDraftEvaluation) bool {
				return evaluation.missingRequiredLocationCount == 0 && evaluation.mostlyNonCombat
			},
			true,
		)
		addMatches(
			minimumCategoryCount-countSelectedQuestArchetypeSuggestionDrafts(selected, func(evaluation questArchetypeSuggestionDraftEvaluation) bool {
				return evaluation.endsInCombat
			}),
			func(evaluation questArchetypeSuggestionDraftEvaluation) bool {
				return evaluation.missingRequiredLocationCount == 0 && evaluation.endsInCombat
			},
			true,
		)
	}

	addMatches(
		targetCount-len(selected),
		func(evaluation questArchetypeSuggestionDraftEvaluation) bool {
			return evaluation.missingRequiredLocationCount == 0
		},
		true,
	)
	addMatches(targetCount-len(selected), nil, true)
	addMatches(targetCount-len(selected), nil, false)

	out := make([]*models.QuestArchetypeSuggestionDraft, 0, len(selected))
	for _, evaluation := range selected {
		out = append(out, evaluation.draft)
	}
	return out
}

func evaluateQuestArchetypeSuggestionDraft(
	draft *models.QuestArchetypeSuggestionDraft,
	recentTokenSets []map[string]struct{},
) questArchetypeSuggestionDraftEvaluation {
	evaluation := questArchetypeSuggestionDraftEvaluation{
		draft:                        draft,
		tokenSet:                     questArchetypeSuggestionDraftTokenSet(draft),
		routeKey:                     questArchetypeSuggestionDraftRouteKey(draft),
		contentTypeCount:             questArchetypeSuggestionDraftContentTypeCount(draft),
		nodeCount:                    questArchetypeSuggestionDraftNodeCount(draft),
		missingRequiredLocationCount: questArchetypeSuggestionMissingRequiredLocationCount(draft),
		placeholderCount:             questArchetypeSuggestionPlaceholderCount(draft),
		failureBranchCount:           questArchetypeSuggestionDraftFailureBranchCount(draft),
		branchPointCount:             questArchetypeSuggestionDraftBranchPointCount(draft),
		familySet:                    questArchetypeSuggestionDraftFamilySet(draft),
	}
	evaluation.warningCount = len(draft.Warnings)
	evaluation.mostlyNonCombat = questArchetypeSuggestionIsMostlyNonCombat(draft)
	evaluation.endsInCombat = questArchetypeSuggestionEndsInCombat(draft)
	evaluation.maxRecentSimilarity = questArchetypeSuggestionMaxSimilarity(
		evaluation.tokenSet,
		recentTokenSets,
	)

	score := 100
	score -= evaluation.warningCount * 8
	score -= evaluation.missingRequiredLocationCount * 18
	score -= evaluation.placeholderCount * 7
	if evaluation.nodeCount == 3 || evaluation.nodeCount == 4 {
		score += 10
	} else if evaluation.nodeCount == 5 {
		score += 4
	} else if evaluation.nodeCount <= 1 {
		score -= 8
	} else if evaluation.nodeCount > 6 {
		score -= (evaluation.nodeCount - 6) * 2
	}
	if strings.TrimSpace(draft.Hook) != "" {
		score += 5
	}
	if strings.TrimSpace(draft.WhyThisScales) != "" {
		score += 4
	}
	score += minInt(3, len(draft.AcceptanceDialogue))
	if evaluation.contentTypeCount >= 2 {
		score += 6
	}
	if evaluation.contentTypeCount >= 3 {
		score += 3
	}
	if evaluation.mostlyNonCombat {
		score += 5
	}
	if evaluation.endsInCombat {
		score += 5
	}
	if evaluation.failureBranchCount == 1 {
		score += 4
	} else if evaluation.failureBranchCount == 2 {
		score += 5
	} else if evaluation.failureBranchCount > 2 {
		score -= (evaluation.failureBranchCount - 2) * 2
	}
	if evaluation.branchPointCount == 1 {
		score += 4
	} else if evaluation.branchPointCount == 2 {
		score += 5
	} else if evaluation.branchPointCount > 2 {
		score -= (evaluation.branchPointCount - 2) * 3
	}
	if evaluation.maxRecentSimilarity > questArchetypeSuggestionRecentSimilarityPenaltyStart {
		score -= int((evaluation.maxRecentSimilarity - questArchetypeSuggestionRecentSimilarityPenaltyStart) * 60.0)
	}
	evaluation.score = score
	return evaluation
}

func countSelectedQuestArchetypeSuggestionDrafts(
	selected []questArchetypeSuggestionDraftEvaluation,
	filter func(questArchetypeSuggestionDraftEvaluation) bool,
) int {
	count := 0
	for _, evaluation := range selected {
		if filter == nil || filter(evaluation) {
			count++
		}
	}
	return count
}

func questArchetypeSuggestionTooSimilarToSelection(
	candidate questArchetypeSuggestionDraftEvaluation,
	selected []questArchetypeSuggestionDraftEvaluation,
) bool {
	for _, existing := range selected {
		if candidate.routeKey != "" && candidate.routeKey == existing.routeKey {
			return true
		}
		if questArchetypeSuggestionTokenSetSimilarity(candidate.tokenSet, existing.tokenSet) >= questArchetypeSuggestionDuplicateSimilarityThreshold {
			return true
		}
	}
	return false
}

type questArchetypeSuggestionFamilyTarget struct {
	slug  string
	count int
}

func orderedQuestArchetypeSuggestionFamilyMixTargets(
	targets models.QuestArchetypeSuggestionFamilyMixTargets,
) []questArchetypeSuggestionFamilyTarget {
	out := make([]questArchetypeSuggestionFamilyTarget, 0, len(targets))
	for _, slug := range models.QuestArchetypeSuggestionKnownFamilySlugs() {
		count := targets[slug]
		if count <= 0 {
			continue
		}
		out = append(out, questArchetypeSuggestionFamilyTarget{
			slug:  slug,
			count: count,
		})
	}
	sort.SliceStable(out, func(left, right int) bool {
		if out[left].count != out[right].count {
			return out[left].count > out[right].count
		}
		return out[left].slug < out[right].slug
	})
	return out
}

func questArchetypeSuggestionEvaluationMatchesFamily(
	evaluation questArchetypeSuggestionDraftEvaluation,
	familySlug string,
) bool {
	familySlug = models.NormalizeQuestArchetypeSuggestionFamilySlug(familySlug)
	if familySlug == "" {
		return false
	}
	if familySlug == "combat_finale" {
		return evaluation.endsInCombat
	}
	_, ok := evaluation.familySet[familySlug]
	return ok
}

func buildRecentQuestArchetypeTokenSets(
	recent []*models.QuestArchetype,
	limit int,
) []map[string]struct{} {
	if len(recent) == 0 {
		return nil
	}
	items := make([]*models.QuestArchetype, 0, len(recent))
	for _, item := range recent {
		if item != nil {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(left, right int) bool {
		return items[left].CreatedAt.After(items[right].CreatedAt)
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	tokenSets := make([]map[string]struct{}, 0, len(items))
	for _, item := range items {
		tokenSet := tokenSet(strings.Join(
			[]string{
				strings.TrimSpace(item.Name),
				strings.TrimSpace(item.Description),
				strings.Join(item.CharacterTags, " "),
				strings.Join(item.InternalTags, " "),
				strings.TrimSpace(item.ZoneKind),
			},
			" ",
		))
		if len(tokenSet) == 0 {
			continue
		}
		tokenSets = append(tokenSets, tokenSet)
	}
	return tokenSets
}

func questArchetypeSuggestionDraftTokenSet(
	draft *models.QuestArchetypeSuggestionDraft,
) map[string]struct{} {
	if draft == nil {
		return map[string]struct{}{}
	}
	nodes := questArchetypeSuggestionDraftNodesForAnalysis(draft)
	nodeIndexByKey := questArchetypeSuggestionNodeIndexByKey(nodes)
	parts := []string{
		strings.TrimSpace(draft.Name),
		strings.TrimSpace(draft.Hook),
		strings.TrimSpace(draft.Description),
		strings.TrimSpace(draft.WhyThisScales),
		strings.TrimSpace(draft.ZoneKind),
		strings.Join(draft.CharacterTags, " "),
		strings.Join(draft.InternalTags, " "),
		strings.Join(draft.ChallengeTemplateSeeds, " "),
		strings.Join(draft.ScenarioTemplateSeeds, " "),
		strings.Join(draft.MonsterTemplateSeeds, " "),
	}
	for index, node := range nodes {
		parts = append(
			parts,
			node.Source,
			node.Content,
			strings.TrimSpace(node.LocationConcept),
			strings.TrimSpace(node.LocationArchetypeName),
			strings.Join(node.LocationMetadataTags, " "),
			strings.TrimSpace(node.TemplateConcept),
			strings.Join(node.PotentialContent, " "),
			strings.TrimSpace(node.ChallengeQuestion),
			strings.TrimSpace(node.ChallengeDescription),
			strings.TrimSpace(node.ScenarioPrompt),
			strings.Join(node.ScenarioBeats, " "),
			strings.TrimSpace(node.ExpositionTitle),
			strings.TrimSpace(node.ExpositionDescription),
			strings.Join(node.ExpositionDialogue, " "),
			strings.Join(node.MonsterTemplateNames, " "),
			strings.Join(node.EncounterTone, " "),
			fmt.Sprintf("node count %d", len(nodes)),
		)
		if len(node.Outcomes) == 0 {
			parts = append(parts, fmt.Sprintf("node %d terminal", index+1))
		}
		for _, outcome := range questArchetypeSuggestionOrderedOutcomes(node.Outcomes) {
			nextIndex, ok := nodeIndexByKey[outcome.NextNodeKey]
			if !ok {
				continue
			}
			parts = append(
				parts,
				fmt.Sprintf("node %d %s branch node %d", index+1, outcome.Outcome, nextIndex+1),
			)
		}
	}
	parts = append(
		parts,
		fmt.Sprintf("failure branch count %d", questArchetypeSuggestionDraftFailureBranchCount(draft)),
		fmt.Sprintf("branch point count %d", questArchetypeSuggestionDraftBranchPointCount(draft)),
	)
	return tokenSet(strings.Join(parts, " "))
}

func questArchetypeSuggestionDraftRouteKey(
	draft *models.QuestArchetypeSuggestionDraft,
) string {
	nodes := questArchetypeSuggestionDraftNodesForAnalysis(draft)
	if len(nodes) == 0 {
		return ""
	}
	nodeIndexByKey := questArchetypeSuggestionNodeIndexByKey(nodes)
	parts := make([]string, 0, len(nodes))
	for index, node := range nodes {
		locationKey := strings.ToLower(strings.TrimSpace(node.LocationArchetypeName))
		if locationKey == "" {
			locationKey = strings.ToLower(strings.TrimSpace(node.LocationConcept))
		}
		outcomeParts := make([]string, 0, len(node.Outcomes))
		for _, outcome := range questArchetypeSuggestionOrderedOutcomes(node.Outcomes) {
			nextIndex, ok := nodeIndexByKey[outcome.NextNodeKey]
			if !ok {
				continue
			}
			outcomeParts = append(
				outcomeParts,
				fmt.Sprintf("%s>%d", strings.TrimSpace(outcome.Outcome), nextIndex+1),
			)
		}
		if len(outcomeParts) == 0 {
			outcomeParts = append(outcomeParts, "end")
		}
		parts = append(
			parts,
			fmt.Sprintf(
				"%d:%s:%s:%s:%s:%s",
				index+1,
				strings.TrimSpace(node.Source),
				strings.TrimSpace(node.Content),
				locationKey,
				strings.ToLower(strings.TrimSpace(node.TemplateConcept)),
				strings.Join(outcomeParts, "|"),
			),
		)
	}
	return strings.Join(parts, " -> ")
}

func questArchetypeSuggestionDraftContentTypeCount(
	draft *models.QuestArchetypeSuggestionDraft,
) int {
	if draft == nil {
		return 0
	}
	seen := map[string]struct{}{}
	for _, node := range questArchetypeSuggestionDraftNodesForAnalysis(draft) {
		content := strings.TrimSpace(node.Content)
		if content == "" {
			continue
		}
		seen[content] = struct{}{}
	}
	return len(seen)
}

func questArchetypeSuggestionMissingRequiredLocationCount(
	draft *models.QuestArchetypeSuggestionDraft,
) int {
	if draft == nil {
		return 0
	}
	count := 0
	for _, warning := range draft.Warnings {
		normalized := strings.ToLower(strings.TrimSpace(warning))
		if strings.Contains(normalized, "required location archetype") &&
			strings.Contains(normalized, "was not used") {
			count++
		}
	}
	return count
}

func questArchetypeSuggestionPlaceholderCount(
	draft *models.QuestArchetypeSuggestionDraft,
) int {
	if draft == nil {
		return 0
	}
	count := 0
	if strings.EqualFold(strings.TrimSpace(draft.Name), "Generated Quest Archetype Draft") {
		count++
	}
	if strings.EqualFold(strings.TrimSpace(draft.Description), "Generated quest archetype draft.") {
		count++
	}
	for _, node := range questArchetypeSuggestionDraftNodesForAnalysis(draft) {
		if strings.EqualFold(strings.TrimSpace(node.ChallengeDescription), "Generated challenge template.") {
			count++
		}
		if strings.EqualFold(strings.TrimSpace(node.ScenarioPrompt), "Generated scenario template.") {
			count++
		}
		if strings.EqualFold(strings.TrimSpace(node.ExpositionDescription), "Generated exposition template.") {
			count++
		}
	}
	return count
}

func questArchetypeSuggestionIsMostlyNonCombat(
	draft *models.QuestArchetypeSuggestionDraft,
) bool {
	nodes := questArchetypeSuggestionDraftNodesForAnalysis(draft)
	if len(nodes) == 0 {
		return false
	}
	monsterCount := 0
	for _, node := range nodes {
		if node.Content == "monster" {
			monsterCount++
		}
	}
	return monsterCount*2 < len(nodes)
}

func questArchetypeSuggestionEndsInCombat(
	draft *models.QuestArchetypeSuggestionDraft,
) bool {
	nodes := questArchetypeSuggestionDraftNodesForAnalysis(draft)
	if len(nodes) == 0 {
		return false
	}
	terminalNodeIndexes := questArchetypeSuggestionTerminalNodeIndexes(nodes)
	for _, index := range terminalNodeIndexes {
		if strings.TrimSpace(nodes[index].Content) == "monster" {
			return true
		}
	}
	return false
}

func questArchetypeSuggestionDraftNodeCount(
	draft *models.QuestArchetypeSuggestionDraft,
) int {
	return len(questArchetypeSuggestionDraftNodesForAnalysis(draft))
}

func questArchetypeSuggestionDraftFailureBranchCount(
	draft *models.QuestArchetypeSuggestionDraft,
) int {
	count := 0
	for _, node := range questArchetypeSuggestionDraftNodesForAnalysis(draft) {
		for _, outcome := range node.Outcomes {
			if strings.TrimSpace(outcome.Outcome) == questArchetypeSuggestionOutcomeFailure {
				count++
			}
		}
	}
	return count
}

func questArchetypeSuggestionDraftBranchPointCount(
	draft *models.QuestArchetypeSuggestionDraft,
) int {
	count := 0
	for _, node := range questArchetypeSuggestionDraftNodesForAnalysis(draft) {
		if len(node.Outcomes) > 1 {
			count++
		}
	}
	return count
}

func questArchetypeSuggestionDraftFamilySet(
	draft *models.QuestArchetypeSuggestionDraft,
) map[string]struct{} {
	out := map[string]struct{}{}
	if draft == nil {
		return out
	}
	for _, rawTag := range draft.InternalTags {
		slug := models.NormalizeQuestArchetypeSuggestionFamilySlug(rawTag)
		if slug == "" || slug == "combat_finale" {
			continue
		}
		out[slug] = struct{}{}
	}
	haystack := strings.ToLower(strings.Join([]string{
		strings.TrimSpace(draft.Name),
		strings.TrimSpace(draft.Hook),
		strings.TrimSpace(draft.Description),
		strings.TrimSpace(draft.WhyThisScales),
		strings.Join(draft.InternalTags, " "),
	}, " "))
	familyKeywords := map[string][]string{
		"investigation":       {"investigate", "clue", "mystery", "discover", "trace", "uncover", "evidence"},
		"delivery":            {"deliver", "courier", "package", "parcel", "shipment", "handoff", "drop"},
		"negotiation":         {"negotiate", "mediate", "broker", "bargain", "convince", "persuade", "truce"},
		"pursuit":             {"pursuit", "chase", "follow", "hunt", "track down", "race"},
		"containment":         {"contain", "seal", "quarantine", "hold back", "stabilize", "containment"},
		"omen_chasing":        {"omen", "portent", "augury", "sign", "prophecy"},
		"ritual_interruption": {"ritual", "rite", "ceremony", "sigil", "summoning", "circle"},
		"survival":            {"survive", "endure", "escape", "hold out", "last until"},
		"rescue":              {"rescue", "save", "extract", "evacuate", "free the", "free a"},
	}
	for family, keywords := range familyKeywords {
		for _, keyword := range keywords {
			if strings.Contains(haystack, keyword) {
				out[family] = struct{}{}
				break
			}
		}
	}
	return out
}

func questArchetypeSuggestionDraftNodesForAnalysis(
	draft *models.QuestArchetypeSuggestionDraft,
) models.QuestArchetypeSuggestionNodes {
	if draft == nil {
		return nil
	}
	if len(draft.Nodes) > 0 {
		nodes := make(models.QuestArchetypeSuggestionNodes, 0, len(draft.Nodes))
		for index, node := range draft.Nodes {
			if strings.TrimSpace(node.NodeKey) == "" {
				node.NodeKey = fmt.Sprintf("node_%d", index+1)
			}
			nodes = append(nodes, node)
		}
		return nodes
	}
	if len(draft.Steps) == 0 {
		return nil
	}
	nodes := make(models.QuestArchetypeSuggestionNodes, 0, len(draft.Steps))
	for index, step := range draft.Steps {
		node := models.QuestArchetypeSuggestionNode{
			NodeKey:               fmt.Sprintf("node_%d", index+1),
			Source:                step.Source,
			Content:               step.Content,
			LocationConcept:       step.LocationConcept,
			LocationArchetypeName: step.LocationArchetypeName,
			LocationArchetypeID:   step.LocationArchetypeID,
			LocationMetadataTags:  append([]string(nil), step.LocationMetadataTags...),
			DistanceMeters:        step.DistanceMeters,
			TemplateConcept:       step.TemplateConcept,
			PotentialContent:      append([]string(nil), step.PotentialContent...),
			ChallengeQuestion:     step.ChallengeQuestion,
			ChallengeDescription:  step.ChallengeDescription,
			ScenarioPrompt:        step.ScenarioPrompt,
				ScenarioOpenEnded:     step.ScenarioOpenEnded,
				ScenarioBeats:         append([]string(nil), step.ScenarioBeats...),
				ExpositionTitle:       step.ExpositionTitle,
				ExpositionDescription: step.ExpositionDescription,
				ExpositionSpeakerName: step.ExpositionSpeakerName,
				ExpositionPortraitURL: step.ExpositionPortraitURL,
				ExpositionDialogue:    append([]string(nil), step.ExpositionDialogue...),
				MonsterTemplateNames:  append([]string(nil), step.MonsterTemplateNames...),
				MonsterTemplateIDs:    append([]string(nil), step.MonsterTemplateIDs...),
			EncounterTone:         append([]string(nil), step.EncounterTone...),
		}
		if index+1 < len(draft.Steps) {
			node.Outcomes = models.QuestArchetypeSuggestionNodeOutcomes{
				{
					Outcome:     questArchetypeSuggestionOutcomeSuccess,
					NextNodeKey: fmt.Sprintf("node_%d", index+2),
				},
			}
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func questArchetypeSuggestionNodeIndexByKey(
	nodes models.QuestArchetypeSuggestionNodes,
) map[string]int {
	out := make(map[string]int, len(nodes))
	for index, node := range nodes {
		out[node.NodeKey] = index
	}
	return out
}

func questArchetypeSuggestionTerminalNodeIndexes(
	nodes models.QuestArchetypeSuggestionNodes,
) []int {
	if len(nodes) == 0 {
		return nil
	}
	terminal := make([]int, 0, len(nodes))
	for index, node := range nodes {
		if len(node.Outcomes) == 0 {
			terminal = append(terminal, index)
		}
	}
	if len(terminal) == 0 {
		return []int{len(nodes) - 1}
	}
	return terminal
}

func questArchetypeSuggestionOrderedOutcomes(
	outcomes models.QuestArchetypeSuggestionNodeOutcomes,
) models.QuestArchetypeSuggestionNodeOutcomes {
	if len(outcomes) <= 1 {
		return outcomes
	}
	ordered := append(models.QuestArchetypeSuggestionNodeOutcomes{}, outcomes...)
	sort.SliceStable(ordered, func(left, right int) bool {
		return ordered[left].Outcome < ordered[right].Outcome
	})
	return ordered
}

func questArchetypeSuggestionMaxSimilarity(
	candidate map[string]struct{},
	recentTokenSets []map[string]struct{},
) float64 {
	maxSimilarity := 0.0
	for _, recent := range recentTokenSets {
		similarity := questArchetypeSuggestionTokenSetSimilarity(candidate, recent)
		if similarity > maxSimilarity {
			maxSimilarity = similarity
		}
	}
	return maxSimilarity
}

func questArchetypeSuggestionTokenSetSimilarity(
	left map[string]struct{},
	right map[string]struct{},
) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	intersection := 0
	for token := range left {
		if _, exists := right[token]; exists {
			intersection++
		}
	}
	union := len(left) + len(right) - intersection
	if union <= 0 {
		return 0
	}
	return float64(intersection) / float64(union)
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

func selectQuestArchetypeSuggestionLocationArchetypesForPrompt(
	items []*models.LocationArchetype,
	requiredIDs []string,
	limit int,
) []*models.LocationArchetype {
	if limit <= 0 || len(items) == 0 {
		return nil
	}
	requiredSet := map[uuid.UUID]struct{}{}
	for _, rawID := range requiredIDs {
		parsedID, err := uuid.Parse(strings.TrimSpace(rawID))
		if err != nil || parsedID == uuid.Nil {
			continue
		}
		requiredSet[parsedID] = struct{}{}
	}

	required := make([]*models.LocationArchetype, 0, len(requiredSet))
	others := make([]*models.LocationArchetype, 0, len(items))
	seen := map[uuid.UUID]struct{}{}
	for _, item := range items {
		if item == nil || item.ID == uuid.Nil {
			continue
		}
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		if _, exists := seen[item.ID]; exists {
			continue
		}
		seen[item.ID] = struct{}{}
		if _, exists := requiredSet[item.ID]; exists {
			required = append(required, item)
			continue
		}
		others = append(others, item)
	}

	sort.Slice(required, func(left, right int) bool {
		return strings.ToLower(strings.TrimSpace(required[left].Name)) <
			strings.ToLower(strings.TrimSpace(required[right].Name))
	})
	sort.Slice(others, func(left, right int) bool {
		return strings.ToLower(strings.TrimSpace(others[left].Name)) <
			strings.ToLower(strings.TrimSpace(others[right].Name))
	})

	selected := make([]*models.LocationArchetype, 0, minInt(limit, len(required)+len(others)))
	selected = append(selected, required...)
	if len(selected) >= limit {
		return selected[:limit]
	}
	remaining := limit - len(selected)
	if remaining > len(others) {
		remaining = len(others)
	}
	selected = append(selected, others[:remaining]...)
	return selected
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

func selectQuestArchetypeSuggestionMonsterTemplatesForPrompt(
	items []models.MonsterTemplate,
	preferredZoneKind string,
	limit int,
) []models.MonsterTemplate {
	if limit <= 0 || len(items) == 0 {
		return nil
	}
	normalizedZoneKind := models.NormalizeZoneKind(preferredZoneKind)
	if normalizedZoneKind == "" {
		sorted := filterAndSortQuestArchetypeSuggestionMonsterTemplates(items)
		if len(sorted) > limit {
			return sorted[:limit]
		}
		return sorted
	}

	zoneMatches := make([]models.MonsterTemplate, 0, len(items))
	generic := make([]models.MonsterTemplate, 0, len(items))
	others := make([]models.MonsterTemplate, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		switch models.NormalizeZoneKind(item.ZoneKind) {
		case normalizedZoneKind:
			zoneMatches = append(zoneMatches, item)
		case "":
			generic = append(generic, item)
		default:
			others = append(others, item)
		}
	}
	zoneMatches = filterAndSortQuestArchetypeSuggestionMonsterTemplates(zoneMatches)
	generic = filterAndSortQuestArchetypeSuggestionMonsterTemplates(generic)
	others = filterAndSortQuestArchetypeSuggestionMonsterTemplates(others)

	selected := make([]models.MonsterTemplate, 0, minInt(limit, len(zoneMatches)+len(generic)+len(others)))
	appendLimitedQuestArchetypeSuggestionMonsterTemplates := func(source []models.MonsterTemplate) {
		if len(selected) >= limit || len(source) == 0 {
			return
		}
		remaining := limit - len(selected)
		if remaining > len(source) {
			remaining = len(source)
		}
		selected = append(selected, source[:remaining]...)
	}
	appendLimitedQuestArchetypeSuggestionMonsterTemplates(zoneMatches)
	appendLimitedQuestArchetypeSuggestionMonsterTemplates(generic)
	appendLimitedQuestArchetypeSuggestionMonsterTemplates(others)
	return selected
}

func filterAndSortQuestArchetypeSuggestionMonsterTemplates(
	items []models.MonsterTemplate,
) []models.MonsterTemplate {
	if len(items) == 0 {
		return nil
	}
	filtered := make([]models.MonsterTemplate, 0, len(items))
	seen := map[uuid.UUID]struct{}{}
	for _, item := range items {
		if item.ID == uuid.Nil || strings.TrimSpace(item.Name) == "" {
			continue
		}
		if _, exists := seen[item.ID]; exists {
			continue
		}
		seen[item.ID] = struct{}{}
		filtered = append(filtered, item)
	}
	sort.Slice(filtered, func(left, right int) bool {
		leftWeight := questArchetypeSuggestionMonsterTemplatePromptWeight(filtered[left])
		rightWeight := questArchetypeSuggestionMonsterTemplatePromptWeight(filtered[right])
		if leftWeight != rightWeight {
			return leftWeight < rightWeight
		}
		return strings.ToLower(strings.TrimSpace(filtered[left].Name)) <
			strings.ToLower(strings.TrimSpace(filtered[right].Name))
	})
	return filtered
}

func questArchetypeSuggestionMonsterTemplatePromptWeight(
	item models.MonsterTemplate,
) int {
	switch item.MonsterType {
	case models.MonsterTemplateTypeBoss:
		return 1
	case models.MonsterTemplateTypeRaid:
		return 2
	default:
		return 0
	}
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

const (
	questArchetypeSuggestionOutcomeSuccess = "success"
	questArchetypeSuggestionOutcomeFailure = "failure"
)

func sanitizeQuestArchetypeSuggestionDraft(
	payload questArchetypeSuggestionDraftPayload,
	zoneKind string,
	locationIndex map[string]locationArchetypeIndexEntry,
	monsterIndex map[string]monsterTemplateIndexEntry,
	requiredLocationArchetypes []locationArchetypeIndexEntry,
) *models.QuestArchetypeSuggestionDraft {
	now := time.Now()
	warnings := models.StringArray{}
	nodes, nodeWarnings := sanitizeQuestArchetypeSuggestionNodes(
		payload,
		locationIndex,
		monsterIndex,
	)
	for _, warning := range nodeWarnings {
		warnings = append(warnings, warning)
	}
	steps := questArchetypeSuggestionNodesAsSteps(nodes)
	if len(steps) == 0 {
		warnings = append(warnings, "no usable quest nodes were generated")
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
	acceptanceDialogue := normalizeSuggestionLines(payload.AcceptanceDialogue)

	draft := &models.QuestArchetypeSuggestionDraft{
		ID:                          uuid.New(),
		CreatedAt:                   now,
		UpdatedAt:                   now,
		Status:                      models.QuestArchetypeSuggestionDraftStatusSuggested,
		Name:                        name,
		Hook:                        hook,
		Description:                 description,
		ZoneKind:                    models.NormalizeZoneKind(zoneKind),
		AcceptanceDialogue:          acceptanceDialogue,
		CharacterTags:               normalizeSuggestionTags(payload.CharacterTags),
		InternalTags:                normalizeSuggestionTags(payload.InternalTags),
		DifficultyMode:              difficultyMode,
		Difficulty:                  difficulty,
		MonsterEncounterTargetLevel: monsterLevel,
		WhyThisScales:               whyThisScales,
		Steps:                       steps,
		Nodes:                       nodes,
		ChallengeTemplateSeeds:      normalizeSuggestionLines(payload.ChallengeTemplateSeeds),
		ScenarioTemplateSeeds:       normalizeSuggestionLines(payload.ScenarioTemplateSeeds),
		MonsterTemplateSeeds:        normalizeSuggestionLines(payload.MonsterTemplateSeeds),
		Warnings:                    normalizeSuggestionLines(warnings),
	}
	enrichQuestArchetypeSuggestionDraftNarrative(draft)
	return draft
}

type sanitizedQuestArchetypeSuggestionNodeInput struct {
	node        models.QuestArchetypeSuggestionNode
	rawOutcomes []questArchetypeSuggestionOutcomePayload
}

func sanitizeQuestArchetypeSuggestionNodes(
	payload questArchetypeSuggestionDraftPayload,
	locationIndex map[string]locationArchetypeIndexEntry,
	monsterIndex map[string]monsterTemplateIndexEntry,
) (models.QuestArchetypeSuggestionNodes, []string) {
	nodePayloads := payload.Nodes
	if len(nodePayloads) == 0 {
		nodePayloads = questArchetypeSuggestionNodePayloadsFromLegacySteps(payload.Steps)
	}
	if len(nodePayloads) == 0 {
		return models.QuestArchetypeSuggestionNodes{}, nil
	}

	inputs := make([]sanitizedQuestArchetypeSuggestionNodeInput, 0, len(nodePayloads))
	warnings := make([]string, 0, len(nodePayloads))
	for index, rawNode := range nodePayloads {
		node, nodeWarnings := sanitizeQuestArchetypeSuggestionNode(
			rawNode,
			index,
			locationIndex,
			monsterIndex,
		)
		for _, warning := range nodeWarnings {
			warnings = append(warnings, fmt.Sprintf("node %d: %s", index+1, warning))
		}
		inputs = append(inputs, sanitizedQuestArchetypeSuggestionNodeInput{
			node:        node,
			rawOutcomes: rawNode.Outcomes,
		})
	}

	keyToIndex := make(map[string]int, len(inputs))
	for index := range inputs {
		key := normalizeSuggestionNodeKey(inputs[index].node.NodeKey, index)
		if previousIndex, exists := keyToIndex[key]; exists {
			rekeyed := fmt.Sprintf("node_%d", index+1)
			warnings = append(
				warnings,
				fmt.Sprintf(
					"node %d key %q duplicated node %d and was rewritten to %q",
					index+1,
					key,
					previousIndex+1,
					rekeyed,
				),
			)
			key = rekeyed
		}
		inputs[index].node.NodeKey = key
		keyToIndex[key] = index
	}

	nodes := make(models.QuestArchetypeSuggestionNodes, 0, len(inputs))
	for _, input := range inputs {
		nodes = append(nodes, input.node)
	}
	for index := range nodes {
		outcomes, outcomeWarnings := sanitizeQuestArchetypeSuggestionNodeOutcomes(
			inputs[index].rawOutcomes,
			index,
			len(nodes),
			keyToIndex,
		)
		for _, warning := range outcomeWarnings {
			warnings = append(warnings, fmt.Sprintf("node %d: %s", index+1, warning))
		}
		nodes[index].Outcomes = outcomes
	}
	incomingCounts := suggestionNodeIncomingEdgeCounts(nodes)
	defaultConnected := make([]bool, len(nodes))
	for index := range nodes {
		if len(nodes[index].Outcomes) != 0 || index+1 >= len(nodes) {
			continue
		}
		if len(inputs[index].rawOutcomes) != 0 {
			continue
		}
		if index > 0 && !defaultConnected[index] {
			continue
		}
		if incomingCounts[index+1] != 0 {
			continue
		}
		nodes[index].Outcomes = append(nodes[index].Outcomes, models.QuestArchetypeSuggestionNodeOutcome{
			Outcome:     questArchetypeSuggestionOutcomeSuccess,
			NextNodeKey: nodes[index+1].NodeKey,
		})
		incomingCounts[index+1]++
		defaultConnected[index+1] = true
		warnings = append(
			warnings,
			fmt.Sprintf("node %d: success branch was missing and defaulted to node %q", index+1, nodes[index+1].NodeKey),
		)
	}

	if len(nodes) > 0 && nodes[0].Source == "proximity" {
		warnings = append(warnings, "node 1: root node should not use proximity")
	}

	reachable := reachableSuggestionNodeIndexes(nodes)
	if len(reachable) > 0 && len(reachable) < len(nodes) {
		filtered := make(models.QuestArchetypeSuggestionNodes, 0, len(reachable))
		for index, node := range nodes {
			if _, ok := reachable[index]; ok {
				filtered = append(filtered, node)
				continue
			}
			warnings = append(
				warnings,
				fmt.Sprintf("node %d: node %q was unreachable from the root and was dropped", index+1, node.NodeKey),
			)
		}
		nodes = filtered
	}

	return nodes, warnings
}

func questArchetypeSuggestionNodePayloadsFromLegacySteps(
	steps []questArchetypeSuggestionStepPayload,
) []questArchetypeSuggestionNodePayload {
	if len(steps) == 0 {
		return nil
	}
	nodes := make([]questArchetypeSuggestionNodePayload, 0, len(steps))
	for index, step := range steps {
		node := questArchetypeSuggestionNodePayload{
			NodeKey:                             fmt.Sprintf("node_%d", index+1),
			questArchetypeSuggestionStepPayload: step,
		}
		if index+1 < len(steps) {
			node.Outcomes = []questArchetypeSuggestionOutcomePayload{
				{
					Outcome:     questArchetypeSuggestionOutcomeSuccess,
					NextNodeKey: fmt.Sprintf("node_%d", index+2),
				},
			}
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func sanitizeQuestArchetypeSuggestionNode(
	payload questArchetypeSuggestionNodePayload,
	index int,
	locationIndex map[string]locationArchetypeIndexEntry,
	monsterIndex map[string]monsterTemplateIndexEntry,
) (models.QuestArchetypeSuggestionNode, []string) {
	step, warnings := sanitizeQuestArchetypeSuggestionStep(
		payload.questArchetypeSuggestionStepPayload,
		locationIndex,
		monsterIndex,
	)
	return models.QuestArchetypeSuggestionNode{
		NodeKey:                 normalizeSuggestionNodeKey(payload.NodeKey, index),
		Source:                  step.Source,
		Content:                 step.Content,
		LocationConcept:         step.LocationConcept,
		LocationArchetypeName:   step.LocationArchetypeName,
		LocationArchetypeID:     step.LocationArchetypeID,
		LocationMetadataTags:    step.LocationMetadataTags,
		DistanceMeters:          step.DistanceMeters,
		TemplateConcept:         step.TemplateConcept,
		PotentialContent:        step.PotentialContent,
		ChallengeQuestion:       step.ChallengeQuestion,
		ChallengeDescription:    step.ChallengeDescription,
		ChallengeSubmissionType: step.ChallengeSubmissionType,
		ChallengeProficiency:    step.ChallengeProficiency,
		ChallengeStatTags:       step.ChallengeStatTags,
			ScenarioPrompt:          step.ScenarioPrompt,
			ScenarioOpenEnded:       step.ScenarioOpenEnded,
			ScenarioBeats:           step.ScenarioBeats,
			ExpositionTitle:         step.ExpositionTitle,
			ExpositionDescription:   step.ExpositionDescription,
			ExpositionSpeakerName:   step.ExpositionSpeakerName,
			ExpositionPortraitURL:   step.ExpositionPortraitURL,
			ExpositionDialogue:      step.ExpositionDialogue,
			MonsterTemplateNames:    step.MonsterTemplateNames,
			MonsterTemplateIDs:      step.MonsterTemplateIDs,
		EncounterTone:           step.EncounterTone,
	}, warnings
}

func sanitizeQuestArchetypeSuggestionNodeOutcomes(
	payloads []questArchetypeSuggestionOutcomePayload,
	nodeIndex int,
	totalNodes int,
	keyToIndex map[string]int,
) (models.QuestArchetypeSuggestionNodeOutcomes, []string) {
	outcomes := make(models.QuestArchetypeSuggestionNodeOutcomes, 0, len(payloads))
	warnings := make([]string, 0, len(payloads))
	seen := map[string]struct{}{}
	for _, payload := range payloads {
		outcome := normalizeSuggestionOutcome(payload.Outcome)
		if outcome == "" {
			warnings = append(warnings, fmt.Sprintf("outcome %q was ignored", strings.TrimSpace(payload.Outcome)))
			continue
		}
		if _, exists := seen[outcome]; exists {
			warnings = append(warnings, fmt.Sprintf("%s branch was duplicated and only the first was kept", outcome))
			continue
		}
		nextNodeKey := normalizeSuggestionNodeKey(strings.TrimSpace(payload.NextNodeKey), -1)
		if nextNodeKey == "" {
			warnings = append(warnings, fmt.Sprintf("%s branch was missing nextNodeKey", outcome))
			continue
		}
		nextIndex, exists := keyToIndex[nextNodeKey]
		if !exists {
			warnings = append(warnings, fmt.Sprintf("%s branch pointed to unknown node %q", outcome, nextNodeKey))
			continue
		}
		if nextIndex <= nodeIndex {
			warnings = append(warnings, fmt.Sprintf("%s branch pointed backward to %q and was dropped", outcome, nextNodeKey))
			continue
		}
		seen[outcome] = struct{}{}
		outcomes = append(outcomes, models.QuestArchetypeSuggestionNodeOutcome{
			Outcome:     outcome,
			NextNodeKey: nextNodeKey,
		})
	}
	if len(outcomes) == 0 && nodeIndex+1 >= totalNodes {
		return outcomes, warnings
	}
	return outcomes, warnings
}

func normalizeSuggestionOutcome(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case questArchetypeSuggestionOutcomeFailure:
		return questArchetypeSuggestionOutcomeFailure
	case questArchetypeSuggestionOutcomeSuccess:
		return questArchetypeSuggestionOutcomeSuccess
	default:
		return ""
	}
}

func normalizeSuggestionNodeKey(raw string, index int) string {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		if index >= 0 {
			return fmt.Sprintf("node_%d", index+1)
		}
		return ""
	}
	var builder strings.Builder
	lastUnderscore := false
	for _, r := range trimmed {
		isAlphaNum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlphaNum {
			builder.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore && builder.Len() > 0 {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}
	normalized := strings.Trim(builder.String(), "_")
	if normalized == "" && index >= 0 {
		return fmt.Sprintf("node_%d", index+1)
	}
	return normalized
}

func reachableSuggestionNodeIndexes(
	nodes models.QuestArchetypeSuggestionNodes,
) map[int]struct{} {
	if len(nodes) == 0 {
		return nil
	}
	keyToIndex := make(map[string]int, len(nodes))
	for index, node := range nodes {
		keyToIndex[node.NodeKey] = index
	}
	reachable := map[int]struct{}{}
	queue := []int{0}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if _, exists := reachable[current]; exists {
			continue
		}
		reachable[current] = struct{}{}
		for _, outcome := range nodes[current].Outcomes {
			nextIndex, ok := keyToIndex[outcome.NextNodeKey]
			if !ok {
				continue
			}
			if _, exists := reachable[nextIndex]; exists {
				continue
			}
			queue = append(queue, nextIndex)
		}
	}
	return reachable
}

func suggestionNodeIncomingEdgeCounts(
	nodes models.QuestArchetypeSuggestionNodes,
) []int {
	if len(nodes) == 0 {
		return nil
	}
	keyToIndex := make(map[string]int, len(nodes))
	for index, node := range nodes {
		keyToIndex[node.NodeKey] = index
	}
	counts := make([]int, len(nodes))
	for _, node := range nodes {
		for _, outcome := range node.Outcomes {
			nextIndex, ok := keyToIndex[outcome.NextNodeKey]
			if !ok {
				continue
			}
			counts[nextIndex]++
		}
	}
	return counts
}

func questArchetypeSuggestionNodesAsSteps(
	nodes models.QuestArchetypeSuggestionNodes,
) models.QuestArchetypeSuggestionSteps {
	steps := make(models.QuestArchetypeSuggestionSteps, 0, len(nodes))
	for _, node := range nodes {
		steps = append(steps, models.QuestArchetypeSuggestionStep{
			Source:                  node.Source,
			Content:                 node.Content,
			LocationConcept:         node.LocationConcept,
			LocationArchetypeName:   node.LocationArchetypeName,
			LocationArchetypeID:     node.LocationArchetypeID,
			LocationMetadataTags:    append([]string(nil), node.LocationMetadataTags...),
			DistanceMeters:          node.DistanceMeters,
			TemplateConcept:         node.TemplateConcept,
			PotentialContent:        append([]string(nil), node.PotentialContent...),
			ChallengeQuestion:       node.ChallengeQuestion,
			ChallengeDescription:    node.ChallengeDescription,
			ChallengeSubmissionType: node.ChallengeSubmissionType,
			ChallengeProficiency:    node.ChallengeProficiency,
			ChallengeStatTags:       append([]string(nil), node.ChallengeStatTags...),
				ScenarioPrompt:          node.ScenarioPrompt,
				ScenarioOpenEnded:       node.ScenarioOpenEnded,
				ScenarioBeats:           append([]string(nil), node.ScenarioBeats...),
				ExpositionTitle:         node.ExpositionTitle,
				ExpositionDescription:   node.ExpositionDescription,
				ExpositionSpeakerName:   node.ExpositionSpeakerName,
				ExpositionPortraitURL:   node.ExpositionPortraitURL,
				ExpositionDialogue:      append([]string(nil), node.ExpositionDialogue...),
				MonsterTemplateNames:    append([]string(nil), node.MonsterTemplateNames...),
				MonsterTemplateIDs:      append([]string(nil), node.MonsterTemplateIDs...),
			EncounterTone:           append([]string(nil), node.EncounterTone...),
		})
	}
	return steps
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
	if content == "exposition" && source != "location" {
		source = "location"
		warnings = append(warnings, "exposition step must use a location source and was coerced to location")
	}

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
		} else if suggestionChallengeNeedsRealityBridge(step.ChallengeQuestion, step.ChallengeDescription) {
			step.ChallengeQuestion, step.ChallengeDescription = bridgeSuggestionChallengeToReality(step)
		}
	case "scenario":
		step.ScenarioPrompt = strings.TrimSpace(payload.ScenarioPrompt)
		if step.ScenarioPrompt == "" {
			step.ScenarioPrompt = strings.TrimSpace(step.TemplateConcept)
			warnings = append(warnings, "scenario prompt was empty and fell back to template concept")
		}
		if suggestionScenarioPromptNeedsExpansion(step.ScenarioPrompt) {
			step.ScenarioPrompt = enrichSuggestionScenarioPrompt(step, "")
		}
	case "exposition":
		step.ExpositionTitle = strings.TrimSpace(payload.ExpositionTitle)
		step.ExpositionDescription = strings.TrimSpace(payload.ExpositionDescription)
		step.ExpositionSpeakerName = strings.TrimSpace(payload.ExpositionSpeakerName)
		step.ExpositionPortraitURL = normalizeSuggestionAbsoluteURL(payload.ExpositionPortraitURL)
		step.ExpositionDialogue = []string(normalizeSuggestionLines(payload.ExpositionDialogue))
		if step.ExpositionTitle == "" {
			step.ExpositionTitle = buildSuggestionExpositionFallbackTitle(step)
			warnings = append(warnings, "exposition title was empty and fell back to a generated title")
		}
		if step.ExpositionSpeakerName == "" {
			step.ExpositionSpeakerName = buildSuggestionExpositionSpeakerName(step)
		}
		if step.ExpositionPortraitURL == "" {
			step.ExpositionPortraitURL = buildSuggestionExpositionPortraitURL(step)
		}
		if suggestionNarrativeTextNeedsExpansion(step.ExpositionDescription, 16, 2, nil) {
			step.ExpositionDescription = buildSuggestionExpositionDescription(step)
			if strings.TrimSpace(payload.ExpositionDescription) == "" {
				warnings = append(warnings, "exposition description was empty")
			}
		}
		if suggestionExpositionDialogueNeedsExpansion(step.ExpositionDialogue) {
			step.ExpositionDialogue = buildSuggestionExpositionDialogue(step)
			if len(payload.ExpositionDialogue) == 0 {
				warnings = append(warnings, "exposition dialogue was empty")
			}
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

var suggestionChallengeRealityBridgePhrases = []string{
	"what looks like",
	"could pass for",
	"that resembles",
	"that feels like",
	"of your choice",
	"actually present on site",
	"real-world",
	"something at the location",
}

var suggestionChallengeFantasyCuePhrases = []string{
	"sigil",
	"rune",
	"glyph",
	"tome",
	"grimoire",
	"spellbook",
	"relic",
	"artifact",
	"ritual circle",
	"altar",
	"wardstone",
	"omen",
	"prophecy",
	"totem",
	"idol",
}

var suggestionScenarioPromptGenericLeadPhrases = []string{
	"the ritual is underway",
	"a ritual is underway",
	"a confrontation is brewing",
	"the confrontation is brewing",
	"tension is rising",
	"a tense situation is unfolding",
	"a complication unfolds",
	"something is wrong",
	"trouble is brewing",
	"an exchange goes wrong",
	"a deal is about to go bad",
}

var suggestionDraftDescriptionGenericLeadPhrases = []string{
	"generated quest archetype draft",
	"a quest about",
	"a live problem unfolds",
	"a tense situation unfolds",
	"something is wrong",
	"trouble is brewing",
}

var suggestionDraftHookGenericLeadPhrases = []string{
	"generated quest archetype draft",
	"help solve",
	"help deal with",
	"something is wrong",
	"trouble is brewing",
	"a tense situation",
}

var suggestionDraftWhyThisScalesGenericLeadPhrases = []string{
	"this scales because",
	"it scales because",
	"it can happen anywhere",
	"the quest can scale",
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

func suggestionChallengeNeedsRealityBridge(question string, description string) bool {
	combined := strings.ToLower(strings.TrimSpace(question + " " + description))
	if combined == "" {
		return false
	}
	for _, phrase := range suggestionChallengeRealityBridgePhrases {
		if strings.Contains(combined, phrase) {
			return false
		}
	}
	for _, phrase := range suggestionChallengeFantasyCuePhrases {
		if strings.Contains(combined, phrase) {
			return true
		}
	}
	return false
}

func bridgeSuggestionChallengeToReality(
	step models.QuestArchetypeSuggestionStep,
) (string, string) {
	target := suggestionChallengeRealityBridgeTarget(step)
	if target == "" {
		return step.ChallengeQuestion, step.ChallengeDescription
	}
	return fmt.Sprintf(
			"%s %s.",
			suggestionChallengeRealityBridgeAction(step.ChallengeQuestion, string(step.ChallengeSubmissionType)),
			target,
		),
		buildSuggestionChallengeRealityBridgeDescription(step, target)
}

func suggestionChallengeRealityBridgeAction(question string, submissionType string) string {
	normalizedQuestion := strings.ToLower(strings.TrimSpace(question))
	switch {
	case strings.HasPrefix(normalizedQuestion, "photograph "):
		return "Photograph"
	case strings.HasPrefix(normalizedQuestion, "capture "):
		return "Capture"
	case strings.HasPrefix(normalizedQuestion, "record "):
		return "Record"
	case strings.HasPrefix(normalizedQuestion, "film "):
		return "Film"
	case strings.HasPrefix(normalizedQuestion, "document "):
		return "Document"
	case strings.HasPrefix(normalizedQuestion, "identify "):
		return "Identify"
	case strings.HasPrefix(normalizedQuestion, "describe "):
		return "Describe"
	}
	switch strings.TrimSpace(strings.ToLower(submissionType)) {
	case "video":
		return "Record"
	case "text":
		return "Describe"
	default:
		return "Photograph"
	}
}

func suggestionChallengeRealityBridgeTarget(
	step models.QuestArchetypeSuggestionStep,
) string {
	base := strings.TrimSpace(step.ChallengeQuestion)
	normalized := strings.TrimSpace(strings.TrimRight(base, ".!?"))
	lower := strings.ToLower(normalized)
	for _, prefix := range []string{
		"photograph ",
		"capture ",
		"record ",
		"film ",
		"document ",
		"identify ",
		"describe ",
	} {
		if strings.HasPrefix(lower, prefix) {
			normalized = strings.TrimSpace(normalized[len(prefix):])
			lower = strings.ToLower(normalized)
			break
		}
	}
	if normalized == "" {
		normalized = strings.TrimSpace(step.TemplateConcept)
		lower = strings.ToLower(normalized)
	}
	if normalized == "" {
		return ""
	}

	replaced := lower
	containerReplacements := []struct {
		old string
		new string
	}{
		{"within the tome", "within a book of your choice"},
		{"in the tome", "in a book of your choice"},
		{"inside the tome", "inside a book of your choice"},
		{"within the grimoire", "within a book or display of your choice"},
		{"in the grimoire", "in a book or display of your choice"},
		{"inside the grimoire", "inside a book or display of your choice"},
	}
	for _, replacement := range containerReplacements {
		replaced = strings.ReplaceAll(replaced, replacement.old, replacement.new)
	}
	replaced = strings.TrimPrefix(replaced, "the ")
	replaced = strings.TrimPrefix(replaced, "a ")
	replaced = strings.TrimPrefix(replaced, "an ")
	switch {
	case strings.Contains(replaced, "hidden sigil"):
		replaced = strings.Replace(replaced, "hidden sigil", "what looks like a hidden sigil", 1)
	case strings.Contains(replaced, "sigil"):
		replaced = strings.Replace(replaced, "sigil", "what looks like a sigil", 1)
	case strings.Contains(replaced, "runes"):
		replaced = strings.Replace(replaced, "runes", "details that look like arcane runes", 1)
	case strings.Contains(replaced, "rune"):
		replaced = strings.Replace(replaced, "rune", "what looks like an arcane rune", 1)
	case strings.Contains(replaced, "glyph"):
		replaced = strings.Replace(replaced, "glyph", "what looks like a glyph", 1)
	case strings.Contains(replaced, "relic"):
		replaced = strings.Replace(replaced, "relic", "an object that feels like a relic", 1)
	case strings.Contains(replaced, "artifact"):
		replaced = strings.Replace(replaced, "artifact", "an object that feels like an artifact", 1)
	case strings.Contains(replaced, "altar"):
		replaced = strings.Replace(replaced, "altar", "a display or arrangement that could pass for an altar", 1)
	case strings.Contains(replaced, "ritual circle"):
		replaced = strings.Replace(replaced, "ritual circle", "an arrangement that could pass for a ritual circle", 1)
	case strings.Contains(replaced, "wardstone"):
		replaced = strings.Replace(replaced, "wardstone", "a protective-looking stone or detail", 1)
	case strings.Contains(replaced, "omen"):
		replaced = strings.Replace(replaced, "omen", "a sign or detail that feels like an omen", 1)
	case strings.Contains(replaced, "tome"):
		replaced = strings.Replace(replaced, "tome", "a book of your choice", 1)
	case strings.Contains(replaced, "grimoire"):
		replaced = strings.Replace(replaced, "grimoire", "a book or display of your choice", 1)
	}

	replaced = strings.TrimSpace(strings.Trim(replaced, ".!?"))
	if replaced == "" {
		return ""
	}
	if !strings.HasPrefix(replaced, "what looks like") &&
		!strings.HasPrefix(replaced, "something at the location that could pass for") &&
		!strings.HasPrefix(replaced, "details that look like") &&
		!strings.HasPrefix(replaced, "an object that feels like") &&
		!strings.HasPrefix(replaced, "a display or arrangement that could pass for") &&
		!strings.HasPrefix(replaced, "a sign or detail that feels like") &&
		!strings.HasPrefix(replaced, "a protective-looking") {
		replaced = "something at the location that could pass for " + strings.TrimPrefix(replaced, "the ")
	}
	return replaced
}

func buildSuggestionChallengeRealityBridgeDescription(
	step models.QuestArchetypeSuggestionStep,
	target string,
) string {
	referent := strings.TrimSpace(strings.Trim(target, ".!?"))
	for _, prefix := range []string{
		"what looks like ",
		"something at the location that could pass for ",
		"details that look like ",
		"an object that feels like ",
		"a display or arrangement that could pass for ",
		"a sign or detail that feels like ",
	} {
		if strings.HasPrefix(referent, prefix) {
			referent = strings.TrimSpace(strings.TrimPrefix(referent, prefix))
			break
		}
	}

	combined := strings.ToLower(strings.Join([]string{
		step.ChallengeQuestion,
		step.ChallengeDescription,
		step.TemplateConcept,
		step.LocationConcept,
		strings.Join(step.PotentialContent, " "),
	}, " "))
	scope := "a real-world detail, object, sign, pattern, or decoration"
	switch {
	case strings.Contains(combined, "book"), strings.Contains(combined, "tome"), strings.Contains(combined, "grimoire"), strings.Contains(combined, "page"):
		scope = "a real-world book, page, shelf display, sign, or decorative detail"
	case strings.Contains(combined, "altar"), strings.Contains(combined, "ritual"), strings.Contains(combined, "shrine"):
		scope = "a real-world arrangement, display, architectural detail, or decorative setup"
	case strings.Contains(combined, "relic"), strings.Contains(combined, "artifact"), strings.Contains(combined, "idol"):
		scope = "a real-world object, display, label, or decorative detail"
	}

	finish := "capture the clearest match you can find"
	switch strings.TrimSpace(strings.ToLower(string(step.ChallengeSubmissionType))) {
	case "video":
		finish = "record the clearest match you can find"
	case "text":
		finish = "describe the clearest match you can find"
	}

	return fmt.Sprintf(
		"Look for %s at the location that could pass for %s. Use something actually present on site, then %s.",
		scope,
		referent,
		finish,
	)
}

func suggestionExpositionDialogueNeedsExpansion(lines []string) bool {
	if len(lines) < 2 {
		return true
	}
	combined := strings.Join(lines, " ")
	return suggestionNarrativeTextNeedsExpansion(combined, 10, 1, nil)
}

func buildSuggestionExpositionFallbackTitle(
	step models.QuestArchetypeSuggestionStep,
) string {
	concept := strings.TrimSpace(step.TemplateConcept)
	if concept != "" {
		words := strings.Fields(strings.ReplaceAll(normalizeSuggestionScenarioFragment(concept), "_", " "))
		if len(words) > 5 {
			words = words[:5]
		}
		if len(words) >= 2 {
			return titleCaseSuggestionWords(words)
		}
	}

	combined := strings.ToLower(strings.Join([]string{
		step.TemplateConcept,
		step.LocationConcept,
		strings.Join(step.LocationMetadataTags, " "),
		strings.Join(step.PotentialContent, " "),
	}, " "))
	switch {
	case strings.Contains(combined, "omen"), strings.Contains(combined, "prophecy"):
		return "Omen Trace"
	case strings.Contains(combined, "ritual"), strings.Contains(combined, "sigil"), strings.Contains(combined, "rune"):
		return "Lingering Ward"
	case strings.Contains(combined, "book"), strings.Contains(combined, "archive"), strings.Contains(combined, "tome"):
		return "Marginal Warning"
	case strings.Contains(combined, "witness"), strings.Contains(combined, "testimony"), strings.Contains(combined, "echo"):
		return "Witness Echo"
	case len(step.LocationMetadataTags) > 0:
		return titleCaseSuggestionWords([]string{humanizeSuggestionTag(step.LocationMetadataTags[0]), "echo"})
	default:
		return "Lingering Echo"
	}
}

func buildSuggestionExpositionSpeakerName(
	step models.QuestArchetypeSuggestionStep,
) string {
	if name := strings.TrimSpace(step.ExpositionSpeakerName); name != "" {
		return name
	}
	if title := strings.TrimSpace(step.ExpositionTitle); title != "" {
		lower := strings.ToLower(title)
		if strings.Contains(lower, "echo") ||
			strings.Contains(lower, "warning") ||
			strings.Contains(lower, "note") ||
			strings.Contains(lower, "voice") ||
			strings.Contains(lower, "whisper") ||
			strings.Contains(lower, "imprint") {
			return title
		}
	}
	combined := strings.ToLower(strings.Join([]string{
		step.ExpositionTitle,
		step.TemplateConcept,
		step.LocationConcept,
		strings.Join(step.LocationMetadataTags, " "),
		strings.Join(step.PotentialContent, " "),
		strings.Join(step.ExpositionDialogue, " "),
	}, " "))
	switch {
	case strings.Contains(combined, "witness"), strings.Contains(combined, "testimony"), strings.Contains(combined, "message"):
		return "Witness Echo"
	case strings.Contains(combined, "book"), strings.Contains(combined, "archive"), strings.Contains(combined, "tome"), strings.Contains(combined, "margin"):
		return "Marginal Warning"
	case strings.Contains(combined, "omen"), strings.Contains(combined, "prophecy"), strings.Contains(combined, "star"), strings.Contains(combined, "celestial"):
		return "Omen Reader"
	case strings.Contains(combined, "ritual"), strings.Contains(combined, "sigil"), strings.Contains(combined, "rune"), strings.Contains(combined, "ward"):
		return "Ward Echo"
	case strings.Contains(combined, "shrine"), strings.Contains(combined, "prayer"), strings.Contains(combined, "altar"), strings.Contains(combined, "bell"):
		return "Shrine Whisper"
	case strings.Contains(combined, "market"), strings.Contains(combined, "street"), strings.Contains(combined, "alley"), strings.Contains(combined, "vendor"):
		return "Street Witness"
	default:
		return "Witness Echo"
	}
}

func buildSuggestionExpositionPortraitURL(
	step models.QuestArchetypeSuggestionStep,
) string {
	if portraitURL := normalizeSuggestionAbsoluteURL(step.ExpositionPortraitURL); portraitURL != "" {
		return portraitURL
	}
	return questArchetypeSuggestionSpeakerPortraitPlaceholderURL
}

func buildSuggestionExpositionDescription(
	step models.QuestArchetypeSuggestionStep,
) string {
	location := strings.TrimSpace(step.LocationConcept)
	if location == "" {
		location = "site"
	}

	fragments := make([]string, 0, 6)
	appendSuggestionScenarioFragment(&fragments, step.ExpositionDescription)
	appendSuggestionScenarioFragment(&fragments, step.TemplateConcept)
	for _, item := range step.PotentialContent {
		appendSuggestionScenarioFragment(&fragments, item)
	}
	for _, line := range step.ExpositionDialogue {
		appendSuggestionScenarioFragment(&fragments, line)
	}
	if len(fragments) == 0 {
		fragments = append(fragments, "a discoverable magical trace is still hanging over the scene")
	}

	opening := buildSuggestionScenarioOpeningSentence(location, fragments[0])
	detail := ""
	for _, fragment := range fragments[1:] {
		candidate := ensureSuggestionSentence(normalizeSuggestionScenarioFragment(fragment))
		if candidate == "" || strings.EqualFold(candidate, opening) {
			continue
		}
		detail = candidate
		break
	}
	if detail == "" {
		detail = buildSuggestionExpositionFallbackDetail(step)
	}

	closing := buildSuggestionExpositionFallbackClosing(step)
	parts := []string{opening}
	if detail != "" && !strings.EqualFold(detail, opening) {
		parts = append(parts, detail)
	}
	if closing != "" && !strings.EqualFold(closing, opening) && !strings.EqualFold(closing, detail) {
		parts = append(parts, closing)
	}
	return strings.Join(parts, " ")
}

func buildSuggestionExpositionDialogue(
	step models.QuestArchetypeSuggestionStep,
) []string {
	lineOne := fmt.Sprintf(
		"The %s still feels wrong if you stop long enough to read it.",
		strings.TrimSpace(step.LocationConcept),
	)
	if strings.TrimSpace(step.LocationConcept) == "" {
		lineOne = "Something here is still trying to tell the story of what went wrong."
	}

	lineTwo := ""
	for _, fragment := range append([]string{step.TemplateConcept}, step.PotentialContent...) {
		candidate := ensureSuggestionSentence(normalizeSuggestionScenarioFragment(fragment))
		if candidate == "" {
			continue
		}
		lineTwo = candidate
		break
	}
	if lineTwo == "" {
		lineTwo = buildSuggestionExpositionFallbackDetail(step)
	}

	lineThree := buildSuggestionExpositionFallbackClosing(step)
	return normalizeSuggestionLines([]string{lineOne, lineTwo, lineThree})
}

func normalizeSuggestionAbsoluteURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "http://") {
		return value
	}
	return ""
}

func buildSuggestionExpositionFallbackDetail(
	step models.QuestArchetypeSuggestionStep,
) string {
	if len(step.LocationMetadataTags) >= 2 {
		return fmt.Sprintf(
			"The %s and %s details around it make the warning feel immediate instead of decorative.",
			humanizeSuggestionTag(step.LocationMetadataTags[0]),
			humanizeSuggestionTag(step.LocationMetadataTags[1]),
		)
	}
	if len(step.LocationMetadataTags) == 1 {
		return fmt.Sprintf(
			"The %s details around it make the warning feel immediate instead of decorative.",
			humanizeSuggestionTag(step.LocationMetadataTags[0]),
		)
	}
	return "The scene reads like a discoverable warning left behind for whoever reaches it next."
}

func buildSuggestionExpositionFallbackClosing(
	step models.QuestArchetypeSuggestionStep,
) string {
	combined := strings.ToLower(strings.Join([]string{
		step.TemplateConcept,
		step.ExpositionTitle,
		strings.Join(step.PotentialContent, " "),
	}, " "))
	switch {
	case strings.Contains(combined, "omen"), strings.Contains(combined, "prophecy"):
		return "Taken together, the signs read like a warning about what is already moving through the district."
	case strings.Contains(combined, "ritual"), strings.Contains(combined, "summon"), strings.Contains(combined, "circle"):
		return "Whatever started here is still pushing outward, and the residue makes that impossible to ignore."
	case strings.Contains(combined, "witness"), strings.Contains(combined, "testimony"), strings.Contains(combined, "message"):
		return "It feels less like a solved clue than a message meant to reach the next person in time."
	default:
		return "It gives the route a clearer sense of what happened here and why the pressure is still rising."
	}
}

func titleCaseSuggestionWords(words []string) string {
	if len(words) == 0 {
		return ""
	}
	parts := make([]string, 0, len(words))
	for _, raw := range words {
		for _, token := range strings.Fields(strings.ReplaceAll(raw, "_", " ")) {
			word := strings.TrimSpace(token)
			if word == "" {
				continue
			}
			runes := []rune(strings.ToLower(word))
			if len(runes) == 0 {
				continue
			}
			runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
			parts = append(parts, string(runes))
		}
	}
	return strings.Join(parts, " ")
}

func enrichQuestArchetypeSuggestionDraftNarrative(
	draft *models.QuestArchetypeSuggestionDraft,
) {
	if draft == nil {
		return
	}
	if suggestionNarrativeTextNeedsExpansion(
		draft.Description,
		18,
		2,
		suggestionDraftDescriptionGenericLeadPhrases,
	) {
		draft.Description = buildQuestArchetypeSuggestionDescription(draft)
	}
	if suggestionNarrativeTextNeedsExpansion(
		draft.Hook,
		10,
		1,
		suggestionDraftHookGenericLeadPhrases,
	) {
		draft.Hook = buildQuestArchetypeSuggestionHook(draft)
	}
	if suggestionAcceptanceDialogueNeedsExpansion(draft.AcceptanceDialogue) {
		draft.AcceptanceDialogue = buildQuestArchetypeSuggestionAcceptanceDialogue(draft)
	}
	if suggestionNarrativeTextNeedsExpansion(
		draft.WhyThisScales,
		12,
		1,
		suggestionDraftWhyThisScalesGenericLeadPhrases,
	) {
		draft.WhyThisScales = buildQuestArchetypeSuggestionWhyThisScales(draft)
	}
}

func suggestionAcceptanceDialogueNeedsExpansion(lines models.StringArray) bool {
	if len(lines) < 2 {
		return true
	}
	return suggestionNarrativeTextNeedsExpansion(
		strings.Join(lines, " "),
		18,
		2,
		nil,
	)
}

func suggestionNarrativeTextNeedsExpansion(
	text string,
	minWords int,
	minSentences int,
	genericLeadPhrases []string,
) bool {
	normalized := suggestionScenarioComparableText(text)
	if normalized == "" {
		return true
	}
	if minWords > 0 && len(strings.Fields(normalized)) < minWords {
		return true
	}
	if minSentences > 0 && suggestionScenarioSentenceCount(text) < minSentences {
		return true
	}
	for _, phrase := range genericLeadPhrases {
		if strings.HasPrefix(normalized, phrase) {
			return true
		}
	}
	return false
}

func buildQuestArchetypeSuggestionDescription(
	draft *models.QuestArchetypeSuggestionDraft,
) string {
	location := questArchetypeSuggestionPrimaryLocationLabel(draft)
	fragments := questArchetypeSuggestionNarrativeFragments(draft)
	opening := ""
	if len(fragments) > 0 {
		if location != "" {
			opening = buildSuggestionScenarioOpeningSentence(location, fragments[0])
		} else {
			opening = ensureSuggestionSentence(fragments[0])
		}
	}
	if opening == "" {
		if location != "" {
			opening = fmt.Sprintf("At the %s, a live problem is already unfolding.", location)
		} else {
			opening = "A live urban-fantasy problem is already unfolding."
		}
	}

	detail := ""
	for _, fragment := range fragments[1:] {
		candidate := ensureSuggestionSentence(normalizeSuggestionScenarioFragment(fragment))
		if candidate == "" || strings.EqualFold(candidate, opening) {
			continue
		}
		detail = candidate
		break
	}
	if detail == "" {
		detail = questArchetypeSuggestionClosingPressureLine(draft)
	}

	closing := questArchetypeSuggestionClosingPressureLine(draft)
	parts := []string{opening}
	if detail != "" && !strings.EqualFold(detail, opening) {
		parts = append(parts, detail)
	}
	if closing != "" && !strings.EqualFold(closing, opening) && !strings.EqualFold(closing, detail) {
		parts = append(parts, closing)
	}
	return strings.Join(parts, " ")
}

func buildQuestArchetypeSuggestionHook(
	draft *models.QuestArchetypeSuggestionDraft,
) string {
	if first := suggestionFirstSentence(draft.Description); first != "" {
		return first
	}
	location := questArchetypeSuggestionPrimaryLocationLabel(draft)
	fragments := questArchetypeSuggestionNarrativeFragments(draft)
	if len(fragments) > 0 {
		if location != "" {
			return buildSuggestionScenarioOpeningSentence(location, fragments[0])
		}
		return ensureSuggestionSentence(fragments[0])
	}
	if location != "" {
		return fmt.Sprintf("At the %s, a live problem is already demanding intervention.", location)
	}
	return "A live urban-fantasy problem is already demanding intervention."
}

func buildQuestArchetypeSuggestionAcceptanceDialogue(
	draft *models.QuestArchetypeSuggestionDraft,
) models.StringArray {
	lineOne := suggestionFirstSentence(draft.Description)
	if lineOne == "" {
		lineOne = buildQuestArchetypeSuggestionHook(draft)
	}

	lineTwo := ""
	fragments := questArchetypeSuggestionNarrativeFragments(draft)
	for _, fragment := range fragments[1:] {
		candidate := ensureSuggestionSentence(normalizeSuggestionScenarioFragment(fragment))
		if candidate == "" || strings.EqualFold(candidate, lineOne) {
			continue
		}
		lineTwo = candidate
		break
	}
	if lineTwo == "" {
		lineTwo = questArchetypeSuggestionClosingPressureLine(draft)
	}

	lineThree := "Step in, read the scene quickly, and keep the whole mess from getting worse."
	switch {
	case questArchetypeSuggestionEndsInCombat(draft):
		lineThree = "Get there fast and keep it from breaking into open violence."
	case questArchetypeSuggestionDraftFailureBranchCount(draft) > 0:
		lineThree = "I need someone who can improvise when the first clean answer falls apart."
	case questArchetypeSuggestionIsMostlyNonCombat(draft):
		lineThree = "I need someone who can read the room before panic or pride makes this worse."
	}

	return normalizeSuggestionLines([]string{lineOne, lineTwo, lineThree})
}

func buildQuestArchetypeSuggestionWhyThisScales(
	draft *models.QuestArchetypeSuggestionDraft,
) string {
	locationScope := "different city landmarks"
	if normalizedZoneKind := models.NormalizeZoneKind(draft.ZoneKind); normalizedZoneKind != "" {
		locationScope = fmt.Sprintf("different %s landmarks", humanizeSuggestionTag(normalizedZoneKind))
	}

	opening := fmt.Sprintf(
		"This premise scales cleanly because the same kind of live crisis can erupt around %s without losing its identity.",
		locationScope,
	)
	detail := "Each step gives the problem more room to spread, so higher-level versions can widen the consequences and pressure without rewriting the quest."
	switch {
	case questArchetypeSuggestionEndsInCombat(draft):
		detail = "As stakes rise, the route can grow from tense investigation or social pressure into open violence without changing the core problem."
	case questArchetypeSuggestionDraftFailureBranchCount(draft) > 0:
		detail = "Its fail-forward branches let setbacks turn into detours and escalations instead of dead ends, so the quest keeps feeling alive at higher pressure."
	case questArchetypeSuggestionIsMostlyNonCombat(draft):
		detail = "As player power rises, the tension can deepen through harder choices, sharper observation, and more dangerous social consequences instead of just bigger numbers."
	}
	return strings.Join([]string{opening, detail}, " ")
}

func questArchetypeSuggestionPrimaryLocationLabel(
	draft *models.QuestArchetypeSuggestionDraft,
) string {
	if draft == nil {
		return ""
	}
	for _, step := range draft.Steps {
		if location := strings.TrimSpace(step.LocationConcept); location != "" {
			return location
		}
		if location := strings.TrimSpace(step.LocationArchetypeName); location != "" {
			return location
		}
	}
	if normalizedZoneKind := models.NormalizeZoneKind(draft.ZoneKind); normalizedZoneKind != "" {
		return fmt.Sprintf("%s district", humanizeSuggestionTag(normalizedZoneKind))
	}
	return ""
}

func questArchetypeSuggestionNarrativeFragments(
	draft *models.QuestArchetypeSuggestionDraft,
) []string {
	if draft == nil {
		return nil
	}
	fragments := make([]string, 0, len(draft.Steps)*3)
	for _, step := range draft.Steps {
		appendSuggestionScenarioFragment(
			&fragments,
			questArchetypeSuggestionNarrativeFragmentForStep(step),
		)
		for _, beat := range step.ScenarioBeats {
			appendSuggestionScenarioFragment(&fragments, beat)
		}
		for _, option := range step.PotentialContent {
			appendSuggestionScenarioFragment(&fragments, option)
		}
		if concept := strings.TrimSpace(step.TemplateConcept); concept != "" {
			appendSuggestionScenarioFragment(&fragments, concept)
		}
	}
	return fragments
}

func questArchetypeSuggestionNarrativeFragmentForStep(
	step models.QuestArchetypeSuggestionStep,
) string {
	switch step.Content {
	case "scenario":
		if body, _ := splitSuggestionScenarioPrompt(step.ScenarioPrompt); strings.TrimSpace(body) != "" {
			return body
		}
		if len(step.ScenarioBeats) > 0 {
			return strings.TrimSpace(step.ScenarioBeats[0])
		}
	case "exposition":
		if description := strings.TrimSpace(step.ExpositionDescription); description != "" {
			return description
		}
		if len(step.ExpositionDialogue) > 0 {
			return strings.TrimSpace(step.ExpositionDialogue[0])
		}
	case "challenge":
		if description := strings.TrimSpace(step.ChallengeDescription); description != "" {
			return description
		}
	case "monster":
		if len(step.PotentialContent) > 0 {
			return strings.TrimSpace(step.PotentialContent[0])
		}
	}
	if concept := strings.TrimSpace(step.TemplateConcept); concept != "" {
		return concept
	}
	if len(step.PotentialContent) > 0 {
		return strings.TrimSpace(step.PotentialContent[0])
	}
	return ""
}

func questArchetypeSuggestionClosingPressureLine(
	draft *models.QuestArchetypeSuggestionDraft,
) string {
	if draft == nil {
		return ""
	}
	switch {
	case questArchetypeSuggestionEndsInCombat(draft):
		return "If nobody gets ahead of it quickly, the whole situation is going to break into open violence."
	case questArchetypeSuggestionDraftFailureBranchCount(draft) > 0:
		return "Every wrong move has room to turn into a harsher detour, so the pressure keeps climbing even when the first plan fails."
	case questArchetypeSuggestionIsMostlyNonCombat(draft):
		return "The real danger is letting hesitation, pride, or confusion give the problem room to spread."
	default:
		return "The route is built to keep the pressure rising until someone steps in and changes the situation."
	}
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
	scenarioStep := step
	scenarioStep.ScenarioPrompt = problem
	if scenarioStep.TemplateConcept == "" {
		scenarioStep.TemplateConcept = strings.TrimSpace(step.ChallengeQuestion)
	}
	if len(scenarioStep.ScenarioBeats) == 0 && len(step.PotentialContent) > 0 {
		scenarioStep.ScenarioBeats = append([]string(nil), step.PotentialContent...)
	}
	prompt := enrichSuggestionScenarioPrompt(scenarioStep, "What do you do?")
	if prompt != "" {
		return prompt
	}
	return fmt.Sprintf("At the %s, a volatile complication is already unfolding. What do you do?", location)
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

func suggestionFirstSentence(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	for index, char := range trimmed {
		if char == '.' || char == '!' || char == '?' {
			return strings.TrimSpace(trimmed[:index+1])
		}
	}
	return ensureSuggestionSentence(trimmed)
}

func ensureSuggestionQuestion(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.TrimRight(trimmed, ".!")
	if strings.HasSuffix(trimmed, "?") {
		return trimmed
	}
	return trimmed + "?"
}

func suggestionScenarioPromptNeedsExpansion(prompt string) bool {
	normalized := suggestionScenarioComparableText(prompt)
	if normalized == "" {
		return true
	}
	if len(strings.Fields(normalized)) < 18 {
		return true
	}
	if suggestionScenarioSentenceCount(normalized) < 2 {
		return true
	}
	for _, phrase := range suggestionScenarioPromptGenericLeadPhrases {
		if strings.HasPrefix(normalized, phrase) {
			return true
		}
	}
	return false
}

func enrichSuggestionScenarioPrompt(
	step models.QuestArchetypeSuggestionStep,
	fallbackQuestion string,
) string {
	body, question := splitSuggestionScenarioPrompt(step.ScenarioPrompt)
	fragments := make([]string, 0, 5)
	appendSuggestionScenarioFragment(&fragments, body)
	for _, beat := range step.ScenarioBeats {
		appendSuggestionScenarioFragment(&fragments, beat)
	}
	for _, option := range step.PotentialContent {
		appendSuggestionScenarioFragment(&fragments, option)
	}
	appendSuggestionScenarioFragment(&fragments, step.TemplateConcept)

	location := strings.TrimSpace(step.LocationConcept)
	if location == "" {
		location = "site"
	}

	if len(fragments) == 0 {
		fragments = append(fragments, "a volatile magical complication is already in motion")
	}

	opening := buildSuggestionScenarioOpeningSentence(location, fragments[0])
	detail := ""
	for _, fragment := range fragments[1:] {
		candidate := ensureSuggestionSentence(normalizeSuggestionScenarioFragment(fragment))
		if candidate == "" {
			continue
		}
		if strings.EqualFold(candidate, opening) {
			continue
		}
		detail = candidate
		break
	}
	if detail == "" {
		detail = buildSuggestionScenarioFallbackDetail(step)
	}

	if question == "" {
		question = fallbackQuestion
	}
	if question == "" {
		question = buildSuggestionScenarioFallbackQuestion(step)
	}
	question = ensureSuggestionQuestion(question)

	parts := []string{}
	if opening != "" {
		parts = append(parts, opening)
	}
	if detail != "" {
		parts = append(parts, detail)
	}
	if question != "" {
		parts = append(parts, question)
	}
	return strings.Join(parts, " ")
}

func splitSuggestionScenarioPrompt(prompt string) (string, string) {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return "", ""
	}
	lastQuestionMark := strings.LastIndex(trimmed, "?")
	if lastQuestionMark < 0 {
		return trimmed, ""
	}
	sentenceStart := 0
	for index := lastQuestionMark - 1; index >= 0; index-- {
		switch trimmed[index] {
		case '.', '!', '?':
			sentenceStart = index + 1
			index = -1
		}
	}
	question := strings.TrimSpace(trimmed[sentenceStart : lastQuestionMark+1])
	body := strings.TrimSpace(trimmed[:sentenceStart])
	if body == "" && question != trimmed {
		body = strings.TrimSpace(strings.Replace(trimmed, question, "", 1))
	}
	return body, question
}

func appendSuggestionScenarioFragment(fragments *[]string, raw string) {
	normalized := normalizeSuggestionScenarioFragment(raw)
	if normalized == "" {
		return
	}
	for _, existing := range *fragments {
		if strings.EqualFold(existing, normalized) {
			return
		}
	}
	*fragments = append(*fragments, normalized)
}

func normalizeSuggestionScenarioFragment(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.Trim(trimmed, " \t\r\n.,;:!?")
	if trimmed == "" {
		return ""
	}
	return strings.Join(strings.Fields(trimmed), " ")
}

func buildSuggestionScenarioOpeningSentence(location string, fragment string) string {
	normalized := normalizeSuggestionScenarioFragment(fragment)
	if normalized == "" {
		return ""
	}
	lowerLead := strings.ToLower(normalized)
	if strings.HasPrefix(lowerLead, "the ") ||
		strings.HasPrefix(lowerLead, "a ") ||
		strings.HasPrefix(lowerLead, "an ") {
		normalized = lowerCaseSuggestionFirstRune(normalized)
	}
	return fmt.Sprintf("At the %s, %s.", location, normalized)
}

func buildSuggestionScenarioFallbackDetail(step models.QuestArchetypeSuggestionStep) string {
	if len(step.LocationMetadataTags) >= 2 {
		return fmt.Sprintf(
			"The %s and %s details around the scene make it clear the problem is already spilling into the space around you.",
			humanizeSuggestionTag(step.LocationMetadataTags[0]),
			humanizeSuggestionTag(step.LocationMetadataTags[1]),
		)
	}
	if len(step.LocationMetadataTags) == 1 {
		return fmt.Sprintf(
			"The %s details around the scene make it clear the problem is already escalating.",
			humanizeSuggestionTag(step.LocationMetadataTags[0]),
		)
	}
	return "The pressure is immediate, and every moment of hesitation gives the situation more room to spread."
}

func buildSuggestionScenarioFallbackQuestion(step models.QuestArchetypeSuggestionStep) string {
	templateConcept := strings.ToLower(strings.TrimSpace(step.TemplateConcept))
	switch {
	case strings.Contains(templateConcept, "stop"), strings.Contains(templateConcept, "interrupt"):
		return "How do you stop it?"
	case strings.Contains(templateConcept, "convince"),
		strings.Contains(templateConcept, "persuade"),
		strings.Contains(templateConcept, "negotiate"),
		strings.Contains(templateConcept, "mediate"):
		return "How do you handle the negotiation?"
	case strings.Contains(templateConcept, "escape"),
		strings.Contains(templateConcept, "survive"),
		strings.Contains(templateConcept, "evacuate"):
		return "How do you get everyone through it?"
	default:
		return "What do you do?"
	}
}

func humanizeSuggestionTag(tag string) string {
	trimmed := strings.TrimSpace(tag)
	if trimmed == "" {
		return "ambient"
	}
	return strings.ReplaceAll(trimmed, "_", " ")
}

func suggestionScenarioComparableText(input string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(input)), " "))
}

func suggestionScenarioSentenceCount(input string) int {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return 0
	}
	count := 0
	open := false
	for _, char := range trimmed {
		if char != ' ' && char != '\n' && char != '\t' && char != '\r' {
			open = true
		}
		if char == '.' || char == '!' || char == '?' {
			if open {
				count++
				open = false
			}
		}
	}
	if open {
		count++
	}
	return count
}

func lowerCaseSuggestionFirstRune(input string) string {
	runes := []rune(input)
	if len(runes) == 0 {
		return ""
	}
	runes[0] = []rune(strings.ToLower(string(runes[0])))[0]
	return string(runes)
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
	case "exposition":
		return "exposition"
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

func renderQuestArchetypeSuggestionZoneKind(zoneKind *models.ZoneKind) string {
	if zoneKind == nil {
		return "none"
	}
	label := strings.TrimSpace(models.ZoneKindPromptLabel(zoneKind))
	slug := strings.TrimSpace(models.ZoneKindPromptSlug(zoneKind))
	seed := strings.TrimSpace(models.ZoneKindPromptSeed(zoneKind))
	parts := make([]string, 0, 3)
	if label != "" {
		parts = append(parts, label)
	}
	if slug != "" && slug != label {
		parts = append(parts, "slug="+slug)
	}
	if seed != "" {
		parts = append(parts, "seed="+seed)
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, " | ")
}

func renderTagList(tags []string) string {
	normalized := normalizeSuggestionTags(tags)
	if len(normalized) == 0 {
		return "none"
	}
	return strings.Join(normalized, ", ")
}

func renderQuestArchetypeSuggestionFamilyMixTargets(
	targets models.QuestArchetypeSuggestionFamilyMixTargets,
) string {
	ordered := orderedQuestArchetypeSuggestionFamilyMixTargets(targets)
	if len(ordered) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(ordered))
	for _, target := range ordered {
		parts = append(parts, fmt.Sprintf("%s x%d", target.slug, target.count))
	}
	return strings.Join(parts, ", ")
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
