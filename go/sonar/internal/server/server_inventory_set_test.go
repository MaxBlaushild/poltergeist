package server

import "testing"

func TestNormalizeInventorySetRarityTier(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{input: "Common", expected: "Common"},
		{input: " uncommon ", expected: "Uncommon"},
		{input: "EPIC", expected: "Epic"},
		{input: "mythic", expected: "Mythic"},
		{input: "Not Droppable", expected: ""},
		{input: "", expected: ""},
	}

	for _, tc := range testCases {
		actual := normalizeInventorySetRarityTier(tc.input)
		if actual != tc.expected {
			t.Fatalf("normalizeInventorySetRarityTier(%q) = %q, expected %q", tc.input, actual, tc.expected)
		}
	}
}

func TestInventorySetPrimaryStatPointsForTargetLevelScalesWithRarity(t *testing.T) {
	level := 50
	common := inventorySetPrimaryStatPointsForTargetLevel(level, "Common")
	uncommon := inventorySetPrimaryStatPointsForTargetLevel(level, "Uncommon")
	epic := inventorySetPrimaryStatPointsForTargetLevel(level, "Epic")
	mythic := inventorySetPrimaryStatPointsForTargetLevel(level, "Mythic")

	if !(common < uncommon && uncommon < epic && epic < mythic) {
		t.Fatalf(
			"expected rarity scaling for level %d, got Common=%d Uncommon=%d Epic=%d Mythic=%d",
			level,
			common,
			uncommon,
			epic,
			mythic,
		)
	}
}

func TestInventorySetPrimaryStatPointsForTargetLevelScalesWithLevel(t *testing.T) {
	lowLevel := inventorySetPrimaryStatPointsForTargetLevel(10, "Mythic")
	highLevel := inventorySetPrimaryStatPointsForTargetLevel(80, "Mythic")
	if lowLevel >= highLevel {
		t.Fatalf("expected higher target levels to produce higher points: level10=%d level80=%d", lowLevel, highLevel)
	}
}
