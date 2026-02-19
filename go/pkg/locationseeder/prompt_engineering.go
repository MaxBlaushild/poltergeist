package locationseeder

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

type PromptText struct {
	Text string `json:"text"`
}

const premise = `
	You are a video game designer tasked with converting real-world locations into points of interest on a fantasy RPG map.

	Describe how %s would appear if it was in a fantasy role playing video game.

	An editorial summary of %s is: %s.

	Some categories that people use to describe %s are: %v.

	The sophistication of %s is considered to be %s.

	Here is a bit about the region %s is in:

	Name: %s
	Description: %s

	Do not use the location's name in your response.
`

const generatePointOfInterestPromptTemplate = premise + `
	Please try to keep the description to 50 words or less.

	Please format your response as a JSON object with the following fields:
	
	{
		"name": "string", // The fantasy name of the point of interest
		"description": "string", // A description of the appearance of the fantasy point of interest and a bit of made up lore about it
		"clue": "string", // A clue that can be used to find the point of interest in the real world
	}
`

const generateImagePromptPromptTemplate = premise + `
	Please describe how the location would look from the outside if it was in a fantasy role playing video game.

	The image should match the aesthetic of retro 16-bit RPG pixel art item and character images:
	- Crisp outlines, limited color palette, clean background
	- Centered subject, readable silhouette
	- No text, no logos, no UI
	- Exterior view, 3/4 angle or slight isometric perspective
	- Keep the prompt focused on a single iconic exterior scene

	Please format your response as a JSON object with the following fields:
	
	{
		"text": "string", // A single concise image prompt in the above style
	}
`

// const generateFantasyImagePromptTemplate = premise + `
// 	The goal is to take these real-world values and translate them into fantasy-themed locations while maintaining their core concept but enhancing them with magical, mythical, and pixelated video game-style elements. Each location should evoke a sense of nostalgia for retro video games, with blocky shapes, pixelated visuals, and vibrant colors that evoke classic RPG vibes.
// `

const style = ""

func (c *client) generateFantasyTheming(place googlemaps.Place, zone *models.Zone) (*FantasyPointOfInterest, error) {
	answer, err := c.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: c.makeFantasyThemingPrompt(place, zone),
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

func (c *client) generateFantasyImage(ctx context.Context, place googlemaps.Place, zone *models.Zone) (string, error) {
	prompt, err := c.generateFantasyImagePrompt(place, zone)
	if err != nil {
		log.Printf("Error generating fantasy image prompt: %v", err)
		return "", err
	}
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("generated image prompt was empty")
	}

	request := deep_priest.GenerateImageRequest{
		Prompt: prompt,
	}
	deep_priest.ApplyGenerateImageDefaults(&request)
	base64Image, err := c.deepPriest.GenerateImage(request)
	if err != nil {
		log.Printf("Error generating image: %v", err)
		return "", err
	}

	imageUrl, err := c.UploadImage(ctx, place.ID, base64Image)
	if err != nil {
		log.Printf("Error processing image: %v", err)
		return "", err
	}
	log.Printf("Uploaded image to S3: %s", imageUrl)

	return imageUrl, nil
}

func (c *client) generateFantasyImagePrompt(place googlemaps.Place, zone *models.Zone) (string, error) {
	answer, err := c.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: c.makeFantasyImagePromptPrompt(place, zone),
	})
	if err != nil {
		log.Printf("Error getting response from DeepPriest: %v", err)
		return "", err
	}

	var promptText PromptText
	if err := json.Unmarshal([]byte(answer.Answer), &promptText); err != nil {
		log.Printf("Error unmarshaling prompt text: %v", err)
		return "", err
	}

	return promptText.Text, nil
}

func (c *client) makeFantasyImagePromptPrompt(place googlemaps.Place, zone *models.Zone) string {
	prompt := fmt.Sprintf(
		generateImagePromptPromptTemplate,
		place.DisplayName.Text,
		place.DisplayName.Text,
		place.EditorialSummary.Text,
		place.DisplayName.Text,
		place.Types,
		place.DisplayName.Text,
		c.generateSophistication(place),
		zone.Name,
		zone.Description,
	)
	return prompt
}

// func (c *client) makeFantasyImagePrompt(place googlemaps.Place) string {
// 	prompt := fmt.Sprintf(
// 		generateFantasyImagePromptTemplate,
// 		place.DisplayName.Text,
// 		place.Types,
// 		c.generateSophistication(place),
// 	)
// 	log.Printf("Generated fantasy image prompt: %s", prompt)
// 	return prompt
// }

func (c *client) makeFantasyThemingPrompt(place googlemaps.Place, zone *models.Zone) string {
	prompt := fmt.Sprintf(
		generatePointOfInterestPromptTemplate,
		place.DisplayName.Text,
		place.DisplayName.Text,
		place.EditorialSummary.Text,
		place.DisplayName.Text,
		place.Types,
		place.DisplayName.Text,
		c.generateSophistication(place),
		zone.Name,
		zone.Description,
	)
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
