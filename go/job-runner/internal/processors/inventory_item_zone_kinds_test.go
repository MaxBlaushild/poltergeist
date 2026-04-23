package processors

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestDeriveInventoryItemSuggestionZoneKindHeuristically(t *testing.T) {
	zoneKinds := []models.ZoneKind{
		{Slug: "swamp_bog", Name: "Swamp Bog", Description: "Rotting marshland and toxic reeds."},
		{Slug: "city_streets", Name: "City Streets", Description: "Dense urban alleys, markets, and neon storefronts."},
	}

	got := deriveInventoryItemSuggestionZoneKindHeuristically(
		zoneKinds,
		"",
		"Bogglass Tonic",
		"A murky remedy brewed from reeds and swamp venom.",
		"Restores mana and hardens the drinker against marsh toxins.",
	)

	if got != "swamp-bog" {
		t.Fatalf("expected swamp-bog, got %q", got)
	}
}
