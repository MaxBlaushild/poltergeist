package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type pointOfInterestRewardConfigRequest struct {
	RewardMode       string                       `json:"rewardMode"`
	RandomRewardSize string                       `json:"randomRewardSize"`
	RewardExperience int                          `json:"rewardExperience"`
	RewardGold       int                          `json:"rewardGold"`
	MaterialRewards  []baseMaterialRewardPayload  `json:"materialRewards"`
	ItemRewards      []scenarioRewardItemPayload  `json:"itemRewards"`
	SpellRewards     []scenarioRewardSpellPayload `json:"spellRewards"`
}

type pointOfInterestRewardConfig struct {
	RewardMode       models.RewardMode
	RandomRewardSize models.RandomRewardSize
	RewardExperience int
	RewardGold       int
	MaterialRewards  models.BaseMaterialRewards
	ItemRewards      []models.PointOfInterestItemReward
	SpellRewards     []models.PointOfInterestSpellReward
}

func (s *server) parsePointOfInterestItemRewards(
	input []scenarioRewardItemPayload,
) ([]models.PointOfInterestItemReward, error) {
	rewards := make([]models.PointOfInterestItemReward, 0, len(input))
	for _, reward := range input {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			return nil, fmt.Errorf("itemRewards require inventoryItemId and positive quantity")
		}
		rewards = append(rewards, models.PointOfInterestItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return rewards, nil
}

func (s *server) parsePointOfInterestSpellRewards(
	ctx context.Context,
	input []scenarioRewardSpellPayload,
) ([]models.PointOfInterestSpellReward, error) {
	rewards := make([]models.PointOfInterestSpellReward, 0, len(input))
	for _, reward := range input {
		spellID, err := uuid.Parse(strings.TrimSpace(reward.SpellID))
		if err != nil {
			return nil, fmt.Errorf("invalid spellId")
		}
		spell, err := s.dbClient.Spell().FindByID(ctx, spellID)
		if err != nil {
			return nil, err
		}
		if spell == nil {
			return nil, fmt.Errorf("spellId not found")
		}
		rewards = append(rewards, models.PointOfInterestSpellReward{SpellID: spellID})
	}
	return rewards, nil
}

func (s *server) parsePointOfInterestRewardConfig(
	ctx context.Context,
	body pointOfInterestRewardConfigRequest,
) (*pointOfInterestRewardConfig, error) {
	if body.RewardExperience < 0 || body.RewardGold < 0 {
		return nil, fmt.Errorf("reward values must be zero or greater")
	}
	materialRewards, err := parseBaseMaterialRewards(body.MaterialRewards, "materialRewards")
	if err != nil {
		return nil, err
	}
	itemRewards, err := s.parsePointOfInterestItemRewards(body.ItemRewards)
	if err != nil {
		return nil, err
	}
	spellRewards, err := s.parsePointOfInterestSpellRewards(ctx, body.SpellRewards)
	if err != nil {
		return nil, err
	}
	rewardMode := models.NormalizeRewardMode(body.RewardMode)
	if strings.TrimSpace(body.RewardMode) == "" {
		if body.RewardExperience > 0 ||
			body.RewardGold > 0 ||
			len(materialRewards) > 0 ||
			len(itemRewards) > 0 ||
			len(spellRewards) > 0 {
			rewardMode = models.RewardModeExplicit
		}
	}
	config := &pointOfInterestRewardConfig{
		RewardMode:       rewardMode,
		RandomRewardSize: models.NormalizeRandomRewardSize(body.RandomRewardSize),
		RewardExperience: body.RewardExperience,
		RewardGold:       body.RewardGold,
		MaterialRewards:  materialRewards,
		ItemRewards:      itemRewards,
		SpellRewards:     spellRewards,
	}
	if config.RewardMode == models.RewardModeRandom {
		config.RewardExperience = 0
		config.RewardGold = 0
		config.MaterialRewards = models.BaseMaterialRewards{}
		config.SpellRewards = []models.PointOfInterestSpellReward{}
	}
	return config, nil
}

func pointOfInterestRewardItemsFromRewards(
	rewards []models.PointOfInterestItemReward,
) []scenarioRewardItem {
	out := make([]scenarioRewardItem, 0, len(rewards))
	for _, reward := range rewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		out = append(out, scenarioRewardItem{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return out
}

func pointOfInterestRewardSpellsFromRewards(
	rewards []models.PointOfInterestSpellReward,
) []scenarioRewardSpell {
	out := make([]scenarioRewardSpell, 0, len(rewards))
	for _, reward := range rewards {
		if reward.SpellID == uuid.Nil {
			continue
		}
		out = append(out, scenarioRewardSpell{SpellID: reward.SpellID})
	}
	return out
}
