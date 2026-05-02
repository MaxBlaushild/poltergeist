package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestBuildRandomBaseMaterialRewardsForContextPrefersMonsterPartsForCombat(t *testing.T) {
	rewards := buildRandomBaseMaterialRewardsForContext(
		"monster-seed",
		&models.RandomRewardContext{
			ContentKind: models.RandomRewardContentMonsterEncounter,
		},
	)
	if len(rewards) == 0 {
		t.Fatalf("expected at least one base material reward")
	}
	if rewards[0].ResourceKey != models.BaseResourceMonsterParts {
		t.Fatalf("expected combat rewards to start with monster_parts, got %q", rewards[0].ResourceKey)
	}
}

func TestBuildRandomBaseMaterialRewardsForContextPrefersRelicsForLore(t *testing.T) {
	rewards := buildRandomBaseMaterialRewardsForContext(
		"lore-seed",
		&models.RandomRewardContext{
			ContentKind: models.RandomRewardContentExposition,
			ZoneKind:    "haunted-archive",
		},
	)
	if len(rewards) == 0 {
		t.Fatalf("expected at least one base material reward")
	}
	if rewards[0].ResourceKey != models.BaseResourceRelicShards {
		t.Fatalf("expected exposition rewards to start with relic_shards, got %q", rewards[0].ResourceKey)
	}
}

func TestBuildRandomBaseMaterialRewardsForContextUsesPreferredMaterialKeys(t *testing.T) {
	rewards := buildRandomBaseMaterialRewardsForContext(
		"profile-seed",
		&models.RandomRewardContext{
			ContentKind:           models.RandomRewardContentMonsterEncounter,
			PreferredMaterialKeys: []models.BaseResourceKey{models.BaseResourceTimber, models.BaseResourceStone},
		},
	)
	if len(rewards) == 0 {
		t.Fatalf("expected at least one base material reward")
	}
	if rewards[0].ResourceKey != models.BaseResourceTimber {
		t.Fatalf("expected preferred material order to win, got %q", rewards[0].ResourceKey)
	}
}
