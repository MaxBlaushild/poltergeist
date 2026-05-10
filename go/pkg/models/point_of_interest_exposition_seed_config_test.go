package models

import "testing"

func TestResolvePointOfInterestExpositionSeedProfilesOverlaysCustomProfiles(t *testing.T) {
	t.Helper()

	resolved := ResolvePointOfInterestExpositionSeedProfiles([]PointOfInterestExpositionSeedProfile{
		{
			Category:                     PointOfInterestMarkerCategoryMarket,
			FirstSpawnChanceBasisPoints:  1111,
			SecondSpawnChanceBasisPoints: 2222,
		},
		{
			Category:                     PointOfInterestMarkerCategory("not-real"),
			FirstSpawnChanceBasisPoints:  9999,
			SecondSpawnChanceBasisPoints: 9999,
		},
	})

	if len(resolved) != len(AllPointOfInterestMarkerCategories()) {
		t.Fatalf("expected one resolved profile per POI marker category, got %d", len(resolved))
	}

	market := PointOfInterestExpositionSeedProfileForCategory(resolved, PointOfInterestMarkerCategoryMarket)
	if market.FirstSpawnChanceBasisPoints != 1111 || market.SecondSpawnChanceBasisPoints != 2222 {
		t.Fatalf("expected market exposition ratios to be overridden, got %+v", market)
	}

	archive := PointOfInterestExpositionSeedProfileForCategory(resolved, PointOfInterestMarkerCategoryArchive)
	if archive.FirstSpawnChanceBasisPoints != 7000 || archive.SecondSpawnChanceBasisPoints != 2400 {
		t.Fatalf("expected untouched archive defaults, got %+v", archive)
	}
}
