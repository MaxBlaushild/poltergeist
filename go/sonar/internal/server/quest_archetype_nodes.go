package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type questArchetypeNodePayload struct {
	NodeType                   string                       `json:"nodeType"`
	LocationArchetypeID        *uuid.UUID                   `json:"locationArchetypeID"`
	LocationSelectionMode      string                       `json:"locationSelectionMode"`
	ChallengeTemplateID        *uuid.UUID                   `json:"challengeTemplateId"`
	ScenarioTemplateID         *uuid.UUID                   `json:"scenarioTemplateId"`
	FetchCharacterID           *uuid.UUID                   `json:"fetchCharacterId"`
	FetchRequirements          []scenarioRewardItemPayload  `json:"fetchRequirements"`
	ObjectiveDescription       string                       `json:"objectiveDescription"`
	StoryFlagKey               string                       `json:"storyFlagKey"`
	MonsterTemplateIDs         []string                     `json:"monsterTemplateIds"`
	MonsterIDs                 []string                     `json:"monsterIds"`
	TargetLevel                *int                         `json:"targetLevel"`
	EncounterProximityMeters   *int                         `json:"encounterProximityMeters"`
	Difficulty                 *int                         `json:"difficulty"`
	ExpositionTitle            string                       `json:"expositionTitle"`
	ExpositionDescription      string                       `json:"expositionDescription"`
	ExpositionDialogue         []models.DialogueMessage     `json:"expositionDialogue"`
	ExpositionRewardMode       string                       `json:"expositionRewardMode"`
	ExpositionRandomRewardSize string                       `json:"expositionRandomRewardSize"`
	ExpositionRewardExperience int                          `json:"expositionRewardExperience"`
	ExpositionRewardGold       int                          `json:"expositionRewardGold"`
	ExpositionMaterialRewards  []baseMaterialRewardPayload  `json:"expositionMaterialRewards"`
	ExpositionItemRewards      []scenarioRewardItemPayload  `json:"expositionItemRewards"`
	ExpositionSpellRewards     []scenarioRewardSpellPayload `json:"expositionSpellRewards"`
}

func (p questArchetypeNodePayload) hasExplicitConfig() bool {
	return strings.TrimSpace(p.NodeType) != "" ||
		p.LocationArchetypeID != nil ||
		strings.TrimSpace(p.LocationSelectionMode) != "" ||
		p.ChallengeTemplateID != nil ||
		p.ScenarioTemplateID != nil ||
		p.hasFetchQuestConfig() ||
		p.hasStoryFlagConfig() ||
		len(p.MonsterTemplateIDs) > 0 ||
		len(p.MonsterIDs) > 0 ||
		p.TargetLevel != nil ||
		p.EncounterProximityMeters != nil ||
		p.hasExpositionConfig()
}

func (p questArchetypeNodePayload) hasStoryFlagConfig() bool {
	return strings.TrimSpace(p.StoryFlagKey) != ""
}

func (p questArchetypeNodePayload) hasFetchQuestConfig() bool {
	return p.FetchCharacterID != nil || len(p.FetchRequirements) > 0
}

func (p questArchetypeNodePayload) hasExpositionConfig() bool {
	return strings.TrimSpace(p.ExpositionTitle) != "" ||
		strings.TrimSpace(p.ExpositionDescription) != "" ||
		len(p.ExpositionDialogue) > 0 ||
		strings.TrimSpace(p.ExpositionRewardMode) != "" ||
		strings.TrimSpace(p.ExpositionRandomRewardSize) != "" ||
		p.ExpositionRewardExperience > 0 ||
		p.ExpositionRewardGold > 0 ||
		len(p.ExpositionMaterialRewards) > 0 ||
		len(p.ExpositionItemRewards) > 0 ||
		len(p.ExpositionSpellRewards) > 0
}

func (p questArchetypeNodePayload) inferredNodeType() models.QuestArchetypeNodeType {
	if strings.TrimSpace(p.NodeType) != "" {
		return models.NormalizeQuestArchetypeNodeType(p.NodeType)
	}
	if p.ScenarioTemplateID != nil {
		return models.QuestArchetypeNodeTypeScenario
	}
	if p.hasFetchQuestConfig() {
		return models.QuestArchetypeNodeTypeFetchQuest
	}
	if p.hasStoryFlagConfig() {
		return models.QuestArchetypeNodeTypeStoryFlag
	}
	if p.hasExpositionConfig() {
		return models.QuestArchetypeNodeTypeExposition
	}
	if len(p.MonsterTemplateIDs) > 0 ||
		len(p.MonsterIDs) > 0 ||
		p.TargetLevel != nil {
		return models.QuestArchetypeNodeTypeMonsterEncounter
	}
	if p.ChallengeTemplateID != nil ||
		p.LocationArchetypeID != nil ||
		strings.TrimSpace(p.LocationSelectionMode) != "" ||
		p.EncounterProximityMeters != nil {
		return models.QuestArchetypeNodeTypeChallenge
	}
	return models.QuestArchetypeNodeTypeChallenge
}

