package processors

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestNormalizeScenarioGeneratedZoneKindFallsBackToPreferredKind(t *testing.T) {
	zoneKinds := []models.ZoneKind{
		{Slug: "forest", Name: "Forest"},
		{Slug: "swamp", Name: "Swamp"},
	}

	got := normalizeScenarioGeneratedZoneKind("invalid-kind", zoneKinds, "forest")
	if got != "forest" {
		t.Fatalf("expected fallback forest, got %q", got)
	}
}

func TestDeriveScenarioZoneKindHeuristicallyMatchesEnvironmentalCue(t *testing.T) {
	zoneKinds := []models.ZoneKind{
		{Slug: "forest", Name: "Forest", Description: "Ancient woods and tangled groves."},
		{Slug: "swamp", Name: "Swamp", Description: "Boggy wetlands full of mire and reeds."},
	}

	got := deriveScenarioZoneKindHeuristically(
		zoneKinds,
		"",
		"A ferryman begs for help guiding villagers across a foggy bog while something moves beneath the reeds.",
	)
	if got != "swamp" {
		t.Fatalf("expected swamp, got %q", got)
	}
}
