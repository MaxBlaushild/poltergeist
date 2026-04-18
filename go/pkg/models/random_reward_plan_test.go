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

func TestMergeRandomRewardItemGrantsMergesAndSorts(t *testing.T) {
	grants := MergeRandomRewardItemGrants(
		[]RandomRewardItemGrant{
			{InventoryItemID: 5, Quantity: 1},
			{InventoryItemID: 2, Quantity: 3},
		},
		[]RandomRewardItemGrant{
			{InventoryItemID: 5, Quantity: 2},
			{InventoryItemID: 0, Quantity: 1},
			{InventoryItemID: 7, Quantity: -1},
		},
	)

	if len(grants) != 2 {
		t.Fatalf("expected 2 merged grants, got %d", len(grants))
	}
	if grants[0].InventoryItemID != 2 || grants[0].Quantity != 3 {
		t.Fatalf("expected first merged grant to be item 2 x3, got %+v", grants[0])
	}
	if grants[1].InventoryItemID != 5 || grants[1].Quantity != 3 {
		t.Fatalf("expected second merged grant to be item 5 x3, got %+v", grants[1])
	}
}

func stringPtr(value string) *string {
	return &value
}
