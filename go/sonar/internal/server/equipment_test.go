package server

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestCanEquipInventoryItemToSlot(t *testing.T) {
	dominantHand := string(models.EquipmentSlotDominantHand)
	offHand := string(models.EquipmentSlotOffHand)
	weapon := string(models.HandItemCategoryWeapon)
	shield := string(models.HandItemCategoryShield)
	oneHanded := string(models.HandednessOneHanded)
	twoHanded := string(models.HandednessTwoHanded)

	makeItem := func(equipSlot string, category string, handedness string) *models.InventoryItem {
		return &models.InventoryItem{
			EquipSlot:        &equipSlot,
			HandItemCategory: &category,
			Handedness:       &handedness,
		}
	}

	testCases := []struct {
		name          string
		item          *models.InventoryItem
		requestedSlot string
		expected      bool
	}{
		{
			name:          "dominant hand weapon can stay in dominant hand",
			item:          makeItem(dominantHand, weapon, oneHanded),
			requestedSlot: dominantHand,
			expected:      true,
		},
		{
			name:          "one handed weapon can move to off hand",
			item:          makeItem(dominantHand, weapon, oneHanded),
			requestedSlot: offHand,
			expected:      true,
		},
		{
			name:          "two handed weapon cannot move to off hand",
			item:          makeItem(dominantHand, weapon, twoHanded),
			requestedSlot: offHand,
			expected:      false,
		},
		{
			name:          "shield stays restricted to off hand",
			item:          makeItem(offHand, shield, oneHanded),
			requestedSlot: dominantHand,
			expected:      false,
		},
		{
			name:          "off hand shield can equip to off hand",
			item:          makeItem(offHand, shield, oneHanded),
			requestedSlot: offHand,
			expected:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := canEquipInventoryItemToSlot(tc.item, tc.requestedSlot)
			if actual != tc.expected {
				t.Fatalf("canEquipInventoryItemToSlot() = %v, expected %v", actual, tc.expected)
			}
		})
	}
}
