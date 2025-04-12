package locationseeder

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

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
}

type FantasyPointOfInterest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Clue        string `json:"clue"`
	Challenge   string `json:"challenge"`
}

const premise = `
	You are a video game designer. You are tasked with converting real world locations into points of interest on fantasy role playing game's map.
	These points of interest should be analogous to the real world location, but should be in a fantasy setting.
	For example, if the real world location is a cafe, the point of interest should be a tavern. 

	Here are some details about the real world location:

	Name: %s
	Vicinity: %s
	Rating: %.1f stars from %d reviews
	Types: %v
	Sophistication: %s
	Business Status: %s
`

const generatePointOfInterestPromptTemplate = premise + `
	Please format your response as a JSON object with the following fields:
	
	{
		"name": "string",
		"description": "string",
		"clue": "string",
	}

	Name should the name of the point of interest. The name should be fantasy themed and rooted in the name and function of the real world location.
	Description should be a short description of the point of interest. Please make up a fantasy themed background for the point of interest. Feel free to make up proper nouns or lore for a fake fantasy universe.
	Clue should be a short clue that hints at the location of the point of interest. It should relate to the real world location, not the one you're making up.
`

const generateFantasyImagePromptTemplate = premise + `
	The image should be a fantasy themed, retro video game vibes, and pixelated.
`

const style = "fantasy, retro video game vibes, pixelated"

type Client interface {
	GeneratePointOfInterest(ctx context.Context, place googlemaps.Place) (*models.PointOfInterest, error)
	SeedPointsOfInterest(ctx context.Context, zone models.Zone, locationType googlemaps.PlaceType) ([]*models.PointOfInterest, error)
}

func NewClient(googlemapsClient googlemaps.Client, dbClient db.DbClient, deepPriest deep_priest.DeepPriest) Client {
	return &client{
		googlemapsClient: googlemapsClient,
		dbClient:         dbClient,
		deepPriest:       deepPriest,
	}
}

func (c *client) SeedPointsOfInterest(ctx context.Context, zone models.Zone, locationType googlemaps.PlaceType) ([]*models.PointOfInterest, error) {
	places, err := c.googlemapsClient.FindPlaces(googlemaps.PlaceQuery{
		Lat:      zone.Latitude,
		Long:     zone.Longitude,
		Radius:   int(zone.Radius),
		Category: string(locationType),
	})
	if err != nil {
		return nil, err
	}

	pointsOfInterest := make([]*models.PointOfInterest, len(places))
	for i, place := range places {
		poi, err := c.GeneratePointOfInterest(ctx, place)
		if err != nil {
			return nil, err
		}
		pointsOfInterest[i] = poi
	}

	return pointsOfInterest, nil
}

func (c *client) GeneratePointOfInterest(ctx context.Context, place googlemaps.Place) (*models.PointOfInterest, error) {
	fantasyPointOfInterest, err := c.generateFantasyTheming(place)
	if err != nil {
		return nil, err
	}

	imageUrl, err := c.generateFantasyImage(place)
	if err != nil {
		return nil, err
	}

	poi := &models.PointOfInterest{
		ID:          uuid.New(),
		Name:        fantasyPointOfInterest.Name,
		Description: fantasyPointOfInterest.Description,
		Clue:        fantasyPointOfInterest.Clue,
		ImageUrl:    imageUrl,
		Lat:         strconv.FormatFloat(place.Geometry.Location.Lat, 'f', -1, 64),
		Lng:         strconv.FormatFloat(place.Geometry.Location.Lng, 'f', -1, 64),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tags := make([]*models.Tag, len(place.Types))
	for i, t := range place.Types {
		tag, err := c.mapTypeToTag(googlemaps.PlaceType(t))
		if err != nil {
			return nil, err
		}

		if err := c.dbClient.Tag().Upsert(ctx, tag); err != nil {
			return nil, err
		}

		tags[i] = tag
	}

	if err := c.dbClient.PointOfInterest().Create(ctx, *poi); err != nil {
		return nil, err
	}

	for _, tag := range tags {
		if err := c.dbClient.Tag().AddTagToPointOfInterest(ctx, tag.ID, poi.ID); err != nil {
			return nil, err
		}

		poi.Tags = append(poi.Tags, *tag)
	}

	return poi, nil
}

func (c *client) generateFantasyTheming(place googlemaps.Place) (*FantasyPointOfInterest, error) {
	answer, err := c.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: c.makeFantasyThemingPrompt(place),
	})
	if err != nil {
		return nil, err
	}

	var fantasyPointOfInterest FantasyPointOfInterest
	if err := json.Unmarshal([]byte(answer.Answer), &fantasyPointOfInterest); err != nil {
		return nil, err
	}

	return &fantasyPointOfInterest, nil
}

func (c *client) generateFantasyImage(place googlemaps.Place) (string, error) {
	res, err := c.deepPriest.GenerateImage(deep_priest.GenerateImageRequest{
		Prompt:         c.makeFantasyImagePrompt(place),
		Style:          style,
		Size:           "1024x1024",
		N:              1,
		ResponseFormat: "b64_json",
		User:           "poltergeist",
		Model:          "dall-e-3",
		Quality:        "standard",
	})
	if err != nil {
		return "", err
	}

	return res, nil
}

func (c *client) makeFantasyImagePrompt(place googlemaps.Place) string {
	return fmt.Sprintf(
		generateFantasyImagePromptTemplate,
		place.Name,
		place.Vicinity,
		place.Rating,
		place.UserRatingsTotal,
		place.Types,
		c.generateSophistication(place),
		place.BusinessStatus,
	)
}

func (c *client) makeFantasyThemingPrompt(place googlemaps.Place) string {
	return fmt.Sprintf(
		generatePointOfInterestPromptTemplate,
		place.Name,
		place.Vicinity,
		place.Rating,
		place.UserRatingsTotal,
		place.Types,
		c.generateSophistication(place),
		place.BusinessStatus,
	)
}

func (c *client) generateSophistication(place googlemaps.Place) string {
	switch place.PriceLevel {
	case 0:
		return "free"
	case 1:
		return "casual"
	case 2:
		return "mid-tier"
	case 3:
		return "high-end"
	case 4:
		return "luxury"
	}

	return "casual"
}
