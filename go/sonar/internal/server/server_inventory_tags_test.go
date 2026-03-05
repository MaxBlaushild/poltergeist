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
