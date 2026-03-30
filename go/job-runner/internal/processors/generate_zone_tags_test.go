package processors

import (
	"reflect"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestSanitizeSelectedZoneTagsNormalizesAndFiltersToPool(t *testing.T) {
	candidatePool := []string{
		"market_square",
		"dockside",
		"nightlife",
		"working_class",
		"ritual_sites",
	}

	selected, err := sanitizeSelectedZoneTags(
		[]string{"Market Square", "dockside", "night-life", "working class", "ritual_sites", "not_in_pool"},
		candidatePool,
		5,
	)
	if err != nil {
		t.Fatalf("expected valid selection, got error: %v", err)
	}

	expected := []string{
		"market_square",
		"dockside",
		"nightlife",
		"working_class",
		"ritual_sites",
	}
	if !reflect.DeepEqual(selected, expected) {
		t.Fatalf("expected %v, got %v", expected, selected)
	}
}

func TestApplyGeneratedZoneTagsPreservesManualTagsAndReplacesPoolTags(t *testing.T) {
	existing := models.StringArray{"custom_manual_tag", "dockside", "nightlife"}
	generated := []string{"working_class", "market_square", "ritual_sites", "dockside", "harborfront"}
	candidatePool := []string{"dockside", "nightlife", "working_class", "market_square", "ritual_sites", "harborfront"}

	merged := applyGeneratedZoneTags(existing, generated, candidatePool)
	expected := models.StringArray{
		"custom_manual_tag",
		"working_class",
		"market_square",
		"ritual_sites",
		"dockside",
		"harborfront",
	}
	if !reflect.DeepEqual(merged, expected) {
		t.Fatalf("expected %v, got %v", expected, merged)
	}
}
