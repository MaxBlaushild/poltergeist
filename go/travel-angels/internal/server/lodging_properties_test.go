package server

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
)

type lodgingMapsFake struct {
	googlemaps.Client
	place      *googlemaps.LodgingPlace
	err        error
	contexts   []context.Context
	queries    []string
	ids        []string
	photoNames []string
}

type nearbyLodgingMapsFake struct {
	*lodgingMapsFake
	latitude, longitude float64
	nearQueries         []string
}

func (m *nearbyLodgingMapsFake) SearchLodgingNearContext(ctx context.Context, query string, latitude, longitude float64) ([]googlemaps.LodgingPlace, error) {
	m.contexts = append(m.contexts, ctx)
	m.nearQueries = append(m.nearQueries, query)
	m.latitude, m.longitude = latitude, longitude
	if m.err != nil {
		return nil, m.err
	}
	return []googlemaps.LodgingPlace{*m.place}, nil
}

func TestLodgingAdapterSupportsOptionalSearchBias(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, "nearby")
	maps := &nearbyLodgingMapsFake{lodgingMapsFake: &lodgingMapsFake{place: &googlemaps.LodgingPlace{ID: "hotel", DisplayName: googlemaps.LocalizedText{Text: "Casa Lisboa"}}}}
	provider := googleLodgingPropertyProvider(maps).(*googleLodgingAdapter)
	rows, err := provider.SearchNear(ctx, "Casa Lisboa", 38.72, -9.14)
	if err != nil || len(rows) != 1 || rows[0].Provider != "google" || rows[0].PlaceID != "hotel" {
		t.Fatalf("nearby mapping: %v %v", rows, err)
	}
	if maps.latitude != 38.72 || maps.longitude != -9.14 || len(maps.nearQueries) != 1 || maps.nearQueries[0] != "Casa Lisboa" || maps.contexts[0] != ctx || len(maps.queries) != 0 {
		t.Fatal("adapter lost bias/context or also called unbiased search")
	}
	want := errors.New("nearby provider unavailable")
	maps.err = want
	if _, err := provider.SearchNear(ctx, "hotel", 0, 0); !errors.Is(err, want) || len(maps.queries) != 0 {
		t.Fatalf("biased failure was hidden or retried without bias: %v", err)
	}
}

func TestLodgingAdapterNearFallsBackForExistingProviders(t *testing.T) {
	maps := &lodgingMapsFake{place: &googlemaps.LodgingPlace{ID: "hotel", DisplayName: googlemaps.LocalizedText{Text: "Stay"}}}
	provider := googleLodgingPropertyProvider(maps).(*googleLodgingAdapter)
	rows, err := provider.SearchNear(context.Background(), "Stay", 0, 0)
	if err != nil || len(rows) != 1 || len(maps.queries) != 1 || maps.queries[0] != "Stay" {
		t.Fatalf("legacy fallback: %v %v", rows, err)
	}
	for _, point := range [][2]float64{{math.NaN(), 0}, {0, math.Inf(1)}, {91, 0}, {0, 181}} {
		if _, err := provider.SearchNear(context.Background(), "Stay", point[0], point[1]); err == nil {
			t.Fatal("invalid coordinates accepted by legacy fallback")
		}
	}
	if len(maps.queries) != 1 {
		t.Fatal("invalid bias still called provider")
	}
}

func (m *lodgingMapsFake) SearchLodgingContext(ctx context.Context, query string) ([]googlemaps.LodgingPlace, error) {
	m.contexts = append(m.contexts, ctx)
	m.queries = append(m.queries, query)
	if m.err != nil {
		return nil, m.err
	}
	if m.place == nil {
		return nil, nil
	}
	return []googlemaps.LodgingPlace{*m.place}, nil
}
func (m *lodgingMapsFake) GetLodgingContext(ctx context.Context, id string) (*googlemaps.LodgingPlace, error) {
	m.contexts = append(m.contexts, ctx)
	m.ids = append(m.ids, id)
	return m.place, m.err
}
func (m *lodgingMapsFake) GetLodgingPhotoContext(ctx context.Context, name string) ([]byte, string, error) {
	m.contexts = append(m.contexts, ctx)
	m.photoNames = append(m.photoNames, name)
	return []byte("image"), "image/jpeg", m.err
}

