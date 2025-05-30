package dungeonmaster

import (
	"context"
	"errors"
	"fmt"
	"log"

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
	GenerateQuest(ctx context.Context, zone *models.Zone, questArchetypeID uuid.UUID) (*models.PointOfInterestGroup, error)
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

// generateQuestCopyInternalFunc and generateQuestImageInternalFunc are package-level
// variables used by the test suite to replace the actual implementations of
// generateQuestCopy and generateQuestImage (defined in prompt_engineering.go).
var (
	generateQuestCopyInternalFunc func(ctx context.Context, locations []string, descriptions []string, challenges []string) (*QuestCopy, error)
	generateQuestImageInternalFunc func(ctx context.Context, questCopy QuestCopy) (string, error)
)

func (c *client) GenerateQuest(
	ctx context.Context,
	zone *models.Zone,
	questArchetypeID uuid.UUID,
) (*models.PointOfInterestGroup, error) {
	log.Printf("Generating quest for zone %s with quest arch type %+v", zone.Name, questArchetypeID)

	questArchType, err := c.dbClient.QuestArchetype().FindByID(ctx, questArchetypeID)
	if err != nil {
		log.Printf("Error finding quest arch type: %v", err)
		return nil, err
	}

	locations := make([]string, 0)
	descriptions := make([]string, 0)
	challenges := make([]string, 0)

	log.Println("Creating initial quest point of interest group")
	quest, err := c.dbClient.PointOfInterestGroup().Create(
		ctx,
		"Quest",
		"A quest to complete",
		"",
		models.PointOfInterestGroupTypeQuest,
	)
	if err != nil {
		log.Printf("Error creating quest group: %v", err)
		return nil, err
	}

	// Track used POIs at the quest level
	usedPOIs := make(map[uuid.UUID]bool)

	log.Println("Processing quest nodes")
	if err := c.processNode(ctx, zone, &questArchType.Root, &locations, &descriptions, &challenges, quest, nil, usedPOIs, nil); err != nil {
		log.Printf("Error processing quest nodes: %v", err)
		if deleteErr := c.dbClient.PointOfInterestGroup().Delete(ctx, quest.ID); deleteErr != nil {
			log.Printf("Error deleting quest group after node processing failure: %v", deleteErr)
		}
		return nil, err
	}

	log.Println("Generating quest copy")
	// This will call (c *client) generateQuestCopy defined in prompt_engineering.go,
	// which in turn calls generateQuestCopyInternalFunc if it's set by a test.
	questCopy, err := c.generateQuestCopy(ctx, locations, descriptions, challenges)
	if err != nil {
		log.Printf("Error generating quest copy: %v", err)
		if deleteErr := c.dbClient.PointOfInterestGroup().Delete(ctx, quest.ID); deleteErr != nil {
			log.Printf("Error deleting quest group after copy generation failure: %v", deleteErr)
		}
		return nil, err
	}

	log.Println("Generating quest image")
	// This will call (c *client) generateQuestImage defined in prompt_engineering.go,
	// which in turn calls generateQuestImageInternalFunc if it's set by a test.
	questImage, err := c.generateQuestImage(ctx, *questCopy)
	if err != nil {
		log.Printf("Error generating quest image: %v", err)
		if deleteErr := c.dbClient.PointOfInterestGroup().Delete(ctx, quest.ID); deleteErr != nil {
			log.Printf("Error deleting quest group after image generation failure: %v", deleteErr)
		}
		return nil, err
	}

	log.Println("Updating quest with generated content")
	if err := c.dbClient.PointOfInterestGroup().Update(ctx, quest.ID, &models.PointOfInterestGroup{
		Name:        questCopy.Name,
		Description: questCopy.Description,
		ImageUrl:    questImage,
	}); err != nil {
		log.Printf("Error updating quest: %v", err)
		if deleteErr := c.dbClient.PointOfInterestGroup().Delete(ctx, quest.ID); deleteErr != nil {
			log.Printf("Error deleting quest group after update failure: %v", deleteErr)
		}
		return nil, err
	}

	log.Printf("Successfully generated quest %s", quest.ID)
	return quest, nil
}

func (c *client) processNode(
	ctx context.Context,
	zone *models.Zone,
	questArchTypeNode *models.QuestArchetypeNode,
	locations *[]string,
	descriptions *[]string,
	challenges *[]string,
	quest *models.PointOfInterestGroup,
	member *models.PointOfInterestGroupMember,
	usedPOIs map[uuid.UUID]bool,
	prevChallenge *models.PointOfInterestChallenge,
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

	// Ensure includedTypes and excludedTypes are []googlemaps.PlaceType for SeedPointsOfInterest
	var googleIncludedTypes []googlemaps.PlaceType
	for _, t := range locationArchetype.IncludedTypes {
		googleIncludedTypes = append(googleIncludedTypes, googlemaps.PlaceType(t))
	}
	var googleExcludedTypes []googlemaps.PlaceType
	for _, t := range locationArchetype.ExcludedTypes {
		googleExcludedTypes = append(googleExcludedTypes, googlemaps.PlaceType(t))
	}

	pointsOfInterest, err := c.locationSeeder.SeedPointsOfInterest(ctx, *zone, googleIncludedTypes, googleExcludedTypes, 1)
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

	newMember, err := c.dbClient.PointOfInterestGroup().AddMember(ctx, pointOfInterest.ID, quest.ID)
	if err != nil {
		log.Printf("Error adding member to group: %v", err)
		return err
	}

	if prevChallenge != nil && member != nil {
		if err := c.dbClient.PointOfInterestChildren().Create(ctx, member.ID, newMember.ID, prevChallenge.ID); err != nil {
			log.Printf("Error creating point of interest children: %v", err)
			return err
		}
	}

	*locations = append(*locations, pointOfInterest.Name)
	*descriptions = append(*descriptions, pointOfInterest.Description)

	for i, allotedChallenge := range questArchTypeNode.Challenges {
		log.Printf("Processing challenge %d", i)

		randomChallengeText, err := questArchTypeNode.GetRandomChallenge() // This should return string
		if err != nil {
			log.Printf("Error getting random challenge: %v", err)
			return err
		}
		*challenges = append(*challenges, randomChallengeText)

		log.Printf("Creating challenge: %s", randomChallengeText)
		// The db.PointOfInterestChallengeHandle().Create() expects:
		// pointOfInterestID uuid.UUID, tier int, question string, inventoryItemID int, pointOfInterestGroupID *uuid.UUID
		// allotedChallenge is models.QuestArchetypeNodeChallenge. It should have Reward and potentially Type/Data for challenge text.
		// Assuming allotedChallenge.Challenge is the string and allotedChallenge.Reward is the int for inventoryItemID (as per test setup).
		challenge, err := c.dbClient.PointOfInterestChallenge().Create(
			ctx,
			pointOfInterest.ID,
			i, // tier
			randomChallengeText, // question
			allotedChallenge.Reward, // inventoryItemID (using Reward as per previous logic)
			&quest.ID, // pointOfInterestGroupID
		)
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
			if err := c.processNode(ctx, zone, unlockedNode, locations, descriptions, challenges, quest, newMember, usedPOIs, challenge); err != nil {
				log.Printf("Error processing child node: %v", err)
				return err
			}
		}
	}

	log.Printf("Successfully processed node for point of interest %s", pointOfInterest.ID)
	return nil
}
