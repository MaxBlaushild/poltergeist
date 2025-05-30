package dungeonmaster

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
)

// QuestCopy is the single source of truth for this struct definition.
type QuestCopy struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type QuestImagePrompt struct {
	Description string `json:"description"`
}

const generateQuestPromptTemplate = `
	You are a video game designer tasked with converting real-world tasks and chores into quests for a fantasy RPG.

	The tasks and chores for the quest you are designing are as follows:

	%s

	Please come up with copy for the quest that incorporates the task and chores into a cohesive quest that might appear in a fantasy role playing game. Try to keep the description to 50 words or less. Feel free to make up names or proper nouns.

	Please format your response as a JSON object with the following fields:
	{
		"name": "string", // The name of the overarching quest that includes all of the above tasks and chores
		"description": "string", // A description of why the chores need to be done in the made-up fantasy world and a bit of lore about the quest
	}
`

const generateQuestImagePromptTemplate = `
	You are a video game designer tasked with creating visual assets for quests in a fantasy role playing game.

	The quest is this:

	%s

	Please describe what an iconic moment from this quest would look like to an outside observer.

	Please format your response as a JSON object with the following fields:
	{
		"description": "string", // A description of what the hero image for the quest would look like
	}
`

const style = "natural"

// generateQuestCopy is the primary implementation.
// It calls generateQuestCopyInternalFunc if set (for testing).
func (c *client) generateQuestCopy(
	ctx context.Context,
	locations []string,
	descriptions []string,
	challenges []string,
) (*QuestCopy, error) {
	if generateQuestCopyInternalFunc != nil {
		return generateQuestCopyInternalFunc(ctx, locations, descriptions, challenges)
	}

	log.Printf("Generating quest copy for locations: %v, descriptions: %v, and challenges: %v", locations, descriptions, challenges)

	tasks := ""
	for i, location := range locations {
		challenge := challenges[i]
		description := descriptions[i]
		tasks += fmt.Sprintf(`
			TASK #%d
			Location: %s
			Challenge: %s
			Description: %s
			END_TASK

		`, i+1, location, challenge, description)
	}

	answer, err := c.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: fmt.Sprintf(generateQuestPromptTemplate, tasks),
	})
	if err != nil {
		log.Printf("Error getting response from DeepPriest: %v", err)
		return nil, err
	}

	var questCopy QuestCopy
	if err := json.Unmarshal([]byte(answer.Answer), &questCopy); err != nil {
		log.Printf("Error unmarshaling quest copy: %v", err)
		return nil, err
	}

	log.Printf("Successfully generated quest copy")
	return &questCopy, nil
}

// generateQuestImage is the primary implementation.
// It calls generateQuestImageInternalFunc if set (for testing).
func (c *client) generateQuestImage(ctx context.Context, questCopy QuestCopy) (string, error) {
	if generateQuestImageInternalFunc != nil {
		return generateQuestImageInternalFunc(ctx, questCopy)
	}

	log.Printf("Generating quest image for quest: %s", questCopy.Name)

	prompt, err := c.generateQuestImagePrompt(questCopy)
	if err != nil {
		log.Printf("Error generating quest image prompt: %v", err)
		return "", err
	}

	log.Printf("Generated quest image prompt: %s", prompt)

	base64Image, err := c.deepPriest.GenerateImage(deep_priest.GenerateImageRequest{
		Prompt:         prompt,
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

	log.Printf("Generated image: %s", base64Image)

	imageUrl, err := c.UploadImage(ctx, base64Image) // Assuming c.UploadImage exists
	if err != nil {
		log.Printf("Error processing image: %v", err)
		return "", err
	}
	log.Printf("Uploaded image to S3: %s", imageUrl)

	return imageUrl, nil
}

func (c *client) generateQuestImagePrompt(questCopy QuestCopy) (string, error) {
	answer, err := c.deepPriest.PetitionTheFount(&deep_priest.Question{
		// The original prompt here had two %s placeholders but only passed questCopy.Name.
		// Assuming it meant to use both Name and Description.
		Question: fmt.Sprintf(generateQuestImagePromptTemplate, fmt.Sprintf("Name: %s\nDescription: %s", questCopy.Name, questCopy.Description)),
	})
	if err != nil {
		log.Printf("Error getting response from DeepPriest: %v", err)
		return "", err
	}

	var questImagePrompt QuestImagePrompt
	if err := json.Unmarshal([]byte(answer.Answer), &questImagePrompt); err != nil {
		log.Printf("Error unmarshaling quest image prompt: %v", err)
		return "", err
	}

	return questImagePrompt.Description, nil
}

// UploadImage method is defined in image_processing.go
