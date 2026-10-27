package server

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
)

type destinationTestMaps struct {
	googlemaps.Client
	candidates []googlemaps.Candidate
	searchErr  error
	details    map[string]*googlemaps.Place
	detailsErr error
	queries    []string
	detailIDs  []string
}

func (m *destinationTestMaps) FindCandidatesByQuery(query string) ([]googlemaps.Candidate, error) {
	m.queries = append(m.queries, query)
	return m.candidates, m.searchErr
}

func (m *destinationTestMaps) FindPlaceByID(id string) (*googlemaps.Place, error) {
	m.detailIDs = append(m.detailIDs, id)
	return m.details[id], m.detailsErr
}

type destinationContextTestMaps struct {
	*destinationTestMaps
	search func(context.Context, string) ([]googlemaps.Candidate, error)
	detail func(context.Context, string) (*googlemaps.Place, error)
}

func (m *destinationContextTestMaps) FindCandidatesByQueryContext(ctx context.Context, query string) ([]googlemaps.Candidate, error) {
	return m.search(ctx, query)
}

func (m *destinationContextTestMaps) FindPlaceByIDContext(ctx context.Context, id string) (*googlemaps.Place, error) {
	return m.detail(ctx, id)
}

func destinationTestCandidate(id, name string, types ...string) googlemaps.Candidate {
	c := googlemaps.Candidate{PlaceID: id, Name: name, Types: types, FormattedAddress: "Tokyo, Japan"}
	c.Geometry.Location.Lat, c.Geometry.Location.Lng = 35.6762, 139.6503
	return c
}

func TestGoogleDestinationSearchUsesStructuredCountryAndCityTimeZone(t *testing.T) {
	maps := &destinationTestMaps{
		candidates: []googlemaps.Candidate{destinationTestCandidate(" tokyo ", " Tokyo ", "political")},
		details: map[string]*googlemaps.Place{"tokyo": {
			Types: []string{"political", "locality"},
			AddressComponents: []googlemaps.AddressComponent{
				{LongText: "Tokyo", ShortText: "13", Types: []string{"administrative_area_level_1"}},
				{LongText: " Japan ", ShortText: " jp ", Types: []string{"political", "country"}},
			},
			TimeZone: &googlemaps.PlaceTimeZone{ID: " Asia/Tokyo "},
		}},
	}
	rows, err := googleDestinationSearch(maps)(context.Background(), "Tokyo")
	if err != nil || len(rows) != 1 {
		t.Fatalf("search = %v, %v", rows, err)
	}
	got := rows[0]
	if got.Name != "Tokyo" || got.Provider != "google" || got.ProviderID != "tokyo" || got.Kind != "city" || got.CountryCode != "JP" || got.CountryName != "Japan" || got.TimeZone != "Asia/Tokyo" {
		t.Fatalf("incorrect structured destination: %+v", got)
	}
	if got.ID != nil || got.Latitude == nil || got.Longitude == nil || *got.Latitude != 35.6762 || *got.Longitude != 139.6503 {
		t.Fatalf("provider invented catalog identity or lost coordinates: %+v", got)
	}
}

func TestGoogleDestinationSearchDoesNotGuessCountryOrBroadAreaTimeZone(t *testing.T) {
	for _, tc := range []struct {
		name                        string
		types                       []string
		zone, countryCode, wantKind string
	}{
		{"country", []string{"political", "country"}, "Asia/Tokyo", "", "country"},
		{"region", []string{"administrative_area_level_1"}, "Asia/Tokyo", "", "region"},
		{"district", []string{"administrative_area_level_2"}, "Asia/Tokyo", "", "region"},
		{"invalid timezone", []string{"locality"}, "Not/AZone", "", "city"},
		{"machine timezone", []string{"postal_town"}, "Local", "", "city"},
		{"invalid country code", []string{"museum"}, "", "JPN", "place"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			maps := &destinationTestMaps{
				candidates: []googlemaps.Candidate{destinationTestCandidate("place", "Tokyo Japan", tc.types...)},
				details: map[string]*googlemaps.Place{"place": {
					TimeZone:          &googlemaps.PlaceTimeZone{ID: tc.zone},
					AddressComponents: []googlemaps.AddressComponent{{LongText: "Japan", ShortText: tc.countryCode, Types: []string{"country"}}},
				}},
			}
			rows, err := googleDestinationSearch(maps)(context.Background(), "Tokyo Japan")
			if err != nil || len(rows) != 1 {
				t.Fatalf("search = %v, %v", rows, err)
			}
			if rows[0].Kind != tc.wantKind || rows[0].TimeZone != "" || rows[0].CountryCode != "" || rows[0].CountryName != "" {
				t.Fatalf("unsupported geography was inferred: %+v", rows[0])
			}
		})
	}
}

