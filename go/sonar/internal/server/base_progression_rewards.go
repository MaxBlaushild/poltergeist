package server

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

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

func baseRewardTierFromResolvedRewards(
	rewardMode models.RewardMode,
	rewardSize models.RandomRewardSize,
	rewardExperience int,
	rewardGold int,
	itemCount int,
) int {
	if models.NormalizeRewardMode(string(rewardMode)) == models.RewardModeRandom {
		switch models.NormalizeRandomRewardSize(string(rewardSize)) {
		case models.RandomRewardSizeLarge:
			return 3
		case models.RandomRewardSizeMedium:
			return 2
		default:
			return 1
		}
	}
	switch {
	case rewardExperience >= 600 || rewardGold >= 225 || itemCount >= 2:
		return 3
	case rewardExperience >= 250 || rewardGold >= 90 || itemCount >= 1:
		return 2
	default:
		return 1
	}
}

func baseResourceGrantsForScenario(
	rewardMode models.RewardMode,
	rewardSize models.RandomRewardSize,
	rewardExperience int,
	rewardGold int,
	itemCount int,
) []models.BaseResourceDelta {
	tier := baseRewardTierFromResolvedRewards(rewardMode, rewardSize, rewardExperience, rewardGold, itemCount)
	grants := []models.BaseResourceDelta{
		{ResourceKey: models.BaseResourceHerbs, Amount: tier * 2},
		{ResourceKey: models.BaseResourceArcaneDust, Amount: tier},
	}
	if tier >= 3 {
		grants = append(grants, models.BaseResourceDelta{
			ResourceKey: models.BaseResourceRelicShards,
			Amount:      1,
		})
	}
	return grants
}

func baseResourceGrantsForChallenge(
	rewardMode models.RewardMode,
	rewardSize models.RandomRewardSize,
	rewardExperience int,
	rewardGold int,
	itemCount int,
) []models.BaseResourceDelta {
	tier := baseRewardTierFromResolvedRewards(rewardMode, rewardSize, rewardExperience, rewardGold, itemCount)
	return []models.BaseResourceDelta{
		{ResourceKey: models.BaseResourceTimber, Amount: tier},
		{ResourceKey: models.BaseResourceArcaneDust, Amount: tier},
	}
}

func baseResourceGrantsForMonster(
	rewardMode models.RewardMode,
	rewardSize models.RandomRewardSize,
	rewardExperience int,
	rewardGold int,
	itemCount int,
) []models.BaseResourceDelta {
	tier := baseRewardTierFromResolvedRewards(rewardMode, rewardSize, rewardExperience, rewardGold, itemCount)
	return []models.BaseResourceDelta{
		{ResourceKey: models.BaseResourceMonsterParts, Amount: tier * 2},
		{ResourceKey: models.BaseResourceIron, Amount: tier},
	}
}

func baseResourceGrantsForTreasureChest(
	rewardMode models.RewardMode,
	rewardSize models.RandomRewardSize,
	rewardExperience int,
	rewardGold int,
	itemCount int,
	unlockTier *int,
) []models.BaseResourceDelta {
	tier := baseRewardTierFromResolvedRewards(rewardMode, rewardSize, rewardExperience, rewardGold, itemCount)
	if unlockTier != nil {
		switch {
		case *unlockTier >= 76:
			tier = max(tier, 3)
		case *unlockTier >= 26:
			tier = max(tier, 2)
		default:
			tier = max(tier, 1)
		}
	}
	grants := []models.BaseResourceDelta{
		{ResourceKey: models.BaseResourceTimber, Amount: tier},
		{ResourceKey: models.BaseResourceStone, Amount: tier},
	}
	if unlockTier != nil {
		grants = append(grants, models.BaseResourceDelta{
			ResourceKey: models.BaseResourceIron,
			Amount:      max(1, tier-1),
		})
	}
	return grants
}

func baseResourceGrantsForQuest(
	rewardMode models.RewardMode,
	rewardSize models.RandomRewardSize,
	rewardGold int,
	itemCount int,
	spellCount int,
) []models.BaseResourceDelta {
	tier := baseRewardTierFromResolvedRewards(rewardMode, rewardSize, 0, rewardGold, itemCount+spellCount)
	grants := []models.BaseResourceDelta{
		{ResourceKey: models.BaseResourceTimber, Amount: tier * 2},
		{ResourceKey: models.BaseResourceStone, Amount: tier * 2},
	}
	if tier >= 2 {
		grants = append(grants, models.BaseResourceDelta{
			ResourceKey: models.BaseResourceRelicShards,
			Amount:      1,
		})
	}
	return grants
}
