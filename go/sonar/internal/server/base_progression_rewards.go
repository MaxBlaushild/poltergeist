package server

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
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

func resolveBaseMaterialRewards(
	rewardMode models.RewardMode,
	explicit models.BaseMaterialRewards,
	randomSeed string,
) []models.BaseResourceDelta {
	if models.NormalizeRewardMode(string(rewardMode)) == models.RewardModeRandom {
		return buildRandomBaseMaterialRewards(randomSeed)
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
