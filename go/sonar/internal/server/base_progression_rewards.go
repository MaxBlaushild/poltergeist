package server

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/sonar/internal/rewards"
	"github.com/google/uuid"
)

type baseMaterialRewardPayload struct {
	ResourceKey string `json:"resourceKey"`
	Amount      int    `json:"amount"`
}

var baseMaterialRewardKeyOrder = []models.BaseResourceKey{
	models.BaseResourceTimber,
	models.BaseResourceStone,
	models.BaseResourceIron,
	models.BaseResourceHerbs,
	models.BaseResourceMonsterParts,
	models.BaseResourceArcaneDust,
	models.BaseResourceRelicShards,
}

func serializeBaseResourceDeltas(deltas []models.BaseResourceDelta) []map[string]interface{} {
	response := make([]map[string]interface{}, 0, len(deltas))
	for _, delta := range deltas {
		if delta.ResourceKey == "" || delta.Amount <= 0 {
			continue
		}
		response = append(response, map[string]interface{}{
			"resourceKey": delta.ResourceKey,
			"amount":      delta.Amount,
		})
	}
	return response
}

func parseBaseMaterialRewards(
	input []baseMaterialRewardPayload,
	fieldName string,
) (models.BaseMaterialRewards, error) {
	normalized := make([]models.BaseResourceDelta, 0, len(input))
	for idx, entry := range input {
		resourceKey := models.NormalizeBaseResourceKey(entry.ResourceKey)
		if resourceKey == "" && strings.TrimSpace(entry.ResourceKey) == "" && entry.Amount == 0 {
			continue
		}
		if resourceKey == "" {
			return nil, fmt.Errorf("%s[%d].resourceKey is invalid", fieldName, idx)
		}
		if entry.Amount <= 0 {
			return nil, fmt.Errorf("%s[%d].amount must be positive", fieldName, idx)
		}
		normalized = append(normalized, models.BaseResourceDelta{
			ResourceKey: resourceKey,
			Amount:      entry.Amount,
		})
	}
	return models.BaseMaterialRewards(normalizeBaseMaterialRewards(normalized)), nil
}

func normalizeBaseMaterialRewards(input []models.BaseResourceDelta) []models.BaseResourceDelta {
	if len(input) == 0 {
		return []models.BaseResourceDelta{}
	}
	totals := map[models.BaseResourceKey]int{}
	for _, delta := range input {
		resourceKey := models.NormalizeBaseResourceKey(string(delta.ResourceKey))
		if resourceKey == "" || delta.Amount <= 0 {
			continue
		}
		totals[resourceKey] += delta.Amount
	}
	if len(totals) == 0 {
		return []models.BaseResourceDelta{}
	}

	response := make([]models.BaseResourceDelta, 0, len(totals))
	for _, resourceKey := range baseMaterialRewardKeyOrder {
		if totals[resourceKey] <= 0 {
			continue
		}
		response = append(response, models.BaseResourceDelta{
			ResourceKey: resourceKey,
			Amount:      totals[resourceKey],
		})
		delete(totals, resourceKey)
	}
	if len(totals) > 0 {
		extraKeys := make([]string, 0, len(totals))
		for resourceKey := range totals {
			extraKeys = append(extraKeys, string(resourceKey))
		}
		sort.Strings(extraKeys)
		for _, rawKey := range extraKeys {
			resourceKey := models.BaseResourceKey(rawKey)
			response = append(response, models.BaseResourceDelta{
				ResourceKey: resourceKey,
				Amount:      totals[resourceKey],
			})
		}
	}
	return response
}

func randomBaseMaterialRewardSeed(seed string) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(strings.TrimSpace(seed)))
	value := int64(hasher.Sum64())
	if value == 0 {
		return 1
	}
	return value
}

func buildRandomBaseMaterialRewards(seed string) []models.BaseResourceDelta {
	rng := rand.New(rand.NewSource(randomBaseMaterialRewardSeed(seed)))
	firstIndex := rng.Intn(len(baseMaterialRewardKeyOrder))
	rewards := []models.BaseResourceDelta{{
		ResourceKey: baseMaterialRewardKeyOrder[firstIndex],
		Amount:      1 + rng.Intn(3),
	}}
	if len(baseMaterialRewardKeyOrder) <= 1 || rng.Float64() >= 0.5 {
		return rewards
	}
	secondIndex := rng.Intn(len(baseMaterialRewardKeyOrder) - 1)
	if secondIndex >= firstIndex {
		secondIndex++
	}
	rewards = append(rewards, models.BaseResourceDelta{
		ResourceKey: baseMaterialRewardKeyOrder[secondIndex],
		Amount:      1 + rng.Intn(3),
	})
	return rewards
}

