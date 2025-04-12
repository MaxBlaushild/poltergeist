package dungeonmaster

import (
	"context"
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

type client struct {
	googlemapsClient googlemaps.Client
	dbClient         db.Client
	deepPriest       deep_priest.DeepPriest
}

const (
	GeneratePointOfInterestPromptTemplate = `
	You are a dungeon master on a gameshow. You are tasked with generating a prompt for a quest that would be interesting to complete. The prompt should be a single sentence that captures the essence of the place.

	Here is the challenge: %s

	%s

	%s

	Please answer in the form of a JSON object with the following fields:
	
		{
			"name": "string",
			"description": "string"
		}
	`
)

type Client interface {
	GenerateQuests() ([]*models.PointOfInterest, error)
}

func NewClient(googlemapsClient googlemaps.Client, dbClient db.Client, deepPriest deep_priest.DeepPriest) Client {
	return &client{
		googlemapsClient: googlemapsClient,
		dbClient:         dbClient,
		deepPriest:       deepPriest,
	}
}

func (c *client) SeedPointsOfInterest(ctx context.Context, zone models.Zone, locationType string) ([]*models.PointOfInterest, error) {
	places, err := c.googlemapsClient.FindPlaces(ctx, googlemaps.PlaceQuery{
		Lat:      zone.Latitude,
		Long:     zone.Longitude,
		Radius:   1000,
		Category: locationType,
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (c *client) generatePointOfInterest(ctx context.Context, place googlemaps.Place) (*models.PointOfInterest, error) {
	answer, err := c.deepPriest.ConsultTheDeep(ctx, deep_priest.Question{
		Question: place.Name,
	})
	if err != nil {
		return nil, err
	}

	return &models.PointOfInterest{
		Name:        place.Name,
		Description: answer.Answer,
	}, nil
}

func (c *client) makePrompt(place googlemaps.Place) string {
	return fmt.Sprintf("The following is a description of a place: %s. Please generate a prompt for a quest that would be interesting to complete. The prompt should be a single sentence that captures the essence of the place.", place.Name)
}

func (c *client) GenerateQuests(ctx context.Context) ([]*models.PointOfInterest, error) {
	zones, err := c.dbClient.Zone().FindAll(ctx)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
