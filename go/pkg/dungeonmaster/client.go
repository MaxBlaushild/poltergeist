package dungeonmaster

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/locationseeder"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type client struct {
	googlemapsClient googlemaps.Client
	dbClient         db.DbClient
	deepPriest       deep_priest.DeepPriest
	locationSeeder   locationseeder.Client
	awsClient        aws.AWSClient
}

type Client interface {
	GenerateQuest(ctx context.Context, zone *models.Zone, questArchetypeID uuid.UUID, questGiverCharacterID *uuid.UUID) (*models.Quest, error)
}

func NewClient(
	googlemapsClient googlemaps.Client,
	dbClient db.DbClient,
	deepPriest deep_priest.DeepPriest,
	locationSeeder locationseeder.Client,
	awsClient aws.AWSClient,
) Client {
	return &client{
		googlemapsClient: googlemapsClient,
		dbClient:         dbClient,
		deepPriest:       deepPriest,
		locationSeeder:   locationSeeder,
		awsClient:        awsClient,
	}
}

func (c *client) GenerateQuest(
	ctx context.Context,
	zone *models.Zone,
	questArchetypeID uuid.UUID,
	questGiverCharacterID *uuid.UUID,
) (*models.Quest, error) {
	log.Printf("Generating quest for zone %s with quest arch type %+v", zone.Name, questArchetypeID)

	questArchType, err := c.dbClient.QuestArchetype().FindByID(ctx, questArchetypeID)
	if err != nil {
		log.Printf("Error finding quest arch type: %v", err)
		return nil, err
	}

	locations := make([]string, 0)
	descriptions := make([]string, 0)
	challenges := make([]string, 0)

	log.Println("Creating quest")
	quest := &models.Quest{
		ID:                    uuid.New(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		Name:                  "Quest",
		Description:           "A quest to complete",
		AcceptanceDialogue:    models.StringArray{},
		ZoneID:                &zone.ID,
		QuestArchetypeID:      &questArchetypeID,
		QuestGiverCharacterID: questGiverCharacterID,
		Gold:                  questArchType.DefaultGold,
	}
	if err := c.dbClient.Quest().Create(ctx, quest); err != nil {
		log.Printf("Error creating quest: %v", err)
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	// Track used POIs at the quest level
	usedPOIs := make(map[uuid.UUID]bool)

	log.Println("Processing quest nodes")
	orderIndex := 0
	nodeMap := make(map[uuid.UUID]uuid.UUID)
	if err := c.processQuestNode(ctx, zone, &questArchType.Root, &locations, &descriptions, &challenges, quest, usedPOIs, &orderIndex, nodeMap); err != nil {
		log.Printf("Error processing quest nodes: %v", err)
		return nil, err
	}

	if err := c.applyQuestArchetypeRewards(ctx, quest.ID, questArchType); err != nil {
		log.Printf("Error applying quest archetype rewards: %v", err)
		if deleteErr := c.dbClient.Quest().Delete(ctx, quest.ID); deleteErr != nil {
			log.Printf("Error deleting quest after reward application failure: %v", deleteErr)
		}
		return nil, err
	}

	log.Println("Generating quest copy")
	questCopy, err := c.generateQuestCopy(ctx, locations, descriptions, challenges)
	if err != nil {
		log.Printf("Error generating quest copy: %v", err)
		if deleteErr := c.dbClient.Quest().Delete(ctx, quest.ID); deleteErr != nil {
			log.Printf("Error deleting quest after copy generation failure: %v", deleteErr)
		}
		return nil, err
	}

	acceptanceDialogue, err := c.generateQuestAcceptanceDialogue(ctx, questCopy, questGiverCharacterID, questArchType, locations, descriptions, challenges)
	if err != nil {
		log.Printf("Error generating quest acceptance dialogue: %v", err)
		acceptanceDialogue = nil
	}
	acceptanceDialogueValue := models.StringArray(acceptanceDialogue)
	if acceptanceDialogueValue == nil {
		acceptanceDialogueValue = models.StringArray{}
	}

	log.Println("Generating quest image")
	questImage, err := c.generateQuestImage(ctx, *questCopy)
	if err != nil {
		log.Printf("Error generating quest image: %v", err)
		if deleteErr := c.dbClient.Quest().Delete(ctx, quest.ID); deleteErr != nil {
			log.Printf("Error deleting quest after image generation failure: %v", deleteErr)
		}
		return nil, err
	}

	log.Println("Updating quest with generated content")
	if err := c.dbClient.Quest().Update(ctx, quest.ID, &models.Quest{
		Name:                  questCopy.Name,
		Description:           questCopy.Description,
		AcceptanceDialogue:    acceptanceDialogueValue,
		ImageURL:              questImage,
		ZoneID:                quest.ZoneID,
		QuestArchetypeID:      quest.QuestArchetypeID,
		QuestGiverCharacterID: quest.QuestGiverCharacterID,
		RecurringQuestID:      quest.RecurringQuestID,
		RecurrenceFrequency:   quest.RecurrenceFrequency,
		NextRecurrenceAt:      quest.NextRecurrenceAt,
		Gold:                  quest.Gold,
		UpdatedAt:             time.Now(),
	}); err != nil {
		log.Printf("Error updating quest: %v", err)
		return nil, err
	}

	if questGiverCharacterID != nil {
		if err := c.ensureQuestActionForCharacter(ctx, quest.ID, *questGiverCharacterID); err != nil {
			log.Printf("Error ensuring quest action for character: %v", err)
		}
	}

	log.Printf("Successfully generated quest %s", quest.ID)
	return quest, nil
}

func (c *client) processQuestNode(
	ctx context.Context,
	zone *models.Zone,
	questArchTypeNode *models.QuestArchetypeNode,
	locations *[]string,
	descriptions *[]string,
	challenges *[]string,
	quest *models.Quest,
	usedPOIs map[uuid.UUID]bool,
	orderIndex *int,
	nodeMap map[uuid.UUID]uuid.UUID,
) error {
	log.Printf("Processing node for zone %s with place type %s", zone.Name, questArchTypeNode.LocationArchetypeID)

	locationArchetype, err := c.dbClient.LocationArchetype().FindByID(ctx, questArchTypeNode.LocationArchetypeID)
	if err != nil {
		log.Printf("Error finding location archetype: %v", err)
		return err
	}

	log.Printf("Location archetype: %+v", locationArchetype)
	for _, includedType := range locationArchetype.IncludedTypes {
		log.Printf("Included type: %s", includedType)
	}
	for _, excludedType := range locationArchetype.ExcludedTypes {
		log.Printf("Excluded type: %s", excludedType)
	}

	pointsOfInterest, err := c.locationSeeder.SeedPointsOfInterest(ctx, *zone, locationArchetype.IncludedTypes, locationArchetype.ExcludedTypes, 1)
	if err != nil {
		log.Printf("Error seeding points of interest: %v", err)
		return err
	}

	if len(pointsOfInterest) == 0 {
		log.Printf("No points of interest found for place type %s", locationArchetype.ID)
		return errors.New("no points of interest found")
	}

	var pointOfInterest *models.PointOfInterest
	for _, poi := range pointsOfInterest {
		if !usedPOIs[poi.ID] {
			pointOfInterest = poi
			usedPOIs[poi.ID] = true
			break
		}
	}
	if pointOfInterest == nil {
		return fmt.Errorf("no unused points of interest found for type %s", locationArchetype.ID)
	}
	log.Printf("Found point of interest: %s", pointOfInterest.Name)

	// Mark this POI as used in a quest
	if err := c.dbClient.PointOfInterest().UpdateLastUsedInQuest(ctx, pointOfInterest.ID); err != nil {
		log.Printf("Warning: failed to update last_used_in_quest_at for POI %s: %v", pointOfInterest.ID, err)
		// Don't fail the quest generation for this, just log the warning
	}

	existingNodeID, ok := nodeMap[questArchTypeNode.ID]
	var questNodeID uuid.UUID
	if ok {
		questNodeID = existingNodeID
	} else {
		questNodeID = uuid.New()
		node := &models.QuestNode{
			ID:                questNodeID,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			QuestID:           quest.ID,
			OrderIndex:        *orderIndex,
			PointOfInterestID: &pointOfInterest.ID,
			SubmissionType:    models.DefaultQuestNodeSubmissionType(),
		}
		if err := c.dbClient.QuestNode().Create(ctx, node); err != nil {
			log.Printf("Error creating quest node: %v", err)
			return err
		}
		nodeMap[questArchTypeNode.ID] = questNodeID
		*orderIndex++
	}

	*locations = append(*locations, pointOfInterest.Name)
	*descriptions = append(*descriptions, pointOfInterest.Description)

	for i, allotedChallenge := range questArchTypeNode.Challenges {
		log.Printf("Processing challenge %d", i)

		randomChallenge, err := questArchTypeNode.GetRandomChallenge()
		if err != nil {
			log.Printf("Error getting random challenge: %v", err)
			return err
		}
		*challenges = append(*challenges, randomChallenge)

		log.Printf("Creating challenge: %s", randomChallenge)
		challenge := &models.QuestNodeChallenge{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			QuestNodeID: questNodeID,
			Tier:        i,
			Question:    randomChallenge,
			Reward:      allotedChallenge.Reward,
			InventoryItemID: allotedChallenge.InventoryItemID,
			Difficulty:  0,
			StatTags:    models.StringArray{},
			Proficiency: allotedChallenge.Proficiency,
		}
		err = c.dbClient.QuestNodeChallenge().Create(ctx, challenge)
		if err != nil {
			log.Printf("Error creating challenge: %v", err)
			return err
		}

		if allotedChallenge.UnlockedNodeID != nil {
			unlockedNode, err := c.dbClient.QuestArchetypeNode().FindByID(ctx, *allotedChallenge.UnlockedNodeID)
			if err != nil {
				log.Printf("Error finding unlocked node: %v", err)
				return err
			}
			log.Printf("Processing child node: %s", unlockedNode.LocationArchetypeID)
			if err := c.processQuestNode(ctx, zone, unlockedNode, locations, descriptions, challenges, quest, usedPOIs, orderIndex, nodeMap); err != nil {
				log.Printf("Error processing child node: %v", err)
				return err
			}
			childNodeID := nodeMap[unlockedNode.ID]
			child := &models.QuestNodeChild{
				ID:                   uuid.New(),
				CreatedAt:            time.Now(),
				UpdatedAt:            time.Now(),
				QuestNodeID:          questNodeID,
				NextQuestNodeID:      childNodeID,
				QuestNodeChallengeID: &challenge.ID,
			}
			if err := c.dbClient.QuestNodeChild().Create(ctx, child); err != nil {
				log.Printf("Error creating quest node child: %v", err)
				return err
			}
		}
	}

	log.Printf("Successfully processed node for point of interest %s", pointOfInterest.ID)
	return nil
}

func (c *client) applyQuestArchetypeRewards(ctx context.Context, questID uuid.UUID, questArchetype *models.QuestArchetype) error {
	if questArchetype == nil {
		return nil
	}
	rewards := questArchetype.ItemRewards
	if len(rewards) == 0 {
		loaded, err := c.dbClient.QuestArchetypeItemReward().FindByQuestArchetypeID(ctx, questArchetype.ID)
		if err == nil {
			rewards = loaded
		}
	}
	if len(rewards) == 0 {
		return nil
	}
	now := time.Now()
	questRewards := make([]models.QuestItemReward, 0, len(rewards))
	for _, reward := range rewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		questRewards = append(questRewards, models.QuestItemReward{
			ID:              uuid.New(),
			CreatedAt:       now,
			UpdatedAt:       now,
			QuestID:         questID,
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	if len(questRewards) == 0 {
		return nil
	}
	return c.dbClient.QuestItemReward().ReplaceForQuest(ctx, questID, questRewards)
}

func (c *client) ensureQuestActionForCharacter(ctx context.Context, questID uuid.UUID, characterID uuid.UUID) error {
	actions, err := c.dbClient.CharacterAction().FindByCharacterID(ctx, characterID)
	if err != nil {
		return err
	}
	questIDStr := questID.String()
	for _, action := range actions {
		if action.ActionType != models.ActionTypeGiveQuest {
			continue
		}
		if action.Metadata == nil {
			continue
		}
		if value, ok := action.Metadata["questId"]; ok && fmt.Sprint(value) == questIDStr {
			return nil
		}
	}

	action := &models.CharacterAction{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CharacterID: characterID,
		ActionType:  models.ActionTypeGiveQuest,
		Dialogue:    []models.DialogueMessage{},
		Metadata:    map[string]interface{}{"questId": questIDStr},
	}
	return c.dbClient.CharacterAction().Create(ctx, action)
}