func buildRandomBaseMaterialRewardsForContext(
	seed string,
	rewardContext *models.RandomRewardContext,
) []models.BaseResourceDelta {
	if rewardContext == nil {
		return buildRandomBaseMaterialRewards(seed)
	}

	priorities := preferredBaseResourceOrderForContext(rewardContext)
	if len(priorities) == 0 {
		return buildRandomBaseMaterialRewards(seed)
	}

	rng := rand.New(rand.NewSource(randomBaseMaterialRewardSeed(seed)))
	rewards := []models.BaseResourceDelta{{
		ResourceKey: priorities[0],
		Amount:      1 + rng.Intn(3),
	}}

	if len(priorities) == 1 || rng.Float64() >= 0.6 {
		return rewards
	}

	window := min(4, len(priorities))
	secondIndex := 1 + rng.Intn(window-1)
	rewards = append(rewards, models.BaseResourceDelta{
		ResourceKey: priorities[secondIndex],
		Amount:      1 + rng.Intn(2),
	})
	return rewards
}

func preferredBaseResourceOrderForContext(
	rewardContext *models.RandomRewardContext,
) []models.BaseResourceKey {
	if rewardContext == nil {
		return append([]models.BaseResourceKey{}, baseMaterialRewardKeyOrder...)
	}

	order := make([]models.BaseResourceKey, 0, len(baseMaterialRewardKeyOrder))
	seen := map[models.BaseResourceKey]struct{}{}
	add := func(resourceKey models.BaseResourceKey) {
		resourceKey = models.NormalizeBaseResourceKey(string(resourceKey))
		if resourceKey == "" {
			return
		}
		if _, exists := seen[resourceKey]; exists {
			return
		}
		seen[resourceKey] = struct{}{}
		order = append(order, resourceKey)
	}
	addMany := func(resourceKeys ...models.BaseResourceKey) {
		for _, resourceKey := range resourceKeys {
			add(resourceKey)
		}
	}

	if len(rewardContext.PreferredMaterialKeys) > 0 {
		addMany(rewardContext.PreferredMaterialKeys...)
	} else {
		switch rewardContext.ContentKind {
		case models.RandomRewardContentMonster, models.RandomRewardContentMonsterEncounter:
			addMany(models.BaseResourceMonsterParts, models.BaseResourceIron, models.BaseResourceHerbs)
		case models.RandomRewardContentExposition:
			addMany(models.BaseResourceRelicShards, models.BaseResourceArcaneDust, models.BaseResourceHerbs)
		case models.RandomRewardContentPointOfInterest:
			switch rewardContext.PointOfInterestCategory {
			case models.PointOfInterestMarkerCategoryArchive, models.PointOfInterestMarkerCategoryMuseum, models.PointOfInterestMarkerCategoryLandmark, models.PointOfInterestMarkerCategoryCivic:
				addMany(models.BaseResourceRelicShards, models.BaseResourceArcaneDust, models.BaseResourceHerbs)
			case models.PointOfInterestMarkerCategoryMarket, models.PointOfInterestMarkerCategoryCoffeehouse, models.PointOfInterestMarkerCategoryTavern, models.PointOfInterestMarkerCategoryEatery:
				addMany(models.BaseResourceHerbs, models.BaseResourceIron, models.BaseResourceTimber)
			case models.PointOfInterestMarkerCategoryPark, models.PointOfInterestMarkerCategoryWaterfront:
				addMany(models.BaseResourceHerbs, models.BaseResourceTimber, models.BaseResourceStone)
			case models.PointOfInterestMarkerCategoryArena:
				addMany(models.BaseResourceIron, models.BaseResourceMonsterParts, models.BaseResourceHerbs)
			}
		case models.RandomRewardContentTreasureChest:
			addMany(models.BaseResourceRelicShards, models.BaseResourceArcaneDust)
		}
	}

	for _, tag := range rewardContext.PreferredRewardTags() {
		switch strings.ToLower(strings.TrimSpace(tag)) {
		case "martial", "frontline", "defender", "hunter":
			addMany(models.BaseResourceMonsterParts, models.BaseResourceIron, models.BaseResourceStone)
		case "rogue", "skirmisher", "street", "scout":
			addMany(models.BaseResourceTimber, models.BaseResourceHerbs, models.BaseResourceIron)
		case "scholar", "arcane", "seer", "ritual", "relic":
			addMany(models.BaseResourceArcaneDust, models.BaseResourceRelicShards, models.BaseResourceHerbs)
		case "social", "leader", "broker", "court", "guide":
			addMany(models.BaseResourceRelicShards, models.BaseResourceHerbs, models.BaseResourceTimber)
		case "nature", "wild":
			addMany(models.BaseResourceHerbs, models.BaseResourceTimber, models.BaseResourceStone)
		case "fire", "ice", "lightning", "storm", "shadow", "holy":
			addMany(models.BaseResourceArcaneDust, models.BaseResourceMonsterParts)
		case "poison":
			addMany(models.BaseResourceHerbs, models.BaseResourceMonsterParts)
		}
	}

	zoneKind := strings.ToLower(strings.TrimSpace(rewardContext.ZoneKind))
	switch {
	case strings.Contains(zoneKind, "park"), strings.Contains(zoneKind, "garden"), strings.Contains(zoneKind, "wild"), strings.Contains(zoneKind, "meadow"):
		addMany(models.BaseResourceHerbs, models.BaseResourceTimber)
	case strings.Contains(zoneKind, "water"), strings.Contains(zoneKind, "harbor"), strings.Contains(zoneKind, "coast"), strings.Contains(zoneKind, "river"):
		addMany(models.BaseResourceTimber, models.BaseResourceHerbs, models.BaseResourceRelicShards)
	case strings.Contains(zoneKind, "industrial"), strings.Contains(zoneKind, "forge"), strings.Contains(zoneKind, "factory"), strings.Contains(zoneKind, "rail"):
		addMany(models.BaseResourceIron, models.BaseResourceStone)
	case strings.Contains(zoneKind, "haunted"), strings.Contains(zoneKind, "occult"), strings.Contains(zoneKind, "grave"), strings.Contains(zoneKind, "ruin"):
		addMany(models.BaseResourceArcaneDust, models.BaseResourceRelicShards)
	}

	for _, resourceKey := range baseMaterialRewardKeyOrder {
		add(resourceKey)
	}
	return order
}

