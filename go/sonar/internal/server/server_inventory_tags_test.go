package server

import "testing"

func TestParseInventoryInternalTagsNormalizesAndDeduplicates(t *testing.T) {
	result := parseInventoryInternalTags([]string{
		"  Consumable  ",
		"seed_drop_only",
		"consumable",
		"",
		"  ",
		"SEED_DROP_ONLY",
		"  healing  ",
	})

	expected := []string{"consumable", "seed_drop_only", "healing"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d tags, got %d (%v)", len(expected), len(result), result)
	}
	for idx := range expected {
		if result[idx] != expected[idx] {
			t.Fatalf("expected tag[%d] = %q, got %q", idx, expected[idx], result[idx])
		}
	}
}

func TestParseCharacterInternalTagsNormalizesAndDeduplicates(t *testing.T) {
	result := parseCharacterInternalTags([]string{
		"  Merchant  ",
		"starter_quest",
		"merchant",
		"",
		"  ",
		"STARTER_QUEST",
		"  blacksmith  ",
	})

	expected := []string{"merchant", "starter_quest", "blacksmith"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d tags, got %d (%v)", len(expected), len(result), result)
	}
	for idx := range expected {
		if result[idx] != expected[idx] {
			t.Fatalf("expected tag[%d] = %q, got %q", idx, expected[idx], result[idx])
		}
	}
}

func TestParseZoneInternalTagsNormalizesAndDeduplicates(t *testing.T) {
	result := parseZoneInternalTags([]string{
		"  Downtown  ",
		"starter_region",
		"downtown",
		"",
		"  ",
		"STARTER_REGION",
		"  rail_hub  ",
	})

	expected := []string{"downtown", "starter_region", "rail_hub"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d tags, got %d (%v)", len(expected), len(result), result)
	}
	for idx := range expected {
		if result[idx] != expected[idx] {
			t.Fatalf("expected tag[%d] = %q, got %q", idx, expected[idx], result[idx])
		}
	}
}

func TestParseScenarioInternalTagsNormalizesAndDeduplicates(t *testing.T) {
	result := parseScenarioInternalTags([]string{
		"  Tutorial  ",
		"boss_intro",
		"tutorial",
		"",
		"  ",
		"BOSS_INTRO",
		"  social_check  ",
	})

	expected := []string{"tutorial", "boss_intro", "social_check"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d tags, got %d (%v)", len(expected), len(result), result)
	}
	for idx := range expected {
		if result[idx] != expected[idx] {
			t.Fatalf("expected tag[%d] = %q, got %q", idx, expected[idx], result[idx])
		}
	}
}
