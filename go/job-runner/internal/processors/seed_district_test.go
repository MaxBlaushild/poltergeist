package processors

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestDistrictSeedMatchCountDeduplicatesTags(t *testing.T) {
	score := districtSeedMatchCount(
		models.StringArray{"Market", "market", "harbor", "  "},
		map[string]struct{}{
			"market": {},
			"harbor": {},
		},
	)

	if score != 2 {
		t.Fatalf("expected score 2, got %d", score)
	}
}

func TestSelectBestDistrictSeedZonePicksHighestOverlap(t *testing.T) {
	zones := []models.Zone{
		{ID: uuid.New(), Name: "Harbor", InternalTags: models.StringArray{"harbor", "smuggler"}},
		{ID: uuid.New(), Name: "Market", InternalTags: models.StringArray{"market", "trade", "festival"}},
		{ID: uuid.New(), Name: "Temple", InternalTags: models.StringArray{"ritual"}},
	}

	selected, score, err := selectBestDistrictSeedZone(
		zones,
		models.StringArray{"market", "festival"},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if score != 2 {
		t.Fatalf("expected score 2, got %d", score)
	}
	if selected == nil || selected.ID != zones[1].ID {
		t.Fatalf("expected Market zone to be selected, got %+v", selected)
	}
}

func TestSelectBestDistrictSeedZoneAllowsZeroTagTie(t *testing.T) {
	zones := []models.Zone{
		{ID: uuid.New(), Name: "North"},
		{ID: uuid.New(), Name: "South"},
	}

	selected, score, err := selectBestDistrictSeedZone(zones, models.StringArray{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if score != 0 {
		t.Fatalf("expected zero score, got %d", score)
	}
	if selected == nil {
		t.Fatal("expected a selected zone")
	}
	if selected.ID != zones[0].ID && selected.ID != zones[1].ID {
		t.Fatalf("selected unexpected zone: %+v", selected)
	}
}
