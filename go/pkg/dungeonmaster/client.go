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
	log.Println("Creating new dungeonmaster client")
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

	foundPlaces := make(map[uuid.UUID]map[string]bool)

	log.Println("Processing quest nodes")
	if err := c.processNode(ctx, zone, &questArchType.Root, &locations, &descriptions, &challenges, quest, nil, foundPlaces); err != nil {
		log.Printf("Error processing quest nodes: %v", err)
		if deleteErr := c.dbClient.PointOfInterestGroup().Delete(ctx, quest.ID); deleteErr != nil {
			log.Printf("Error deleting quest group after node processing failure: %v", deleteErr)
		}
		return nil, err
	}

	log.Println("Generating quest copy")
	questCopy, err := c.generateQuestCopy(ctx, locations, descriptions, challenges)
	if err != nil {
		log.Printf("Error generating quest copy: %v", err)
		if deleteErr := c.dbClient.PointOfInterestGroup().Delete(ctx, quest.ID); deleteErr != nil {
			log.Printf("Error deleting quest group after copy generation failure: %v", deleteErr)
		}
		return nil, err
	}

	log.Println("Generating quest image")
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
	foundPlaces map[uuid.UUID]map[string]bool,
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

	count := 1
	if _, ok := foundPlaces[locationArchetype.ID]; !ok {
		foundPlaces[locationArchetype.ID] = make(map[string]bool)
	} else {
		count = len(foundPlaces[locationArchetype.ID]) + 1
	}

	pointsOfInterest, err := c.locationSeeder.SeedPointsOfInterest(ctx, *zone, locationArchetype.IncludedTypes, locationArchetype.ExcludedTypes, int32(count))
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
		if !foundPlaces[locationArchetype.ID][poi.ID.String()] {
			pointOfInterest = poi
			foundPlaces[locationArchetype.ID][poi.ID.String()] = true
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
		challenge, err := c.dbClient.PointOfInterestChallenge().Create(
			ctx,
			pointOfInterest.ID,
			i,
			randomChallenge,
			allotedChallenge.Reward,
			&quest.ID,
		)
		if err != nil {
			log.Printf("Error creating challenge: %v", err)
			return err
		}

		if member != nil {
			log.Printf("Creating point of interest children for member %s", member.ID)
			if err := c.dbClient.PointOfInterestChildren().Create(ctx, member.ID, pointOfInterest.ID, challenge.ID); err != nil {
				log.Printf("Error creating point of interest children: %v", err)
				return err
			}
		}

		if allotedChallenge.UnlockedNodeID != nil {
			unlockedNode, err := c.dbClient.QuestArchetypeNode().FindByID(ctx, *allotedChallenge.UnlockedNodeID)
			if err != nil {
				log.Printf("Error finding unlocked node: %v", err)
				return err
			}
			log.Printf("Processing child node: %s", unlockedNode.LocationArchetypeID)
			if err := c.processNode(ctx, zone, unlockedNode, locations, descriptions, challenges, quest, newMember, foundPlaces); err != nil {
				log.Printf("Error processing child node: %v", err)
				return err
			}
		}
	}

	log.Printf("Successfully processed node for point of interest %s", pointOfInterest.ID)
	return nil
}
