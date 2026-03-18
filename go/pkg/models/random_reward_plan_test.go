package models

import "testing"

func TestBuildRandomRewardPlanExcludesNotDroppableItems(t *testing.T) {
	items := []InventoryItem{
		{
			ID:         1,
			Name:       "Quest Trophy",
			RarityTier: "Not Droppable",
			ItemLevel:  25,
		},
		{
			ID:         2,
			Name:       "Potion",
			RarityTier: "Common",
			ItemLevel:  25,
		},
		{
			ID:         3,
			Name:       "Sword",
			RarityTier: "Not Droppable",
			ItemLevel:  25,
			EquipSlot:  stringPtr("dominant_hand"),
		},
		{
			ID:         4,
			Name:       "Helm",
			RarityTier: "Uncommon",
			ItemLevel:  25,
			EquipSlot:  stringPtr("hat"),
		},
	}

	plan := BuildRandomRewardPlan(25, RandomRewardSizeLarge, "reward-seed", items)
	if len(plan.ItemGrants) == 0 {
		t.Fatalf("expected at least one item grant in large reward plan")
	}

	for _, grant := range plan.ItemGrants {
		if grant.InventoryItemID == 1 || grant.InventoryItemID == 3 {
			t.Fatalf("expected not droppable items to be excluded, got item ID %d", grant.InventoryItemID)
		}
	}
}

func stringPtr(value string) *string {
	return &value
}
