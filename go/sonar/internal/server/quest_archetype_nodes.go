package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type questArchetypeNodePayload struct {
	NodeType                  string                      `json:"nodeType"`
	LocationArchetypeID       *uuid.UUID                  `json:"locationArchetypeID"`
	ScenarioTemplateID        *uuid.UUID                  `json:"scenarioTemplateId"`
	MonsterTemplateIDs        []string                    `json:"monsterTemplateIds"`
	MonsterIDs                []string                    `json:"monsterIds"`
	TargetLevel               *int                        `json:"targetLevel"`
	EncounterRewardMode       string                      `json:"encounterRewardMode"`
	EncounterRandomRewardSize string                      `json:"encounterRandomRewardSize"`
	EncounterRewardExperience *int                        `json:"encounterRewardExperience"`
	EncounterRewardGold       *int                        `json:"encounterRewardGold"`
	EncounterMaterialRewards  []baseMaterialRewardPayload `json:"encounterMaterialRewards"`
	EncounterItemRewards      []scenarioRewardItemPayload `json:"encounterItemRewards"`
	EncounterProximityMeters  *int                        `json:"encounterProximityMeters"`
	Difficulty                *int                        `json:"difficulty"`
}

func (p questArchetypeNodePayload) hasExplicitConfig() bool {
	return strings.TrimSpace(p.NodeType) != "" ||
		p.LocationArchetypeID != nil ||
		p.ScenarioTemplateID != nil ||
		len(p.MonsterTemplateIDs) > 0 ||
		len(p.MonsterIDs) > 0 ||
		p.TargetLevel != nil ||
		strings.TrimSpace(p.EncounterRewardMode) != "" ||
		strings.TrimSpace(p.EncounterRandomRewardSize) != "" ||
		p.EncounterRewardExperience != nil ||
		p.EncounterRewardGold != nil ||
		len(p.EncounterMaterialRewards) > 0 ||
		len(p.EncounterItemRewards) > 0 ||
		p.EncounterProximityMeters != nil
}

