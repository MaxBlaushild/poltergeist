package dungeonmaster

import (
	"errors"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestMakeQuestNodeChallengeUsesPointOfInterestID(t *testing.T) {
	c := &client{}
	zoneID := uuid.New()
	poiID := uuid.New()

	challenge, err := c.makeQuestNodeChallenge(
		zoneID,
		nil,
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
		t.Fatalf("makeQuestNodeChallenge returned error: %v", err)
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

func TestMakeQuestNodeChallengeSupportsCoordinateAnchors(t *testing.T) {
	c := &client{}
	zoneID := uuid.New()

	challenge, err := c.makeQuestNodeChallenge(
		zoneID,
		&questNodeAnchor{
			Latitude:  40.1001,
			Longitude: -73.9002,
		},
		nil,
		models.QuestNodeSubmissionTypePhoto,
		false,
		2,
	)
	if err != nil {
		t.Fatalf("makeQuestNodeChallenge returned error: %v", err)
	}
	if challenge.PointOfInterestID != nil {
		t.Fatalf("expected coordinate-backed challenge to omit PointOfInterestID")
	}
	if challenge.Latitude != 40.1001 || challenge.Longitude != -73.9002 {
		t.Fatalf("expected coordinates to match anchor, got %f,%f", challenge.Latitude, challenge.Longitude)
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

func TestSelectClosestUnusedPointOfInterestUsesReferenceAnchor(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	selected := selectClosestUnusedPointOfInterest(
		[]*models.PointOfInterest{
			{ID: firstID, Name: "Far", Lat: "40.0000", Lng: "-73.0000"},
			{ID: secondID, Name: "Near", Lat: "40.0005", Lng: "-73.0005"},
		},
		map[uuid.UUID]bool{},
		&questNodeAnchor{Latitude: 40.0004, Longitude: -73.0004},
	)
	if selected == nil {
		t.Fatalf("expected a point of interest to be selected")
	}
	if selected.ID != secondID {
		t.Fatalf("expected nearest point of interest %s, got %s", secondID, selected.ID)
	}
}

func TestResolveQuestNodeAnchorSameAsPreviousReusesPreviousAnchor(t *testing.T) {
	c := &client{}
	previousAnchor := &questNodeAnchor{Latitude: 40.1234, Longitude: -73.9876}

	anchor, pointOfInterest, err := c.resolveQuestNodeAnchor(
		t.Context(),
		nil,
		&models.QuestArchetypeNode{
			LocationSelectionMode: models.QuestArchetypeNodeLocationSelectionModeSameAsPrevious,
		},
		nil,
		map[uuid.UUID]bool{},
		previousAnchor,
	)
	if err != nil {
		t.Fatalf("resolveQuestNodeAnchor returned error: %v", err)
	}
	if pointOfInterest != nil {
		t.Fatalf("expected same_as_previous to avoid selecting a point of interest")
	}
	if anchor == nil {
		t.Fatalf("expected same_as_previous to return an anchor")
	}
	if anchor == previousAnchor {
		t.Fatalf("expected same_as_previous to copy the previous anchor instead of reusing the pointer")
	}
	if anchor.Latitude != previousAnchor.Latitude || anchor.Longitude != previousAnchor.Longitude {
		t.Fatalf("expected anchor %f,%f, got %f,%f", previousAnchor.Latitude, previousAnchor.Longitude, anchor.Latitude, anchor.Longitude)
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
