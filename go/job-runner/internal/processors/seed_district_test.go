package processors

import (
	"fmt"
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

func TestFinalizeDistrictSeedJobKeepsCompletedStatusWhenSomeResultsFail(t *testing.T) {
	message := "older error"
	job := &models.DistrictSeedJob{
		ID:           uuid.New(),
		Status:       models.DistrictSeedJobStatusInProgress,
		ErrorMessage: &message,
	}

	finalizeDistrictSeedJob(job, 2, 5)

	if job.Status != models.DistrictSeedJobStatusCompleted {
		t.Fatalf("expected completed status, got %q", job.Status)
	}
	if job.ErrorMessage != nil {
		t.Fatalf("expected error message to be cleared, got %v", *job.ErrorMessage)
	}
}

func TestShouldRetryDistrictSeedQuestInAnotherZone(t *testing.T) {
	err := fmt.Errorf("no unused points of interest found for location archetype 123 in zone 456 after checking 4 candidates")
	if !shouldRetryDistrictSeedQuestInAnotherZone(err) {
		t.Fatal("expected zone compatibility error to be retried in another zone")
	}
}

func TestShouldNotRetryDistrictSeedQuestInAnotherZoneForOtherErrors(t *testing.T) {
	err := fmt.Errorf("location archetype 123 not found")
	if shouldRetryDistrictSeedQuestInAnotherZone(err) {
		t.Fatal("expected non-POI configuration errors to avoid retrying other zones")
	}
	if shouldRetryDistrictSeedQuestInAnotherZone(fmt.Errorf("temporary network failure")) {
		t.Fatal("expected general errors to avoid district zone fallback retry")
	}
}
