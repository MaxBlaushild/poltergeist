package server

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type questTemplateGeneratorSource string
type questTemplateGeneratorContent string

const (
	questTemplateGeneratorSourceLocation   questTemplateGeneratorSource  = "location_archetype"
	questTemplateGeneratorSourceProximity  questTemplateGeneratorSource  = "proximity"
	questTemplateGeneratorContentChallenge questTemplateGeneratorContent = "challenge"
	questTemplateGeneratorContentScenario  questTemplateGeneratorContent = "scenario"
	questTemplateGeneratorContentMonster   questTemplateGeneratorContent = "monster"
)

type questTemplateGeneratorStepRequest struct {
	Source              string     `json:"source"`
	Content             string     `json:"content"`
	LocationArchetypeID *uuid.UUID `json:"locationArchetypeId"`
	ProximityMeters     *int       `json:"proximityMeters"`
}

type questTemplateGeneratorRequest struct {
	Name          string                              `json:"name"`
	ThemePrompt   string                              `json:"themePrompt"`
	CharacterTags []string                            `json:"characterTags"`
	InternalTags  []string                            `json:"internalTags"`
	Steps         []questTemplateGeneratorStepRequest `json:"steps"`
}

type normalizedQuestTemplateGeneratorStep struct {
	Source            questTemplateGeneratorSource
	Content           questTemplateGeneratorContent
	LocationArchetype *models.LocationArchetype
	ProximityMeters   int
}

var questTemplateGeneratorTokenPattern = regexp.MustCompile(`[^a-z0-9]+`)

func normalizeQuestTemplateGeneratorSource(raw string) questTemplateGeneratorSource {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(questTemplateGeneratorSourceProximity):
		return questTemplateGeneratorSourceProximity
	default:
		return questTemplateGeneratorSourceLocation
	}
}

func normalizeQuestTemplateGeneratorContent(raw string) questTemplateGeneratorContent {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(questTemplateGeneratorContentScenario):
		return questTemplateGeneratorContentScenario
	case string(questTemplateGeneratorContentMonster):
		return questTemplateGeneratorContentMonster
	default:
		return questTemplateGeneratorContentChallenge
	}
}

func (s *server) createQuestArchetypeFromGenerator(ctx *gin.Context) {
	var requestBody questTemplateGeneratorRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	created, err := s.generateQuestArchetypeFromSteps(ctx, requestBody)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, created)
}

func (s *server) generateQuestArchetypeFromSteps(
	ctx *gin.Context,
	requestBody questTemplateGeneratorRequest,
) (*models.QuestArchetype, error) {
	steps, err := s.normalizeQuestTemplateGeneratorSteps(ctx, requestBody.Steps)
	if err != nil {
		return nil, err
	}
	if len(steps) == 0 {
		return nil, fmt.Errorf("at least one step is required")
	}

	monsterTemplates, err := s.dbClient.MonsterTemplate().FindAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load monster templates")
	}

	var rootNodeID uuid.UUID
	var previousNode *models.QuestArchetypeNode
	for idx, step := range steps {
		node, err := s.createGeneratedQuestTemplateStepNode(ctx, requestBody.ThemePrompt, idx, step, &monsterTemplates)
		if err != nil {
			return nil, err
		}
		if idx == 0 {
			rootNodeID = node.ID
		}
		if previousNode != nil {
			if err := s.linkGeneratedQuestTemplateNodes(
				ctx,
				previousNode,
				node.ID,
			); err != nil {
				return nil, err
			}
		}
		previousNode = node
	}

	name := buildGeneratedQuestTemplateName(requestBody.Name, requestBody.ThemePrompt, steps)
	description := buildGeneratedQuestTemplateDescription(requestBody.ThemePrompt, steps)
	questArchetype := &models.QuestArchetype{
		ID:                          uuid.New(),
		Name:                        name,
		Description:                 description,
		AcceptanceDialogue:          dialogueSequenceFromLines(buildGeneratedQuestTemplateAcceptanceDialogue(requestBody.ThemePrompt, steps)),
		ImageURL:                    "",
		DifficultyMode:              models.QuestDifficultyModeScale,
		Difficulty:                  1,
		MonsterEncounterTargetLevel: 1,
		DefaultGold:                 0,
		RewardMode:                  models.RewardModeRandom,
		RandomRewardSize:            models.RandomRewardSizeSmall,
		RewardExperience:            0,
		MaterialRewards:             models.BaseMaterialRewards{},
		CharacterTags:               normalizeQuestTemplateCharacterTags(requestBody.CharacterTags),
		InternalTags:                normalizeQuestTemplateInternalTags(requestBody.InternalTags),
		RootID:                      rootNodeID,
		ItemRewards:                 []models.QuestArchetypeItemReward{},
		SpellRewards:                []models.QuestArchetypeSpellReward{},
	}
	if err := s.dbClient.QuestArchetype().Create(ctx, questArchetype); err != nil {
		return nil, fmt.Errorf("failed to create generated quest template")
	}
	created, err := s.dbClient.QuestArchetype().FindByID(ctx, questArchetype.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load generated quest template")
	}
	return created, nil
}

