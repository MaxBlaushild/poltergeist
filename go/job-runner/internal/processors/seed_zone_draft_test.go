package processors

import (
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestGenerateZoneSeedExpositionsBuildsRequestedCount(t *testing.T) {
	jobID := uuid.New()
	zone := models.Zone{
		ID:        uuid.New(),
		Name:      "Harbor District",
		Latitude:  40.7128,
		Longitude: -74.0060,
	}
	branding := &zoneBrandingResponse{
		FantasyName:     "Gullmarket",
		ZoneDescription: "A salt-stained quarter of smugglers, ferrymen, and hidden errands.",
	}
	pois := []models.ZoneSeedPointOfInterestDraft{
		{
			DraftID:          uuid.New(),
			PlaceID:          "dock-place",
			Name:             "Pier Cafe",
			Types:            []string{"cafe", "marina"},
			Latitude:         40.7131,
			Longitude:        -74.0055,
			Address:          "1 Dock Way",
			EditorialSummary: "Bustling waterfront cafe",
		},
		{
			DraftID:          uuid.New(),
			PlaceID:          "market-place",
			Name:             "Night Market",
			Types:            []string{"market", "store"},
			Latitude:         40.7124,
			Longitude:        -74.0065,
			Address:          "2 Lantern Row",
			EditorialSummary: "Crowded evening bazaar",
		},
	}
	characters := []models.ZoneSeedCharacterDraft{
		{
			DraftID: uuid.New(),
			Name:    "Captain Mire",
			PlaceID: "dock-place",
		},
		{
			DraftID: uuid.New(),
			Name:    "Tallow Venn",
			PlaceID: "market-place",
		},
	}

	expositions := generateZoneSeedExpositions(
		jobID,
		zone,
		branding,
		"port",
		pois,
		characters,
		3,
	)

	if len(expositions) != 3 {
		t.Fatalf("expected 3 exposition drafts, got %d", len(expositions))
	}

	allowedPlaceIDs := map[string]struct{}{
		"dock-place":   {},
		"market-place": {},
	}
	for index, exposition := range expositions {
		if exposition.DraftID == uuid.Nil {
			t.Fatalf("expected exposition %d to have a draft id", index)
		}
		if exposition.Title == "" {
			t.Fatalf("expected exposition %d to have a title", index)
		}
		if exposition.Description == "" {
			t.Fatalf("expected exposition %d to have a description", index)
		}
		if len(exposition.Dialogue) < 3 {
			t.Fatalf("expected exposition %d to have at least 3 dialogue lines, got %+v", index, exposition.Dialogue)
		}
		if _, ok := allowedPlaceIDs[exposition.PlaceID]; !ok {
			t.Fatalf("expected exposition %d to use a known place id, got %q", index, exposition.PlaceID)
		}
		for _, line := range exposition.Dialogue {
			if line.SpeakerName == "" || line.Text == "" {
				t.Fatalf("expected exposition dialogue lines to keep speaker names and text, got %+v", line)
			}
		}
	}
}

func TestGenerateZoneSeedExpositionsFallsBackWithoutPOIs(t *testing.T) {
	expositions := generateZoneSeedExpositions(
		uuid.New(),
		models.Zone{
			ID:        uuid.New(),
			Name:      "Silent Moor",
			Latitude:  41.0,
			Longitude: -73.5,
		},
		&zoneBrandingResponse{FantasyName: "Silent Moor"},
		"swamp",
		nil,
		nil,
		1,
	)

	if len(expositions) != 1 {
		t.Fatalf("expected 1 exposition draft, got %d", len(expositions))
	}
	if expositions[0].Latitude == nil || expositions[0].Longitude == nil {
		t.Fatalf("expected exposition fallback coordinates, got %+v", expositions[0])
	}
	if len(expositions[0].Dialogue) == 0 {
		t.Fatalf("expected exposition fallback dialogue, got %+v", expositions[0])
	}
}

func TestZoneSeedPointOfInterestExpositionCountForRollsSupportsSecondSpawn(t *testing.T) {
	if got := zoneSeedPointOfInterestExpositionCountForRolls(
		models.PointOfInterestMarkerCategoryMarket,
		0,
		0,
	); got != 2 {
		t.Fatalf("expected market exposition rolls to allow a second exposition, got %d", got)
	}

	if got := zoneSeedPointOfInterestExpositionCountForRolls(
		models.PointOfInterestMarkerCategoryMarket,
		0,
		9999,
	); got != 1 {
		t.Fatalf("expected market exposition rolls to keep one exposition when the second roll misses, got %d", got)
	}

	if got := zoneSeedPointOfInterestExpositionCountForRolls(
		models.PointOfInterestMarkerCategoryMarket,
		9999,
		0,
	); got != 0 {
		t.Fatalf("expected market exposition rolls to miss entirely when the first roll misses, got %d", got)
	}
}

func TestZoneSeedPointOfInterestExpositionCountForRollsUsesConfiguredProfiles(t *testing.T) {
	if got := zoneSeedPointOfInterestExpositionCountForRollsWithProfiles(
		models.PointOfInterestMarkerCategoryGeneric,
		0,
		0,
		[]models.PointOfInterestExpositionSeedProfile{
			{
				Category:                     models.PointOfInterestMarkerCategoryGeneric,
				FirstSpawnChanceBasisPoints:  10_000,
				SecondSpawnChanceBasisPoints: 10_000,
			},
		},
	); got != 2 {
		t.Fatalf("expected configured exposition profile to allow two expositions, got %d", got)
	}
}

func TestBuildZoneSeedExpositionDraftTailorsToPOIName(t *testing.T) {
	zone := models.Zone{
		ID:        uuid.New(),
		Name:      "Harbor District",
		Latitude:  40.7128,
		Longitude: -74.0060,
	}
	branding := &zoneBrandingResponse{
		FantasyName:     "Gullmarket",
		ZoneDescription: "A salt-stained quarter of smugglers, ferrymen, and hidden errands.",
	}
	poi := &models.ZoneSeedPointOfInterestDraft{
		DraftID:         uuid.New(),
		PlaceID:         "market-place",
		Name:            "Night Market",
		MarkerCategory:  models.PointOfInterestMarkerCategoryMarket,
		Types:           []string{"market", "store"},
		Latitude:        40.7124,
		Longitude:       -74.0065,
		Address:         "2 Lantern Row",
		UserRatingCount: 120,
	}

	exposition := buildZoneSeedExpositionDraft(0, zone, branding, "port", poi, nil)

	if exposition.PlaceID != poi.PlaceID {
		t.Fatalf("expected exposition to stay attached to %q, got %q", poi.PlaceID, exposition.PlaceID)
	}
	if !strings.Contains(strings.ToLower(exposition.Description), "night market") {
		t.Fatalf("expected exposition description to mention the POI name, got %q", exposition.Description)
	}
	if len(exposition.Dialogue) < 3 {
		t.Fatalf("expected tailored exposition dialogue, got %+v", exposition.Dialogue)
	}
	foundLocationReference := false
	for _, line := range exposition.Dialogue {
		if strings.Contains(strings.ToLower(line.Text), "night market") {
			foundLocationReference = true
			break
		}
	}
	if !foundLocationReference {
		t.Fatalf("expected exposition dialogue to mention the POI name, got %+v", exposition.Dialogue)
	}
}
