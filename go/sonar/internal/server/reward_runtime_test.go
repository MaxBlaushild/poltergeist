package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestEnsureMonsterEncounterRandomRewardHasItemPreservesExistingGrants(t *testing.T) {
	plan := models.RandomRewardPlan{
		Experience: 100,
		Gold:       25,
		ItemGrants: []models.RandomRewardItemGrant{
			{InventoryItemID: 99, Quantity: 2},
		},
	}

	got := ensureMonsterEncounterRandomRewardHasItem(plan, map[int]models.InventoryItem{}, 25, "seed")
	if len(got.ItemGrants) != 1 {
		t.Fatalf("expected existing item grants to remain untouched, got %d", len(got.ItemGrants))
	}
	if got.ItemGrants[0].InventoryItemID != 99 || got.ItemGrants[0].Quantity != 2 {
		t.Fatalf("expected existing grant to be preserved, got %+v", got.ItemGrants[0])
	}
}

func TestEnsureMonsterEncounterRandomRewardHasItemAddsConsumableFallback(t *testing.T) {
	itemByID := map[int]models.InventoryItem{
		1: {ID: 1, Name: "Potion", ItemLevel: 20, RarityTier: "Common"},
		2: {ID: 2, Name: "Elixir", ItemLevel: 40, RarityTier: "Common"},
		3: {
			ID:         3,
			Name:       "Sword",
			ItemLevel:  20,
			RarityTier: "Common",
			EquipSlot:  stringPtrRewardRuntime("dominant_hand"),
		},
	}

	got := ensureMonsterEncounterRandomRewardHasItem(models.RandomRewardPlan{}, itemByID, 22, "encounter-seed")
	if len(got.ItemGrants) != 1 {
		t.Fatalf("expected one fallback item grant, got %d", len(got.ItemGrants))
	}
	if got.ItemGrants[0].InventoryItemID != 1 {
		t.Fatalf("expected consumable fallback item 1, got %d", got.ItemGrants[0].InventoryItemID)
	}
	if got.ItemGrants[0].Quantity != 1 {
		t.Fatalf("expected fallback quantity 1, got %d", got.ItemGrants[0].Quantity)
	}
}

func TestEnsureMonsterEncounterRandomRewardHasItemFallsBackToEquippable(t *testing.T) {
	itemByID := map[int]models.InventoryItem{
		7: {
			ID:         7,
			Name:       "Bronze Sword",
			ItemLevel:  18,
			RarityTier: "Common",
			EquipSlot:  stringPtrRewardRuntime("dominant_hand"),
		},
	}

	got := ensureMonsterEncounterRandomRewardHasItem(models.RandomRewardPlan{}, itemByID, 20, "encounter-seed")
	if len(got.ItemGrants) != 1 {
		t.Fatalf("expected one fallback item grant, got %d", len(got.ItemGrants))
	}
	if got.ItemGrants[0].InventoryItemID != 7 {
		t.Fatalf("expected equippable fallback item 7, got %d", got.ItemGrants[0].InventoryItemID)
	}
}

func TestMergeScenarioRewardItemsMergesDuplicateItemIDs(t *testing.T) {
	got := mergeScenarioRewardItems(
		[]scenarioRewardItem{
			{InventoryItemID: 5, Quantity: 1},
			{InventoryItemID: 2, Quantity: 2},
		},
		[]scenarioRewardItem{
			{InventoryItemID: 5, Quantity: 3},
			{InventoryItemID: 0, Quantity: 1},
		},
	)

	if len(got) != 2 {
		t.Fatalf("expected 2 merged reward items, got %d", len(got))
	}
	if got[0].InventoryItemID != 2 || got[0].Quantity != 2 {
		t.Fatalf("expected first merged reward to be item 2 x2, got %+v", got[0])
	}
	if got[1].InventoryItemID != 5 || got[1].Quantity != 4 {
		t.Fatalf("expected second merged reward to be item 5 x4, got %+v", got[1])
	}
}

func stringPtrRewardRuntime(value string) *string {
	return &value
}