func (s *server) normalizeQuestTemplateGeneratorSteps(
	ctx *gin.Context,
	input []questTemplateGeneratorStepRequest,
) ([]normalizedQuestTemplateGeneratorStep, error) {
	steps := make([]normalizedQuestTemplateGeneratorStep, 0, len(input))
	for index, raw := range input {
		source := normalizeQuestTemplateGeneratorSource(raw.Source)
		content := normalizeQuestTemplateGeneratorContent(raw.Content)
		if source == questTemplateGeneratorSourceProximity && index == 0 {
			return nil, fmt.Errorf("steps[0] cannot use proximity without a previous node")
		}
		if source == questTemplateGeneratorSourceProximity && content == questTemplateGeneratorContentChallenge {
			return nil, fmt.Errorf("steps[%d] challenge steps must use a location archetype", index)
		}
		step := normalizedQuestTemplateGeneratorStep{
			Source:          source,
			Content:         content,
			ProximityMeters: 100,
		}
		if source == questTemplateGeneratorSourceLocation {
			if raw.LocationArchetypeID == nil || *raw.LocationArchetypeID == uuid.Nil {
				return nil, fmt.Errorf("steps[%d].locationArchetypeId is required", index)
			}
			locationArchetype, err := s.dbClient.LocationArchetype().FindByID(ctx, *raw.LocationArchetypeID)
			if err != nil {
				return nil, fmt.Errorf("steps[%d].locationArchetypeId could not be loaded", index)
			}
			if locationArchetype == nil {
				return nil, fmt.Errorf("steps[%d].locationArchetypeId could not be loaded", index)
			}
			step.LocationArchetype = locationArchetype
			step.ProximityMeters = 0
		} else {
			if raw.ProximityMeters != nil {
				step.ProximityMeters = *raw.ProximityMeters
			}
			if step.ProximityMeters < 0 {
				return nil, fmt.Errorf("steps[%d].proximityMeters must be zero or greater", index)
			}
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func (s *server) createGeneratedQuestTemplateStepNode(
	ctx *gin.Context,
	themePrompt string,
	stepIndex int,
	step normalizedQuestTemplateGeneratorStep,
	monsterTemplates *[]models.MonsterTemplate,
) (*models.QuestArchetypeNode, error) {
	payload := questArchetypeNodePayload{
		Difficulty: intPtr(0),
	}
	if step.LocationArchetype != nil {
		locationID := step.LocationArchetype.ID
		payload.LocationArchetypeID = &locationID
	}
	switch step.Content {
	case questTemplateGeneratorContentScenario:
		template, err := s.createGeneratedScenarioTemplate(ctx, themePrompt, stepIndex, step)
		if err != nil {
			return nil, err
		}
		payload.NodeType = string(models.QuestArchetypeNodeTypeScenario)
		payload.ScenarioTemplateID = &template.ID
		payload.EncounterProximityMeters = intPtr(step.ProximityMeters)
	case questTemplateGeneratorContentMonster:
		templateIDs, err := s.ensureQuestMonsterTemplateIDs(
			ctx,
			monsterTemplates,
			questMonsterTemplateRequest{
				Count:             1,
				MonsterType:       models.MonsterTemplateTypeMonster,
				ThemePrompt:       themePrompt,
				EncounterConcept:  buildGeneratedQuestTemplateName("", themePrompt, []normalizedQuestTemplateGeneratorStep{step}),
				LocationConcept:   generatedQuestMonsterLocationConcept(step),
				LocationArchetype: step.LocationArchetype,
			},
			nil,
		)
		if err != nil {
			return nil, err
		}
		payload.NodeType = string(models.QuestArchetypeNodeTypeMonsterEncounter)
		payload.MonsterTemplateIDs = templateIDs
		payload.TargetLevel = intPtr(maxInt(1, 1+stepIndex))
		payload.EncounterProximityMeters = intPtr(step.ProximityMeters)
	default:
		template, err := s.createGeneratedChallengeTemplate(
			ctx,
			themePrompt,
			stepIndex,
			step,
			nil,
		)
		if err != nil {
			return nil, err
		}
		payload.NodeType = string(models.QuestArchetypeNodeTypeChallenge)
		payload.ChallengeTemplateID = &template.ID
		payload.EncounterProximityMeters = intPtr(step.ProximityMeters)
	}
	node := &models.QuestArchetypeNode{
		ID:         uuid.New(),
		NodeType:   models.QuestArchetypeNodeTypeChallenge,
		Difficulty: 0,
	}
	if err := s.applyQuestArchetypeNodePayload(ctx, node, payload, true); err != nil {
		return nil, err
	}
	if err := s.dbClient.QuestArchetypeNode().Create(ctx, node); err != nil {
		return nil, fmt.Errorf("failed to create quest template node")
	}
	return node, nil
}

func (s *server) linkGeneratedQuestTemplateNodes(
	ctx *gin.Context,
	parentNode *models.QuestArchetypeNode,
	childNodeID uuid.UUID,
) error {
	challenge := &models.QuestArchetypeChallenge{
		ID:             uuid.New(),
		Reward:         0,
		Difficulty:     0,
		UnlockedNodeID: &childNodeID,
	}
	if err := s.dbClient.QuestArchetypeChallenge().Create(ctx, challenge); err != nil {
		return fmt.Errorf("failed to create quest template link")
	}
	return s.dbClient.QuestArchetypeNodeChallenge().Create(ctx, &models.QuestArchetypeNodeChallenge{
		ID:                        uuid.New(),
		QuestArchetypeChallengeID: challenge.ID,
		QuestArchetypeNodeID:      parentNode.ID,
	})
}

func (s *server) createGeneratedScenarioTemplate(
	ctx *gin.Context,
	themePrompt string,
	stepIndex int,
	step normalizedQuestTemplateGeneratorStep,
) (*models.ScenarioTemplate, error) {
	prompt := buildGeneratedScenarioTemplatePrompt(themePrompt, stepIndex, step)
	template := &models.ScenarioTemplate{
		Prompt:                    prompt,
		ImageURL:                  "",
		ThumbnailURL:              "",
		ScaleWithUserLevel:        false,
		RewardMode:                models.RewardModeRandom,
		RandomRewardSize:          models.RandomRewardSizeSmall,
		Difficulty:                maxInt(0, 8+(stepIndex*2)),
		RewardExperience:          0,
		RewardGold:                0,
		OpenEnded:                 true,
		FailurePenaltyMode:        models.ScenarioFailurePenaltyModeShared,
		FailureHealthDrainType:    models.ScenarioFailureDrainTypeNone,
		FailureHealthDrainValue:   0,
		FailureManaDrainType:      models.ScenarioFailureDrainTypeNone,
		FailureManaDrainValue:     0,
		FailureStatuses:           models.ScenarioFailureStatusTemplates{},
		SuccessRewardMode:         models.ScenarioSuccessRewardModeShared,
		SuccessHealthRestoreType:  models.ScenarioFailureDrainTypeNone,
		SuccessHealthRestoreValue: 0,
		SuccessManaRestoreType:    models.ScenarioFailureDrainTypeNone,
		SuccessManaRestoreValue:   0,
		SuccessStatuses:           models.ScenarioFailureStatusTemplates{},
		Options:                   models.ScenarioTemplateOptions{},
		ItemRewards:               models.ScenarioTemplateRewards{},
		ItemChoiceRewards:         models.ScenarioTemplateRewards{},
		SpellRewards:              models.ScenarioTemplateSpellRewards{},
	}
	if err := s.dbClient.ScenarioTemplate().Create(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to create generated scenario template")
	}
	return template, nil
}

func buildGeneratedScenarioTemplatePrompt(
	themePrompt string,
	stepIndex int,
	step normalizedQuestTemplateGeneratorStep,
) string {
	theme := strings.TrimSpace(themePrompt)
	if theme == "" {
		theme = "a rising mystery"
	}
	if step.LocationArchetype != nil {
		return fmt.Sprintf(
			"At the %s, a scene tied to %s unfolds and demands a response. Decide how you want to intervene before the trail goes cold.",
			strings.TrimSpace(step.LocationArchetype.Name),
			theme,
		)
	}
	return fmt.Sprintf(
		"A fresh development tied to %s emerges %d steps into the trail. Decide how you want to respond before the situation slips away.",
		theme,
		stepIndex+1,
	)
}

func (s *server) createGeneratedChallengeTemplate(
	ctx *gin.Context,
	themePrompt string,
	stepIndex int,
	step normalizedQuestTemplateGeneratorStep,
	nextStep *normalizedQuestTemplateGeneratorStep,
) (*models.ChallengeTemplate, error) {
	if step.LocationArchetype == nil {
		return nil, fmt.Errorf("generated challenge templates require a location archetype")
	}
	template := &models.ChallengeTemplate{
		LocationArchetypeID: step.LocationArchetype.ID,
		Question:            buildGeneratedChallengeTemplateQuestion(themePrompt, step, nextStep),
		Description:         buildGeneratedChallengeTemplateDescription(themePrompt, step, nextStep),
		ImageURL:            "",
		ThumbnailURL:        "",
		ScaleWithUserLevel:  false,
		RewardMode:          models.RewardModeRandom,
		RandomRewardSize:    models.RandomRewardSizeSmall,
		RewardExperience:    0,
		Reward:              0,
		InventoryItemID:     nil,
		ItemChoiceRewards:   models.ChallengeTemplateItemChoiceRewards{},
		SubmissionType:      models.DefaultQuestNodeSubmissionType(),
		Difficulty:          maxInt(0, 5+(stepIndex*2)),
		StatTags:            models.StringArray{},
		Proficiency:         nil,
	}
	if err := s.dbClient.ChallengeTemplate().Create(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to create generated challenge template")
	}
	return template, nil
}

func buildGeneratedChallengeTemplateQuestion(
	themePrompt string,
	step normalizedQuestTemplateGeneratorStep,
	nextStep *normalizedQuestTemplateGeneratorStep,
) string {
	locationName := strings.TrimSpace(step.LocationArchetype.Name)
	if locationName == "" {
		locationName = "the site"
	}
	theme := strings.TrimSpace(themePrompt)
	switch {
	case nextStep != nil && nextStep.Content == questTemplateGeneratorContentScenario:
		if theme != "" {
			return fmt.Sprintf("Photograph a detail at %s that feels tied to %s.", locationName, theme)
		}
		return fmt.Sprintf("Photograph a detail at %s that hints a story is unfolding there.", locationName)
	case nextStep != nil && nextStep.Content == questTemplateGeneratorContentMonster:
		if theme != "" {
			return fmt.Sprintf("Photograph a detail at %s that would make travelers wary of %s.", locationName, theme)
		}
		return fmt.Sprintf("Photograph a detail at %s that suggests danger could be near.", locationName)
	default:
		if theme != "" {
			return fmt.Sprintf("Photograph the detail at %s that best captures the mood of %s.", locationName, theme)
		}
		return fmt.Sprintf("Photograph the detail at %s that best captures the spirit of the place.", locationName)
	}
}

func buildGeneratedChallengeTemplateDescription(
	themePrompt string,
	step normalizedQuestTemplateGeneratorStep,
	nextStep *normalizedQuestTemplateGeneratorStep,
) string {
	locationName := strings.TrimSpace(step.LocationArchetype.Name)
	if locationName == "" {
		locationName = "this place"
	}
	theme := strings.TrimSpace(themePrompt)
	if nextStep == nil {
		if theme != "" {
			return fmt.Sprintf("A concrete on-site photo challenge at %s. The player should submit a detail that clearly expresses %s and can be judged from the image alone.", locationName, theme)
		}
		return fmt.Sprintf("A concrete on-site photo challenge at %s that can be judged from the submission alone.", locationName)
	}
	if theme != "" {
		return fmt.Sprintf("A concrete on-site photo challenge at %s that bridges toward the next %s beat in %s. The task should be enjoyable, locally grounded, and gradeable from the submission alone.", locationName, nextStep.Content, theme)
	}
	return fmt.Sprintf("A concrete on-site photo challenge at %s that bridges toward the next %s beat. The task should be enjoyable, locally grounded, and gradeable from the submission alone.", locationName, nextStep.Content)
}

func buildGeneratedQuestTemplateName(
	rawName string,
	themePrompt string,
	steps []normalizedQuestTemplateGeneratorStep,
) string {
	if trimmed := strings.TrimSpace(rawName); trimmed != "" {
		return trimmed
	}
	if trimmed := strings.TrimSpace(themePrompt); trimmed != "" {
		words := strings.Fields(trimmed)
		if len(words) > 5 {
			words = words[:5]
		}
		return strings.Join(words, " ")
	}
	if len(steps) > 0 && steps[0].LocationArchetype != nil {
		return strings.TrimSpace(steps[0].LocationArchetype.Name) + " Trail"
	}
	return "Generated Quest Template"
}

func buildGeneratedQuestTemplateDescription(
	themePrompt string,
	steps []normalizedQuestTemplateGeneratorStep,
) string {
	parts := make([]string, 0, len(steps))
	for _, step := range steps {
		label := string(step.Content)
		if step.LocationArchetype != nil {
			label = fmt.Sprintf("%s at %s", label, strings.TrimSpace(step.LocationArchetype.Name))
		} else {
			label = fmt.Sprintf("%s within %dm", label, step.ProximityMeters)
		}
		parts = append(parts, label)
	}
	if theme := strings.TrimSpace(themePrompt); theme != "" {
		return fmt.Sprintf("%s. Generated flow: %s.", theme, strings.Join(parts, " -> "))
	}
	return fmt.Sprintf("Generated flow: %s.", strings.Join(parts, " -> "))
}

func buildGeneratedQuestTemplateAcceptanceDialogue(
	themePrompt string,
	steps []normalizedQuestTemplateGeneratorStep,
) models.StringArray {
	lines := models.StringArray{}
	if theme := strings.TrimSpace(themePrompt); theme != "" {
		lines = append(lines, fmt.Sprintf("There is a trail tied to %s that needs following.", theme))
	}
	if len(steps) > 0 && steps[0].LocationArchetype != nil {
		lines = append(lines, fmt.Sprintf("Start at the %s and follow each lead in order.", strings.TrimSpace(steps[0].LocationArchetype.Name)))
	} else {
		lines = append(lines, "Follow each lead in order and keep the thread intact.")
	}
	return lines
}

func generatedQuestTemplateTokens(input string) []string {
	normalized := strings.ToLower(strings.TrimSpace(input))
	if normalized == "" {
		return nil
	}
	normalized = questTemplateGeneratorTokenPattern.ReplaceAllString(normalized, " ")
	words := strings.Fields(normalized)
	out := make([]string, 0, len(words))
	for _, word := range words {
		if len(word) < 3 {
			continue
		}
		out = append(out, word)
	}
	return out
}

func generatedQuestMonsterLocationConcept(step normalizedQuestTemplateGeneratorStep) string {
	if step.LocationArchetype == nil {
		return ""
	}
	return strings.TrimSpace(step.LocationArchetype.Name)
}
