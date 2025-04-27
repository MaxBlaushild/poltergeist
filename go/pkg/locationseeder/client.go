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
	GeneratePointOfInterest(ctx context.Context, place googlemaps.Place, zone models.Zone) (*models.PointOfInterest, error)
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

	poi, err := c.GeneratePointOfInterest(ctx, *place, zone)
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

	place, err := c.googlemapsClient.FindPlaceByID(*poi.GoogleMapsPlaceID)
	if err != nil {
		log.Printf("Error finding place by ID: %v", err)
		return err
	}

	fantasyPointOfInterest, err := c.generateFantasyTheming(*place)
	if err != nil {
		log.Printf("Error generating fantasy theming: %v", err)
		return err
	}

	imageUrl, err := c.generateFantasyImage(ctx, *place)
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

	log.Printf("Refreshing point of interest image for %s", poi.Name)

	place, err := c.googlemapsClient.FindPlaceByID(*poi.GoogleMapsPlaceID)
	if err != nil {
		log.Printf("Error finding place by ID: %v", err)

		return err
	}

	if place == nil {
		log.Printf("Place not found")
		return fmt.Errorf("place not found")
	}

	imageUrl, err := c.generateFantasyImage(ctx, *place)
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

func (c *client) SeedPointsOfInterest(ctx context.Context, zone models.Zone, includedTypes []googlemaps.PlaceType, excludedTypes []googlemaps.PlaceType, numberOfPlaces int32) ([]*models.PointOfInterest, error) {
	log.Printf("Starting to seed points of interest for zone %s with included types %v and excluded types %v", zone.Name, includedTypes, excludedTypes)

	log.Printf("Zone latitude: %f, longitude: %f, radius: %d", zone.Latitude, zone.Longitude, int(zone.Radius))

	lat, lng, radius := c.fuzzCoordinates(zone.Latitude, zone.Longitude, zone.Radius)
	log.Printf("Fuzzed latitude: %f, longitude: %f, radius: %f", lat, lng, radius)

	places, err := c.googlemapsClient.FindPlaces(googlemaps.PlaceQuery{
		Lat:            lat,
		Long:           lng,
		Radius:         radius,
		IncludedTypes:  includedTypes,
		ExcludedTypes:  excludedTypes,
		MaxResultCount: numberOfPlaces,
	})
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

		poi, err := c.GeneratePointOfInterest(ctx, place, zone)
		if err != nil {
			log.Printf("Error generating point of interest for place %s: %v", place.Name, err)
			return nil, err
		}
		pointsOfInterest = append(pointsOfInterest, poi)
	}

	log.Printf("Successfully generated %d points of interest", len(pointsOfInterest))
	return pointsOfInterest, nil
}

func (c *client) fuzzCoordinates(lat float64, lng float64, radius float64) (float64, float64, float64) {
	// Convert radius from meters to degrees (approximate)
	radiusDegrees := radius / 111000.0 // 111km per degree

	// Generate random angle and distance within radius
	angle := rand.Float64() * 2 * math.Pi
	distance := rand.Float64() * radiusDegrees

	// Calculate new coordinates
	newLat := lat + (distance * math.Cos(angle))
	newLng := lng + (distance * math.Sin(angle))

	// Calculate what percentage of the radius was used
	return newLat, newLng, radius
}

func (c *client) GeneratePointOfInterest(ctx context.Context, place googlemaps.Place, zone models.Zone) (*models.PointOfInterest, error) {
	log.Printf("Starting to generate point of interest for place: %s", place.Name)

	fantasyPointOfInterest, err := c.generateFantasyTheming(place)
	if err != nil {
		log.Printf("Error generating fantasy theming: %v", err)
		return nil, err
	}
	log.Printf("Generated fantasy theming with name: %s", fantasyPointOfInterest.Name)

	imageUrl, err := c.generateFantasyImage(ctx, place)
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
