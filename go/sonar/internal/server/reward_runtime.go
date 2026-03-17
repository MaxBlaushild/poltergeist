package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func (s *server) randomRewardPlanForUser(
	ctx context.Context,
	userID uuid.UUID,
	size models.RandomRewardSize,
	seed string,
) (models.RandomRewardPlan, map[int]models.InventoryItem, error) {
	userLevel, err := s.currentUserLevel(ctx, userID)
	if err != nil {
		return models.RandomRewardPlan{}, nil, err
	}
	allItems, err := s.dbClient.InventoryItem().FindAllActiveInventoryItems(ctx)
	if err != nil {
		return models.RandomRewardPlan{}, nil, err
	}
	itemByID := make(map[int]models.InventoryItem, len(allItems))
	for _, item := range allItems {
		itemByID[item.ID] = item
	}
	normalizedSeed := strings.TrimSpace(seed)
	if normalizedSeed == "" {
		normalizedSeed = fmt.Sprintf("user:%s", userID)
	}
	plan := models.BuildRandomRewardPlan(
		userLevel,
		models.NormalizeRandomRewardSize(string(size)),
		normalizedSeed,
		allItems,
	)
	return plan, itemByID, nil
}

func randomRewardPlanToScenarioItems(plan models.RandomRewardPlan) []scenarioRewardItem {
	items := make([]scenarioRewardItem, 0, len(plan.ItemGrants))
	for _, grant := range plan.ItemGrants {
		if grant.InventoryItemID <= 0 || grant.Quantity <= 0 {
			continue
		}
		items = append(items, scenarioRewardItem{
			InventoryItemID: grant.InventoryItemID,
			Quantity:        grant.Quantity,
		})
	}
	return items
}

func randomRewardPlanToMonsterItemRewards(
	plan models.RandomRewardPlan,
	itemByID map[int]models.InventoryItem,
	monsterID uuid.UUID,
) []models.MonsterItemReward {
	out := make([]models.MonsterItemReward, 0, len(plan.ItemGrants))
	for _, grant := range plan.ItemGrants {
		if grant.InventoryItemID <= 0 || grant.Quantity <= 0 {
			continue
		}
		reward := models.MonsterItemReward{
			MonsterID:       monsterID,
			InventoryItemID: grant.InventoryItemID,
			Quantity:        grant.Quantity,
		}
		if item, ok := itemByID[grant.InventoryItemID]; ok {
			reward.InventoryItem = item
		}
		out = append(out, reward)
	}
	return out
}

func (s *server) applyMonsterRewardsForUser(
	ctx context.Context,
	userID uuid.UUID,
	monster *models.Monster,
	response *monsterResponse,
) error {
	if monster == nil || response == nil {
		return nil
	}
	rewardMode := models.NormalizeRewardMode(string(monster.RewardMode))
	rewardSize := models.NormalizeRandomRewardSize(string(monster.RandomRewardSize))
	response.RewardMode = rewardMode
	response.RandomRewardSize = rewardSize

	if rewardMode == models.RewardModeRandom {
		plan, itemByID, err := s.randomRewardPlanForUser(
			ctx,
			userID,
			rewardSize,
			fmt.Sprintf("monster:%s:user:%s", monster.ID, userID),
		)
		if err != nil {
			return err
		}
		response.RewardExperience = plan.Experience
		response.RewardGold = plan.Gold
		response.ItemRewards = randomRewardPlanToMonsterItemRewards(plan, itemByID, monster.ID)
		return nil
	}

	response.RewardExperience = monster.RewardExperience
	if response.RewardExperience < 0 {
		response.RewardExperience = 0
	}
	response.RewardGold = monster.RewardGold
	if response.RewardGold < 0 {
		response.RewardGold = 0
	}
	response.ItemRewards = monster.ItemRewards
	return nil
}
