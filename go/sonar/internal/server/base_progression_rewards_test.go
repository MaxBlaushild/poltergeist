package server

import (
	"fmt"
	"testing"
)

func TestBuildRandomBaseMaterialRewardsAlwaysAwardsAtLeastOneMaterial(t *testing.T) {
	for i := 0; i < 256; i++ {
		rewards := buildRandomBaseMaterialRewards(fmt.Sprintf("seed-%d", i))
		if len(rewards) == 0 {
			t.Fatalf("expected at least one material reward for seed %d", i)
		}
		if len(rewards) > 2 {
			t.Fatalf("expected at most two material rewards for seed %d, got %d", i, len(rewards))
		}
		seen := map[string]struct{}{}
		for _, reward := range rewards {
			if reward.ResourceKey == "" {
				t.Fatalf("expected resource key for seed %d", i)
			}
			if reward.Amount < 1 || reward.Amount > 3 {
				t.Fatalf("expected reward amount between 1 and 3 for seed %d, got %d", i, reward.Amount)
			}
			if _, exists := seen[string(reward.ResourceKey)]; exists {
				t.Fatalf("expected distinct material kinds for seed %d", i)
			}
			seen[string(reward.ResourceKey)] = struct{}{}
		}
	}
}

func TestBuildRandomBaseMaterialRewardsCanAwardSecondMaterial(t *testing.T) {
	foundSecondMaterial := false
	for i := 0; i < 512; i++ {
		rewards := buildRandomBaseMaterialRewards(fmt.Sprintf("chance-seed-%d", i))
		if len(rewards) == 2 {
			foundSecondMaterial = true
			break
		}
	}
	if !foundSecondMaterial {
		t.Fatal("expected some seeds to award a second material")
	}
}