func TestGoogleDestinationSearchRetainsBasicResultWhenDetailsUnavailable(t *testing.T) {
	for _, detailsErr := range []error{nil, errors.New("provider details unavailable")} {
		maps := &destinationTestMaps{candidates: []googlemaps.Candidate{destinationTestCandidate("tokyo", "Tokyo", "locality")}, detailsErr: detailsErr}
		rows, err := googleDestinationSearch(maps)(context.Background(), "Tokyo")
		if err != nil || len(rows) != 1 || rows[0].Name != "Tokyo" || rows[0].CountryCode != "" || rows[0].TimeZone != "" {
			t.Fatalf("basic candidate should survive details failure: %+v, %v", rows, err)
		}
	}
}

func TestGoogleDestinationSearchSkipsDuplicatesAndBoundsDetailsRequests(t *testing.T) {
	maps := &destinationTestMaps{candidates: []googlemaps.Candidate{
		destinationTestCandidate("", "missing ID"), destinationTestCandidate("blank", " "),
		destinationTestCandidate("one", "First"), destinationTestCandidate(" one ", "Duplicate"),
	}}
	for i := 0; i < 9; i++ {
		maps.candidates = append(maps.candidates, destinationTestCandidate(fmt.Sprint(i), fmt.Sprint("Place ", i)))
	}
	rows, err := googleDestinationSearch(maps)(context.Background(), "places")
	if err != nil || len(rows) != 5 || len(maps.detailIDs) != 5 || rows[0].ProviderID != "one" {
		t.Fatalf("lookup should have five distinct valid results and detail requests: %+v, %v, %v", rows, maps.detailIDs, err)
	}
}

func TestGoogleDestinationSearchPropagatesSearchErrorAndCancellation(t *testing.T) {
	want := errors.New("search unavailable")
	if _, err := googleDestinationSearch(&destinationTestMaps{searchErr: want})(context.Background(), "Tokyo"); !errors.Is(err, want) {
		t.Fatalf("search error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	maps := &destinationTestMaps{}
	if _, err := googleDestinationSearch(maps)(ctx, "Tokyo"); !errors.Is(err, context.Canceled) || len(maps.queries) != 0 {
		t.Fatalf("canceled search made a provider call: %v, %v", err, maps.queries)
	}
	if rows, err := googleDestinationSearch(nil)(context.Background(), "Tokyo"); err != nil || len(rows) != 0 {
		t.Fatalf("missing provider should leave catalog fallback available: %v, %v", rows, err)
	}
}

func TestGoogleDestinationSearchUsesRequestContextThroughDetails(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	legacy := &destinationTestMaps{}
	maps := &destinationContextTestMaps{
		destinationTestMaps: legacy,
		search: func(got context.Context, query string) ([]googlemaps.Candidate, error) {
			if got != ctx || query != "Tokyo" {
				t.Fatalf("search lost request context or query")
			}
			return []googlemaps.Candidate{destinationTestCandidate("tokyo", "Tokyo")}, nil
		},
		detail: func(got context.Context, id string) (*googlemaps.Place, error) {
			if got != ctx || id != "tokyo" {
				t.Fatalf("details lost request context or identity")
			}
			cancel()
			return nil, got.Err()
		},
	}
	_, err := googleDestinationSearch(maps)(ctx, "Tokyo")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("details cancellation was swallowed: %v", err)
	}
	if len(legacy.queries) != 0 || len(legacy.detailIDs) != 0 {
		t.Fatal("context-aware provider fell back to uncancellable methods")
	}
}
