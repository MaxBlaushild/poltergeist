package server

import (
	"math/rand"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

type treasureChestLockDistributionRequest struct {
	UnlockedPercentage    int `json:"unlockedPercentage"`
	Tier1To10Percentage   int `json:"tier1To10Percentage"`
	Tier11To25Percentage  int `json:"tier11To25Percentage"`
	Tier26To50Percentage  int `json:"tier26To50Percentage"`
	Tier51To75Percentage  int `json:"tier51To75Percentage"`
	Tier76To100Percentage int `json:"tier76To100Percentage"`
}

func treasureChestRewardSizeForUnlockTier(tier *int) models.RandomRewardSize {
	if tier == nil || *tier <= 25 {
		return models.RandomRewardSizeSmall
	}
	if *tier <= 50 {
		return models.RandomRewardSizeMedium
	}
	return models.RandomRewardSizeLarge
}

func allocateTreasureChestDistributionCounts(total int, percentages []int) []int {
	counts := make([]int, len(percentages))
	if total <= 0 || len(percentages) == 0 {
		return counts
	}

	type remainderEntry struct {
		index     int
		remainder int
	}

	remainders := make([]remainderEntry, len(percentages))
	assigned := 0
	for i, percentage := range percentages {
		numerator := total * percentage
		counts[i] = numerator / 100
		remainders[i] = remainderEntry{
			index:     i,
			remainder: numerator % 100,
		}
		assigned += counts[i]
	}

	for assigned < total {
		bestIndex := -1
		bestRemainder := -1
		for _, entry := range remainders {
			if entry.remainder > bestRemainder {
				bestRemainder = entry.remainder
				bestIndex = entry.index
			}
		}
		if bestIndex < 0 {
			bestIndex = 0
		}
		counts[bestIndex]++
		remainders[bestIndex].remainder = -1
		assigned++
	}

	return counts
}

func treasureChestUnlockTierPtr(v int) *int {
	return &v
}

func randomTreasureChestUnlockStrength(min int, max int) *int {
	if max < min {
		max = min
	}
	value := min
	if max > min {
		value = rand.Intn(max-min+1) + min
	}
	return treasureChestUnlockTierPtr(value)
}

func (s *server) reconfigureTreasureChestLockDistribution(ctx *gin.Context) {
	var requestBody treasureChestLockDistributionRequest
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	percentages := []int{
		requestBody.UnlockedPercentage,
		requestBody.Tier1To10Percentage,
		requestBody.Tier11To25Percentage,
		requestBody.Tier26To50Percentage,
		requestBody.Tier51To75Percentage,
		requestBody.Tier76To100Percentage,
	}
	totalPercentage := 0
	for _, percentage := range percentages {
		if percentage < 0 || percentage > 100 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "percentages must each be between 0 and 100"})
			return
		}
		totalPercentage += percentage
	}
	if totalPercentage != 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "percentages must sum to 100"})
		return
	}

	allChests, err := s.dbClient.TreasureChest().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch treasure chests: " + err.Error()})
		return
	}

	activeChests := make([]models.TreasureChest, 0, len(allChests))
	for _, chest := range allChests {
		if chest.Invalidated {
			continue
		}
		activeChests = append(activeChests, chest)
	}

	counts := allocateTreasureChestDistributionCounts(len(activeChests), percentages)
	rand.Shuffle(len(activeChests), func(i, j int) {
		activeChests[i], activeChests[j] = activeChests[j], activeChests[i]
	})

	assignedTiers := make([]*int, 0, len(activeChests))
	for i := 0; i < counts[0]; i++ {
		assignedTiers = append(assignedTiers, nil)
	}
	for i := 0; i < counts[1]; i++ {
		assignedTiers = append(assignedTiers, randomTreasureChestUnlockStrength(1, 10))
	}
	for i := 0; i < counts[2]; i++ {
		assignedTiers = append(assignedTiers, randomTreasureChestUnlockStrength(11, 25))
	}
	for i := 0; i < counts[3]; i++ {
		assignedTiers = append(assignedTiers, randomTreasureChestUnlockStrength(26, 50))
	}
	for i := 0; i < counts[4]; i++ {
		assignedTiers = append(assignedTiers, randomTreasureChestUnlockStrength(51, 75))
	}
	for i := 0; i < counts[5]; i++ {
		assignedTiers = append(assignedTiers, randomTreasureChestUnlockStrength(76, 100))
	}

	for i, chest := range activeChests {
		var assignedTier *int
		if i < len(assignedTiers) {
			assignedTier = assignedTiers[i]
		}
		updates := &models.TreasureChest{
			ID:               chest.ID,
			ZoneID:           chest.ZoneID,
			Latitude:         chest.Latitude,
			Longitude:        chest.Longitude,
			UnlockTier:       assignedTier,
			RewardMode:       chest.RewardMode,
			RandomRewardSize: treasureChestRewardSizeForUnlockTier(assignedTier),
			RewardExperience: chest.RewardExperience,
			Gold:             chest.Gold,
			Invalidated:      chest.Invalidated,
		}
		if err := s.dbClient.TreasureChest().Update(ctx, chest.ID, updates); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update treasure chest distribution: " + err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "treasure chest lock tier distribution updated",
		"updatedCount": len(activeChests),
		"counts": gin.H{
			"unlocked":    counts[0],
			"tier1To10":   counts[1],
			"tier11To25":  counts[2],
			"tier26To50":  counts[3],
			"tier51To75":  counts[4],
			"tier76To100": counts[5],
		},
		"rewardSizes": gin.H{
			"small":  counts[0] + counts[1] + counts[2],
			"medium": counts[3],
			"large":  counts[4] + counts[5],
		},
	})
}
