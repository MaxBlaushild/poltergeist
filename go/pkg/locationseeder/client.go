package locationseeder

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type client struct {
	googlemapsClient googlemaps.Client
	dbClient         db.DbClient
	deepPriest       deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

type FantasyPointOfInterest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Clue        string `json:"clue"`
	Challenge   string `json:"challenge"`
}

type Client interface {
	GeneratePointOfInterest(ctx context.Context, place googlemaps.Place, zone *models.Zone) (*models.PointOfInterest, error)
	SeedPointsOfInterest(ctx context.Context, zone models.Zone, includedTypes []googlemaps.PlaceType, excludedTypes []googlemaps.PlaceType, numberOfPlaces int32) ([]*models.PointOfInterest, error)
	RefreshPointOfInterestImage(ctx context.Context, poi *models.PointOfInterest) error
	RefreshPointOfInterest(ctx context.Context, poi *models.PointOfInterest) error
	ImportPlace(ctx context.Context, placeID string, zone models.Zone) (*models.PointOfInterest, error)
}

func NewClient(
	googlemapsClient googlemaps.Client,
	dbClient db.DbClient,
	deepPriest deep_priest.DeepPriest,
	awsClient aws.AWSClient,
) Client {
	log.Println("Creating new locationseeder client")
	return &client{
		googlemapsClient: googlemapsClient,
		dbClient:         dbClient,
		deepPriest:       deepPriest,
		awsClient:        awsClient,
	}
}

func (c *client) ImportPlace(ctx context.Context, placeID string, zone models.Zone) (*models.PointOfInterest, error) {
	place, err := c.googlemapsClient.FindPlaceByID(placeID)
	if err != nil {
		log.Printf("Error finding place by ID: %v", err)
		return nil, err
	}

	poi, err := c.GeneratePointOfInterest(ctx, *place, &zone)
	if err != nil {
		log.Printf("Error generating point of interest: %v", err)
		return nil, err
	}

	return poi, nil
}

func (c *client) RefreshPointOfInterest(ctx context.Context, poi *models.PointOfInterest) error {
	if poi == nil || poi.GoogleMapsPlaceID == nil {
		log.Printf("Point of interest %s has no Google Maps place ID", poi.Name)
		return fmt.Errorf("point of interest has no Google Maps place ID")
	}

	zone, err := c.dbClient.Zone().FindByPointOfInterestID(ctx, poi.ID)
	if err != nil {
		log.Printf("Error finding zone by ID: %v", err)
		return err
	}

	place, err := c.googlemapsClient.FindPlaceByID(*poi.GoogleMapsPlaceID)
	if err != nil {
		log.Printf("Error finding place by ID: %v", err)
		return err
	}

	fantasyPointOfInterest, err := c.generateFantasyTheming(*place, zone)
	if err != nil {
		log.Printf("Error generating fantasy theming: %v", err)
		return err
	}

	imageUrl, err := c.generateFantasyImage(ctx, *place, zone)
	if err != nil {
		log.Printf("Error generating fantasy image: %v", err)
		return err
	}

	if err := c.dbClient.PointOfInterest().Update(ctx, poi.ID, &models.PointOfInterest{
		Name:         fantasyPointOfInterest.Name,
		Description:  fantasyPointOfInterest.Description,
		Clue:         fantasyPointOfInterest.Clue,
		ImageUrl:     imageUrl,
		OriginalName: place.DisplayName.Text,
		Geometry:     poi.Geometry,
		UpdatedAt:    time.Now(),
	}); err != nil {
		log.Printf("Error updating point of interest: %v", err)
		return err
	}

	tags, err := c.ProccessPlaceTypes(ctx, place.Types)
	if err != nil {
		log.Printf("Error processing place types: %v", err)
		return err
	}

	for _, tag := range tags {
		log.Printf("Adding tag %s to point of interest", tag.Value)
		if err := c.dbClient.Tag().AddTagToPointOfInterest(ctx, tag.ID, poi.ID); err != nil {
			log.Printf("Error adding tag to point of interest: %v", err)
			return err
		}

		poi.Tags = append(poi.Tags, *tag)
	}

	log.Printf("Successfully updated point of interest %s", poi.Name)
	return nil
}

