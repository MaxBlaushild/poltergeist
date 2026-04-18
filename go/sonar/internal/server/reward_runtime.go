package server

import (
	"context"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func (s *server) randomRewardPlanForUser(
	ctx context.Context,
	userID uuid.UUID,
	size models.RandomRewardSize,
	seed string,
) (models.RandomRewardPlan, map[int]models.InventoryItem, int, error) {
	userLevel, err := s.currentUserLevel(ctx, userID)
	if err != nil {
		return models.RandomRewardPlan{}, nil, 0, err
	}
	allItems, err := s.dbClient.InventoryItem().FindAllActiveInventoryItems(ctx)
	if err != nil {
		return models.RandomRewardPlan{}, nil, 0, err
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
	return plan, itemByID, userLevel, nil
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

func scenarioRewardItemsToRandomRewardItemGrants(
	items []scenarioRewardItem,
) []models.RandomRewardItemGrant {
	grants := make([]models.RandomRewardItemGrant, 0, len(items))
	for _, item := range items {
		if item.InventoryItemID <= 0 || item.Quantity <= 0 {
			continue
		}
		grants = append(grants, models.RandomRewardItemGrant{
			InventoryItemID: item.InventoryItemID,
			Quantity:        item.Quantity,
		})
	}
	return grants
}

func mergeScenarioRewardItems(itemSets ...[]scenarioRewardItem) []scenarioRewardItem {
	grantSets := make([][]models.RandomRewardItemGrant, 0, len(itemSets))
	for _, itemSet := range itemSets {
		grantSets = append(grantSets, scenarioRewardItemsToRandomRewardItemGrants(itemSet))
	}
	merged := models.MergeRandomRewardItemGrants(grantSets...)
	items := make([]scenarioRewardItem, 0, len(merged))
	for _, grant := range merged {
		items = append(items, scenarioRewardItem{
			InventoryItemID: grant.InventoryItemID,
			Quantity:        grant.Quantity,
		})
	}
	return items
}

func challengeRewardItemsFromInventoryItemID(
	inventoryItemID *int,
) []scenarioRewardItem {
	if inventoryItemID == nil || *inventoryItemID <= 0 {
		return []scenarioRewardItem{}
	}
	return []scenarioRewardItem{{
		InventoryItemID: *inventoryItemID,
		Quantity:        1,
	}}
}

func randomRewardPlanToMonsterItemRewards(
	plan models.RandomRewardPlan,
	itemByID map[int]models.InventoryItem,
	monsterID uuid.UUID,
) []models.MonsterItemReward {
	return randomRewardItemGrantsToMonsterItemRewards(
		plan.ItemGrants,
		itemByID,
		monsterID,
	)
}

func randomRewardItemGrantsToMonsterItemRewards(
	grants []models.RandomRewardItemGrant,
	itemByID map[int]models.InventoryItem,
	monsterID uuid.UUID,
) []models.MonsterItemReward {
	out := make([]models.MonsterItemReward, 0, len(grants))
	for _, grant := range grants {
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

func ensureMonsterEncounterRandomRewardHasItem(
	plan models.RandomRewardPlan,
	itemByID map[int]models.InventoryItem,
	userLevel int,
	seed string,
) models.RandomRewardPlan {
	if len(plan.ItemGrants) > 0 {
		return plan
	}

	item := pickDeterministicEncounterFallbackItem(itemByID, userLevel, seed, false)
	if item == nil {
		item = pickDeterministicEncounterFallbackItem(itemByID, userLevel, seed, true)
	}
	if item == nil {
		return plan
	}

	plan.ItemGrants = append(plan.ItemGrants, models.RandomRewardItemGrant{
		InventoryItemID: item.ID,
		Quantity:        1,
	})
	return plan
}

func pickDeterministicEncounterFallbackItem(
	itemByID map[int]models.InventoryItem,
	userLevel int,
	seed string,
	allowEquippable bool,
) *models.InventoryItem {
	candidates := make([]models.InventoryItem, 0, len(itemByID))
	for _, item := range itemByID {
		if item.ID <= 0 || item.Archived || item.IsCaptureType {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(item.RarityTier), "Not Droppable") {
			continue
		}
		isEquippable := false
		if item.EquipSlot != nil && strings.TrimSpace(*item.EquipSlot) != "" {
			isEquippable = true
		}
		if isEquippable != allowEquippable {
			continue
		}
		candidates = append(candidates, item)
	}
	if len(candidates) == 0 {
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].ItemLevel == candidates[j].ItemLevel {
			return candidates[i].ID < candidates[j].ID
		}
		return candidates[i].ItemLevel < candidates[j].ItemLevel
	})

	near := make([]models.InventoryItem, 0, len(candidates))
	withLevels := make([]models.InventoryItem, 0, len(candidates))
	for _, item := range candidates {
		if item.ItemLevel > 0 {
			withLevels = append(withLevels, item)
			if absEncounterRewardInt(item.ItemLevel-userLevel) <= 6 {
				near = append(near, item)
			}
		}
	}

	pool := near
	if len(pool) == 0 && len(withLevels) > 0 {
		bestDelta := absEncounterRewardInt(withLevels[0].ItemLevel - userLevel)
		for _, item := range withLevels[1:] {
			delta := absEncounterRewardInt(item.ItemLevel - userLevel)
			if delta < bestDelta {
				bestDelta = delta
			}
		}
		for _, item := range withLevels {
			if absEncounterRewardInt(item.ItemLevel-userLevel) == bestDelta {
				pool = append(pool, item)
			}
		}
	}
	if len(pool) == 0 {
		pool = candidates
	}

	index := deterministicEncounterRewardIndex(seed, userLevel, len(pool), allowEquippable)
	chosen := pool[index]
	return &chosen
}

func deterministicEncounterRewardIndex(
	seed string,
	userLevel int,
	count int,
	allowEquippable bool,
) int {
	if count <= 1 {
		return 0
	}
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(strings.TrimSpace(seed)))
	_, _ = hasher.Write([]byte{0})
	if allowEquippable {
		_, _ = hasher.Write([]byte("equippable"))
	} else {
		_, _ = hasher.Write([]byte("consumable"))
	}
	_, _ = hasher.Write([]byte{0})
	_, _ = hasher.Write([]byte{byte(userLevel & 0xff), byte((userLevel >> 8) & 0xff)})
	return int(hasher.Sum64() % uint64(count))
}

func absEncounterRewardInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
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
		plan, itemByID, _, err := s.randomRewardPlanForUser(
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
		for _, reward := range monster.ItemRewards {
			if reward.InventoryItem.ID != 0 {
				itemByID[reward.InventoryItemID] = reward.InventoryItem
				continue
			}
			if reward.InventoryItemID <= 0 {
				continue
			}
			if _, ok := itemByID[reward.InventoryItemID]; ok {
				continue
			}
			item, err := s.dbClient.InventoryItem().FindInventoryItemByID(
				ctx,
				reward.InventoryItemID,
			)
			if err != nil {
				return err
			}
			if item != nil {
				itemByID[reward.InventoryItemID] = *item
			}
		}
		itemGrants := models.MergeRandomRewardItemGrants(
			plan.ItemGrants,
			scenarioRewardItemsToRandomRewardItemGrants(
				monsterRewardItemsToScenarioRewards(monster.ItemRewards),
			),
		)
		response.ItemRewards = randomRewardItemGrantsToMonsterItemRewards(
			itemGrants,
			itemByID,
			monster.ID,
		)
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