func clearQuestArchetypeNodeExposition(node *models.QuestArchetypeNode) {
	if node == nil {
		return
	}
	node.ExpositionTitle = ""
	node.ExpositionDescription = ""
	node.ExpositionDialogue = models.DialogueSequence{}
	node.ExpositionRewardMode = models.RewardModeRandom
	node.ExpositionRandomRewardSize = models.RandomRewardSizeSmall
	node.ExpositionRewardExperience = 0
	node.ExpositionRewardGold = 0
	node.ExpositionMaterialRewards = models.BaseMaterialRewards{}
	node.ExpositionItemRewards = models.QuestArchetypeExpositionItemRewards{}
	node.ExpositionSpellRewards = models.QuestArchetypeExpositionSpellRewards{}
}

func clearQuestArchetypeNodeChallenge(node *models.QuestArchetypeNode) {
	if node == nil {
		return
	}
	node.ChallengeTemplateID = nil
	node.ChallengeTemplate = nil
}

func clearQuestArchetypeNodeStoryFlag(node *models.QuestArchetypeNode) {
	if node == nil {
		return
	}
	node.StoryFlagKey = ""
}

func clearQuestArchetypeNodeFetchQuest(node *models.QuestArchetypeNode) {
	if node == nil {
		return
	}
	node.FetchCharacterID = nil
	node.FetchCharacter = nil
	node.FetchRequirements = models.FetchQuestRequirements{}
}

