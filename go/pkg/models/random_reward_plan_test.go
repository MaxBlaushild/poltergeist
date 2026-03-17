package models

import "testing"

func TestFilterRewardItemsExcludesArchived(t *testing.T) {
	equipSlot := "hat"
	items := []InventoryItem{
		{ID: 1, ItemLevel: 10, Archived: false},
		{ID: 2, ItemLevel: 10, Archived: true},
		{ID: 3, ItemLevel: 10, IsCaptureType: true},
		{ID: 4, ItemLevel: 10, EquipSlot: &equipSlot},
	}

	consumables := filterRewardItems(items, 10, false)
	if len(consumables) != 1 || consumables[0].ID != 1 {
		t.Fatalf("expected only active non-capture consumable item, got %+v", consumables)
	}

	equippables := filterRewardItems(items, 10, true)
	if len(equippables) != 1 || equippables[0].ID != 4 {
		t.Fatalf("expected only active equippable item, got %+v", equippables)
	}
}