func (p questArchetypeNodePayload) inferredNodeType() models.QuestArchetypeNodeType {
	if strings.TrimSpace(p.NodeType) != "" {
		return models.NormalizeQuestArchetypeNodeType(p.NodeType)
	}
	if p.LocationArchetypeID != nil {
		return models.QuestArchetypeNodeTypeLocation
	}
	if p.ScenarioTemplateID != nil {
		return models.QuestArchetypeNodeTypeScenario
	}
	if len(p.MonsterTemplateIDs) > 0 ||
		len(p.MonsterIDs) > 0 ||
		p.TargetLevel != nil ||
		p.EncounterRewardExperience != nil ||
		p.EncounterRewardGold != nil ||
		len(p.EncounterMaterialRewards) > 0 ||
		len(p.EncounterItemRewards) > 0 ||
		p.EncounterProximityMeters != nil ||
		strings.TrimSpace(p.EncounterRewardMode) != "" ||
		strings.TrimSpace(p.EncounterRandomRewardSize) != "" {
		return models.QuestArchetypeNodeTypeMonsterEncounter
	}
	return models.QuestArchetypeNodeTypeLocation
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
	if !payload.hasExplicitConfig() {
		if requireConfig {
			return fmt.Errorf("quest archetype node configuration is required")
		}
		return nil
	}

	nodeType := payload.inferredNodeType()
	switch nodeType {
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
		var locationArchetypeID *uuid.UUID
		if payload.LocationArchetypeID != nil {
			if *payload.LocationArchetypeID == uuid.Nil {
				return fmt.Errorf("locationArchetypeID must be a valid UUID when provided")
			}
			locationArchetype, err := s.dbClient.LocationArchetype().FindByID(ctx, *payload.LocationArchetypeID)
			if err != nil {
				return fmt.Errorf("locationArchetypeID could not be loaded")
			}
			if locationArchetype == nil {
				return fmt.Errorf("locationArchetypeID could not be loaded")
			}
			locationArchetypeID = payload.LocationArchetypeID
		}
		proximityMeters := 100
		if payload.EncounterProximityMeters != nil {
			proximityMeters = *payload.EncounterProximityMeters
		}
		if proximityMeters < 0 {
			return fmt.Errorf("encounterProximityMeters must be zero or greater")
		}

		node.NodeType = models.QuestArchetypeNodeTypeScenario
		node.LocationArchetypeID = locationArchetypeID
		node.LocationArchetype = nil
		node.ScenarioTemplateID = payload.ScenarioTemplateID
		node.ScenarioTemplate = nil
		node.MonsterTemplateIDs = models.StringArray{}
		node.TargetLevel = 1
		node.EncounterRewardMode = models.RewardModeRandom
		node.EncounterRandomRewardSize = models.RandomRewardSizeSmall
		node.EncounterRewardExperience = 0
		node.EncounterRewardGold = 0
		node.EncounterMaterialRewards = models.BaseMaterialRewards{}
		node.EncounterItemRewards = models.MonsterEncounterRewardItems{}
		node.EncounterProximityMeters = proximityMeters
	case models.QuestArchetypeNodeTypeMonsterEncounter:
		var locationArchetypeID *uuid.UUID
		if payload.LocationArchetypeID != nil {
			if *payload.LocationArchetypeID == uuid.Nil {
				return fmt.Errorf("locationArchetypeID must be a valid UUID when provided")
			}
			locationArchetype, err := s.dbClient.LocationArchetype().FindByID(ctx, *payload.LocationArchetypeID)
			if err != nil {
				return fmt.Errorf("locationArchetypeID could not be loaded")
			}
			if locationArchetype == nil {
				return fmt.Errorf("locationArchetypeID could not be loaded")
			}
			locationArchetypeID = payload.LocationArchetypeID
		}
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
		proximityMeters := 100
		if payload.EncounterProximityMeters != nil {
			proximityMeters = *payload.EncounterProximityMeters
		}
		if proximityMeters < 0 {
			return fmt.Errorf("encounterProximityMeters must be zero or greater")
		}
		rewardExperience := 0
		if payload.EncounterRewardExperience != nil {
			rewardExperience = *payload.EncounterRewardExperience
		}
		if rewardExperience < 0 {
			return fmt.Errorf("encounterRewardExperience must be zero or greater")
		}
		rewardGold := 0
		if payload.EncounterRewardGold != nil {
			rewardGold = *payload.EncounterRewardGold
		}
		if rewardGold < 0 {
			return fmt.Errorf("encounterRewardGold must be zero or greater")
		}
		materialRewards, err := parseBaseMaterialRewards(payload.EncounterMaterialRewards, "encounterMaterialRewards")
		if err != nil {
			return err
		}
		itemRewards, err := buildQuestArchetypeNodeEncounterItemRewards(payload.EncounterItemRewards)
		if err != nil {
			return err
		}

		rewardMode := models.RewardModeRandom
		if strings.TrimSpace(payload.EncounterRewardMode) != "" {
			rewardMode = models.NormalizeRewardMode(payload.EncounterRewardMode)
		} else if rewardExperience > 0 || rewardGold > 0 || len(materialRewards) > 0 || len(itemRewards) > 0 {
			rewardMode = models.RewardModeExplicit
		}

		node.NodeType = models.QuestArchetypeNodeTypeMonsterEncounter
		node.LocationArchetypeID = locationArchetypeID
		node.LocationArchetype = nil
		node.ScenarioTemplateID = nil
		node.ScenarioTemplate = nil
		node.MonsterTemplateIDs = models.StringArray(monsterTemplateIDs)
		node.TargetLevel = targetLevel
		node.EncounterRewardMode = rewardMode
		node.EncounterRandomRewardSize = models.NormalizeRandomRewardSize(payload.EncounterRandomRewardSize)
		node.EncounterRewardExperience = rewardExperience
		node.EncounterRewardGold = rewardGold
		node.EncounterMaterialRewards = materialRewards
		node.EncounterItemRewards = itemRewards
		node.EncounterProximityMeters = proximityMeters
	default:
		if payload.LocationArchetypeID == nil || *payload.LocationArchetypeID == uuid.Nil {
			return fmt.Errorf("locationArchetypeID is required for location nodes")
		}
		node.NodeType = models.QuestArchetypeNodeTypeLocation
		node.LocationArchetypeID = payload.LocationArchetypeID
		node.LocationArchetype = nil
		node.ScenarioTemplateID = nil
		node.ScenarioTemplate = nil
		node.MonsterTemplateIDs = models.StringArray{}
		node.TargetLevel = 1
		node.EncounterRewardMode = models.RewardModeRandom
		node.EncounterRandomRewardSize = models.RandomRewardSizeSmall
		node.EncounterRewardExperience = 0
		node.EncounterRewardGold = 0
		node.EncounterMaterialRewards = models.BaseMaterialRewards{}
		node.EncounterItemRewards = models.MonsterEncounterRewardItems{}
		node.EncounterProximityMeters = 100
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

func buildQuestArchetypeNodeEncounterItemRewards(
	input []scenarioRewardItemPayload,
) (models.MonsterEncounterRewardItems, error) {
	rewards := make(models.MonsterEncounterRewardItems, 0, len(input))
	for idx, reward := range input {
		if reward.InventoryItemID == 0 && reward.Quantity == 0 {
			continue
		}
		if reward.InventoryItemID <= 0 {
			return nil, fmt.Errorf("encounterItemRewards[%d].inventoryItemId must be positive", idx)
		}
		if reward.Quantity <= 0 {
			return nil, fmt.Errorf("encounterItemRewards[%d].quantity must be positive", idx)
		}
		rewards = append(rewards, models.MonsterEncounterRewardItem{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	if rewards == nil {
		return models.MonsterEncounterRewardItems{}, nil
	}
	return rewards, nil
}