func (c *client) RefreshPointOfInterestImage(ctx context.Context, poi *models.PointOfInterest) error {
	if poi == nil || poi.GoogleMapsPlaceID == nil {
		log.Printf("Point of interest %s has no Google Maps place ID", poi.Name)
		return fmt.Errorf("point of interest has no Google Maps place ID")
	}

	zone, err := c.dbClient.Zone().FindByPointOfInterestID(ctx, poi.ID)
	if err != nil {
		log.Printf("Error finding zone by ID: %v", err)
		return err
	}

	place, err := c.googlemapsClient.FindPlaceByID(*poi.GoogleMapsPlaceID)
	if err != nil {
		log.Printf("Error finding place by ID: %v", err)

		return err
	}

	if place == nil {
		log.Printf("Place not found")
		return fmt.Errorf("place not found")
	}

	imageUrl, err := c.generateFantasyImage(ctx, *place, zone)
	if err != nil {
		log.Printf("Error generating fantasy image: %v", err)
		return err
	}

	if err := c.dbClient.PointOfInterest().UpdateImageUrl(ctx, poi.ID, imageUrl); err != nil {
		log.Printf("Error updating point of interest image URL: %v", err)
		return err
	}

	log.Printf("Successfully updated point of interest image URL for %s", poi.Name)

	return nil
}

