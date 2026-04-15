package server

import (
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/paulmach/orb"
)

func TestZoneSeedInferAutoCountsAppliesMinimumFloors(t *testing.T) {
	counts, warnings := zoneSeedInferAutoCounts(0, nil)

	if counts.PlaceCount != 1 ||
		counts.MonsterCount != 1 ||
		counts.BossEncounterCount != 1 ||
		counts.RaidEncounterCount != 1 ||
		counts.InputEncounterCount != 1 ||
		counts.OptionEncounterCount != 1 ||
		counts.TreasureChestCount != 1 ||
		counts.HealingFountainCount != 1 {
		t.Fatalf("expected minimum floor of 1 for every count, got %+v", counts)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings without required tags, got %v", warnings)
	}
}

func TestZoneSeedInferAutoCountsBumpsPlacesForRequiredTags(t *testing.T) {
	counts, warnings := zoneSeedInferAutoCounts(0.25, []string{"cafe", "park", "museum"})

	if counts.PlaceCount != 3 {
		t.Fatalf("expected place count to bump to 3, got %d", counts.PlaceCount)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "Increased POI recommendation") {
		t.Fatalf("expected bump warning, got %v", warnings)
	}
}

func TestZoneSeedAreaForAuditRejectsMissingBoundary(t *testing.T) {
	if _, _, err := zoneSeedAreaForAudit(&models.Zone{}); err == nil {
		t.Fatal("expected missing boundary error")
	}
}

func TestZoneSeedAreaForAuditReturnsAreaForPolygon(t *testing.T) {
	polygon := orb.Polygon{{
		{-73.99, 40.75},
		{-73.99, 40.751},
		{-73.989, 40.751},
		{-73.989, 40.75},
		{-73.99, 40.75},
	}}
	zone := &models.Zone{
		Polygon: &polygon,
	}

	squareFeet, acres, err := zoneSeedAreaForAudit(zone)
	if err != nil {
		t.Fatalf("expected polygon area, got error: %v", err)
	}
	if squareFeet <= 0 {
		t.Fatalf("expected square feet to be positive, got %f", squareFeet)
	}
	if acres <= 0 {
		t.Fatalf("expected acres to be positive, got %f", acres)
	}
}
