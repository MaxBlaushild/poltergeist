package locationseeder

import (
	"context"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type fakePointOfInterestFinder struct {
	existingByPlaceID map[string]*models.PointOfInterest
	calls             []string
	err               error
}

func (f *fakePointOfInterestFinder) FindByGoogleMapsPlaceID(
	_ context.Context,
	googleMapsPlaceID string,
) (*models.PointOfInterest, error) {
	f.calls = append(f.calls, googleMapsPlaceID)
	if f.err != nil {
		return nil, f.err
	}
	if f.existingByPlaceID == nil {
		return nil, nil
	}
	return f.existingByPlaceID[googleMapsPlaceID], nil
}

type zoneAttachment struct {
	zoneID            uuid.UUID
	pointOfInterestID uuid.UUID
}

type fakeZoneAttacher struct {
	attachments []zoneAttachment
	err         error
}

func (f *fakeZoneAttacher) AddPointOfInterestToZone(
	_ context.Context,
	zoneID uuid.UUID,
	pointOfInterestID uuid.UUID,
) error {
	f.attachments = append(f.attachments, zoneAttachment{
		zoneID:            zoneID,
		pointOfInterestID: pointOfInterestID,
	})
	return f.err
}

type fakeGooglePlaceFinder struct {
	place *googlemaps.Place
	calls []string
	err   error
}

func (f *fakeGooglePlaceFinder) FindPlaceByID(id string) (*googlemaps.Place, error) {
	f.calls = append(f.calls, id)
	if f.err != nil {
		return nil, f.err
	}
	return f.place, nil
}

func TestImportPlaceWithReuseReturnsExistingPointOfInterest(t *testing.T) {
	ctx := context.Background()
	zoneID := uuid.New()
	existingPOI := &models.PointOfInterest{ID: uuid.New(), Name: "Existing"}
	poiFinder := &fakePointOfInterestFinder{
		existingByPlaceID: map[string]*models.PointOfInterest{
			"place-123": existingPOI,
		},
	}
	zoneAttacher := &fakeZoneAttacher{}
	googleFinder := &fakeGooglePlaceFinder{
		place: &googlemaps.Place{ID: "place-123", Name: "Should Not Load"},
	}
	generated := false

	poi, err := importPlaceWithReuse(
		ctx,
		"  place-123  ",
		models.Zone{ID: zoneID},
		nil,
		poiFinder,
		zoneAttacher,
		googleFinder,
		func(context.Context, googlemaps.Place, *models.Zone, *models.ZoneGenre) (*models.PointOfInterest, error) {
			generated = true
			return nil, nil
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if poi != existingPOI {
		t.Fatalf("expected existing point of interest to be reused")
	}
	if generated {
		t.Fatalf("expected generator not to run when an existing point of interest is found")
	}
	if len(googleFinder.calls) != 0 {
		t.Fatalf("expected google place lookup to be skipped, got %d call(s)", len(googleFinder.calls))
	}
	if len(zoneAttacher.attachments) != 1 {
		t.Fatalf("expected existing point of interest to be attached to the zone once, got %d", len(zoneAttacher.attachments))
	}
	if zoneAttacher.attachments[0].zoneID != zoneID {
		t.Fatalf("expected zone attachment to use zone %s, got %s", zoneID, zoneAttacher.attachments[0].zoneID)
	}
	if zoneAttacher.attachments[0].pointOfInterestID != existingPOI.ID {
		t.Fatalf(
			"expected attached point of interest %s, got %s",
			existingPOI.ID,
			zoneAttacher.attachments[0].pointOfInterestID,
		)
	}
}

func TestImportPlaceWithReuseGeneratesPointOfInterestWhenMissing(t *testing.T) {
	ctx := context.Background()
	zoneID := uuid.New()
	poiFinder := &fakePointOfInterestFinder{}
	zoneAttacher := &fakeZoneAttacher{}
	googlePlace := &googlemaps.Place{ID: "place-456", Name: "Lantern Cafe"}
	googleFinder := &fakeGooglePlaceFinder{place: googlePlace}
	generatedPOI := &models.PointOfInterest{ID: uuid.New(), Name: "Generated"}
	generated := false

	poi, err := importPlaceWithReuse(
		ctx,
		"place-456",
		models.Zone{ID: zoneID},
		nil,
		poiFinder,
		zoneAttacher,
		googleFinder,
		func(_ context.Context, place googlemaps.Place, zone *models.Zone, genre *models.ZoneGenre) (*models.PointOfInterest, error) {
			generated = true
			if place.ID != googlePlace.ID {
				t.Fatalf("expected place ID %q, got %q", googlePlace.ID, place.ID)
			}
			if zone == nil || zone.ID != zoneID {
				t.Fatalf("expected generator to receive zone %s", zoneID)
			}
			if genre != nil {
				t.Fatalf("expected nil genre for default import flow")
			}
			return generatedPOI, nil
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if poi != generatedPOI {
		t.Fatalf("expected generated point of interest to be returned")
	}
	if !generated {
		t.Fatalf("expected generator to run when no existing point of interest is found")
	}
	if len(googleFinder.calls) != 1 || googleFinder.calls[0] != "place-456" {
		t.Fatalf("expected a google place lookup for place-456, got %v", googleFinder.calls)
	}
	if len(zoneAttacher.attachments) != 0 {
		t.Fatalf("expected no zone reattachment for a newly generated point of interest, got %d", len(zoneAttacher.attachments))
	}
}

func TestImportPlaceWithReuseRejectsBlankPlaceID(t *testing.T) {
	ctx := context.Background()
	poiFinder := &fakePointOfInterestFinder{}
	zoneAttacher := &fakeZoneAttacher{}
	googleFinder := &fakeGooglePlaceFinder{}
	generated := false

	poi, err := importPlaceWithReuse(
		ctx,
		"   ",
		models.Zone{ID: uuid.New()},
		nil,
		poiFinder,
		zoneAttacher,
		googleFinder,
		func(context.Context, googlemaps.Place, *models.Zone, *models.ZoneGenre) (*models.PointOfInterest, error) {
			generated = true
			return nil, nil
		},
	)
	if err == nil {
		t.Fatalf("expected an error for a blank place ID")
	}
	if poi != nil {
		t.Fatalf("expected no point of interest to be returned")
	}
	if generated {
		t.Fatalf("expected generator to be skipped for a blank place ID")
	}
	if len(googleFinder.calls) != 0 {
		t.Fatalf("expected google place lookup to be skipped for a blank place ID")
	}
	if len(zoneAttacher.attachments) != 0 {
		t.Fatalf("expected no zone attachments for a blank place ID")
	}
}