// haversineDistance calculates the distance in meters between two lat/lng points
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371000 // Earth's radius in meters

	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLng := (lng2 - lng1) * math.Pi / 180.0

	lat1Rad := lat1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLng/2)*math.Sin(dLng/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// selectPlaceByDistanceWeight selects a place from the list using distance-based weighted random selection.
// Places closer to the search center have exponentially higher probability of being selected.
func selectPlaceByDistanceWeight(places []googlemaps.Place, centerLat, centerLng, radius float64) *googlemaps.Place {
	if len(places) == 0 {
		return nil
	}

	if len(places) == 1 {
		return &places[0]
	}

	// Calculate weights based on distance (closer = higher weight)
	weights := make([]float64, len(places))
	totalWeight := 0.0

	for i, place := range places {
		distance := haversineDistance(centerLat, centerLng, place.Location.Latitude, place.Location.Longitude)
		// Use exponential decay: weight = e^(-distance/radius)
		// This gives much higher weight to closer places
		weight := math.Exp(-distance / radius)
		weights[i] = weight
		totalWeight += weight
	}

	// Select randomly based on cumulative probability
	randValue := rand.Float64() * totalWeight
	cumulative := 0.0

	for i, weight := range weights {
		cumulative += weight
		if randValue <= cumulative {
			return &places[i]
		}
	}

	// Fallback (should rarely happen due to floating point precision)
	return &places[len(places)-1]
}

func (c *client) GetPlacesInZone(ctx context.Context, zone models.Zone, includedTypes []googlemaps.PlaceType, excludedTypes []googlemaps.PlaceType, numberOfPlaces int32) ([]googlemaps.Place, error) {
	var placesInZone []googlemaps.Place
	var recentlyUsedPlaces []googlemaps.Place
	seenPlaceIDs := make(map[string]bool)
	attempts := 0
	maxAttempts := 20
	totalPlacesFound := 0
	totalPlacesInBoundary := 0

	// Request 3x the needed places to account for boundary filtering and weighted selection
	requestCount := numberOfPlaces * 3
	if requestCount > 20 {
		requestCount = 20 // Google Maps API typically limits to 20 results
	}

	// Get recently used places with fallback windows (7 days, 3 days, 1 day)
	exclusionWindows := []time.Duration{
		7 * 24 * time.Hour,
		3 * 24 * time.Hour,
		1 * 24 * time.Hour,
	}

	var recentlyUsed map[string]bool
	var err error
	for _, window := range exclusionWindows {
		since := time.Now().Add(-window)
		recentlyUsed, err = c.dbClient.PointOfInterest().FindRecentlyUsedInZone(ctx, zone.ID, since)
		if err != nil {
			log.Printf("Error finding recently used places: %v", err)
			recentlyUsed = make(map[string]bool)
			break
		}

		// If we have few recently used places, this window is good
		if len(recentlyUsed) < 10 {
			log.Printf("Using %d day exclusion window, found %d recently used places", int(window.Hours()/24), len(recentlyUsed))
			break
		}
	}

	log.Printf("Starting search for %d places (requesting %d per attempt, max %d attempts, excluding %d recently used)",
		numberOfPlaces, requestCount, maxAttempts, len(recentlyUsed))

	for attempts < maxAttempts && int32(len(placesInZone)) < numberOfPlaces {
		randomPoint := zone.GetRandomPoint()
		centerLat := randomPoint.Y()
		centerLng := randomPoint.X()

		log.Printf("Attempt %d/%d - Searching at lat: %f, lng: %f, radius: %f", attempts+1, maxAttempts, centerLat, centerLng, zone.Radius)

		places, err := c.googlemapsClient.FindPlaces(googlemaps.PlaceQuery{
			Lat:            centerLat,
			Long:           centerLng,
			Radius:         1000,
			IncludedTypes:  includedTypes,
			ExcludedTypes:  excludedTypes,
			MaxResultCount: requestCount,
			RankPreference: googlemaps.RankPreferenceDistance,
		})
		if err != nil {
			log.Printf("Error finding places on attempt %d: %v", attempts+1, err)
			return nil, err
		}

		totalPlacesFound += len(places)
		newPlacesThisAttempt := 0
		duplicatesThisAttempt := 0
		recentlyUsedSkipped := 0

		// Filter places to only include valid candidates
		var validPlaces []googlemaps.Place
		for _, place := range places {
			log.Printf("Place: %s", place.Name)
			// Skip if we've already seen this place
			if seenPlaceIDs[place.ID] {
				duplicatesThisAttempt++
				continue
			}

			// Skip if recently used in a quest, but store it for potential fallback
			if recentlyUsed[place.ID] {
				recentlyUsedSkipped++
				// Store this place in case we need it later
				if zone.IsPointInBoundary(place.Location.Latitude, place.Location.Longitude) {
					recentlyUsedPlaces = append(recentlyUsedPlaces, place)
				}
				continue
			}

			// Check if place is within zone boundary
			if zone.IsPointInBoundary(place.Location.Latitude, place.Location.Longitude) {
				validPlaces = append(validPlaces, place)
			}
		}

		// Use weighted random selection to pick from valid places
		for int32(len(placesInZone)) < numberOfPlaces && len(validPlaces) > 0 {
			selectedPlace := selectPlaceByDistanceWeight(validPlaces, centerLat, centerLng, zone.Radius)
			if selectedPlace == nil {
				break
			}

			seenPlaceIDs[selectedPlace.ID] = true
			placesInZone = append(placesInZone, *selectedPlace)
			totalPlacesInBoundary++
			newPlacesThisAttempt++

			// Remove the selected place from valid places
			newValidPlaces := make([]googlemaps.Place, 0, len(validPlaces)-1)
			for _, p := range validPlaces {
				if p.ID != selectedPlace.ID {
					newValidPlaces = append(newValidPlaces, p)
				}
			}
			validPlaces = newValidPlaces
		}

		log.Printf("Attempt %d: Found %d API results, %d new unique places added (%d duplicates, %d recently used, %d outside boundary). Total progress: %d/%d",
			attempts+1, len(places), newPlacesThisAttempt, duplicatesThisAttempt, recentlyUsedSkipped,
			len(places)-newPlacesThisAttempt-duplicatesThisAttempt-recentlyUsedSkipped, len(placesInZone), numberOfPlaces)

		if int32(len(placesInZone)) >= numberOfPlaces {
			log.Printf("Success! Found %d unique places in zone after %d attempts (total API results: %d, in boundary: %d, duplicates: %d, recently used filtered: %d)",
				len(placesInZone), attempts+1, totalPlacesFound, totalPlacesInBoundary, len(seenPlaceIDs)-totalPlacesInBoundary, recentlyUsedSkipped)
			return placesInZone, nil
		}

		attempts++
	}

	// If we still don't have enough places, try using recently used ones
	if int32(len(placesInZone)) < numberOfPlaces {
		needed := int(numberOfPlaces) - len(placesInZone)
		log.Printf("Could not find enough fresh places. Need %d more. Attempting to use recently used places (%d available)", needed, len(recentlyUsedPlaces))

		if len(recentlyUsedPlaces) == 0 {
			return nil, fmt.Errorf("could not find enough places in zone after %d attempts. Found %d/%d places and no recently used places available as fallback", attempts, len(placesInZone), numberOfPlaces)
		}

		// Randomly shuffle and select from recently used places
		rand.Shuffle(len(recentlyUsedPlaces), func(i, j int) {
			recentlyUsedPlaces[i], recentlyUsedPlaces[j] = recentlyUsedPlaces[j], recentlyUsedPlaces[i]
		})

		for i := 0; i < needed && i < len(recentlyUsedPlaces); i++ {
			if !seenPlaceIDs[recentlyUsedPlaces[i].ID] {
				placesInZone = append(placesInZone, recentlyUsedPlaces[i])
				seenPlaceIDs[recentlyUsedPlaces[i].ID] = true
			}
		}

		log.Printf("Added %d recently used places as fallback. Total: %d/%d", len(placesInZone)-int(numberOfPlaces)+needed, len(placesInZone), numberOfPlaces)
	}

	log.Printf("Found %d places in zone after %d attempts (total API results: %d)", len(placesInZone), attempts, totalPlacesFound)
	return placesInZone, nil
}

func (c *client) SeedPointsOfInterest(ctx context.Context, zone models.Zone, includedTypes []googlemaps.PlaceType, excludedTypes []googlemaps.PlaceType, numberOfPlaces int32) ([]*models.PointOfInterest, error) {
	log.Printf("Starting to seed points of interest for zone %s with included types %v and excluded types %v", zone.Name, includedTypes, excludedTypes)

	randomPoint := zone.GetRandomPoint()
	log.Printf("Fuzzed latitude: %f, longitude: %f, radius: %f", randomPoint.X(), randomPoint.Y(), zone.Radius)

	places, err := c.GetPlacesInZone(ctx, zone, includedTypes, excludedTypes, numberOfPlaces)
	if err != nil {
		log.Printf("Error finding places: %v", err)
		return nil, err
	}

	log.Printf("Found %d places to convert to points of interest", len(places))

	var pointsOfInterest []*models.PointOfInterest
	for i, place := range places {
		log.Printf("Generating point of interest %d/%d for place: %s", i+1, len(places), place.Name)

		existingPointOfInterest, err := c.dbClient.PointOfInterest().FindByGoogleMapsPlaceID(ctx, place.ID)
		if err != nil {
			log.Printf("Error checking if point of interest has been imported: %v", err)
			return nil, err
		}
		if existingPointOfInterest != nil {
			log.Printf("Point of interest %s has already been imported", place.Name)
			pointsOfInterest = append(pointsOfInterest, existingPointOfInterest)
			continue
		}

		poi, err := c.GeneratePointOfInterest(ctx, place, &zone)
		if err != nil {
			log.Printf("Error generating point of interest for place %s: %v", place.Name, err)
			return nil, err
		}
		pointsOfInterest = append(pointsOfInterest, poi)
	}

	log.Printf("Successfully generated %d points of interest", len(pointsOfInterest))
	return pointsOfInterest, nil
}

func (c *client) GeneratePointOfInterest(ctx context.Context, place googlemaps.Place, zone *models.Zone) (*models.PointOfInterest, error) {
	placeDetails, err := c.googlemapsClient.FindPlaceByID(place.ID)
	if err != nil {
		log.Printf("Error getting place details: %v", err)
		return nil, err
	}

	if placeDetails == nil {
		log.Printf("Place details not found")
		return nil, fmt.Errorf("place details not found")
	}

	place = *placeDetails

	log.Printf("Starting to generate point of interest for place: %s", place.Name)

	fantasyPointOfInterest, err := c.generateFantasyTheming(place, zone)
	if err != nil {
		log.Printf("Error generating fantasy theming: %v", err)
		return nil, err
	}
	log.Printf("Generated fantasy theming with name: %s", fantasyPointOfInterest.Name)

	imageUrl, err := c.generateFantasyImage(ctx, place, zone)
	if err != nil {
		log.Printf("Error generating fantasy image: %v", err)
		return nil, err
	}
	log.Printf("Generated fantasy image URL: %s", imageUrl)

	poi := &models.PointOfInterest{
		ID:                uuid.New(),
		Name:              fantasyPointOfInterest.Name,
		OriginalName:      place.DisplayName.Text,
		Description:       fantasyPointOfInterest.Description,
		Clue:              fantasyPointOfInterest.Clue,
		ImageUrl:          imageUrl,
		GoogleMapsPlaceID: &place.ID,
		Lat:               strconv.FormatFloat(place.Location.Latitude, 'f', -1, 64),
		Lng:               strconv.FormatFloat(place.Location.Longitude, 'f', -1, 64),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	log.Printf("Created point of interest object with ID: %s", poi.ID)

	tags, err := c.ProccessPlaceTypes(ctx, place.Types)
	if err != nil {
		log.Printf("Error processing place types: %v", err)
		return nil, err
	}

	if err := c.dbClient.PointOfInterest().Create(ctx, *poi); err != nil {
		log.Printf("Error creating point of interest in database: %v", err)
		return nil, err
	}
	log.Printf("Successfully created point of interest in database")

	for _, tag := range tags {
		log.Printf("Adding tag %s to point of interest", tag.Value)
		if err := c.dbClient.Tag().AddTagToPointOfInterest(ctx, tag.ID, poi.ID); err != nil {
			log.Printf("Error adding tag to point of interest: %v", err)
			return nil, err
		}

		poi.Tags = append(poi.Tags, *tag)
	}

	if err := c.dbClient.Zone().AddPointOfInterestToZone(ctx, zone.ID, poi.ID); err != nil {
		log.Printf("Error adding point of interest to zone: %v", err)
		return nil, err
	}

	log.Printf("Successfully generated point of interest with %d tags", len(poi.Tags))
	return poi, nil
}