func resolveBaseMaterialRewards(
	rewardMode models.RewardMode,
	explicit models.BaseMaterialRewards,
	randomSeed string,
) []models.BaseResourceDelta {
	return resolveBaseMaterialRewardsForContext(rewardMode, explicit, randomSeed, nil)
}

func (s *server) resolveBaseMaterialRewardsForUserContext(
	ctx context.Context,
	rewardMode models.RewardMode,
	explicit models.BaseMaterialRewards,
	randomSeed string,
	rewardContext *models.RandomRewardContext,
) (models.BaseMaterialRewards, error) {
	rewardContext, err := rewards.ApplyDefaultRewardProfiles(ctx, s.dbClient, rewardContext)
	if err != nil {
		return nil, err
	}
	return resolveBaseMaterialRewardsForContext(rewardMode, explicit, randomSeed, rewardContext), nil
}

func resolveBaseMaterialRewardsForContext(
	rewardMode models.RewardMode,
	explicit models.BaseMaterialRewards,
	randomSeed string,
	rewardContext *models.RandomRewardContext,
) []models.BaseResourceDelta {
	if models.NormalizeRewardMode(string(rewardMode)) == models.RewardModeRandom {
		return buildRandomBaseMaterialRewardsForContext(randomSeed, rewardContext)
	}
	return normalizeBaseMaterialRewards(explicit)
}

func (s *server) awardBaseResourcesToUser(
	ctx context.Context,
	userID uuid.UUID,
	deltas []models.BaseResourceDelta,
	sourceType string,
	sourceID *uuid.UUID,
) ([]models.BaseResourceDelta, error) {
	if len(deltas) == 0 {
		return []models.BaseResourceDelta{}, nil
	}
	if err := s.dbClient.BaseResourceBalance().GrantToUser(ctx, userID, deltas, sourceType, sourceID, nil); err != nil {
		return nil, err
	}
	return deltas, nil
}

func (s *server) awardBaseResourcesToParticipants(
	ctx context.Context,
	participantIDs []uuid.UUID,
	submitterID uuid.UUID,
	deltas []models.BaseResourceDelta,
	sourceType string,
	sourceID *uuid.UUID,
) ([]models.BaseResourceDelta, error) {
	if len(deltas) == 0 {
		return []models.BaseResourceDelta{}, nil
	}
	submitterAwarded := []models.BaseResourceDelta{}
	for _, participantID := range participantIDs {
		awarded, err := s.awardBaseResourcesToUser(ctx, participantID, deltas, sourceType, sourceID)
		if err != nil {
			return nil, err
		}
		if participantID == submitterID {
			submitterAwarded = awarded
		}
	}
	return submitterAwarded, nil
}
