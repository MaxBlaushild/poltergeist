package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestBuildTutorialMonsterEncounterCopiesTemplateEncounterRewards(t *testing.T) {
	userID := uuid.New()
	zoneID := uuid.New()
	template := &models.MonsterEncounter{
		Name:               "Crypt Stalker",
		Description:        "A tutorial encounter",
		ImageURL:           "encounter.png",
		ThumbnailURL:       "encounter-thumb.png",
		EncounterType:      models.MonsterEncounterTypeBoss,
		ScaleWithUserLevel: true,
		RewardMode:         models.RewardModeRandom,
		RandomRewardSize:   models.RandomRewardSizeLarge,
		RewardExperience:   42,
		RewardGold:         17,
		MaterialRewards: models.BaseMaterialRewards{
			{ResourceKey: models.BaseResourceStone, Amount: 2},
		},
		ItemRewards: models.MonsterEncounterRewardItems{
			{InventoryItemID: 1001, Quantity: 3},
		},
		Members: []models.MonsterEncounterMember{
			{
				Slot: 0,
				Monster: models.Monster{
					Name:             "Bone Wolf",
					RewardExperience: 11,
					RewardGold:       5,
					ItemRewards: []models.MonsterItemReward{
						{InventoryItemID: 2001, Quantity: 1},
					},
				},
			},
		},
	}

	encounter, members, monsters := buildTutorialMonsterEncounter(userID, zoneID, 40.0, -73.0, template, nil)

	if encounter == nil {
		t.Fatal("expected encounter to be built")
	}
	if encounter.EncounterType != template.EncounterType {
		t.Fatalf("expected encounter type %q, got %q", template.EncounterType, encounter.EncounterType)
	}
	if encounter.RewardMode != template.RewardMode {
		t.Fatalf("expected reward mode %q, got %q", template.RewardMode, encounter.RewardMode)
	}
	if encounter.RandomRewardSize != template.RandomRewardSize {
		t.Fatalf("expected reward size %q, got %q", template.RandomRewardSize, encounter.RandomRewardSize)
	}
	if encounter.RewardExperience != template.RewardExperience {
		t.Fatalf("expected reward experience %d, got %d", template.RewardExperience, encounter.RewardExperience)
	}
	if encounter.RewardGold != template.RewardGold {
		t.Fatalf("expected reward gold %d, got %d", template.RewardGold, encounter.RewardGold)
	}
	if len(encounter.ItemRewards) != 1 || encounter.ItemRewards[0].InventoryItemID != 1001 || encounter.ItemRewards[0].Quantity != 3 {
		t.Fatalf("expected encounter item rewards to be copied, got %+v", encounter.ItemRewards)
	}
	if len(encounter.MaterialRewards) != 1 || encounter.MaterialRewards[0].ResourceKey != models.BaseResourceStone || encounter.MaterialRewards[0].Amount != 2 {
		t.Fatalf("expected encounter material rewards to be copied, got %+v", encounter.MaterialRewards)
	}
	if len(members) != 1 || len(monsters) != 1 {
		t.Fatalf("expected one member and one monster, got %d members and %d monsters", len(members), len(monsters))
	}
	if monsters[0].RewardExperience != 11 || monsters[0].RewardGold != 5 {
		t.Fatalf("expected monster rewards to still be copied from template monster, got exp=%d gold=%d", monsters[0].RewardExperience, monsters[0].RewardGold)
	}
	if len(monsters[0].ItemRewards) != 1 || monsters[0].ItemRewards[0].InventoryItemID != 2001 {
		t.Fatalf("expected monster item rewards to be copied, got %+v", monsters[0].ItemRewards)
	}
}

func TestBuildTutorialMonsterEncounterPromotesConfiguredRewardsToEncounter(t *testing.T) {
	userID := uuid.New()
	zoneID := uuid.New()
	template := &models.MonsterEncounter{
		RewardMode:       models.RewardModeRandom,
		RandomRewardSize: models.RandomRewardSizeLarge,
		RewardExperience: 42,
		RewardGold:       17,
		MaterialRewards: models.BaseMaterialRewards{
			{ResourceKey: models.BaseResourceStone, Amount: 2},
		},
		ItemRewards: models.MonsterEncounterRewardItems{
			{InventoryItemID: 1001, Quantity: 3},
		},
		Members: []models.MonsterEncounterMember{
			{
				Slot: 0,
				Monster: models.Monster{
					Name:             "Bone Wolf",
					RewardExperience: 11,
					RewardGold:       5,
				},
			},
			{
				Slot: 1,
				Monster: models.Monster{
					Name:             "Bone Bat",
					RewardExperience: 7,
					RewardGold:       2,
					ItemRewards: []models.MonsterItemReward{
						{InventoryItemID: 3001, Quantity: 1},
					},
				},
			},
		},
	}
	config := &models.TutorialConfig{
		MonsterRewardExperience: 90,
		MonsterRewardGold:       33,
		MonsterItemRewards: []models.TutorialItemReward{
			{InventoryItemID: 4001, Quantity: 2},
		},
	}

	encounter, _, monsters := buildTutorialMonsterEncounter(userID, zoneID, 40.0, -73.0, template, config)

	if encounter.RewardMode != models.RewardModeExplicit {
		t.Fatalf("expected configured tutorial rewards to force explicit encounter rewards, got %q", encounter.RewardMode)
	}
	if encounter.RewardExperience != 90 || encounter.RewardGold != 33 {
		t.Fatalf("expected encounter reward override to be applied, got exp=%d gold=%d", encounter.RewardExperience, encounter.RewardGold)
	}
	if len(encounter.ItemRewards) != 1 || encounter.ItemRewards[0].InventoryItemID != 4001 || encounter.ItemRewards[0].Quantity != 2 {
		t.Fatalf("expected configured tutorial encounter item rewards, got %+v", encounter.ItemRewards)
	}
	if len(encounter.MaterialRewards) != 1 || encounter.MaterialRewards[0].ResourceKey != models.BaseResourceStone {
		t.Fatalf("expected template material rewards to be preserved, got %+v", encounter.MaterialRewards)
	}
	if len(monsters) != 2 {
		t.Fatalf("expected two monsters, got %d", len(monsters))
	}
	if monsters[0].RewardExperience != 90 || monsters[0].RewardGold != 33 {
		t.Fatalf("expected primary monster to mirror configured rewards, got exp=%d gold=%d", monsters[0].RewardExperience, monsters[0].RewardGold)
	}
	if len(monsters[0].ItemRewards) != 1 || monsters[0].ItemRewards[0].InventoryItemID != 4001 {
		t.Fatalf("expected primary monster configured item rewards, got %+v", monsters[0].ItemRewards)
	}
	if monsters[1].RewardExperience != 0 || monsters[1].RewardGold != 0 || len(monsters[1].ItemRewards) != 0 {
		t.Fatalf("expected non-primary monster rewards to be cleared, got exp=%d gold=%d items=%+v", monsters[1].RewardExperience, monsters[1].RewardGold, monsters[1].ItemRewards)
	}
}
