package models

import "testing"

func TestResolvePointOfInterestShopkeeperSeedProfilesOverlaysCustomProfiles(t *testing.T) {
	t.Helper()

	resolved := ResolvePointOfInterestShopkeeperSeedProfiles([]PointOfInterestShopkeeperSeedProfile{
		{
			Category:               PointOfInterestMarkerCategoryMarket,
			SpawnChanceBasisPoints: 1111,
			Candidates: []PointOfInterestShopkeeperSeedCandidate{
				{Tag: " Merchant ", Weight: 2},
				{Tag: "merchant", Weight: 3},
				{Tag: "relic", Weight: 1},
				{Tag: "", Weight: 4},
				{Tag: "guide", Weight: 0},
			},
		},
		{
			Category:               PointOfInterestMarkerCategory("not-real"),
			SpawnChanceBasisPoints: 9999,
			Candidates: []PointOfInterestShopkeeperSeedCandidate{
				{Tag: "ignored", Weight: 1},
			},
		},
	})

	if len(resolved) != len(AllPointOfInterestMarkerCategories()) {
		t.Fatalf("expected one resolved profile per POI marker category, got %d", len(resolved))
	}

	market := PointOfInterestShopkeeperSeedProfileForCategory(resolved, PointOfInterestMarkerCategoryMarket)
	if market.SpawnChanceBasisPoints != 1111 {
		t.Fatalf("expected market spawn chance to be overridden, got %d", market.SpawnChanceBasisPoints)
	}
	if len(market.Candidates) != 2 {
		t.Fatalf("expected normalized market candidates to dedupe to 2 entries, got %d", len(market.Candidates))
	}
	if market.Candidates[0].Tag != "merchant" || market.Candidates[0].Weight != 5 {
		t.Fatalf("expected merged merchant candidate weight 5, got %+v", market.Candidates[0])
	}
	if market.Candidates[1].Tag != "relic" || market.Candidates[1].Weight != 1 {
		t.Fatalf("expected relic candidate to remain, got %+v", market.Candidates[1])
	}

	archive := PointOfInterestShopkeeperSeedProfileForCategory(resolved, PointOfInterestMarkerCategoryArchive)
	if archive.SpawnChanceBasisPoints != 7600 {
		t.Fatalf("expected untouched archive default chance 7600, got %d", archive.SpawnChanceBasisPoints)
	}
	if len(archive.Candidates) == 0 || archive.Candidates[0].Tag != "arcane" {
		t.Fatalf("expected untouched archive defaults, got %+v", archive.Candidates)
	}
}