func TestLodgingAdapterPreservesAttributionAndHidesProviderPhotoName(t *testing.T) {
	maps := &lodgingMapsFake{place: &googlemaps.LodgingPlace{
		ID: "hotel", DisplayName: googlemaps.LocalizedText{Text: "Casa Lisboa"}, FormattedAddress: "Lisbon", Types: []string{"lodging"},
		WebsiteURI: "https://hotel.example", GoogleMapsURI: "https://maps.google.com/?cid=1", Rating: 4.7, UserRatingCount: 120,
		Attributions: []googlemaps.LodgingAttribution{{Provider: "Local data", ProviderURI: "https://data.example"}},
		Photos: []googlemaps.LodgingPhoto{{Name: "places/hotel/photos/photo", Width: 1600, Height: 1200, GoogleMapsURI: "https://maps.google.com/photo/1", AuthorAttributions: []googlemaps.LodgingAuthorAttribution{{DisplayName: "Author", URI: "//maps.google.com/contrib/1", PhotoURI: "//lh3.googleusercontent.com/avatar"}}},
			{Name: "places/other/photos/private"}},
	}}
	provider := googleLodgingPropertyProvider(maps)
	ctx := context.WithValue(context.Background(), struct{}{}, "caller")
	rows, err := provider.Search(ctx, "Casa Lisboa")
	if err != nil || len(rows) != 1 {
		t.Fatalf("search: %+v %v", rows, err)
	}
	row := rows[0]
	if row.Provider != "google" || row.PlaceID != "hotel" || row.Name != "Casa Lisboa" || row.Address != "Lisbon" || row.Rating != 4.7 || row.RatingCount != 120 || row.WebsiteURL != "https://hotel.example" || len(row.Attributions) != 1 || row.Attributions[0].DisplayName != "Local data" {
		t.Fatalf("property mapping lost content: %+v", row)
	}
	if len(row.Photos) != 1 || row.Photos[0].Name != "places/hotel/photos/photo" || row.Photos[0].URL != "" || row.Photos[0].GoogleMapsURI != "https://maps.google.com/photo/1" || row.Photos[0].Attributions[0].URI != "https://maps.google.com/contrib/1" || row.Photos[0].Attributions[0].PhotoURI != "https://lh3.googleusercontent.com/avatar" {
		t.Fatalf("photo mapping: %+v", row.Photos)
	}
	encoded, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "places/hotel/photos") || !strings.Contains(string(encoded), "authorAttributions") {
		t.Fatalf("provider photo resource leaked or attribution lost: %s", encoded)
	}
	if _, err = provider.Details(ctx, "hotel"); err != nil {
		t.Fatal(err)
	}
	photo, mime, err := provider.Photo(ctx, row.Photos[0].Name)
	if err != nil || string(photo) != "image" || mime != "image/jpeg" {
		t.Fatalf("photo: %s %s %v", photo, mime, err)
	}
	if len(maps.contexts) != 3 || maps.contexts[0] != ctx || maps.contexts[1] != ctx || maps.contexts[2] != ctx || maps.queries[0] != "Casa Lisboa" || maps.ids[0] != "hotel" || maps.photoNames[0] != "places/hotel/photos/photo" {
		t.Fatal("adapter lost request context or selected identity")
	}
}

func TestLodgingAdapterDoesNotRequireNewMethodsFromLegacyMaps(t *testing.T) {
	if googleLodgingPropertyProvider(nil) != nil || googleLodgingPropertyProvider(&destinationTestMaps{}) != nil {
		t.Fatal("legacy provider should leave lodging integration unconfigured")
	}
}

func TestLodgingAdapterPropagatesFailuresAndMissingProperty(t *testing.T) {
	want := errors.New("provider unavailable")
	maps := &lodgingMapsFake{err: want}
	provider := googleLodgingPropertyProvider(maps)
	if _, err := provider.Search(context.Background(), "hotel"); !errors.Is(err, want) {
		t.Fatalf("search: %v", err)
	}
	if _, err := provider.Details(context.Background(), "hotel"); !errors.Is(err, want) {
		t.Fatalf("details: %v", err)
	}
	if _, _, err := provider.Photo(context.Background(), "places/hotel/photos/p"); !errors.Is(err, want) {
		t.Fatalf("photo: %v", err)
	}
	maps.err = nil
	if _, err := provider.Details(context.Background(), "hotel"); err == nil {
		t.Fatal("nil property accepted")
	}
	maps.place = &googlemaps.LodgingPlace{ID: "other"}
	if _, err := provider.Details(context.Background(), "hotel"); err == nil {
		t.Fatal("mismatched property accepted")
	}
}

func TestLodgingAdapterRejectsExecutableAndCredentialLinks(t *testing.T) {
	for _, raw := range []string{"javascript:alert(1)", "data:text/html,hello", "file:///private/data", "https://user:secret@hotel.example", "/relative"} {
		if got := lodgingSourceURL(raw); got != "" {
			t.Fatalf("unsafe source link retained: %q", got)
		}
	}
	if got := lodgingSourceURL(" //maps.google.com/contrib/1 "); got != "https://maps.google.com/contrib/1" {
		t.Fatalf("scheme-relative attribution = %q", got)
	}
}
