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
		counts.HealingFountainCount != 1 ||
		counts.ResourceCount != 1 {
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

func TestNormalizeZoneSeedDraftCountModeDefaultsToAbsolute(t *testing.T) {
	mode, err := normalizeZoneSeedDraftCountMode("")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mode != models.ZoneSeedCountModeAbsolute {
		t.Fatalf("expected absolute default, got %q", mode)
	}
}

func TestZoneSeedResolveCurrentAwareCountsSubtractsExisting(t *testing.T) {
	target := models.ZoneSeedResolvedCounts{
		PlaceCount:           15,
		MonsterCount:         6,
		BossEncounterCount:   2,
		RaidEncounterCount:   1,
		InputEncounterCount:  4,
		OptionEncounterCount: 3,
		TreasureChestCount:   5,
		HealingFountainCount: 2,
		ResourceCount:        7,
	}
	snapshot := zoneSeedCurrentContentSnapshot{
		ExistingCounts: models.ZoneSeedResolvedCounts{
			PlaceCount:           3,
			MonsterCount:         1,
			BossEncounterCount:   1,
			RaidEncounterCount:   0,
			InputEncounterCount:  1,
			OptionEncounterCount: 2,
			TreasureChestCount:   4,
			HealingFountainCount: 1,
			ResourceCount:        2,
		},
	}

	queued, audit := zoneSeedResolveCurrentAwareCounts(target, snapshot)

	if queued.PlaceCount != 12 ||
		queued.MonsterCount != 5 ||
		queued.BossEncounterCount != 1 ||
		queued.RaidEncounterCount != 1 ||
		queued.InputEncounterCount != 3 ||
		queued.OptionEncounterCount != 1 ||
		queued.TreasureChestCount != 1 ||
		queued.HealingFountainCount != 1 ||
		queued.ResourceCount != 5 {
		t.Fatalf("unexpected queued counts: %+v", queued)
	}
	if audit.QueuedCounts.PlaceCount != 12 {
		t.Fatalf("expected audit queued counts to be stored, got %+v", audit.QueuedCounts)
	}
	if len(audit.Warnings) == 0 {
		t.Fatal("expected current-aware warnings when existing content reduces counts")
	}
}

func TestZoneSeedResolveCurrentAwareCountsBumpsPlacesForUnmetRequiredTags(t *testing.T) {
	target := models.ZoneSeedResolvedCounts{PlaceCount: 4}
	snapshot := zoneSeedCurrentContentSnapshot{
		ExistingCounts:             models.ZoneSeedResolvedCounts{PlaceCount: 3},
		RemainingRequiredPlaceTags: []string{"cafe", "park", "museum"},
	}

	queued, audit := zoneSeedResolveCurrentAwareCounts(target, snapshot)

	if queued.PlaceCount != 3 {
		t.Fatalf("expected place count to bump to remaining required tags, got %d", queued.PlaceCount)
	}
	if len(audit.RemainingRequiredPlaceTags) != 3 {
		t.Fatalf("expected remaining required tags in audit, got %v", audit.RemainingRequiredPlaceTags)
	}
	if len(audit.Warnings) == 0 || !strings.Contains(audit.Warnings[len(audit.Warnings)-1], "still unmet") {
		t.Fatalf("expected unmet-tag warning, got %v", audit.Warnings)
	}
}

func TestZoneSeedRemainingRequiredPlaceTagsUsesExistingPOIs(t *testing.T) {
	pois := []models.PointOfInterest{
		{
			Name:           "Morning Roast",
			MarkerCategory: models.PointOfInterestMarkerCategoryCoffeehouse,
		},
		{
			Name:           "Central Library",
			MarkerCategory: models.PointOfInterestMarkerCategoryArchive,
		},
	}

	remaining := zoneSeedRemainingRequiredPlaceTags(
		[]string{"cafe", "library", "park"},
		pois,
	)

	if len(remaining) != 1 || remaining[0] != "park" {
		t.Fatalf("expected only park to remain unmet, got %v", remaining)
	}
}
