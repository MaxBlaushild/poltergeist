package poilocals

import "testing"

func TestGenerateDraftsFallsBackWithoutDeepPriest(t *testing.T) {
	t.Helper()

	places := []PlaceContext{
		{ID: "poi-1", Name: "Gilded Bakery", Types: []string{"bakery", "cafe"}},
		{ID: "poi-2", Name: "Moonlit Park", Types: []string{"park"}},
	}

	drafts := GenerateDrafts(nil, ZoneContext{Name: "Lantern Ward"}, places)
	expected := DesiredLocalCount("poi-1") + DesiredLocalCount("poi-2")
	if len(drafts) != expected {
		t.Fatalf("expected %d drafts, got %d", expected, len(drafts))
	}

	for _, draft := range drafts {
		if draft.Name == "" {
			t.Fatalf("expected fallback draft to have a name")
		}
		if draft.Description == "" {
			t.Fatalf("expected fallback draft %q to have a description", draft.Name)
		}
		if len(draft.Dialogue) < 2 {
			t.Fatalf("expected fallback draft %q to have at least 2 dialogue lines", draft.Name)
		}
		if draft.PlaceID == "" {
			t.Fatalf("expected fallback draft %q to retain place id", draft.Name)
		}
	}
}

func TestDesiredLocalCountIsStableAndBounded(t *testing.T) {
	t.Helper()

	keys := []string{"poi-alpha", "poi-beta", "poi-gamma", "poi-delta"}
	for _, key := range keys {
		first := DesiredLocalCount(key)
		second := DesiredLocalCount(key)
		if first != second {
			t.Fatalf("expected stable desired count for %q", key)
		}
		if first < 1 || first > 2 {
			t.Fatalf("expected desired count for %q to be 1 or 2, got %d", key, first)
		}
	}
}
