package dungeonmaster

import (
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
