// Mock Google Maps client for testing purposes.
package mocks

import (
	"context"

	"googlemaps.github.io/maps"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
)

// MockGoogleMapsClient is a mock implementation of the Google Maps client.
type MockGoogleMapsClient struct {
	GetPlaceDetailsFunc  func(ctx context.Context, placeID string) (*googlemaps.Place, error)
	SearchNearbyFunc     func(ctx context.Context, req *maps.NearbySearchRequest) ([]googlemaps.Candidate, error)
	GetPlacePhotoFunc    func(ctx context.Context, photoReference string, maxHeight, maxWidth uint) (*maps.PlacePhotoResponse, error)
	GetStaticMapFunc     func(ctx context.Context, req *maps.StaticMapRequest) ([]byte, error)
	GeocodeFunc          func(ctx context.Context, req *maps.GeocodingRequest) ([]maps.GeocodingResult, error)
	// ReverseGeocodeFunc   func(ctx context.Context, req *maps.ReverseGeocodingRequest) ([]maps.GeocodingResult, error) // Commented out
	AutocompleteFunc     func(ctx context.Context, req *maps.PlaceAutocompleteRequest) ([]maps.AutocompletePrediction, error)
	// QueryAutocompleteFunc func(ctx context.Context, req *maps.QueryAutocompleteRequest) ([]maps.QueryAutocompletePrediction, error) // Commented out
}

// GetPlaceDetails mocks the GetPlaceDetails operation.
func (m *MockGoogleMapsClient) GetPlaceDetails(ctx context.Context, placeID string) (*googlemaps.Place, error) {
	if m.GetPlaceDetailsFunc != nil {
		return m.GetPlaceDetailsFunc(ctx, placeID)
	}
	return &googlemaps.Place{}, nil
}

// SearchNearby mocks the SearchNearby operation.
func (m *MockGoogleMapsClient) SearchNearby(ctx context.Context, req *maps.NearbySearchRequest) ([]googlemaps.Candidate, error) {
	if m.SearchNearbyFunc != nil {
		return m.SearchNearbyFunc(ctx, req)
	}
	return []googlemaps.Candidate{}, nil
}

// GetPlacePhoto mocks the GetPlacePhoto operation.
func (m *MockGoogleMapsClient) GetPlacePhoto(ctx context.Context, photoReference string, maxHeight, maxWidth uint) (*maps.PlacePhotoResponse, error) {
	if m.GetPlacePhotoFunc != nil {
		return m.GetPlacePhotoFunc(ctx, photoReference, maxHeight, maxWidth)
	}
	return &maps.PlacePhotoResponse{}, nil
}

// GetStaticMap mocks the GetStaticMap operation.
func (m *MockGoogleMapsClient) GetStaticMap(ctx context.Context, req *maps.StaticMapRequest) ([]byte, error) {
	if m.GetStaticMapFunc != nil {
		return m.GetStaticMapFunc(ctx, req)
	}
	return []byte{}, nil
}

// Geocode mocks the Geocode operation.
func (m *MockGoogleMapsClient) Geocode(ctx context.Context, req *maps.GeocodingRequest) ([]maps.GeocodingResult, error) {
	if m.GeocodeFunc != nil {
		return m.GeocodeFunc(ctx, req)
	}
	return []maps.GeocodingResult{}, nil
}

/* // Commented out
// ReverseGeocode mocks the ReverseGeocode operation.
func (m *MockGoogleMapsClient) ReverseGeocode(ctx context.Context, req *maps.ReverseGeocodingRequest) ([]maps.GeocodingResult, error) {
	if m.ReverseGeocodeFunc != nil {
		return m.ReverseGeocodeFunc(ctx, req)
	}
	return []maps.GeocodingResult{}, nil
}
*/

// Autocomplete mocks the Autocomplete operation.
func (m *MockGoogleMapsClient) Autocomplete(ctx context.Context, req *maps.PlaceAutocompleteRequest) ([]maps.AutocompletePrediction, error) {
	if m.AutocompleteFunc != nil {
		return m.AutocompleteFunc(ctx, req)
	}
	return []maps.AutocompletePrediction{}, nil
}

/* // Commented out
// QueryAutocomplete mocks the QueryAutocomplete operation.
func (m *MockGoogleMapsClient) QueryAutocomplete(ctx context.Context, req *maps.QueryAutocompleteRequest) ([]maps.QueryAutocompletePrediction, error) {
	if m.QueryAutocompleteFunc != nil {
		return m.QueryAutocompleteFunc(ctx, req)
	}
	return []maps.QueryAutocompletePrediction{}, nil
}
*/
