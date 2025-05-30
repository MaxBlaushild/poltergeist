// Mock LocationSeeder client for testing purposes.
package mocks

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/locationseeder"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

// MockLocationSeederClient is a mock implementation of the locationseeder.Client interface.
type MockLocationSeederClient struct {
	GeneratePointOfInterestFunc   func(ctx context.Context, place googlemaps.Place, zone *models.Zone) (*models.PointOfInterest, error)
	SeedPointsOfInterestFunc      func(ctx context.Context, zone models.Zone, includedTypes []googlemaps.PlaceType, excludedTypes []googlemaps.PlaceType, numberOfPlaces int32) ([]*models.PointOfInterest, error)
	RefreshPointOfInterestImageFunc func(ctx context.Context, poi *models.PointOfInterest) error
	RefreshPointOfInterestFunc    func(ctx context.Context, poi *models.PointOfInterest) error
	ImportPlaceFunc               func(ctx context.Context, placeID string, zone models.Zone) (*models.PointOfInterest, error)
}

// GeneratePointOfInterest mocks the GeneratePointOfInterest operation.
func (m *MockLocationSeederClient) GeneratePointOfInterest(ctx context.Context, place googlemaps.Place, zone *models.Zone) (*models.PointOfInterest, error) {
	if m.GeneratePointOfInterestFunc != nil {
		return m.GeneratePointOfInterestFunc(ctx, place, zone)
	}
	return &models.PointOfInterest{}, nil
}

// SeedPointsOfInterest mocks the SeedPointsOfInterest operation.
func (m *MockLocationSeederClient) SeedPointsOfInterest(ctx context.Context, zone models.Zone, includedTypes []googlemaps.PlaceType, excludedTypes []googlemaps.PlaceType, numberOfPlaces int32) ([]*models.PointOfInterest, error) {
	if m.SeedPointsOfInterestFunc != nil {
		return m.SeedPointsOfInterestFunc(ctx, zone, includedTypes, excludedTypes, numberOfPlaces)
	}
	return []*models.PointOfInterest{}, nil
}

// RefreshPointOfInterestImage mocks the RefreshPointOfInterestImage operation.
func (m *MockLocationSeederClient) RefreshPointOfInterestImage(ctx context.Context, poi *models.PointOfInterest) error {
	if m.RefreshPointOfInterestImageFunc != nil {
		return m.RefreshPointOfInterestImageFunc(ctx, poi)
	}
	return nil
}

// RefreshPointOfInterest mocks the RefreshPointOfInterest operation.
func (m *MockLocationSeederClient) RefreshPointOfInterest(ctx context.Context, poi *models.PointOfInterest) error {
	if m.RefreshPointOfInterestFunc != nil {
		return m.RefreshPointOfInterestFunc(ctx, poi)
	}
	return nil
}

// ImportPlace mocks the ImportPlace operation.
func (m *MockLocationSeederClient) ImportPlace(ctx context.Context, placeID string, zone models.Zone) (*models.PointOfInterest, error) {
	if m.ImportPlaceFunc != nil {
		return m.ImportPlaceFunc(ctx, placeID, zone)
	}
	return &models.PointOfInterest{}, nil
}

// Ensure MockLocationSeederClient implements locationseeder.Client
var _ locationseeder.Client = &MockLocationSeederClient{}
