package server

import (
	"strings"
	"testing"
)

func TestInventoryCoreSeedPackRequestsAreUniqueAndLarge(t *testing.T) {
	requests := inventoryCoreSeedPackRequests()
	if len(requests) < 250 {
		t.Fatalf("expected a large inventory seed pack, got %d items", len(requests))
	}

	seen := map[string]struct{}{}
	equipmentCount := 0
	materialCount := 0
	utilityCount := 0

	for _, seed := range requests {
		name := strings.TrimSpace(seed.Request.Name)
		if name == "" {
			t.Fatalf("found seed item with empty name")
		}
		lower := strings.ToLower(name)
		if strings.Contains(lower, "charm") || strings.Contains(lower, "potion") || strings.Contains(lower, "tome") {
			t.Fatalf("seed pack should skip charms, potions, and tomes, got %q", name)
		}
		if _, exists := seen[lower]; exists {
			t.Fatalf("duplicate inventory seed name %q", name)
		}
		seen[lower] = struct{}{}

		switch seed.Category {
		case "equipment":
			equipmentCount++
			if seed.Request.EquipSlot == nil || strings.TrimSpace(*seed.Request.EquipSlot) == "" {
				t.Fatalf("equipment seed %q is missing equip slot", name)
			}
		case "material":
			materialCount++
			if seed.Request.EquipSlot != nil {
				t.Fatalf("material seed %q should not have an equip slot", name)
			}
		case "utility":
			utilityCount++
			if seed.Request.EquipSlot != nil {
				t.Fatalf("utility seed %q should not have an equip slot", name)
			}
		default:
			t.Fatalf("unexpected category %q for %q", seed.Category, name)
		}
	}

	expectedEquipment := len(inventorySeedSetConfigs()) * len(inventorySetAllEquippableSlots())
	if equipmentCount != expectedEquipment {
		t.Fatalf("expected %d equipment items, got %d", expectedEquipment, equipmentCount)
	}
	if materialCount < 70 {
		t.Fatalf("expected at least 70 materials, got %d", materialCount)
	}
	if utilityCount < 10 {
		t.Fatalf("expected at least 10 utility items, got %d", utilityCount)
	}
}

func TestInventorySeedSetRequestsCoverAllSlots(t *testing.T) {
	configs := inventorySeedSetConfigs()
	if len(configs) == 0 {
		t.Fatal("expected at least one inventory seed set config")
	}

	requests := buildInventorySeedSetRequests(configs[0])
	expectedCount := len(inventorySetAllEquippableSlots())
	if len(requests) != expectedCount {
		t.Fatalf("expected %d slot requests, got %d", expectedCount, len(requests))
	}

	seenSlots := map[string]struct{}{}
	for _, seed := range requests {
		if seed.Category != "equipment" {
			t.Fatalf("expected equipment category, got %q", seed.Category)
		}
		if seed.Request.EquipSlot == nil {
			t.Fatalf("expected equip slot for %q", seed.Request.Name)
		}
		slot := strings.TrimSpace(*seed.Request.EquipSlot)
		if slot == "" {
			t.Fatalf("expected non-empty equip slot for %q", seed.Request.Name)
		}
		if _, exists := seenSlots[slot]; exists {
			t.Fatalf("duplicate slot %q in generated seed set", slot)
		}
		seenSlots[slot] = struct{}{}
	}

	for _, slot := range inventorySetAllEquippableSlots() {
		if _, exists := seenSlots[slot]; !exists {
			t.Fatalf("missing slot %q from generated seed set", slot)
		}
	}
}

func TestSeededUtilityRequestsUseSupportedMechanics(t *testing.T) {
	for _, spec := range inventorySeedUtilitySpecs() {
		request := seededUtilityRequest(spec)
		if request.EquipSlot != nil {
			t.Fatalf("utility item %q should not be equippable", request.Name)
		}
		if request.ConsumeCreateBase && request.UnlockLocksStrength != nil {
			t.Fatalf("utility item %q should not be both a base deed and a lock tool", request.Name)
		}
		if !request.ConsumeCreateBase && request.UnlockLocksStrength == nil {
			t.Fatalf("utility item %q should map to an existing supported utility mechanic", request.Name)
		}
	}
}
