package models

import "testing"

func TestInferPointOfInterestMarkerCategoryPrefersPrimaryType(t *testing.T) {
	category := InferPointOfInterestMarkerCategory(
		"coffee_shop",
		[]string{"restaurant", "food", "store"},
		"Coffee Shop",
		false,
		false,
		false,
		false,
		false,
	)

	if category != PointOfInterestMarkerCategoryCoffeehouse {
		t.Fatalf("expected coffeehouse, got %q", category)
	}
}

func TestInferPointOfInterestMarkerCategoryUsesFoodDrinkHints(t *testing.T) {
	tavern := InferPointOfInterestMarkerCategory(
		"restaurant",
		[]string{"restaurant", "food"},
		"Late Night Kitchen",
		false,
		true,
		false,
		true,
		true,
	)
	if tavern != PointOfInterestMarkerCategoryTavern {
		t.Fatalf("expected tavern, got %q", tavern)
	}

	coffeehouse := InferPointOfInterestMarkerCategory(
		"restaurant",
		[]string{"restaurant", "food"},
		"Morning Counter",
		true,
		false,
		false,
		false,
		false,
	)
	if coffeehouse != PointOfInterestMarkerCategoryCoffeehouse {
		t.Fatalf("expected coffeehouse, got %q", coffeehouse)
	}
}

func TestInferPointOfInterestMarkerCategoryFallsBackToTypeList(t *testing.T) {
	category := InferPointOfInterestMarkerCategory(
		"",
		[]string{"book_store", "store"},
		"",
		false,
		false,
		false,
		false,
		false,
	)

	if category != PointOfInterestMarkerCategoryArchive {
		t.Fatalf("expected archive, got %q", category)
	}
}

func TestInferPointOfInterestMarkerCategoryFallsBackToDisplayText(t *testing.T) {
	category := InferPointOfInterestMarkerCategory(
		"",
		nil,
		"Observation Deck",
		false,
		false,
		false,
		false,
		false,
	)

	if category != PointOfInterestMarkerCategoryLandmark {
		t.Fatalf("expected landmark, got %q", category)
	}
}

func TestNormalizePointOfInterestMarkerCategoryDefaultsToGeneric(t *testing.T) {
	if category := NormalizePointOfInterestMarkerCategory("unknown"); category != PointOfInterestMarkerCategoryGeneric {
		t.Fatalf("expected generic, got %q", category)
	}
}
