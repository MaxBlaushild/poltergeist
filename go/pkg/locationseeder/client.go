package locationseeder

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	log.Println("Creating new locationseeder client")
	return &client{
		googlemapsClient: googlemapsClient,
		dbClient:         dbClient,
		deepPriest:       deepPriest,
	}
}

func (c *client) SeedPointsOfInterest(ctx context.Context, zone models.Zone, locationType googlemaps.PlaceType) ([]*models.PointOfInterest, error) {
	log.Printf("Starting to seed points of interest for zone %s with location type %s", zone.Name, locationType)

	places, err := c.googlemapsClient.FindPlaces(googlemaps.PlaceQuery{
		Lat:      zone.Latitude,
		Long:     zone.Longitude,
		Radius:   int(zone.Radius),
		Category: string(locationType),
	})
	if err != nil {
		log.Printf("Error finding places: %v", err)
		return nil, err
	}

	log.Printf("Found %d places to convert to points of interest", len(places))

	pointsOfInterest := make([]*models.PointOfInterest, len(places))
	for i, place := range places {
		log.Printf("Generating point of interest %d/%d for place: %s", i+1, len(places), place.Name)
		poi, err := c.GeneratePointOfInterest(ctx, place)
		if err != nil {
			log.Printf("Error generating point of interest for place %s: %v", place.Name, err)
			return nil, err
		}
		pointsOfInterest[i] = poi
	}

	log.Printf("Successfully generated %d points of interest", len(pointsOfInterest))
	return pointsOfInterest, nil
}

func (c *client) GeneratePointOfInterest(ctx context.Context, place googlemaps.Place) (*models.PointOfInterest, error) {
	log.Printf("Starting to generate point of interest for place: %s", place.Name)

	fantasyPointOfInterest, err := c.generateFantasyTheming(place)
	if err != nil {
		log.Printf("Error generating fantasy theming: %v", err)
		return nil, err
	}
	log.Printf("Generated fantasy theming with name: %s", fantasyPointOfInterest.Name)

	imageUrl, err := c.generateFantasyImage(place)
	if err != nil {
		log.Printf("Error generating fantasy image: %v", err)
		return nil, err
	}
	log.Printf("Generated fantasy image URL: %s", imageUrl)

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

	log.Printf("Created point of interest object with ID: %s", poi.ID)

	tags := make([]*models.Tag, len(place.Types))
	for i, t := range place.Types {
		log.Printf("Processing tag %d/%d: %s", i+1, len(place.Types), t)
		tag, err := c.mapTypeToTag(googlemaps.PlaceType(t))
		if err != nil {
			log.Printf("Error mapping type to tag: %v", err)
			return nil, err
		}

		if err := c.dbClient.Tag().Upsert(ctx, tag); err != nil {
			log.Printf("Error upserting tag: %v", err)
			return nil, err
		}

		tags[i] = tag
	}

	if err := c.dbClient.PointOfInterest().Create(ctx, *poi); err != nil {
		log.Printf("Error creating point of interest in database: %v", err)
		return nil, err
	}
	log.Printf("Successfully created point of interest in database")

	for _, tag := range tags {
		log.Printf("Adding tag %s to point of interest", tag.Name)
		if err := c.dbClient.Tag().AddTagToPointOfInterest(ctx, tag.ID, poi.ID); err != nil {
			log.Printf("Error adding tag to point of interest: %v", err)
			return nil, err
		}

		poi.Tags = append(poi.Tags, *tag)
	}

	log.Printf("Successfully generated point of interest with %d tags", len(poi.Tags))
	return poi, nil
}

func (c *client) generateFantasyTheming(place googlemaps.Place) (*FantasyPointOfInterest, error) {
	log.Printf("Generating fantasy theming for place: %s", place.Name)

	answer, err := c.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: c.makeFantasyThemingPrompt(place),
	})
	if err != nil {
		log.Printf("Error getting response from DeepPriest: %v", err)
		return nil, err
	}

	var fantasyPointOfInterest FantasyPointOfInterest
	if err := json.Unmarshal([]byte(answer.Answer), &fantasyPointOfInterest); err != nil {
		log.Printf("Error unmarshaling fantasy point of interest: %v", err)
		return nil, err
	}

	log.Printf("Successfully generated fantasy theming")
	return &fantasyPointOfInterest, nil
}

func (c *client) generateFantasyImage(place googlemaps.Place) (string, error) {
	log.Printf("Generating fantasy image for place: %s", place.Name)

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
		log.Printf("Error generating image: %v", err)
		return "", err
	}

	log.Printf("Successfully generated fantasy image")
	return res, nil
}

func (c *client) makeFantasyImagePrompt(place googlemaps.Place) string {
	prompt := fmt.Sprintf(
		generateFantasyImagePromptTemplate,
		place.Name,
		place.Vicinity,
		place.Rating,
		place.UserRatingsTotal,
		place.Types,
		c.generateSophistication(place),
		place.BusinessStatus,
	)
	log.Printf("Generated fantasy image prompt: %s", prompt)
	return prompt
}

func (c *client) makeFantasyThemingPrompt(place googlemaps.Place) string {
	prompt := fmt.Sprintf(
		generatePointOfInterestPromptTemplate,
		place.Name,
		place.Vicinity,
		place.Rating,
		place.UserRatingsTotal,
		place.Types,
		c.generateSophistication(place),
		place.BusinessStatus,
	)
	log.Printf("Generated fantasy theming prompt: %s", prompt)
	return prompt
}

func (c *client) generateSophistication(place googlemaps.Place) string {
	sophistication := "casual"
	switch place.PriceLevel {
	case 0:
		sophistication = "free"
	case 1:
		sophistication = "casual"
	case 2:
		sophistication = "mid-tier"
	case 3:
		sophistication = "high-end"
	case 4:
		sophistication = "luxury"
	}

	log.Printf("Generated sophistication level for place %s: %s", place.Name, sophistication)
	return sophistication
}