func questArchetypeNodeExpositionItemRewards(
	input []models.ExpositionItemReward,
) models.QuestArchetypeExpositionItemRewards {
	rewards := make(models.QuestArchetypeExpositionItemRewards, 0, len(input))
	for _, reward := range input {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		rewards = append(rewards, models.QuestArchetypeExpositionItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return rewards
}

func questArchetypeNodeExpositionSpellRewards(
	input []models.ExpositionSpellReward,
) models.QuestArchetypeExpositionSpellRewards {
	rewards := make(models.QuestArchetypeExpositionSpellRewards, 0, len(input))
	for _, reward := range input {
		if reward.SpellID == uuid.Nil {
			continue
		}
		rewards = append(rewards, models.QuestArchetypeExpositionSpellReward{
			SpellID: reward.SpellID,
		})
	}
	return rewards
}

func (s *server) applyQuestArchetypeNodeStoryFlagPayload(
	node *models.QuestArchetypeNode,
	payload questArchetypeNodePayload,
	requireConfig bool,
) error {
	if node == nil {
		return fmt.Errorf("quest archetype node is required")
	}
	storyFlagKey := models.NormalizeStoryFlagKey(payload.StoryFlagKey)
	if storyFlagKey == "" {
		if requireConfig {
			return fmt.Errorf("storyFlagKey is required for story flag nodes")
		}
		return nil
	}
	node.StoryFlagKey = storyFlagKey
	return nil
}

func (s *server) applyQuestArchetypeNodeFetchQuestPayload(
	ctx context.Context,
	node *models.QuestArchetypeNode,
	payload questArchetypeNodePayload,
	requireConfig bool,
) error {
	if node == nil {
		return fmt.Errorf("quest archetype node is required")
	}
	if !payload.hasFetchQuestConfig() {
		if requireConfig {
			return fmt.Errorf("fetch quest configuration is required for fetch quest nodes")
		}
		return nil
	}
	if payload.FetchCharacterID == nil || *payload.FetchCharacterID == uuid.Nil {
		return fmt.Errorf("fetchCharacterId is required for fetch quest nodes")
	}
	character, err := s.dbClient.Character().FindByID(ctx, *payload.FetchCharacterID)
	if err != nil {
		return fmt.Errorf("fetchCharacterId could not be loaded")
	}
	if character == nil {
		return fmt.Errorf("fetchCharacterId could not be loaded")
	}
	itemRewards, err := s.parseExpositionItemRewards(payload.FetchRequirements)
	if err != nil {
		return err
	}
	requirements := make([]models.FetchQuestRequirement, 0, len(itemRewards))
	for _, reward := range itemRewards {
		requirements = append(requirements, models.FetchQuestRequirement{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	normalizedRequirements := models.NormalizeFetchQuestRequirements(requirements)
	if len(normalizedRequirements) == 0 {
		return fmt.Errorf("fetch quest nodes require at least one item requirement")
	}
	node.FetchCharacterID = payload.FetchCharacterID
	node.FetchCharacter = nil
	node.FetchRequirements = normalizedRequirements
	return nil
}

func (s *server) applyQuestArchetypeNodeExpositionPayload(
	ctx context.Context,
	node *models.QuestArchetypeNode,
	payload questArchetypeNodePayload,
	requireConfig bool,
) error {
	if node == nil {
		return fmt.Errorf("quest archetype node is required")
	}
	if !payload.hasExpositionConfig() {
		if requireConfig {
			return fmt.Errorf("exposition configuration is required for exposition nodes")
		}
		return nil
	}

	title := strings.TrimSpace(payload.ExpositionTitle)
	if title == "" {
		return fmt.Errorf("expositionTitle is required for exposition nodes")
	}
	dialogue, err := s.parseExpositionDialogue(ctx, payload.ExpositionDialogue)
	if err != nil {
		return err
	}
	if payload.ExpositionRewardExperience < 0 || payload.ExpositionRewardGold < 0 {
		return fmt.Errorf("exposition reward values must be zero or greater")
	}
	materialRewards, err := parseBaseMaterialRewards(
		payload.ExpositionMaterialRewards,
		"expositionMaterialRewards",
	)
	if err != nil {
		return err
	}
	itemRewards, err := s.parseExpositionItemRewards(payload.ExpositionItemRewards)
	if err != nil {
		return err
	}
	spellRewards, err := s.parseExpositionSpellRewards(ctx, payload.ExpositionSpellRewards)
	if err != nil {
		return err
	}
	rewardMode := models.NormalizeRewardMode(payload.ExpositionRewardMode)
	if strings.TrimSpace(payload.ExpositionRewardMode) == "" {
		if payload.ExpositionRewardExperience > 0 ||
			payload.ExpositionRewardGold > 0 ||
			len(materialRewards) > 0 ||
			len(itemRewards) > 0 ||
			len(spellRewards) > 0 {
			rewardMode = models.RewardModeExplicit
		}
	}

	node.ExpositionTitle = title
	node.ExpositionDescription = strings.TrimSpace(payload.ExpositionDescription)
	node.ExpositionDialogue = dialogue
	node.ExpositionRewardMode = rewardMode
	node.ExpositionRandomRewardSize = models.NormalizeRandomRewardSize(
		payload.ExpositionRandomRewardSize,
	)
	node.ExpositionRewardExperience = payload.ExpositionRewardExperience
	node.ExpositionRewardGold = payload.ExpositionRewardGold
	node.ExpositionMaterialRewards = materialRewards
	node.ExpositionItemRewards = questArchetypeNodeExpositionItemRewards(itemRewards)
	node.ExpositionSpellRewards = questArchetypeNodeExpositionSpellRewards(spellRewards)
	return nil
}

func (s *server) normalizeQuestArchetypeNodeLocationConfig(
	ctx context.Context,
	payload questArchetypeNodePayload,
) (*uuid.UUID, models.QuestArchetypeNodeLocationSelectionMode, int, error) {
	var locationArchetypeID *uuid.UUID
	if payload.LocationArchetypeID != nil {
		if *payload.LocationArchetypeID == uuid.Nil {
			return nil, "", 0, fmt.Errorf("locationArchetypeID must be a valid UUID when provided")
		}
		locationArchetype, err := s.dbClient.LocationArchetype().FindByID(ctx, *payload.LocationArchetypeID)
		if err != nil {
			return nil, "", 0, fmt.Errorf("locationArchetypeID could not be loaded")
		}
		if locationArchetype == nil {
			return nil, "", 0, fmt.Errorf("locationArchetypeID could not be loaded")
		}
		locationArchetypeID = payload.LocationArchetypeID
	}

	proximityMeters := 100
	if payload.EncounterProximityMeters != nil {
		proximityMeters = *payload.EncounterProximityMeters
	}
	if proximityMeters < 0 {
		return nil, "", 0, fmt.Errorf("encounterProximityMeters must be zero or greater")
	}

	selectionMode := models.NormalizeQuestArchetypeNodeLocationSelectionMode(
		payload.LocationSelectionMode,
	)
	return locationArchetypeID, selectionMode, proximityMeters, nil
}

func (s *server) resolveQuestArchetypeNodeChallengeTemplate(
	ctx context.Context,
	payload questArchetypeNodePayload,
) (*models.ChallengeTemplate, error) {
	if payload.ChallengeTemplateID == nil {
		return nil, nil
	}
	if *payload.ChallengeTemplateID == uuid.Nil {
		return nil, fmt.Errorf("challengeTemplateId must be a valid UUID when provided")
	}
	challengeTemplate, err := s.dbClient.ChallengeTemplate().FindByID(ctx, *payload.ChallengeTemplateID)
	if err != nil {
		return nil, fmt.Errorf("challengeTemplateId could not be loaded")
	}
	if challengeTemplate == nil {
		return nil, fmt.Errorf("challengeTemplateId could not be loaded")
	}
	return challengeTemplate, nil
}

func (s *server) applyQuestArchetypeNodePayload(
	ctx context.Context,
	node *models.QuestArchetypeNode,
	payload questArchetypeNodePayload,
	requireConfig bool,
) error {
	if node == nil {
		return fmt.Errorf("quest archetype node is required")
	}
	if payload.Difficulty != nil {
		if *payload.Difficulty < 0 {
			return fmt.Errorf("difficulty must be zero or greater")
		}
		node.Difficulty = *payload.Difficulty
	}
	node.ObjectiveDescription = strings.TrimSpace(payload.ObjectiveDescription)
	if !payload.hasExplicitConfig() {
		if requireConfig {
			return fmt.Errorf("quest archetype node configuration is required")
		}
		return nil
	}

	nodeType := payload.inferredNodeType()
	locationArchetypeID, locationSelectionMode, proximityMeters, err := s.normalizeQuestArchetypeNodeLocationConfig(
		ctx,
		payload,
	)
	if err != nil {
		return err
	}
	switch nodeType {
	case models.QuestArchetypeNodeTypeChallenge:
		if payload.ChallengeTemplateID == nil &&
			models.NormalizeQuestArchetypeNodeType(string(node.NodeType)) == models.QuestArchetypeNodeTypeChallenge &&
			node.ChallengeTemplateID != nil &&
			*node.ChallengeTemplateID != uuid.Nil {
			existingTemplateID := *node.ChallengeTemplateID
			payload.ChallengeTemplateID = &existingTemplateID
		}
		challengeTemplate, err := s.resolveQuestArchetypeNodeChallengeTemplate(
			ctx,
			payload,
		)
		if err != nil {
			return err
		}
		if challengeTemplate == nil {
			return fmt.Errorf("challengeTemplateId is required for challenge nodes")
		}
		node.NodeType = models.QuestArchetypeNodeTypeChallenge
		node.LocationArchetypeID = locationArchetypeID
		node.LocationArchetype = nil
		node.LocationSelectionMode = locationSelectionMode
		if challengeTemplate != nil {
			node.ChallengeTemplateID = &challengeTemplate.ID
		} else {
			node.ChallengeTemplateID = nil
		}
		node.ChallengeTemplate = nil
		node.ScenarioTemplateID = nil
		node.ScenarioTemplate = nil
		node.MonsterTemplateIDs = models.StringArray{}
		node.TargetLevel = 1
		node.EncounterRewardMode = models.RewardModeExplicit
		node.EncounterRandomRewardSize = models.RandomRewardSizeSmall
		node.EncounterRewardExperience = 0
		node.EncounterRewardGold = 0
		node.EncounterMaterialRewards = models.BaseMaterialRewards{}
		node.EncounterItemRewards = models.MonsterEncounterRewardItems{}
		node.EncounterProximityMeters = proximityMeters
		clearQuestArchetypeNodeExposition(node)
		clearQuestArchetypeNodeFetchQuest(node)
		clearQuestArchetypeNodeStoryFlag(node)
	case models.QuestArchetypeNodeTypeExposition:
		if err := s.applyQuestArchetypeNodeExpositionPayload(
			ctx,
			node,
			payload,
			requireConfig || node.NodeType != models.QuestArchetypeNodeTypeExposition,
		); err != nil {
			return err
		}

		node.NodeType = models.QuestArchetypeNodeTypeExposition
		node.LocationArchetypeID = locationArchetypeID
		node.LocationArchetype = nil
		node.LocationSelectionMode = locationSelectionMode
		clearQuestArchetypeNodeChallenge(node)
		node.ScenarioTemplateID = nil
		node.ScenarioTemplate = nil
		node.MonsterTemplateIDs = models.StringArray{}
		node.TargetLevel = 1
		node.EncounterRewardMode = models.RewardModeExplicit
		node.EncounterRandomRewardSize = models.RandomRewardSizeSmall
		node.EncounterRewardExperience = 0
		node.EncounterRewardGold = 0
		node.EncounterMaterialRewards = models.BaseMaterialRewards{}
		node.EncounterItemRewards = models.MonsterEncounterRewardItems{}
		node.EncounterProximityMeters = proximityMeters
		clearQuestArchetypeNodeFetchQuest(node)
		clearQuestArchetypeNodeStoryFlag(node)
	case models.QuestArchetypeNodeTypeScenario:
		if payload.ScenarioTemplateID == nil || *payload.ScenarioTemplateID == uuid.Nil {
			return fmt.Errorf("scenarioTemplateId is required for scenario nodes")
		}
		scenarioTemplate, err := s.dbClient.ScenarioTemplate().FindByID(ctx, *payload.ScenarioTemplateID)
		if err != nil {
			return fmt.Errorf("scenarioTemplateId could not be loaded")
		}
		if scenarioTemplate == nil {
			return fmt.Errorf("scenarioTemplateId could not be loaded")
		}
		node.NodeType = models.QuestArchetypeNodeTypeScenario
		node.LocationArchetypeID = locationArchetypeID
		node.LocationArchetype = nil
		node.LocationSelectionMode = locationSelectionMode
		clearQuestArchetypeNodeChallenge(node)
		node.ScenarioTemplateID = payload.ScenarioTemplateID
		node.ScenarioTemplate = nil
		node.MonsterTemplateIDs = models.StringArray{}
		node.TargetLevel = 1
		node.EncounterRewardMode = models.RewardModeExplicit
		node.EncounterRandomRewardSize = models.RandomRewardSizeSmall
		node.EncounterRewardExperience = 0
		node.EncounterRewardGold = 0
		node.EncounterMaterialRewards = models.BaseMaterialRewards{}
		node.EncounterItemRewards = models.MonsterEncounterRewardItems{}
		node.EncounterProximityMeters = proximityMeters
		clearQuestArchetypeNodeExposition(node)
		clearQuestArchetypeNodeFetchQuest(node)
		clearQuestArchetypeNodeStoryFlag(node)
	case models.QuestArchetypeNodeTypeFetchQuest:
		if err := s.applyQuestArchetypeNodeFetchQuestPayload(
			ctx,
			node,
			payload,
			requireConfig || node.NodeType != models.QuestArchetypeNodeTypeFetchQuest,
		); err != nil {
			return err
		}
		node.NodeType = models.QuestArchetypeNodeTypeFetchQuest
		node.LocationArchetypeID = locationArchetypeID
		node.LocationArchetype = nil
		node.LocationSelectionMode = locationSelectionMode
		clearQuestArchetypeNodeChallenge(node)
		node.ScenarioTemplateID = nil
		node.ScenarioTemplate = nil
		node.MonsterTemplateIDs = models.StringArray{}
		node.TargetLevel = 1
		node.EncounterRewardMode = models.RewardModeExplicit
		node.EncounterRandomRewardSize = models.RandomRewardSizeSmall
		node.EncounterRewardExperience = 0
		node.EncounterRewardGold = 0
		node.EncounterMaterialRewards = models.BaseMaterialRewards{}
		node.EncounterItemRewards = models.MonsterEncounterRewardItems{}
		node.EncounterProximityMeters = proximityMeters
		clearQuestArchetypeNodeExposition(node)
		clearQuestArchetypeNodeStoryFlag(node)
	case models.QuestArchetypeNodeTypeMonsterEncounter:
		monsterTemplateIDs, err := s.parseQuestArchetypeNodeMonsterTemplateIDs(
			ctx,
			payload.MonsterTemplateIDs,
			payload.MonsterIDs,
		)
		if err != nil {
			return err
		}
		if len(monsterTemplateIDs) == 0 {
			return fmt.Errorf("monsterTemplateIds must contain at least one monster template")
		}
		targetLevel := 1
		if payload.TargetLevel != nil {
			targetLevel = *payload.TargetLevel
		}
		if targetLevel < 1 {
			return fmt.Errorf("targetLevel must be one or greater")
		}

		node.NodeType = models.QuestArchetypeNodeTypeMonsterEncounter
		node.LocationArchetypeID = locationArchetypeID
		node.LocationArchetype = nil
		node.LocationSelectionMode = locationSelectionMode
		clearQuestArchetypeNodeChallenge(node)
		node.ScenarioTemplateID = nil
		node.ScenarioTemplate = nil
		node.MonsterTemplateIDs = models.StringArray(monsterTemplateIDs)
		node.TargetLevel = targetLevel
		node.EncounterRewardMode = models.RewardModeExplicit
		node.EncounterRandomRewardSize = models.RandomRewardSizeSmall
		node.EncounterRewardExperience = 0
		node.EncounterRewardGold = 0
		node.EncounterMaterialRewards = models.BaseMaterialRewards{}
		node.EncounterItemRewards = models.MonsterEncounterRewardItems{}
		node.EncounterProximityMeters = proximityMeters
		clearQuestArchetypeNodeExposition(node)
		clearQuestArchetypeNodeFetchQuest(node)
		clearQuestArchetypeNodeStoryFlag(node)
	case models.QuestArchetypeNodeTypeStoryFlag:
		if err := s.applyQuestArchetypeNodeStoryFlagPayload(
			node,
			payload,
			requireConfig || node.NodeType != models.QuestArchetypeNodeTypeStoryFlag,
		); err != nil {
			return err
		}
		node.NodeType = models.QuestArchetypeNodeTypeStoryFlag
		node.LocationArchetypeID = nil
		node.LocationArchetype = nil
		node.LocationSelectionMode = models.QuestArchetypeNodeLocationSelectionModeRandom
		clearQuestArchetypeNodeChallenge(node)
		node.ScenarioTemplateID = nil
		node.ScenarioTemplate = nil
		node.MonsterTemplateIDs = models.StringArray{}
		node.TargetLevel = 1
		node.EncounterRewardMode = models.RewardModeExplicit
		node.EncounterRandomRewardSize = models.RandomRewardSizeSmall
		node.EncounterRewardExperience = 0
		node.EncounterRewardGold = 0
		node.EncounterMaterialRewards = models.BaseMaterialRewards{}
		node.EncounterItemRewards = models.MonsterEncounterRewardItems{}
		node.EncounterProximityMeters = 100
		clearQuestArchetypeNodeExposition(node)
		clearQuestArchetypeNodeFetchQuest(node)
	default:
		return fmt.Errorf("unsupported quest archetype node type")
	}
	return nil
}

func (s *server) parseQuestArchetypeNodeMonsterTemplateIDs(
	ctx context.Context,
	templateInput []string,
	legacyMonsterInput []string,
) ([]string, error) {
	normalized := make([]string, 0, len(templateInput))
	seen := map[uuid.UUID]struct{}{}
	for idx, raw := range templateInput {
		templateID, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil || templateID == uuid.Nil {
			return nil, fmt.Errorf("monsterTemplateIds[%d] must be a valid UUID", idx)
		}
		if _, ok := seen[templateID]; !ok {
			if _, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID); err != nil {
				return nil, fmt.Errorf("monsterTemplateIds[%d] could not be loaded", idx)
			}
			seen[templateID] = struct{}{}
		}
		normalized = append(normalized, templateID.String())
	}
	if len(normalized) == 0 {
		for idx, raw := range legacyMonsterInput {
			monsterID, err := uuid.Parse(strings.TrimSpace(raw))
			if err != nil || monsterID == uuid.Nil {
				return nil, fmt.Errorf("monsterIds[%d] must be a valid UUID", idx)
			}
			monster, err := s.dbClient.Monster().FindByID(ctx, monsterID)
			if err != nil {
				return nil, fmt.Errorf("monsterIds[%d] could not be loaded", idx)
			}
			if monster.TemplateID == nil || *monster.TemplateID == uuid.Nil {
				return nil, fmt.Errorf("monsterIds[%d] does not reference a monster with a template", idx)
			}
			if _, ok := seen[*monster.TemplateID]; ok {
				continue
			}
			seen[*monster.TemplateID] = struct{}{}
			normalized = append(normalized, monster.TemplateID.String())
		}
	}
	if normalized == nil {
		return []string{}, nil
	}
	return normalized, nil
}
