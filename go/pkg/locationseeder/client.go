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
	You are a video game designer tasked with converting real-world locations into points of interest on a fantasy RPG map. These points of interest should closely mirror their real-world counterparts but be reimagined within a fantasy setting, with a retro, pixelated, video game style.

	Each real-world location is provided with the following details:

	Name: %s (Convert this name into a fitting fantasy title, such as a tavern, guild hall, or enchanted marketplace.)

	Vicinity: %s (Describe the surrounding area in a fantastical way—imagine the real-world location as part of a magical kingdom, ancient forest, or mystical city.)

	Rating: %.1f stars from %d reviews (Reflect the rating as if it were from adventurers, using retro-styled pixelated fonts.)

	Types: %v (For example, a café could become a “Tavern,” a park could become a “Sacred Grove,” or a museum could become a “Library of Lore.”)

	Sophistication: %s (Translate this into fantasy status—'Legendary,' 'Epic,' 'Hidden,' 'Common,' etc.)

	Business Status: %s (Imagine the business status as something more fitting for the fantasy world—e.g., 'Open for quests,' 'Temporarily closed due to dragon attack,' etc.)
`

const generatePointOfInterestPromptTemplate = premise + `
	Please format your response as a JSON object with the following fields:
	
	{
		"name": "string",
		"description": "string",
		"clue": "string",
	}
`

const generateFantasyImagePromptTemplate = premise + `
	The goal is to take these real-world values and translate them into fantasy-themed locations while maintaining their core concept but enhancing them with magical, mythical, and pixelated video game-style elements. Each location should evoke a sense of nostalgia for retro video games, with blocky shapes, pixelated visuals, and vibrant colors that evoke classic RPG vibes.
`

const style = "natural"

type Client interface {
	GeneratePointOfInterest(ctx context.Context, place googlemaps.Place, zone models.Zone) (*models.PointOfInterest, error)
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

	log.Printf("Zone latitude: %f, longitude: %f, radius: %d", zone.Latitude, zone.Longitude, int(zone.Radius))
	log.Printf("Location type: %s", string(locationType))

	places, err := c.googlemapsClient.FindPlaces(googlemaps.PlaceQuery{
		Lat:            zone.Latitude,
		Long:           zone.Longitude,
		Radius:         zone.Radius,
		Category:       string(locationType),
		MaxResultCount: 20,
	})
	if err != nil {
		log.Printf("Error finding places: %v", err)
		return nil, err
	}

	log.Printf("Found %d places to convert to points of interest", len(places))

	var pointsOfInterest []*models.PointOfInterest
	for i, place := range places {
		log.Printf("Generating point of interest %d/%d for place: %s", i+1, len(places), place.Name)

		if hasBeenImported, err := c.dbClient.PointOfInterest().HasBeenImportedByGoogleMaps(ctx, place.ID); err != nil {
			log.Printf("Error checking if point of interest has been imported: %v", err)
			return nil, err
		} else if hasBeenImported {
			log.Printf("Point of interest %s has already been imported", place.Name)
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

func (c *client) GeneratePointOfInterest(ctx context.Context, place googlemaps.Place, zone models.Zone) (*models.PointOfInterest, error) {
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
		Lat:         strconv.FormatFloat(place.Location.Latitude, 'f', -1, 64),
		Lng:         strconv.FormatFloat(place.Location.Longitude, 'f', -1, 64),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	log.Printf("Created point of interest object with ID: %s", poi.ID)

	// tags := make([]*models.Tag, len(place.Types))
	// for i, t := range place.Types {
	// 	log.Printf("Processing tag %d/%d: %s", i+1, len(place.Types), t)
	// 	tag := &models.Tag{
	// 		ID:         uuid.New(),
	// 		Value:      t,
	// 		TagGroupID: uuid.MustParse("520de224-e432-4f7a-82f8-0df1ac62d44b"),
	// 	}

	// 	if err := c.dbClient.Tag().Upsert(ctx, tag); err != nil {
	// 		log.Printf("Error upserting tag: %v", err)
	// 		return nil, err
	// 	}

	// 	tags[i] = tag
	// }

	if err := c.dbClient.PointOfInterest().Create(ctx, *poi); err != nil {
		log.Printf("Error creating point of interest in database: %v", err)
		return nil, err
	}
	log.Printf("Successfully created point of interest in database")

	// for _, tag := range tags {
	// 	log.Printf("Adding tag %s to point of interest", tag.Value)
	// 	if err := c.dbClient.Tag().AddTagToPointOfInterest(ctx, tag.ID, poi.ID); err != nil {
	// 		log.Printf("Error adding tag to point of interest: %v", err)
	// 		return nil, err
	// 	}

	// 	poi.Tags = append(poi.Tags, *tag)
	// }

	if err := c.dbClient.Zone().AddPointOfInterestToZone(ctx, zone.ID, poi.ID); err != nil {
		log.Printf("Error adding point of interest to zone: %v", err)
		return nil, err
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
		place.DisplayName.Text,
		place.FormattedAddress,
		place.Rating,
		place.UserRatingCount,
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
		place.DisplayName.Text,
		place.FormattedAddress,
		place.Rating,
		place.UserRatingCount,
		place.Types,
		c.generateSophistication(place),
		place.BusinessStatus,
	)
	log.Printf("Generated fantasy theming prompt: %s", prompt)
	return prompt
}

func (c *client) generateSophistication(place googlemaps.Place) string {
	switch place.PriceLevel {
	case "PRICE_LEVEL_FREE":
		return "free"
	case "PRICE_LEVEL_INEXPENSIVE":
		return "casual"
	case "PRICE_LEVEL_MODERATE":
		return "mid-tier"
	case "PRICE_LEVEL_EXPENSIVE":
		return "high-end"
	case "PRICE_LEVEL_VERY_EXPENSIVE":
		return "luxury"
	default:
		return "casual" // Default for PRICE_LEVEL_UNSPECIFIED or unknown values
	}
}
