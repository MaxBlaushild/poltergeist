package dungeonmaster

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	mathrand "math/rand"
	"sort"
	"strconv"
	"strings"
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

type questNodeAnchor struct {
	Latitude  float64
	Longitude float64
}

func normalizeQuestProficiency(proficiency *string) *string {
	if proficiency == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*proficiency)
	if trimmed == "" {
		return nil
	}
	return &trimmed
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
	if questGiverCharacterID == nil {
		resolvedQuestGiverID, err := c.resolveQuestTemplateCharacterID(ctx, zone, questArchType)
		if err != nil {
			log.Printf("Error resolving quest giver character: %v", err)
			return nil, err
		}
		questGiverCharacterID = resolvedQuestGiverID
	}

	rewardMode := questArchType.RewardMode
	if strings.TrimSpace(string(rewardMode)) == "" {
		rewardMode = models.RewardModeRandom
	}
	randomRewardSize := questArchType.RandomRewardSize
	if strings.TrimSpace(string(randomRewardSize)) == "" {
		randomRewardSize = models.RandomRewardSizeSmall
	}
	acceptanceDialogue := questArchType.AcceptanceDialogue
	if acceptanceDialogue == nil {
		acceptanceDialogue = models.StringArray{}
	}
	var recurringQuestID *uuid.UUID
	var nextRecurrenceAt *time.Time
	if questArchType.RecurrenceFrequency != nil {
		recurrence := models.NormalizeQuestRecurrenceFrequency(*questArchType.RecurrenceFrequency)
		if recurrence != "" {
			if nextAt, ok := models.NextQuestRecurrenceAt(time.Now(), recurrence); ok {
				recurringID := uuid.New()
				recurringQuestID = &recurringID
				nextRecurrenceAt = &nextAt
			}
		}
	}

	log.Println("Creating quest")
	quest := &models.Quest{
		ID:                    uuid.New(),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		Name:                  strings.TrimSpace(questArchType.Name),
		Description:           questArchType.Description,
		AcceptanceDialogue:    acceptanceDialogue,
		ImageURL:              questArchType.ImageURL,
		ZoneID:                &zone.ID,
		QuestArchetypeID:      &questArchetypeID,
		QuestGiverCharacterID: questGiverCharacterID,
		RecurringQuestID:      recurringQuestID,
		RecurrenceFrequency:   questArchType.RecurrenceFrequency,
		NextRecurrenceAt:      nextRecurrenceAt,
		RewardMode:            rewardMode,
		RandomRewardSize:      randomRewardSize,
		RewardExperience:      questArchType.RewardExperience,
		Gold:                  questArchType.DefaultGold,
		MaterialRewards:       questArchType.MaterialRewards,
	}
	if quest.Name == "" {
		quest.Name = "Quest"
	}
	if strings.TrimSpace(quest.Description) == "" {
		quest.Description = "A quest to complete"
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
	anchorMap := make(map[uuid.UUID]*questNodeAnchor)
	if _, err := c.processQuestNode(ctx, zone, &questArchType.Root, quest, usedPOIs, &orderIndex, nodeMap, anchorMap, nil); err != nil {
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
	quest *models.Quest,
	usedPOIs map[uuid.UUID]bool,
	orderIndex *int,
	nodeMap map[uuid.UUID]uuid.UUID,
	anchorMap map[uuid.UUID]*questNodeAnchor,
	previousAnchor *questNodeAnchor,
) (*questNodeAnchor, error) {
	if questArchTypeNode == nil {
		return previousAnchor, fmt.Errorf("quest archetype node is required")
	}
	if questArchTypeNode.NodeType == models.QuestArchetypeNodeTypeScenario {
		return c.processQuestScenarioNode(
			ctx,
			zone,
			questArchTypeNode,
			quest,
			usedPOIs,
			orderIndex,
			nodeMap,
			anchorMap,
			previousAnchor,
		)
	}
	if questArchTypeNode.NodeType == models.QuestArchetypeNodeTypeMonsterEncounter {
		return c.processQuestMonsterEncounterNode(
			ctx,
			zone,
			questArchTypeNode,
			quest,
			usedPOIs,
			orderIndex,
			nodeMap,
			anchorMap,
			previousAnchor,
		)
	}
	return c.processQuestLocationNode(
		ctx,
		zone,
		questArchTypeNode,
		quest,
		usedPOIs,
		orderIndex,
		nodeMap,
		anchorMap,
		previousAnchor,
	)
}

func (c *client) processQuestLocationNode(
	ctx context.Context,
	zone *models.Zone,
	questArchTypeNode *models.QuestArchetypeNode,
	quest *models.Quest,
	usedPOIs map[uuid.UUID]bool,
	orderIndex *int,
	nodeMap map[uuid.UUID]uuid.UUID,
	anchorMap map[uuid.UUID]*questNodeAnchor,
	previousAnchor *questNodeAnchor,
) (*questNodeAnchor, error) {
	if questArchTypeNode.LocationArchetypeID == nil || *questArchTypeNode.LocationArchetypeID == uuid.Nil {
		return previousAnchor, fmt.Errorf("location node is missing a location archetype")
	}
	log.Printf("Processing location node for zone %s with place type %s", zone.Name, questArchTypeNode.LocationArchetypeID.String())

	locationArchetype, err := c.dbClient.LocationArchetype().FindByID(ctx, *questArchTypeNode.LocationArchetypeID)
	if err != nil {
		log.Printf("Error finding location archetype: %v", err)
		return previousAnchor, err
	}
	pointsOfInterest, err := c.locationSeeder.SeedPointsOfInterest(ctx, *zone, locationArchetype.IncludedTypes, locationArchetype.ExcludedTypes, 1)
	if err != nil {
		log.Printf("Error seeding points of interest: %v", err)
		return previousAnchor, err
	}
	if len(pointsOfInterest) == 0 {
		return previousAnchor, errors.New("no points of interest found")
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
		return previousAnchor, fmt.Errorf("no unused points of interest found for type %s", locationArchetype.ID)
	}
	if err := c.dbClient.PointOfInterest().UpdateLastUsedInQuest(ctx, pointOfInterest.ID); err != nil {
		log.Printf("Warning: failed to update last_used_in_quest_at for POI %s: %v", pointOfInterest.ID, err)
	}

	latitude, longitude, err := pointOfInterestCoordinates(pointOfInterest)
	if err != nil {
		return previousAnchor, err
	}
	currentAnchor := &questNodeAnchor{Latitude: latitude, Longitude: longitude}

	existingNodeID, ok := nodeMap[questArchTypeNode.ID]
	var questNodeID uuid.UUID
	if ok {
		questNodeID = existingNodeID
		if existingAnchor, exists := anchorMap[questArchTypeNode.ID]; exists && existingAnchor != nil {
			currentAnchor = existingAnchor
		}
	} else {
		questNodeID = uuid.New()
		submissionType := models.DefaultQuestNodeSubmissionType()
		locationChallenge, err := c.makeQuestLocationChallenge(zone.ID, pointOfInterest, submissionType)
		if err != nil {
			return previousAnchor, err
		}
		if err := c.dbClient.Challenge().Create(ctx, locationChallenge); err != nil {
			return previousAnchor, err
		}
		node := &models.QuestNode{
			ID:             questNodeID,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			QuestID:        quest.ID,
			OrderIndex:     *orderIndex,
			ChallengeID:    &locationChallenge.ID,
			SubmissionType: submissionType,
		}
		if err := c.dbClient.QuestNode().Create(ctx, node); err != nil {
			return previousAnchor, err
		}
		nodeMap[questArchTypeNode.ID] = questNodeID
		anchorMap[questArchTypeNode.ID] = currentAnchor
		(*orderIndex)++
	}

	for i, allotedChallenge := range questArchTypeNode.Challenges {
		resolvedChallenge, err := c.resolveQuestArchetypeLocationChallenge(
			ctx,
			questArchTypeNode,
			&allotedChallenge,
		)
		if err != nil {
			return currentAnchor, err
		}
		challenge := &models.QuestNodeChallenge{
			ID:              uuid.New(),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			QuestNodeID:     questNodeID,
			Tier:            i,
			Question:        resolvedChallenge.Question,
			Reward:          allotedChallenge.Reward,
			InventoryItemID: allotedChallenge.InventoryItemID,
			Difficulty:      resolvedChallenge.Difficulty,
			StatTags:        resolvedChallenge.StatTags,
			Proficiency:     resolvedChallenge.Proficiency,
			SubmissionType:  resolvedChallenge.SubmissionType,
		}
		if err := c.dbClient.QuestNodeChallenge().Create(ctx, challenge); err != nil {
			return currentAnchor, err
		}
		if allotedChallenge.UnlockedNodeID != nil {
			unlockedNode, err := c.dbClient.QuestArchetypeNode().FindByID(ctx, *allotedChallenge.UnlockedNodeID)
			if err != nil {
				return currentAnchor, err
			}
			if _, err := c.processQuestNode(ctx, zone, unlockedNode, quest, usedPOIs, orderIndex, nodeMap, anchorMap, currentAnchor); err != nil {
				return currentAnchor, err
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
				return currentAnchor, err
			}
		}
	}

	return currentAnchor, nil
}

func (c *client) processQuestScenarioNode(
	ctx context.Context,
	zone *models.Zone,
	questArchTypeNode *models.QuestArchetypeNode,
	quest *models.Quest,
	usedPOIs map[uuid.UUID]bool,
	orderIndex *int,
	nodeMap map[uuid.UUID]uuid.UUID,
	anchorMap map[uuid.UUID]*questNodeAnchor,
	previousAnchor *questNodeAnchor,
) (*questNodeAnchor, error) {
	existingNodeID, ok := nodeMap[questArchTypeNode.ID]
	if ok {
		currentAnchor := anchorMap[questArchTypeNode.ID]
		if currentAnchor == nil {
			currentAnchor = previousAnchor
		}
		return currentAnchor, c.attachQuestMonsterEncounterChildren(
			ctx,
			zone,
			questArchTypeNode,
			quest,
			usedPOIs,
			orderIndex,
			nodeMap,
			anchorMap,
			currentAnchor,
			existingNodeID,
		)
	}

	if questArchTypeNode.ScenarioTemplateID == nil || *questArchTypeNode.ScenarioTemplateID == uuid.Nil {
		return previousAnchor, fmt.Errorf("scenario node requires a scenario template")
	}
	template, err := c.dbClient.ScenarioTemplate().FindByID(ctx, *questArchTypeNode.ScenarioTemplateID)
	if err != nil {
		return previousAnchor, err
	}
	if template == nil {
		return previousAnchor, fmt.Errorf("scenario template %s not found", questArchTypeNode.ScenarioTemplateID.String())
	}

	currentAnchor, _, err := c.resolveQuestNodeAnchor(
		ctx,
		zone,
		questArchTypeNode,
		usedPOIs,
		previousAnchor,
	)
	if err != nil {
		return previousAnchor, err
	}
	thumbnailURL := strings.TrimSpace(template.ThumbnailURL)
	if thumbnailURL == "" {
		thumbnailURL = strings.TrimSpace(template.ImageURL)
	}

	scenario := &models.Scenario{
		ID:                        uuid.New(),
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
		ZoneID:                    zone.ID,
		Latitude:                  currentAnchor.Latitude,
		Longitude:                 currentAnchor.Longitude,
		Prompt:                    strings.TrimSpace(template.Prompt),
		InternalTags:              models.StringArray{},
		ImageURL:                  strings.TrimSpace(template.ImageURL),
		ThumbnailURL:              thumbnailURL,
		ScaleWithUserLevel:        template.ScaleWithUserLevel,
		RewardMode:                template.RewardMode,
		RandomRewardSize:          template.RandomRewardSize,
		Difficulty:                template.Difficulty,
		RewardExperience:          template.RewardExperience,
		RewardGold:                template.RewardGold,
		MaterialRewards:           models.BaseMaterialRewards{},
		OpenEnded:                 template.OpenEnded,
		FailurePenaltyMode:        template.FailurePenaltyMode,
		FailureHealthDrainType:    template.FailureHealthDrainType,
		FailureHealthDrainValue:   template.FailureHealthDrainValue,
		FailureManaDrainType:      template.FailureManaDrainType,
		FailureManaDrainValue:     template.FailureManaDrainValue,
		FailureStatuses:           cloneScenarioFailureStatuses(template.FailureStatuses),
		SuccessRewardMode:         template.SuccessRewardMode,
		SuccessHealthRestoreType:  template.SuccessHealthRestoreType,
		SuccessHealthRestoreValue: template.SuccessHealthRestoreValue,
		SuccessManaRestoreType:    template.SuccessManaRestoreType,
		SuccessManaRestoreValue:   template.SuccessManaRestoreValue,
		SuccessStatuses:           cloneScenarioFailureStatuses(template.SuccessStatuses),
		Ephemeral:                 false,
	}
	if err := c.dbClient.Scenario().Create(ctx, scenario); err != nil {
		return previousAnchor, err
	}
	if err := c.dbClient.Scenario().ReplaceOptions(ctx, scenario.ID, scenarioOptionsFromTemplate(template.Options)); err != nil {
		return previousAnchor, err
	}
	if err := c.dbClient.Scenario().ReplaceItemRewards(ctx, scenario.ID, scenarioItemRewardsFromTemplate(template.ItemRewards)); err != nil {
		return previousAnchor, err
	}
	if err := c.dbClient.Scenario().ReplaceItemChoiceRewards(ctx, scenario.ID, scenarioItemChoiceRewardsFromTemplate(template.ItemChoiceRewards)); err != nil {
		return previousAnchor, err
	}
	if err := c.dbClient.Scenario().ReplaceSpellRewards(ctx, scenario.ID, scenarioSpellRewardsFromTemplate(template.SpellRewards)); err != nil {
		return previousAnchor, err
	}

	questNodeID := uuid.New()
	node := &models.QuestNode{
		ID:             questNodeID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		QuestID:        quest.ID,
		OrderIndex:     *orderIndex,
		ScenarioID:     &scenario.ID,
		SubmissionType: models.DefaultQuestNodeSubmissionType(),
	}
	if err := c.dbClient.QuestNode().Create(ctx, node); err != nil {
		return previousAnchor, err
	}
	nodeMap[questArchTypeNode.ID] = questNodeID
	anchorMap[questArchTypeNode.ID] = currentAnchor
	(*orderIndex)++

	if err := c.attachQuestMonsterEncounterChildren(
		ctx,
		zone,
		questArchTypeNode,
		quest,
		usedPOIs,
		orderIndex,
		nodeMap,
		anchorMap,
		currentAnchor,
		questNodeID,
	); err != nil {
		return currentAnchor, err
	}

	return currentAnchor, nil
}

func (c *client) processQuestMonsterEncounterNode(
	ctx context.Context,
	zone *models.Zone,
	questArchTypeNode *models.QuestArchetypeNode,
	quest *models.Quest,
	usedPOIs map[uuid.UUID]bool,
	orderIndex *int,
	nodeMap map[uuid.UUID]uuid.UUID,
	anchorMap map[uuid.UUID]*questNodeAnchor,
	previousAnchor *questNodeAnchor,
) (*questNodeAnchor, error) {
	existingNodeID, ok := nodeMap[questArchTypeNode.ID]
	if ok {
		currentAnchor := anchorMap[questArchTypeNode.ID]
		if currentAnchor == nil {
			currentAnchor = previousAnchor
		}
		return currentAnchor, c.attachQuestMonsterEncounterChildren(
			ctx,
			zone,
			questArchTypeNode,
			quest,
			usedPOIs,
			orderIndex,
			nodeMap,
			anchorMap,
			currentAnchor,
			existingNodeID,
		)
	}

	sourceTemplates := make([]models.MonsterTemplate, 0, len(questArchTypeNode.MonsterTemplateIDs))
	sourceMonsters := make([]models.Monster, 0, len(questArchTypeNode.MonsterTemplateIDs))
	for _, rawID := range questArchTypeNode.MonsterTemplateIDs {
		templateID, err := uuid.Parse(strings.TrimSpace(rawID))
		if err != nil {
			return previousAnchor, fmt.Errorf("invalid monster template id %q in quest archetype node", rawID)
		}
		template, err := c.dbClient.MonsterTemplate().FindByID(ctx, templateID)
		if err != nil {
			return previousAnchor, err
		}
		sourceTemplates = append(sourceTemplates, *template)
		sourceMonsters = append(sourceMonsters, models.Monster{
			Name:         strings.TrimSpace(template.Name),
			Description:  strings.TrimSpace(template.Description),
			ImageURL:     strings.TrimSpace(template.ImageURL),
			ThumbnailURL: strings.TrimSpace(template.ThumbnailURL),
			TemplateID:   &template.ID,
			Template:     template,
		})
	}
	if len(sourceTemplates) == 0 {
		return previousAnchor, fmt.Errorf("monster encounter node requires at least one monster template")
	}
	if len(sourceTemplates) > 9 {
		return previousAnchor, fmt.Errorf("monster encounter node cannot include more than 9 monster templates")
	}

	currentAnchor, _, err := c.resolveQuestNodeAnchor(
		ctx,
		zone,
		questArchTypeNode,
		usedPOIs,
		previousAnchor,
	)
	if err != nil {
		return previousAnchor, err
	}
	createdMonsters := make([]models.Monster, 0, len(sourceTemplates))
	members := make([]models.MonsterEncounterMember, 0, len(sourceTemplates))
	for index, source := range sourceTemplates {
		templateID := source.ID
		imageURL := strings.TrimSpace(source.ImageURL)
		thumbnailURL := strings.TrimSpace(source.ThumbnailURL)
		if thumbnailURL == "" {
			thumbnailURL = imageURL
		}
		monster := models.Monster{
			ID:               uuid.New(),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			Name:             strings.TrimSpace(source.Name),
			Description:      strings.TrimSpace(source.Description),
			ImageURL:         imageURL,
			ThumbnailURL:     thumbnailURL,
			Ephemeral:        false,
			ZoneID:           zone.ID,
			Latitude:         currentAnchor.Latitude,
			Longitude:        currentAnchor.Longitude,
			TemplateID:       &templateID,
			Level:            maxInt(1, questArchTypeNode.TargetLevel),
			RewardMode:       models.RewardModeExplicit,
			RandomRewardSize: models.RandomRewardSizeSmall,
			RewardExperience: 0,
			RewardGold:       0,
			MaterialRewards:  models.BaseMaterialRewards{},
			ItemRewards:      []models.MonsterItemReward{},
		}
		if imageURL != "" {
			monster.ImageGenerationStatus = models.MonsterImageGenerationStatusComplete
			emptyError := ""
			monster.ImageGenerationError = &emptyError
		} else {
			monster.ImageGenerationStatus = models.MonsterImageGenerationStatusNone
		}
		if err := c.dbClient.Monster().Create(ctx, &monster); err != nil {
			return previousAnchor, err
		}
		createdMonsters = append(createdMonsters, monster)
		members = append(members, models.MonsterEncounterMember{
			Slot:      index + 1,
			MonsterID: monster.ID,
		})
	}

	encounter := &models.MonsterEncounter{
		ID:                 uuid.New(),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		Name:               buildQuestEncounterName(sourceMonsters),
		Description:        buildQuestEncounterDescription(sourceMonsters),
		ImageURL:           firstNonEmptyMonsterImage(createdMonsters),
		ThumbnailURL:       firstNonEmptyMonsterThumbnail(createdMonsters),
		EncounterType:      deriveQuestEncounterType(sourceMonsters),
		Ephemeral:          false,
		ScaleWithUserLevel: false,
		ZoneID:             zone.ID,
		Latitude:           currentAnchor.Latitude,
		Longitude:          currentAnchor.Longitude,
		RewardMode:         questArchTypeNode.EncounterRewardMode,
		RandomRewardSize:   questArchTypeNode.EncounterRandomRewardSize,
		RewardExperience:   questArchTypeNode.EncounterRewardExperience,
		RewardGold:         questArchTypeNode.EncounterRewardGold,
		MaterialRewards:    questArchTypeNode.EncounterMaterialRewards,
		ItemRewards:        questArchTypeNode.EncounterItemRewards,
	}
	if err := c.dbClient.MonsterEncounter().Create(ctx, encounter); err != nil {
		return previousAnchor, err
	}
	if err := c.dbClient.MonsterEncounter().ReplaceMembers(ctx, encounter.ID, members); err != nil {
		return previousAnchor, err
	}

	questNodeID := uuid.New()
	node := &models.QuestNode{
		ID:                 questNodeID,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		QuestID:            quest.ID,
		OrderIndex:         *orderIndex,
		MonsterEncounterID: &encounter.ID,
		SubmissionType:     models.DefaultQuestNodeSubmissionType(),
	}
	if err := c.dbClient.QuestNode().Create(ctx, node); err != nil {
		return previousAnchor, err
	}
	nodeMap[questArchTypeNode.ID] = questNodeID
	anchorMap[questArchTypeNode.ID] = currentAnchor
	(*orderIndex)++

	if err := c.attachQuestMonsterEncounterChildren(
		ctx,
		zone,
		questArchTypeNode,
		quest,
		usedPOIs,
		orderIndex,
		nodeMap,
		anchorMap,
		currentAnchor,
		questNodeID,
	); err != nil {
		return currentAnchor, err
	}

	return currentAnchor, nil
}

func (c *client) attachQuestMonsterEncounterChildren(
	ctx context.Context,
	zone *models.Zone,
	questArchTypeNode *models.QuestArchetypeNode,
	quest *models.Quest,
	usedPOIs map[uuid.UUID]bool,
	orderIndex *int,
	nodeMap map[uuid.UUID]uuid.UUID,
	anchorMap map[uuid.UUID]*questNodeAnchor,
	currentAnchor *questNodeAnchor,
	questNodeID uuid.UUID,
) error {
	for _, archetypeChallenge := range questArchTypeNode.Challenges {
		if archetypeChallenge.UnlockedNodeID == nil {
			continue
		}
		unlockedNode, err := c.dbClient.QuestArchetypeNode().FindByID(ctx, *archetypeChallenge.UnlockedNodeID)
		if err != nil {
			return err
		}
		if _, err := c.processQuestNode(ctx, zone, unlockedNode, quest, usedPOIs, orderIndex, nodeMap, anchorMap, currentAnchor); err != nil {
			return err
		}
		childNodeID := nodeMap[unlockedNode.ID]
		child := &models.QuestNodeChild{
			ID:              uuid.New(),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			QuestNodeID:     questNodeID,
			NextQuestNodeID: childNodeID,
		}
		if err := c.dbClient.QuestNodeChild().Create(ctx, child); err != nil {
			return err
		}
	}
	return nil
}

func pointOfInterestCoordinates(poi *models.PointOfInterest) (float64, float64, error) {
	if poi == nil {
		return 0, 0, fmt.Errorf("point of interest is required")
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(poi.Lat), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid point of interest latitude: %w", err)
	}
	lng, err := strconv.ParseFloat(strings.TrimSpace(poi.Lng), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid point of interest longitude: %w", err)
	}
	return lat, lng, nil
}

func (c *client) resolveQuestNodeAnchor(
	ctx context.Context,
	zone *models.Zone,
	questArchTypeNode *models.QuestArchetypeNode,
	usedPOIs map[uuid.UUID]bool,
	previousAnchor *questNodeAnchor,
) (*questNodeAnchor, *models.PointOfInterest, error) {
	if questArchTypeNode != nil &&
		questArchTypeNode.LocationArchetypeID != nil &&
		*questArchTypeNode.LocationArchetypeID != uuid.Nil {
		locationArchetype, err := c.dbClient.LocationArchetype().FindByID(ctx, *questArchTypeNode.LocationArchetypeID)
		if err != nil {
			return previousAnchor, nil, err
		}
		if locationArchetype == nil {
			return previousAnchor, nil, fmt.Errorf("location archetype %s not found", questArchTypeNode.LocationArchetypeID.String())
		}
		pointsOfInterest, err := c.locationSeeder.SeedPointsOfInterest(
			ctx,
			*zone,
			locationArchetype.IncludedTypes,
			locationArchetype.ExcludedTypes,
			1,
		)
		if err != nil {
			return previousAnchor, nil, err
		}
		if len(pointsOfInterest) == 0 {
			return previousAnchor, nil, errors.New("no points of interest found")
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
			return previousAnchor, nil, fmt.Errorf("no unused points of interest found for type %s", locationArchetype.ID)
		}
		if err := c.dbClient.PointOfInterest().UpdateLastUsedInQuest(ctx, pointOfInterest.ID); err != nil {
			log.Printf("Warning: failed to update last_used_in_quest_at for POI %s: %v", pointOfInterest.ID, err)
		}
		latitude, longitude, err := pointOfInterestCoordinates(pointOfInterest)
		if err != nil {
			return previousAnchor, nil, err
		}
		return &questNodeAnchor{Latitude: latitude, Longitude: longitude}, pointOfInterest, nil
	}
	return randomQuestEncounterPoint(zone, previousAnchor, questArchTypeNode.EncounterProximityMeters), nil, nil
}

func randomQuestEncounterPoint(
	zone *models.Zone,
	previousAnchor *questNodeAnchor,
	proximityMeters int,
) *questNodeAnchor {
	if zone == nil {
		return &questNodeAnchor{}
	}
	if previousAnchor == nil {
		point := zone.GetRandomPoint()
		return &questNodeAnchor{Latitude: point.Y(), Longitude: point.X()}
	}
	maxDistance := maxInt(0, proximityMeters)
	for attempt := 0; attempt < 24; attempt++ {
		lat, lng := randomPointNear(previousAnchor.Latitude, previousAnchor.Longitude, float64(maxDistance))
		if zone.IsPointInBoundary(lat, lng) {
			return &questNodeAnchor{Latitude: lat, Longitude: lng}
		}
	}
	point := zone.GetRandomPoint()
	return &questNodeAnchor{Latitude: point.Y(), Longitude: point.X()}
}

func randomPointNear(latitude float64, longitude float64, maxDistanceMeters float64) (float64, float64) {
	if maxDistanceMeters <= 0 {
		return latitude, longitude
	}
	rng := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	distanceMeters := rng.Float64() * maxDistanceMeters
	bearingRadians := rng.Float64() * 2 * math.Pi
	const earthRadiusMeters = 6371000.0
	angularDistance := distanceMeters / earthRadiusMeters
	startLat := latitude * math.Pi / 180.0
	startLng := longitude * math.Pi / 180.0
	endLat := math.Asin(
		math.Sin(startLat)*math.Cos(angularDistance) +
			math.Cos(startLat)*math.Sin(angularDistance)*math.Cos(bearingRadians),
	)
	endLng := startLng + math.Atan2(
		math.Sin(bearingRadians)*math.Sin(angularDistance)*math.Cos(startLat),
		math.Cos(angularDistance)-math.Sin(startLat)*math.Sin(endLat),
	)
	endLng = math.Mod(endLng+3*math.Pi, 2*math.Pi) - math.Pi
	return endLat * 180.0 / math.Pi, endLng * 180.0 / math.Pi
}

func deriveQuestEncounterType(monsters []models.Monster) models.MonsterEncounterType {
	hasBoss := false
	hasRaid := false
	for _, monster := range monsters {
		if monster.Template == nil {
			continue
		}
		switch monster.Template.MonsterType {
		case models.MonsterTemplateTypeRaid:
			hasRaid = true
		case models.MonsterTemplateTypeBoss:
			hasBoss = true
		}
	}
	if hasRaid {
		return models.MonsterEncounterTypeRaid
	}
	if hasBoss {
		return models.MonsterEncounterTypeBoss
	}
	return models.MonsterEncounterTypeMonster
}

type resolvedQuestArchetypeLocationChallenge struct {
	Question       string
	SubmissionType models.QuestNodeSubmissionType
	Difficulty     int
	StatTags       models.StringArray
	Proficiency    *string
}

func (c *client) resolveQuestArchetypeLocationChallenge(
	ctx context.Context,
	questArchTypeNode *models.QuestArchetypeNode,
	allotedChallenge *models.QuestArchetypeChallenge,
) (*resolvedQuestArchetypeLocationChallenge, error) {
	if questArchTypeNode == nil {
		return nil, fmt.Errorf("quest archetype node is required")
	}
	if allotedChallenge == nil {
		return nil, fmt.Errorf("quest archetype challenge is required")
	}
	if allotedChallenge.ChallengeTemplate != nil {
		return resolvedQuestArchetypeLocationChallengeFromTemplate(allotedChallenge.ChallengeTemplate)
	}
	if allotedChallenge.ChallengeTemplateID != nil && *allotedChallenge.ChallengeTemplateID != uuid.Nil {
		template, err := c.dbClient.ChallengeTemplate().FindByID(ctx, *allotedChallenge.ChallengeTemplateID)
		if err != nil {
			return nil, err
		}
		if template == nil {
			return nil, fmt.Errorf("challenge template %s not found", allotedChallenge.ChallengeTemplateID.String())
		}
		return resolvedQuestArchetypeLocationChallengeFromTemplate(template)
	}

	randomChallenge, err := questArchTypeNode.GetRandomChallenge()
	if err != nil {
		return nil, err
	}
	proficiency := normalizeQuestProficiency(allotedChallenge.Proficiency)
	if randomChallenge.Proficiency != nil {
		proficiency = normalizeQuestProficiency(randomChallenge.Proficiency)
	}
	submissionType := randomChallenge.SubmissionType
	if !submissionType.IsValid() {
		submissionType = models.DefaultQuestNodeSubmissionType()
	}
	return &resolvedQuestArchetypeLocationChallenge{
		Question:       randomChallenge.Question,
		SubmissionType: submissionType,
		Difficulty:     allotedChallenge.Difficulty,
		StatTags:       models.StringArray{},
		Proficiency:    proficiency,
	}, nil
}

func resolvedQuestArchetypeLocationChallengeFromTemplate(
	template *models.ChallengeTemplate,
) (*resolvedQuestArchetypeLocationChallenge, error) {
	if template == nil {
		return nil, fmt.Errorf("challenge template is required")
	}
	question := strings.TrimSpace(template.Question)
	if question == "" {
		return nil, fmt.Errorf("challenge template question is required")
	}
	submissionType := template.SubmissionType
	if !submissionType.IsValid() {
		submissionType = models.DefaultQuestNodeSubmissionType()
	}
	statTags := append(models.StringArray{}, template.StatTags...)
	return &resolvedQuestArchetypeLocationChallenge{
		Question:       question,
		SubmissionType: submissionType,
		Difficulty:     template.Difficulty,
		StatTags:       statTags,
		Proficiency:    normalizeQuestProficiency(template.Proficiency),
	}, nil
}

func buildQuestEncounterName(monsters []models.Monster) string {
	names := make([]string, 0, len(monsters))
	for _, monster := range monsters {
		name := strings.TrimSpace(monster.Name)
		if name == "" {
			continue
		}
		names = append(names, name)
		if len(names) == 3 {
			break
		}
	}
	if len(names) == 0 {
		return "Monster Encounter"
	}
	if len(monsters) > len(names) {
		return fmt.Sprintf("Monster Encounter: %s and more", strings.Join(names, ", "))
	}
	return fmt.Sprintf("Monster Encounter: %s", strings.Join(names, ", "))
}

func buildQuestEncounterDescription(monsters []models.Monster) string {
	names := make([]string, 0, len(monsters))
	for _, monster := range monsters {
		name := strings.TrimSpace(monster.Name)
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return "Hostile monsters block the path forward."
	}
	return fmt.Sprintf("Defeat %s to continue the quest.", strings.Join(names, ", "))
}

func firstNonEmptyMonsterImage(monsters []models.Monster) string {
	for _, monster := range monsters {
		if imageURL := strings.TrimSpace(monster.ImageURL); imageURL != "" {
			return imageURL
		}
	}
	return ""
}

func firstNonEmptyMonsterThumbnail(monsters []models.Monster) string {
	for _, monster := range monsters {
		if thumbnailURL := strings.TrimSpace(monster.ThumbnailURL); thumbnailURL != "" {
			return thumbnailURL
		}
	}
	return firstNonEmptyMonsterImage(monsters)
}

func scenarioOptionsFromTemplate(input models.ScenarioTemplateOptions) []models.ScenarioOption {
	options := make([]models.ScenarioOption, 0, len(input))
	for _, option := range input {
		options = append(options, models.ScenarioOption{
			OptionText:                strings.TrimSpace(option.OptionText),
			SuccessText:               strings.TrimSpace(option.SuccessText),
			FailureText:               strings.TrimSpace(option.FailureText),
			StatTag:                   option.StatTag,
			Proficiencies:             cloneStringArray(option.Proficiencies),
			Difficulty:                cloneOptionalInt(option.Difficulty),
			RewardExperience:          option.RewardExperience,
			RewardGold:                option.RewardGold,
			MaterialRewards:           models.BaseMaterialRewards{},
			FailureHealthDrainType:    option.FailureHealthDrainType,
			FailureHealthDrainValue:   option.FailureHealthDrainValue,
			FailureManaDrainType:      option.FailureManaDrainType,
			FailureManaDrainValue:     option.FailureManaDrainValue,
			FailureStatuses:           cloneScenarioFailureStatuses(option.FailureStatuses),
			SuccessHealthRestoreType:  option.SuccessHealthRestoreType,
			SuccessHealthRestoreValue: option.SuccessHealthRestoreValue,
			SuccessManaRestoreType:    option.SuccessManaRestoreType,
			SuccessManaRestoreValue:   option.SuccessManaRestoreValue,
			SuccessStatuses:           cloneScenarioFailureStatuses(option.SuccessStatuses),
			ItemRewards:               scenarioOptionItemRewardsFromTemplate(option.ItemRewards),
			ItemChoiceRewards:         scenarioOptionItemChoiceRewardsFromTemplate(option.ItemChoiceRewards),
			SpellRewards:              scenarioOptionSpellRewardsFromTemplate(option.SpellRewards),
		})
	}
	return options
}

func scenarioItemRewardsFromTemplate(input models.ScenarioTemplateRewards) []models.ScenarioItemReward {
	rewards := make([]models.ScenarioItemReward, 0, len(input))
	for _, reward := range input {
		if reward.InventoryItemID <= 0 || reward.Quantity <= 0 {
			continue
		}
		rewards = append(rewards, models.ScenarioItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return rewards
}

func scenarioItemChoiceRewardsFromTemplate(input models.ScenarioTemplateRewards) []models.ScenarioItemChoiceReward {
	rewards := make([]models.ScenarioItemChoiceReward, 0, len(input))
	for _, reward := range input {
		if reward.InventoryItemID <= 0 || reward.Quantity <= 0 {
			continue
		}
		rewards = append(rewards, models.ScenarioItemChoiceReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return rewards
}

func scenarioSpellRewardsFromTemplate(input models.ScenarioTemplateSpellRewards) []models.ScenarioSpellReward {
	rewards := make([]models.ScenarioSpellReward, 0, len(input))
	for _, reward := range input {
		if reward.SpellID == uuid.Nil {
			continue
		}
		rewards = append(rewards, models.ScenarioSpellReward{
			SpellID: reward.SpellID,
		})
	}
	return rewards
}

func scenarioOptionItemRewardsFromTemplate(input models.ScenarioTemplateRewards) []models.ScenarioOptionItemReward {
	rewards := make([]models.ScenarioOptionItemReward, 0, len(input))
	for _, reward := range input {
		if reward.InventoryItemID <= 0 || reward.Quantity <= 0 {
			continue
		}
		rewards = append(rewards, models.ScenarioOptionItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return rewards
}

func scenarioOptionItemChoiceRewardsFromTemplate(input models.ScenarioTemplateRewards) []models.ScenarioOptionItemChoiceReward {
	rewards := make([]models.ScenarioOptionItemChoiceReward, 0, len(input))
	for _, reward := range input {
		if reward.InventoryItemID <= 0 || reward.Quantity <= 0 {
			continue
		}
		rewards = append(rewards, models.ScenarioOptionItemChoiceReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	return rewards
}

func scenarioOptionSpellRewardsFromTemplate(input models.ScenarioTemplateSpellRewards) []models.ScenarioOptionSpellReward {
	rewards := make([]models.ScenarioOptionSpellReward, 0, len(input))
	for _, reward := range input {
		if reward.SpellID == uuid.Nil {
			continue
		}
		rewards = append(rewards, models.ScenarioOptionSpellReward{
			SpellID: reward.SpellID,
		})
	}
	return rewards
}

func cloneScenarioFailureStatuses(input models.ScenarioFailureStatusTemplates) models.ScenarioFailureStatusTemplates {
	if len(input) == 0 {
		return models.ScenarioFailureStatusTemplates{}
	}
	out := make(models.ScenarioFailureStatusTemplates, 0, len(input))
	for _, status := range input {
		statusCopy := status
		out = append(out, statusCopy)
	}
	return out
}

func cloneStringArray(input models.StringArray) models.StringArray {
	if len(input) == 0 {
		return models.StringArray{}
	}
	out := make(models.StringArray, 0, len(input))
	out = append(out, input...)
	return out
}

func cloneOptionalInt(value *int) *int {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (c *client) makeQuestLocationChallenge(zoneID uuid.UUID, poi *models.PointOfInterest, submissionType models.QuestNodeSubmissionType) (*models.Challenge, error) {
	if poi == nil {
		return nil, fmt.Errorf("point of interest is required")
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(poi.Lat), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid point of interest latitude: %w", err)
	}
	lng, err := strconv.ParseFloat(strings.TrimSpace(poi.Lng), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid point of interest longitude: %w", err)
	}
	question := fmt.Sprintf("Visit %s and share photo proof of your arrival.", strings.TrimSpace(poi.Name))
	if strings.TrimSpace(poi.Name) == "" {
		question = "Visit this location and share photo proof of your arrival."
	}
	description := strings.TrimSpace(poi.Description)
	now := time.Now()
	return &models.Challenge{
		ID:             uuid.New(),
		CreatedAt:      now,
		UpdatedAt:      now,
		ZoneID:         zoneID,
		Latitude:       lat,
		Longitude:      lng,
		Question:       question,
		Description:    description,
		SubmissionType: submissionType,
		Reward:         0,
		Difficulty:     0,
		StatTags:       models.StringArray{},
	}, nil
}

func (c *client) applyQuestArchetypeRewards(ctx context.Context, questID uuid.UUID, questArchetype *models.QuestArchetype) error {
	if questArchetype == nil {
		return nil
	}
	itemRewards := questArchetype.ItemRewards
	if len(itemRewards) == 0 {
		loaded, err := c.dbClient.QuestArchetypeItemReward().FindByQuestArchetypeID(ctx, questArchetype.ID)
		if err == nil {
			itemRewards = loaded
		}
	}
	now := time.Now()
	questItemRewards := make([]models.QuestItemReward, 0, len(itemRewards))
	for _, reward := range itemRewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		questItemRewards = append(questItemRewards, models.QuestItemReward{
			ID:              uuid.New(),
			CreatedAt:       now,
			UpdatedAt:       now,
			QuestID:         questID,
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	if len(questItemRewards) > 0 {
		if err := c.dbClient.QuestItemReward().ReplaceForQuest(ctx, questID, questItemRewards); err != nil {
			return err
		}
	}

	spellRewards := questArchetype.SpellRewards
	if len(spellRewards) == 0 {
		loaded, err := c.dbClient.QuestArchetypeSpellReward().FindByQuestArchetypeID(ctx, questArchetype.ID)
		if err == nil {
			spellRewards = loaded
		}
	}
	questSpellRewards := make([]models.QuestSpellReward, 0, len(spellRewards))
	for _, reward := range spellRewards {
		if reward.SpellID == uuid.Nil {
			continue
		}
		questSpellRewards = append(questSpellRewards, models.QuestSpellReward{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			QuestID:   questID,
			SpellID:   reward.SpellID,
		})
	}
	if len(questSpellRewards) > 0 {
		if err := c.dbClient.QuestSpellReward().ReplaceForQuest(ctx, questID, questSpellRewards); err != nil {
			return err
		}
	}
	return nil
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

func (c *client) resolveQuestTemplateCharacterID(
	ctx context.Context,
	zone *models.Zone,
	questArchetype *models.QuestArchetype,
) (*uuid.UUID, error) {
	if zone == nil || questArchetype == nil || len(questArchetype.CharacterTags) == 0 {
		return nil, nil
	}

	pointsOfInterest, err := c.dbClient.PointOfInterest().FindAllForZone(ctx, zone.ID)
	if err != nil {
		return nil, err
	}
	pointOfInterestIDs := make(map[uuid.UUID]struct{}, len(pointsOfInterest))
	for _, poi := range pointsOfInterest {
		pointOfInterestIDs[poi.ID] = struct{}{}
	}

	characters, err := c.dbClient.Character().FindAll(ctx)
	if err != nil {
		return nil, err
	}

	desiredTags := make(map[string]struct{}, len(questArchetype.CharacterTags))
	for _, tag := range questArchetype.CharacterTags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		desiredTags[normalized] = struct{}{}
	}
	if len(desiredTags) == 0 {
		return nil, nil
	}

	type candidate struct {
		id       uuid.UUID
		name     string
		priority int
	}
	candidates := make([]candidate, 0)
	for _, character := range characters {
		if character == nil || !questTemplateCharacterMatches(character, desiredTags) {
			continue
		}
		if character.PointOfInterestID != nil {
			if _, ok := pointOfInterestIDs[*character.PointOfInterestID]; ok {
				candidates = append(candidates, candidate{
					id:       character.ID,
					name:     strings.ToLower(strings.TrimSpace(character.Name)),
					priority: 0,
				})
				continue
			}
		}
		if questTemplateCharacterInZone(zone, character) {
			candidates = append(candidates, candidate{
				id:       character.ID,
				name:     strings.ToLower(strings.TrimSpace(character.Name)),
				priority: 1,
			})
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].priority != candidates[j].priority {
			return candidates[i].priority < candidates[j].priority
		}
		if candidates[i].name != candidates[j].name {
			return candidates[i].name < candidates[j].name
		}
		return candidates[i].id.String() < candidates[j].id.String()
	})

	selectedID := candidates[0].id
	return &selectedID, nil
}

func questTemplateCharacterMatches(character *models.Character, desiredTags map[string]struct{}) bool {
	if character == nil {
		return false
	}
	for _, tag := range character.InternalTags {
		if _, ok := desiredTags[strings.ToLower(strings.TrimSpace(tag))]; ok {
			return true
		}
	}
	return false
}

func questTemplateCharacterInZone(zone *models.Zone, character *models.Character) bool {
	if zone == nil || character == nil {
		return false
	}
	for _, location := range character.Locations {
		if zone.IsPointInBoundary(location.Latitude, location.Longitude) {
			return true
		}
	}
	return false
}
