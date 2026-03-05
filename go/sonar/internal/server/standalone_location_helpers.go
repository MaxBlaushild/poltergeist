package server

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func parseStandalonePointOfInterestID(raw string) (*uuid.UUID, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid pointOfInterestId")
	}
	return &parsed, nil
}

func parsePointOfInterestCoordinates(poi *models.PointOfInterest) (float64, float64, error) {
	if poi == nil {
		return 0, 0, fmt.Errorf("point of interest not found")
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(poi.Lat), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("point of interest has invalid latitude")
	}
	lng, err := strconv.ParseFloat(strings.TrimSpace(poi.Lng), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("point of interest has invalid longitude")
	}
	if !isValidStandaloneCoordinate(lat, lng) {
		return 0, 0, fmt.Errorf("point of interest has invalid coordinates")
	}
	return lat, lng, nil
}

func isValidStandaloneCoordinate(latitude float64, longitude float64) bool {
	if math.IsNaN(latitude) || math.IsInf(latitude, 0) {
		return false
	}
	if math.IsNaN(longitude) || math.IsInf(longitude, 0) {
		return false
	}
	if latitude < -90 || latitude > 90 {
		return false
	}
	if longitude < -180 || longitude > 180 {
		return false
	}
	return true
}

func (s *server) resolveStandaloneLocation(
	ctx context.Context,
	expectedZoneID *uuid.UUID,
	pointOfInterestID *uuid.UUID,
	latitude float64,
	longitude float64,
) (*uuid.UUID, float64, float64, error) {
	if pointOfInterestID == nil {
		if !isValidStandaloneCoordinate(latitude, longitude) {
			return nil, 0, 0, fmt.Errorf("latitude/longitude must be valid coordinates")
		}
		return nil, latitude, longitude, nil
	}

	poi, err := s.dbClient.PointOfInterest().FindByID(ctx, *pointOfInterestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, 0, fmt.Errorf("pointOfInterestId not found")
		}
		return nil, 0, 0, err
	}

	resolvedLat, resolvedLng, err := parsePointOfInterestCoordinates(poi)
	if err != nil {
		return nil, 0, 0, err
	}
	if expectedZoneID != nil {
		zone, zoneErr := s.dbClient.Zone().FindByPointOfInterestID(ctx, *pointOfInterestID)
		if zoneErr != nil {
			if errors.Is(zoneErr, gorm.ErrRecordNotFound) {
				return nil, 0, 0, fmt.Errorf("pointOfInterestId is not associated with a zone")
			}
			return nil, 0, 0, zoneErr
		}
		if zone.ID != *expectedZoneID {
			return nil, 0, 0, fmt.Errorf("pointOfInterestId does not belong to zoneId")
		}
	}

	return pointOfInterestID, resolvedLat, resolvedLng, nil
}
