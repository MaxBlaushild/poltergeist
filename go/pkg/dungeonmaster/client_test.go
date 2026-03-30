package dungeonmaster

import (
	"errors"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestMakeQuestLocationChallengeUsesPointOfInterestID(t *testing.T) {
	c := &client{}
	zoneID := uuid.New()
	poiID := uuid.New()

	challenge, err := c.makeQuestLocationChallenge(
		zoneID,
		&models.PointOfInterest{
			ID:   poiID,
			Name: "The Copper Cup",
			Lat:  "40.7128",
			Lng:  "-74.0060",
		},
		models.QuestNodeSubmissionTypePhoto,
		true,
		3,
	)
	if err != nil {
		t.Fatalf("makeQuestLocationChallenge returned error: %v", err)
	}
	if challenge.PointOfInterestID == nil {
		t.Fatalf("expected PointOfInterestID to be set")
	}
	if *challenge.PointOfInterestID != poiID {
		t.Fatalf("expected PointOfInterestID %s, got %s", poiID, *challenge.PointOfInterestID)
	}
	if challenge.ZoneID != zoneID {
		t.Fatalf("expected zone %s, got %s", zoneID, challenge.ZoneID)
	}
	if challenge.Latitude != 40.7128 || challenge.Longitude != -74.0060 {
		t.Fatalf("expected coordinates to match point of interest, got %f,%f", challenge.Latitude, challenge.Longitude)
	}
}

func TestQuestNodePOISearchCountRequestsMoreThanOneCandidate(t *testing.T) {
	if count := questNodePOISearchCount(0); count < 1 {
		t.Fatalf("expected at least one candidate, got %d", count)
	}
	if count := questNodePOISearchCount(1); count <= 1 {
		t.Fatalf("expected more than one candidate once a POI is already used, got %d", count)
	}
	if count := questNodePOISearchCount(50); count > 20 {
		t.Fatalf("expected candidate count to be capped at 20, got %d", count)
	}
}

func TestSelectUnusedPointOfInterestSkipsUsedCandidates(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	selected := selectUnusedPointOfInterest(
		[]*models.PointOfInterest{
			{ID: firstID, Name: "First"},
			{ID: secondID, Name: "Second"},
		},
		map[uuid.UUID]bool{firstID: true},
	)
	if selected == nil {
		t.Fatalf("expected an unused point of interest to be selected")
	}
	if selected.ID != secondID {
		t.Fatalf("expected point of interest %s, got %s", secondID, selected.ID)
	}
}

func TestIsNonRetriableQuestGenerationError(t *testing.T) {
	baseErr := errors.New("deterministic failure")
	if IsNonRetriableQuestGenerationError(baseErr) {
		t.Fatalf("expected base error to be retriable by default")
	}
	if !IsNonRetriableQuestGenerationError(markNonRetriableQuestGenerationError(baseErr)) {
		t.Fatalf("expected marked error to be classified as non-retriable")
	}
}
