package dungeonmaster

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type QuestCopy struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type QuestAcceptanceDialogue struct {
	AcceptanceDialogue []string `json:"acceptanceDialogue"`
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

const generateQuestAcceptanceDialogueTemplate = `
	You are a video game writer tasked with scripting the dialogue a quest giver delivers to a player before they accept a quest.

	Quest name: %s
	Quest description: %s
	Quest giver: %s
	Rewards: %s

	The tasks and chores for the quest are:

	%s

	Write 3 to 6 short lines of dialogue. Keep each line under 18 words, in a warm fantasy RPG tone. Speak directly to the player (use "you"). Mention the rewards if provided.

	Please format your response as a JSON object with the following fields:
	{
		"acceptanceDialogue": ["string"] // The dialogue lines the player sees before accepting the quest
	}
`

const style = "natural"

func buildQuestTasks(locations []string, descriptions []string, challenges []string) string {
	count := len(locations)
	if len(descriptions) < count {
		count = len(descriptions)
	}
	if len(challenges) < count {
		count = len(challenges)
	}
	if count == 0 {
		return ""
	}

	var builder strings.Builder
	for i := 0; i < count; i++ {
		builder.WriteString(fmt.Sprintf(`
			TASK #%d
			Location: %s
			Challenge: %s
			Description: %s
			END_TASK

		`, i+1, locations[i], challenges[i], descriptions[i]))
	}
	return builder.String()
}

func (c *client) generateQuestCopy(
	ctx context.Context,
	locations []string,
	descriptions []string,
	challenges []string,
) (*QuestCopy, error) {
	log.Printf("Generating quest copy for locations: %v, descriptions: %v, and challenges: %v", locations, descriptions, challenges)

	tasks := buildQuestTasks(locations, descriptions, challenges)

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

func (c *client) generateQuestAcceptanceDialogue(
	ctx context.Context,
	questCopy *QuestCopy,
	questGiverCharacterID *uuid.UUID,
	questArchetype *models.QuestArchetype,
	locations []string,
	descriptions []string,
	challenges []string,
) ([]string, error) {
	if questCopy == nil {
		return nil, fmt.Errorf("quest copy is required for acceptance dialogue")
	}

	questGiverName := "a quest giver"
	if questGiverCharacterID != nil {
		if character, err := c.dbClient.Character().FindByID(ctx, *questGiverCharacterID); err == nil && character != nil {
			if strings.TrimSpace(character.Name) != "" {
				questGiverName = character.Name
			}
		}
	}

	rewardSummary := "none"
	if questArchetype != nil {
		parts := []string{}
		if questArchetype.DefaultGold > 0 {
			parts = append(parts, fmt.Sprintf("%d gold", questArchetype.DefaultGold))
		}
		rewards := questArchetype.ItemRewards
		if len(rewards) == 0 {
			if loaded, err := c.dbClient.QuestArchetypeItemReward().FindByQuestArchetypeID(ctx, questArchetype.ID); err == nil {
				rewards = loaded
			}
		}
		for _, reward := range rewards {
			if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
				continue
			}
			name := reward.InventoryItem.Name
			if strings.TrimSpace(name) == "" {
				name = fmt.Sprintf("Item %d", reward.InventoryItemID)
			}
			if reward.Quantity > 1 {
				parts = append(parts, fmt.Sprintf("%dx %s", reward.Quantity, name))
			} else {
				parts = append(parts, name)
			}
		}
		if len(parts) > 0 {
			rewardSummary = strings.Join(parts, ", ")
		}
	}

	tasks := buildQuestTasks(locations, descriptions, challenges)

	answer, err := c.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: fmt.Sprintf(generateQuestAcceptanceDialogueTemplate, questCopy.Name, questCopy.Description, questGiverName, rewardSummary, tasks),
	})
	if err != nil {
		return nil, err
	}

	var parsed QuestAcceptanceDialogue
	if err := json.Unmarshal([]byte(answer.Answer), &parsed); err != nil {
		return nil, err
	}

	dialogue := make([]string, 0, len(parsed.AcceptanceDialogue))
	for _, line := range parsed.AcceptanceDialogue {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		dialogue = append(dialogue, trimmed)
	}

	return dialogue, nil
}

func (c *client) generateQuestImage(ctx context.Context, questCopy QuestCopy) (string, error) {
	log.Printf("Generating quest image for quest: %s", questCopy.Name)

	prompt, err := c.generateQuestImagePrompt(questCopy)
	if err != nil {
		log.Printf("Error generating quest image prompt: %v", err)
		return "", err
	}

	log.Printf("Generated quest image prompt: %s", prompt)

	request := deep_priest.GenerateImageRequest{
		Prompt: prompt,
		Style:  style,
	}
	deep_priest.ApplyGenerateImageDefaults(&request)
	base64Image, err := c.deepPriest.GenerateImage(request)
	if err != nil {
		log.Printf("Error generating image: %v", err)
		return "", err
	}

	imageUrl, err := c.UploadImage(ctx, base64Image)
	if err != nil {
		log.Printf("Error processing image: %v", err)
		return "", err
	}
	log.Printf("Uploaded image to S3: %s", imageUrl)

	return imageUrl, nil
}

func (c *client) generateQuestImagePrompt(questCopy QuestCopy) (string, error) {
	answer, err := c.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: fmt.Sprintf(generateQuestImagePromptTemplate, questCopy.Name, questCopy.Description),
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
