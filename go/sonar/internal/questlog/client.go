package questlog

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

type QuestNodeObjectiveType string

const (
	QuestNodeObjectiveTypeChallenge        QuestNodeObjectiveType = "challenge"
	QuestNodeObjectiveTypeFetchQuest       QuestNodeObjectiveType = "fetch_quest"
	QuestNodeObjectiveTypeStoryFlag        QuestNodeObjectiveType = "story_flag"
	QuestNodeObjectiveTypeScenario         QuestNodeObjectiveType = "scenario"
	QuestNodeObjectiveTypeExposition       QuestNodeObjectiveType = "exposition"
	QuestNodeObjectiveTypeMonsterEncounter QuestNodeObjectiveType = "monster_encounter"
	QuestNodeObjectiveTypeMonster          QuestNodeObjectiveType = "monster"
)

type QuestNodeFetchRequirement struct {
	InventoryItemID int                   `json:"inventoryItemId"`
	Quantity        int                   `json:"quantity"`
	InventoryItem   *models.InventoryItem `json:"inventoryItem,omitempty"`
}

type QuestNodeObjective struct {
	ID                uuid.UUID                      `json:"id"`
	Type              QuestNodeObjectiveType         `json:"type"`
	Prompt            string                         `json:"prompt"`
	Description       string                         `json:"description,omitempty"`
	ImageURL          string                         `json:"imageUrl,omitempty"`
	ThumbnailURL      string                         `json:"thumbnailUrl,omitempty"`
	Reward            int                            `json:"reward"`
	InventoryItemID   *int                           `json:"inventoryItemId"`
	StoryFlagKey      string                         `json:"storyFlagKey,omitempty"`
	SubmissionType    models.QuestNodeSubmissionType `json:"submissionType"`
	Difficulty        int                            `json:"difficulty"`
	StatTags          []string                       `json:"statTags,omitempty"`
	Proficiency       *string                        `json:"proficiency,omitempty"`
	CharacterID       *uuid.UUID                     `json:"characterId,omitempty"`
	CharacterName     string                         `json:"characterName,omitempty"`
	FetchRequirements []QuestNodeFetchRequirement    `json:"fetchRequirements,omitempty"`
}

type QuestNode struct {
	ID                   uuid.UUID                      `json:"id"`
	OrderIndex           int                            `json:"orderIndex"`
	ObjectiveText        string                         `json:"objectiveText,omitempty"`
	ObjectiveDescription string                         `json:"objectiveDescription,omitempty"`
	Objective            *QuestNodeObjective            `json:"objective,omitempty"`
	PointOfInterest      *models.PointOfInterest        `json:"pointOfInterest,omitempty"`
	Polygon              []QuestNodePolygonPoint        `json:"polygon,omitempty"`
	ScenarioID           *uuid.UUID                     `json:"scenarioId,omitempty"`
	ExpositionID         *uuid.UUID                     `json:"expositionId,omitempty"`
	FetchCharacterID     *uuid.UUID                     `json:"fetchCharacterId,omitempty"`
	FetchCharacter       *models.Character              `json:"fetchCharacter,omitempty"`
	StoryFlagKey         string                         `json:"storyFlagKey,omitempty"`
	MonsterID            *uuid.UUID                     `json:"monsterId,omitempty"`
	MonsterEncounterID   *uuid.UUID                     `json:"monsterEncounterId,omitempty"`
	ChallengeID          *uuid.UUID                     `json:"challengeId,omitempty"`
	SubmissionType       models.QuestNodeSubmissionType `json:"submissionType"`
}

type QuestNodePolygonPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type QuestItemReward struct {
	InventoryItemID int                   `json:"inventoryItemId"`
	InventoryItem   *models.InventoryItem `json:"inventoryItem,omitempty"`
	Quantity        int                   `json:"quantity"`
}

type QuestSpellReward struct {
	SpellID uuid.UUID     `json:"spellId"`
	Spell   *models.Spell `json:"spell,omitempty"`
}

type Quest struct {
	ID                       uuid.UUID                `json:"id"`
	Name                     string                   `json:"name"`
	Description              string                   `json:"description"`
	Category                 string                   `json:"category"`
	IsTutorial               bool                     `json:"isTutorial,omitempty"`
	AcceptanceDialogue       []models.DialogueMessage `json:"acceptanceDialogue,omitempty"`
	ImageUrl                 string                   `json:"imageUrl"`
	RewardMode               models.RewardMode        `json:"rewardMode"`
	RandomRewardSize         models.RandomRewardSize  `json:"randomRewardSize"`
	Gold                     int                      `json:"gold"`
	ItemRewards              []QuestItemReward        `json:"itemRewards"`
	SpellRewards             []QuestSpellReward       `json:"spellRewards"`
	QuestGiverCharacterID    *uuid.UUID               `json:"questGiverCharacterId,omitempty"`
	MainStoryPreviousQuestID *uuid.UUID               `json:"mainStoryPreviousQuestId,omitempty"`
	MainStoryNextQuestID     *uuid.UUID               `json:"mainStoryNextQuestId,omitempty"`
	RecurringQuestID         *uuid.UUID               `json:"recurringQuestId,omitempty"`
	IsAccepted               bool                     `json:"isAccepted"`
	TurnedInAt               *time.Time               `json:"turnedInAt,omitempty"`
	CompletionCount          int                      `json:"completionCount,omitempty"`
	ReadyToTurnIn            bool                     `json:"readyToTurnIn"`
	CurrentNode              *QuestNode               `json:"currentNode,omitempty"`
}

type QuestLog struct {
	Quests          []Quest     `json:"quests"`
	CompletedQuests []Quest     `json:"completedQuests"`
	TrackedQuestIDs []uuid.UUID `json:"trackedQuestIds"`
}

type QuestlogClient interface {
	GetQuestLog(ctx context.Context, userID uuid.UUID, zoneID uuid.UUID, tags []string) (*QuestLog, error)
	AreQuestObjectivesComplete(ctx context.Context, userID uuid.UUID, questID uuid.UUID) (bool, error)
}

type questlogClient struct {
	dbClient db.DbClient
}

var (
	syntheticTutorialQuestID                = uuid.MustParse("5a4cb6a9-2489-4c83-a5be-b525f4a280e7")
	syntheticTutorialWelcomeNodeID          = uuid.MustParse("6d59d16d-a5bb-4ed2-bc23-9adfe7173eb7")
	syntheticTutorialScenarioNodeID         = uuid.MustParse("9b695530-637f-431c-a10f-47df31b2debf")
	syntheticTutorialLoadoutNodeID          = uuid.MustParse("55b94fe3-d84d-489d-8255-5fd8254b70ce")
	syntheticTutorialMonsterNodeID          = uuid.MustParse("8ff3d79c-c3d2-4ba0-a52e-2d2138a6b93f")
	syntheticTutorialPostMonsterNodeID      = uuid.MustParse("f30ca0ab-cdbb-4a0c-b39b-ab4191e90df1")
	syntheticTutorialBaseKitNodeID          = uuid.MustParse("24289e6e-6c5d-4d7d-9f42-3f3772caa4e5")
	syntheticTutorialPostBaseDialogueNodeID = uuid.MustParse("43c34c30-8e8d-4627-b9b3-a6a5f7a5cb7d")
)

const (
	tutorialQuestName        = "Tutorial"
	tutorialQuestCategory    = "tutorial"
	tutorialQuestDescription = "Follow the guided tutorial to learn the basics of Unclaimed Streets."
)

func NewClient(dbClient db.DbClient) QuestlogClient {
	log.Printf("Creating new QuestlogClient")
	return &questlogClient{dbClient: dbClient}
}

func userStoryFlagMap(flags []models.UserStoryFlag) map[string]bool {
	result := map[string]bool{}
	for _, flag := range flags {
		key := models.NormalizeStoryFlagKey(flag.FlagKey)
		if key == "" {
			continue
		}
		result[key] = flag.Value
	}
	return result
}

func (c *questlogClient) loadUserStoryFlagMap(
	ctx context.Context,
	userID uuid.UUID,
) (map[string]bool, error) {
	flags, err := c.dbClient.UserStoryFlag().FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return userStoryFlagMap(flags), nil
}

func (c *questlogClient) markQuestNodeCompleteForAcceptance(
	ctx context.Context,
	acceptance *models.QuestAcceptanceV2,
	nodeID uuid.UUID,
	completedAt time.Time,
) error {
	if acceptance == nil {
		return nil
	}
	progress, err := c.dbClient.QuestNodeProgress().FindByAcceptanceAndNode(
		ctx,
		acceptance.ID,
		nodeID,
	)
	if err != nil {
		return err
	}
	if progress == nil {
		progress = &models.QuestNodeProgress{
			ID:                uuid.New(),
			CreatedAt:         completedAt,
			UpdatedAt:         completedAt,
			QuestAcceptanceID: acceptance.ID,
			QuestNodeID:       nodeID,
			CompletedAt:       &completedAt,
		}
		return c.dbClient.QuestNodeProgress().Create(ctx, progress)
	}
	if progress.CompletedAt != nil {
		return nil
	}
	return c.dbClient.QuestNodeProgress().MarkCompleted(ctx, progress.ID)
}

func (c *questlogClient) autoCompleteStoryFlagNodes(
	ctx context.Context,
	quest *models.Quest,
	acceptance *models.QuestAcceptanceV2,
	completed map[uuid.UUID]bool,
	activeStoryFlags map[string]bool,
) (map[uuid.UUID]bool, error) {
	if quest == nil || acceptance == nil {
		return completed, nil
	}
	_, autoCompleted := models.ResolveCurrentQuestNode(
		quest.Nodes,
		completed,
		activeStoryFlags,
	)
	if len(autoCompleted) == 0 {
		return completed, nil
	}

	nextCompleted := make(map[uuid.UUID]bool, len(completed)+len(autoCompleted))
	for nodeID, isCompleted := range completed {
		nextCompleted[nodeID] = isCompleted
	}
	completedAt := time.Now()
	for _, nodeID := range autoCompleted {
		if err := c.markQuestNodeCompleteForAcceptance(
			ctx,
			acceptance,
			nodeID,
			completedAt,
		); err != nil {
			return completed, err
		}
		nextCompleted[nodeID] = true
	}

	return nextCompleted, nil
}

func (c *questlogClient) GetQuestLog(ctx context.Context, userID uuid.UUID, zoneID uuid.UUID, tags []string) (*QuestLog, error) {
	log.Printf("Getting quest log for user %s in zone %s with tags %v", userID, zoneID, tags)

	quests, err := c.dbClient.Quest().FindByZoneID(ctx, zoneID)
	if err != nil {
		return nil, err
	}

	acceptances, err := c.dbClient.QuestAcceptanceV2().FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	acceptanceByQuest := map[uuid.UUID]models.QuestAcceptanceV2{}
	for _, acceptance := range acceptances {
		acceptanceByQuest[acceptance.QuestID] = acceptance
	}

	trackedQuestIDs := []uuid.UUID{}
	tracked, err := c.dbClient.TrackedQuest().GetByUserID(ctx, userID)
	if err == nil {
		for _, t := range tracked {
			trackedQuestIDs = append(trackedQuestIDs, t.QuestID)
		}
	}
	activeStoryFlags, err := c.loadUserStoryFlagMap(ctx, userID)
	if err != nil {
		return nil, err
	}

	trackedAccepted := make([]uuid.UUID, 0, len(trackedQuestIDs))
	trackedAcceptedSet := make(map[uuid.UUID]struct{}, len(trackedQuestIDs))
	for _, questID := range trackedQuestIDs {
		if acc, ok := acceptanceByQuest[questID]; ok && acc.TurnedInAt == nil {
			trackedAccepted = append(trackedAccepted, questID)
			trackedAcceptedSet[questID] = struct{}{}
		}
	}

	if len(trackedAccepted) > 0 {
		trackedQuests, err := c.dbClient.Quest().FindByIDs(ctx, trackedAccepted)
		if err != nil {
			return nil, err
		}
		quests = appendTrackedQuests(quests, trackedQuests, trackedAccepted)
	}

	filtered := make([]Quest, 0, len(quests))
	completedBySeries := map[uuid.UUID]Quest{}
	completedCounts := map[uuid.UUID]int{}
	for _, quest := range quests {
		if quest.ZoneID != nil && *quest.ZoneID != zoneID {
			if _, ok := trackedAcceptedSet[quest.ID]; !ok {
				continue
			}
		}

		acceptance, accepted := acceptanceByQuest[quest.ID]
		if !accepted {
			continue
		}

		progress := map[uuid.UUID]bool{}
		nodeProgress, err := c.dbClient.QuestNodeProgress().FindByAcceptanceID(ctx, acceptance.ID)
		if err != nil {
			return nil, err
		}
		for _, p := range nodeProgress {
			if p.CompletedAt != nil {
				progress[p.QuestNodeID] = true
			}
		}
		progress, err = c.autoCompleteStoryFlagNodes(
			ctx,
			&quest,
			&acceptance,
			progress,
			activeStoryFlags,
		)
		if err != nil {
			return nil, err
		}

		currentNode, err := buildCurrentNode(ctx, c.dbClient, &quest, progress)
		if err != nil {
			return nil, err
		}
		allCompleted := len(quest.Nodes) > 0 && allNodesCompleted(&quest, progress)
		itemRewards := make([]QuestItemReward, 0, len(quest.ItemRewards))
		for _, reward := range quest.ItemRewards {
			var invItem *models.InventoryItem
			if reward.InventoryItem.ID != 0 {
				invItem = &reward.InventoryItem
			}
			itemRewards = append(itemRewards, QuestItemReward{
				InventoryItemID: reward.InventoryItemID,
				InventoryItem:   invItem,
				Quantity:        reward.Quantity,
			})
		}
		spellRewards := make([]QuestSpellReward, 0, len(quest.SpellRewards))
		for _, reward := range quest.SpellRewards {
			var spell *models.Spell
			if reward.Spell.ID != uuid.Nil {
				spell = &reward.Spell
			}
			spellRewards = append(spellRewards, QuestSpellReward{
				SpellID: reward.SpellID,
				Spell:   spell,
			})
		}

		seriesID := quest.ID
		if quest.RecurringQuestID != nil {
			seriesID = *quest.RecurringQuestID
		}
		seriesIDCopy := seriesID
		entry := Quest{
			ID:                       quest.ID,
			Name:                     quest.Name,
			Description:              quest.Description,
			Category:                 quest.Category,
			AcceptanceDialogue:       []models.DialogueMessage(quest.AcceptanceDialogue),
			ImageUrl:                 quest.ImageURL,
			RewardMode:               quest.RewardMode,
			RandomRewardSize:         quest.RandomRewardSize,
			Gold:                     quest.Gold,
			ItemRewards:              itemRewards,
			SpellRewards:             spellRewards,
			QuestGiverCharacterID:    quest.QuestGiverCharacterID,
			MainStoryPreviousQuestID: quest.MainStoryPreviousQuestID,
			MainStoryNextQuestID:     quest.MainStoryNextQuestID,
			RecurringQuestID:         &seriesIDCopy,
			IsAccepted:               accepted,
			TurnedInAt:               acceptance.TurnedInAt,
			ReadyToTurnIn:            accepted && acceptance.TurnedInAt == nil && allCompleted,
			CurrentNode:              currentNode,
		}
		if acceptance.TurnedInAt != nil {
			completedCounts[seriesID]++
			existing, ok := completedBySeries[seriesID]
			if !ok || (existing.TurnedInAt != nil && entry.TurnedInAt != nil && entry.TurnedInAt.After(*existing.TurnedInAt)) {
				completedBySeries[seriesID] = entry
			}
		} else {
			filtered = append(filtered, entry)
		}
	}

	completed := make([]Quest, 0, len(completedBySeries))
	for seriesID, entry := range completedBySeries {
		entry.CompletionCount = completedCounts[seriesID]
		completed = append(completed, entry)
	}
	sort.Slice(completed, func(i, j int) bool {
		left := completed[i].TurnedInAt
		right := completed[j].TurnedInAt
		if left == nil && right == nil {
			return completed[i].Name < completed[j].Name
		}
		if left == nil {
			return false
		}
		if right == nil {
			return true
		}
		return left.After(*right)
	})

	tutorialQuest, err := c.loadActiveTutorialQuest(ctx, userID)
	if err != nil {
		return nil, err
	}
	if tutorialQuest != nil {
		filtered = append([]Quest{*tutorialQuest}, filtered...)
		trackedAccepted = prependQuestIDIfMissing(
			trackedAccepted,
			tutorialQuest.ID,
		)
	}

	return &QuestLog{
		Quests:          filtered,
		CompletedQuests: completed,
		TrackedQuestIDs: trackedAccepted,
	}, nil
}

func appendTrackedQuests(
	quests []models.Quest,
	trackedQuests []models.Quest,
	trackedIDs []uuid.UUID,
) []models.Quest {
	if len(trackedQuests) == 0 || len(trackedIDs) == 0 {
		return quests
	}

	existing := make(map[uuid.UUID]struct{}, len(quests))
	for _, quest := range quests {
		existing[quest.ID] = struct{}{}
	}

	trackedByID := make(map[uuid.UUID]models.Quest, len(trackedQuests))
	for _, quest := range trackedQuests {
		trackedByID[quest.ID] = quest
	}

	merged := append([]models.Quest{}, quests...)
	for _, questID := range trackedIDs {
		if _, ok := existing[questID]; ok {
			continue
		}
		quest, ok := trackedByID[questID]
		if !ok {
			continue
		}
		merged = append(merged, quest)
		existing[questID] = struct{}{}
	}

	return merged
}

func prependQuestIDIfMissing(ids []uuid.UUID, id uuid.UUID) []uuid.UUID {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append([]uuid.UUID{id}, ids...)
}

func (c *questlogClient) AreQuestObjectivesComplete(ctx context.Context, userID uuid.UUID, questID uuid.UUID) (bool, error) {
	acceptance, err := c.dbClient.QuestAcceptanceV2().FindByUserAndQuest(ctx, userID, questID)
	if err != nil {
		return false, err
	}
	if acceptance == nil {
		return false, nil
	}

	quest, err := c.dbClient.Quest().FindByID(ctx, questID)
	if err != nil {
		return false, err
	}
	if quest == nil {
		return false, nil
	}

	progressEntries, err := c.dbClient.QuestNodeProgress().FindByAcceptanceID(ctx, acceptance.ID)
	if err != nil {
		return false, err
	}

	completed := map[uuid.UUID]bool{}
	for _, p := range progressEntries {
		if p.CompletedAt != nil {
			completed[p.QuestNodeID] = true
		}
	}
	activeStoryFlags, err := c.loadUserStoryFlagMap(ctx, userID)
	if err != nil {
		return false, err
	}
	completed, err = c.autoCompleteStoryFlagNodes(
		ctx,
		quest,
		acceptance,
		completed,
		activeStoryFlags,
	)
	if err != nil {
		return false, err
	}
	return allNodesCompleted(quest, completed), nil
}

func allNodesCompleted(quest *models.Quest, completed map[uuid.UUID]bool) bool {
	if len(quest.Nodes) == 0 {
		return false
	}
	for _, node := range quest.Nodes {
		if !completed[node.ID] {
			return false
		}
	}
	return true
}

func buildCurrentNode(
	ctx context.Context,
	dbClient db.DbClient,
	quest *models.Quest,
	completed map[uuid.UUID]bool,
) (*QuestNode, error) {
	if quest == nil || len(quest.Nodes) == 0 {
		return nil, nil
	}
	nodes := append([]models.QuestNode(nil), quest.Nodes...)
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].OrderIndex < nodes[j].OrderIndex
	})
	for _, node := range nodes {
		if completed[node.ID] {
			continue
		}
		return buildQuestNodeView(ctx, dbClient, node)
	}
	return nil, nil
}

func (c *questlogClient) loadActiveTutorialQuest(
	ctx context.Context,
	userID uuid.UUID,
) (*Quest, error) {
	config, err := c.dbClient.Tutorial().GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	if config == nil || !config.IsConfigured() {
		return nil, nil
	}

	state, err := c.dbClient.Tutorial().FindStateByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if state == nil || state.CompletedAt != nil {
		return nil, nil
	}
	if state.Stage == models.TutorialStageCompleted {
		return nil, nil
	}

	currentNode := c.buildTutorialQuestNode(ctx, config, state)
	if currentNode == nil {
		return nil, nil
	}

	return &Quest{
		ID:               syntheticTutorialQuestID,
		Name:             tutorialQuestName,
		Description:      tutorialQuestDescription,
		Category:         tutorialQuestCategory,
		IsTutorial:       true,
		RewardMode:       models.RewardModeExplicit,
		RandomRewardSize: models.RandomRewardSizeSmall,
		IsAccepted:       true,
		ReadyToTurnIn:    false,
		CurrentNode:      currentNode,
	}, nil
}

func (c *questlogClient) buildTutorialQuestNode(
	ctx context.Context,
	config *models.TutorialConfig,
	state *models.UserTutorialState,
) *QuestNode {
	if state == nil {
		return nil
	}

	switch state.Stage {
	case models.TutorialStageWelcome:
		return &QuestNode{
			ID:             syntheticTutorialWelcomeNodeID,
			OrderIndex:     0,
			ObjectiveText:  "Begin the tutorial.",
			SubmissionType: models.DefaultQuestNodeSubmissionType(),
		}
	case models.TutorialStageScenario:
		objectiveText := strings.TrimSpace(config.ScenarioPrompt)
		if objectiveText == "" {
			objectiveText = "Complete the opening tutorial scenario."
		}
		return &QuestNode{
			ID:             syntheticTutorialScenarioNodeID,
			OrderIndex:     0,
			ObjectiveText:  objectiveText,
			ScenarioID:     state.TutorialScenarioID,
			SubmissionType: models.DefaultQuestNodeSubmissionType(),
		}
	case models.TutorialStageLoadout:
		return &QuestNode{
			ID:         syntheticTutorialLoadoutNodeID,
			OrderIndex: 0,
			ObjectiveText: c.buildTutorialLoadoutObjectiveText(
				ctx,
				state.RequiredEquipItemIDs,
				state.RequiredUseItemIDs,
			),
			SubmissionType: models.DefaultQuestNodeSubmissionType(),
		}
	case models.TutorialStageMonster:
		return &QuestNode{
			ID:                 syntheticTutorialMonsterNodeID,
			OrderIndex:         0,
			ObjectiveText:      c.buildTutorialMonsterObjectiveText(ctx, state.TutorialMonsterEncounterID),
			MonsterEncounterID: state.TutorialMonsterEncounterID,
			SubmissionType:     models.DefaultQuestNodeSubmissionType(),
		}
	case models.TutorialStagePostMonsterDialogue:
		return &QuestNode{
			ID:             syntheticTutorialPostMonsterNodeID,
			OrderIndex:     0,
			ObjectiveText:  "Continue the tutorial conversation.",
			SubmissionType: models.DefaultQuestNodeSubmissionType(),
		}
	case models.TutorialStageBaseKit:
		return &QuestNode{
			ID:             syntheticTutorialBaseKitNodeID,
			OrderIndex:     0,
			ObjectiveText:  c.buildTutorialBaseKitObjectiveText(ctx, state.RequiredUseItemIDs),
			SubmissionType: models.DefaultQuestNodeSubmissionType(),
		}
	case models.TutorialStagePostBaseDialogue:
		return &QuestNode{
			ID:             syntheticTutorialPostBaseDialogueNodeID,
			OrderIndex:     0,
			ObjectiveText:  "Finish the tutorial conversation.",
			SubmissionType: models.DefaultQuestNodeSubmissionType(),
		}
	default:
		return &QuestNode{
			ID:             syntheticTutorialWelcomeNodeID,
			OrderIndex:     0,
			ObjectiveText:  "Continue the tutorial.",
			SubmissionType: models.DefaultQuestNodeSubmissionType(),
		}
	}
}

func (c *questlogClient) buildTutorialLoadoutObjectiveText(
	ctx context.Context,
	requiredEquipItemIDs []int,
	requiredUseItemIDs []int,
) string {
	clauses := []string{}
	if summary := c.buildTutorialItemRequirementSummary(
		ctx,
		requiredEquipItemIDs,
		"equip",
	); summary != "" {
		clauses = append(clauses, summary)
	}
	if summary := c.buildTutorialItemRequirementSummary(
		ctx,
		requiredUseItemIDs,
		"use",
	); summary != "" {
		clauses = append(clauses, summary)
	}
	if len(clauses) == 0 {
		return "Prepare your loadout."
	}
	return "Prepare your loadout: " + strings.Join(clauses, ". ") + "."
}

func (c *questlogClient) buildTutorialBaseKitObjectiveText(
	ctx context.Context,
	requiredUseItemIDs []int,
) string {
	if summary := c.buildTutorialItemRequirementSummary(
		ctx,
		requiredUseItemIDs,
		"use",
	); summary != "" {
		return "Set up your home base: " + summary + "."
	}
	return "Set up your home base."
}

func (c *questlogClient) buildTutorialMonsterObjectiveText(
	ctx context.Context,
	monsterEncounterID *uuid.UUID,
) string {
	if monsterEncounterID == nil || *monsterEncounterID == uuid.Nil {
		return "Defeat the tutorial monster encounter."
	}
	encounter, err := c.dbClient.MonsterEncounter().FindByID(
		ctx,
		*monsterEncounterID,
	)
	if err != nil {
		log.Printf(
			"loadActiveTutorialQuest: monster encounter lookup failed for %s: %v",
			monsterEncounterID.String(),
			err,
		)
		return "Defeat the tutorial monster encounter."
	}
	if encounter == nil {
		return "Defeat the tutorial monster encounter."
	}
	name := strings.TrimSpace(encounter.Name)
	if name == "" {
		return "Defeat the tutorial monster encounter."
	}
	return "Defeat " + name + "."
}

func (c *questlogClient) buildTutorialItemRequirementSummary(
	ctx context.Context,
	inventoryItemIDs []int,
	verb string,
) string {
	names := c.loadTutorialInventoryItemNames(ctx, inventoryItemIDs)
	if len(names) == 0 {
		return ""
	}
	return strings.TrimSpace(verb) + " " + strings.Join(names, ", ")
}

func (c *questlogClient) loadTutorialInventoryItemNames(
	ctx context.Context,
	inventoryItemIDs []int,
) []string {
	if len(inventoryItemIDs) == 0 {
		return nil
	}
	names := make([]string, 0, len(inventoryItemIDs))
	for _, inventoryItemID := range inventoryItemIDs {
		if inventoryItemID <= 0 {
			continue
		}
		name := fmt.Sprintf("item %d", inventoryItemID)
		item, err := c.dbClient.InventoryItem().FindInventoryItemByID(
			ctx,
			inventoryItemID,
		)
		if err != nil {
			log.Printf(
				"loadActiveTutorialQuest: inventory item lookup failed for %d: %v",
				inventoryItemID,
				err,
			)
		} else if item != nil {
			trimmed := strings.TrimSpace(item.Name)
			if trimmed != "" {
				name = trimmed
			}
		}
		names = append(names, name)
	}
	return names
}

func buildQuestNodeView(
	ctx context.Context,
	dbClient db.DbClient,
	node models.QuestNode,
) (*QuestNode, error) {
	objective, pointOfInterest, polygon, err := buildQuestNodeObjective(
		ctx,
		dbClient,
		node,
	)
	if err != nil {
		return nil, err
	}

	submissionType := node.SubmissionType
	if strings.TrimSpace(string(submissionType)) == "" {
		if objective != nil {
			submissionType = objective.SubmissionType
		}
	}
	if strings.TrimSpace(string(submissionType)) == "" {
		submissionType = models.DefaultQuestNodeSubmissionType()
	}
	objectiveText := strings.TrimSpace(node.ObjectiveDescription)
	if objectiveText == "" && objective != nil {
		objectiveText = strings.TrimSpace(objective.Prompt)
	}

	return &QuestNode{
		ID:                   node.ID,
		OrderIndex:           node.OrderIndex,
		ObjectiveText:        objectiveText,
		ObjectiveDescription: strings.TrimSpace(node.ObjectiveDescription),
		Objective:            objective,
		PointOfInterest:      pointOfInterest,
		Polygon:              polygon,
		ScenarioID:           node.ScenarioID,
		ExpositionID:         node.ExpositionID,
		FetchCharacterID:     node.FetchCharacterID,
		FetchCharacter:       node.FetchCharacter,
		StoryFlagKey:         node.StoryFlagKeyNormalized(),
		MonsterID:            node.MonsterID,
		MonsterEncounterID:   node.MonsterEncounterID,
		ChallengeID:          node.ChallengeID,
		SubmissionType:       submissionType,
	}, nil
}

func buildQuestNodeObjective(
	ctx context.Context,
	dbClient db.DbClient,
	node models.QuestNode,
) (*QuestNodeObjective, *models.PointOfInterest, []QuestNodePolygonPoint, error) {
	if node.IsStoryFlagNode() {
		storyFlagKey := node.StoryFlagKeyNormalized()
		return &QuestNodeObjective{
			ID:             node.ID,
			Type:           QuestNodeObjectiveTypeStoryFlag,
			Prompt:         "Wait for story progress: " + storyFlagKey,
			Description:    "This objective completes automatically once the required story flag becomes active.",
			StoryFlagKey:   storyFlagKey,
			SubmissionType: models.DefaultQuestNodeSubmissionType(),
		}, nil, nil, nil
	}
	if node.IsFetchQuestNode() {
		character := node.FetchCharacter
		if character == nil && node.FetchCharacterID != nil && *node.FetchCharacterID != uuid.Nil {
			loadedCharacter, err := dbClient.Character().FindByID(ctx, *node.FetchCharacterID)
			if err != nil {
				log.Printf(
					"resolveObjectiveText: fetch character lookup failed for %s: %v",
					node.FetchCharacterID.String(),
					err,
				)
				return nil, nil, nil, nil
			}
			character = loadedCharacter
		}

		requirements := make([]QuestNodeFetchRequirement, 0, len(node.FetchRequirements))
		requirementLabels := make([]string, 0, len(node.FetchRequirements))
		for _, requirement := range node.FetchRequirements {
			if requirement.InventoryItemID <= 0 || requirement.Quantity <= 0 {
				continue
			}
			var inventoryItem *models.InventoryItem
			item, err := dbClient.InventoryItem().FindInventoryItemByID(
				ctx,
				requirement.InventoryItemID,
			)
			if err != nil {
				log.Printf(
					"resolveObjectiveText: fetch item lookup failed for %d: %v",
					requirement.InventoryItemID,
					err,
				)
			} else {
				inventoryItem = item
			}
			requirements = append(requirements, QuestNodeFetchRequirement{
				InventoryItemID: requirement.InventoryItemID,
				Quantity:        requirement.Quantity,
				InventoryItem:   inventoryItem,
			})
			itemName := "item"
			if inventoryItem != nil && strings.TrimSpace(inventoryItem.Name) != "" {
				itemName = strings.TrimSpace(inventoryItem.Name)
			}
			requirementLabels = append(
				requirementLabels,
				fmt.Sprintf("%d %s", requirement.Quantity, itemName),
			)
		}

		characterName := ""
		var pointOfInterest *models.PointOfInterest
		if character != nil {
			characterName = strings.TrimSpace(character.Name)
			pointOfInterest = character.PointOfInterest
		}

		prompt := "Deliver the required items"
		if len(requirementLabels) > 0 && characterName != "" {
			prompt = fmt.Sprintf(
				"Bring %s to %s",
				strings.Join(requirementLabels, ", "),
				characterName,
			)
		} else if characterName != "" {
			prompt = fmt.Sprintf("Deliver the required items to %s", characterName)
		}
		description := "Hand over the requested items to continue the quest."
		return &QuestNodeObjective{
			ID:                node.ID,
			Type:              QuestNodeObjectiveTypeFetchQuest,
			Prompt:            prompt,
			Description:       description,
			SubmissionType:    models.DefaultQuestNodeSubmissionType(),
			CharacterID:       node.FetchCharacterID,
			CharacterName:     characterName,
			FetchRequirements: requirements,
		}, pointOfInterest, nil, nil
	}
	if node.ChallengeID != nil {
		challenge, err := dbClient.Challenge().FindByID(ctx, *node.ChallengeID)
		if err != nil {
			log.Printf("resolveObjectiveText: challenge lookup failed for %s: %v", node.ChallengeID.String(), err)
			return nil, nil, nil, nil
		}
		if challenge != nil {
			submissionType := challenge.SubmissionType
			if strings.TrimSpace(string(submissionType)) == "" {
				submissionType = node.SubmissionType
			}
			if strings.TrimSpace(string(submissionType)) == "" {
				submissionType = models.DefaultQuestNodeSubmissionType()
			}
			return &QuestNodeObjective{
					ID:              challenge.ID,
					Type:            QuestNodeObjectiveTypeChallenge,
					Prompt:          strings.TrimSpace(challenge.Question),
					Description:     strings.TrimSpace(challenge.Description),
					ImageURL:        strings.TrimSpace(challenge.ImageURL),
					ThumbnailURL:    strings.TrimSpace(challenge.ThumbnailURL),
					Reward:          challenge.Reward,
					InventoryItemID: challenge.InventoryItemID,
					SubmissionType:  submissionType,
					Difficulty:      challenge.Difficulty,
					StatTags:        []string(challenge.StatTags),
					Proficiency:     challenge.Proficiency,
				},
				challenge.PointOfInterest,
				convertPolygonPoints(challenge.PolygonPoints),
				nil
		}
		return nil, nil, nil, nil
	}

	if node.ScenarioID != nil {
		scenario, err := dbClient.Scenario().FindByID(ctx, *node.ScenarioID)
		if err != nil {
			log.Printf("resolveObjectiveText: scenario lookup failed for %s: %v", node.ScenarioID.String(), err)
			return nil, nil, nil, nil
		}
		if scenario != nil {
			submissionType := node.SubmissionType
			if strings.TrimSpace(string(submissionType)) == "" {
				submissionType = models.DefaultQuestNodeSubmissionType()
			}
			difficulty := scenario.Difficulty
			statTags := []string{}
			seenTags := map[string]struct{}{}
			var proficiency *string
			optionDifficultyTotal := 0
			optionDifficultyCount := 0
			for _, option := range scenario.Options {
				tag := strings.TrimSpace(option.StatTag)
				if tag != "" {
					normalized := strings.ToLower(tag)
					if _, seen := seenTags[normalized]; !seen {
						seenTags[normalized] = struct{}{}
						statTags = append(statTags, normalized)
					}
				}
				if option.Difficulty != nil {
					optionDifficultyTotal += *option.Difficulty
					optionDifficultyCount++
				}
				if proficiency == nil && len(option.Proficiencies) > 0 {
					first := strings.TrimSpace(option.Proficiencies[0])
					if first != "" {
						proficiency = &first
					}
				}
			}
			if optionDifficultyCount > 0 {
				difficulty = int(float64(optionDifficultyTotal) / float64(optionDifficultyCount))
			}
			prompt := strings.TrimSpace(scenario.Prompt)
			if prompt == "" {
				prompt = "Investigate the situation"
			}
			return &QuestNodeObjective{
					ID:             scenario.ID,
					Type:           QuestNodeObjectiveTypeScenario,
					Prompt:         prompt,
					ImageURL:       strings.TrimSpace(scenario.ImageURL),
					ThumbnailURL:   strings.TrimSpace(scenario.ThumbnailURL),
					SubmissionType: submissionType,
					Difficulty:     difficulty,
					StatTags:       statTags,
					Proficiency:    proficiency,
				},
				scenario.PointOfInterest,
				nil,
				nil
		}
		return nil, nil, nil, nil
	}

	if node.ExpositionID != nil {
		exposition, err := dbClient.Exposition().FindByID(ctx, *node.ExpositionID)
		if err != nil {
			log.Printf(
				"resolveObjectiveText: exposition lookup failed for %s: %v",
				node.ExpositionID.String(),
				err,
			)
			return nil, nil, nil, nil
		}
		if exposition != nil {
			submissionType := node.SubmissionType
			if strings.TrimSpace(string(submissionType)) == "" {
				submissionType = models.DefaultQuestNodeSubmissionType()
			}
			title := strings.TrimSpace(exposition.Title)
			prompt := "Complete the exposition dialogue"
			if title != "" {
				prompt = "Complete the dialogue: " + title
			}
			description := strings.TrimSpace(exposition.Description)
			if description == "" {
				description = "Listen through the full dialogue to continue the quest."
			}
			return &QuestNodeObjective{
					ID:             exposition.ID,
					Type:           QuestNodeObjectiveTypeExposition,
					Prompt:         prompt,
					Description:    description,
					ImageURL:       strings.TrimSpace(exposition.ImageURL),
					ThumbnailURL:   strings.TrimSpace(exposition.ThumbnailURL),
					SubmissionType: submissionType,
				},
				exposition.PointOfInterest,
				nil,
				nil
		}
		return nil, nil, nil, nil
	}

	if node.MonsterEncounterID != nil {
		encounter, err := dbClient.MonsterEncounter().FindByID(
			ctx,
			*node.MonsterEncounterID,
		)
		if err != nil {
			log.Printf(
				"resolveObjectiveText: monster encounter lookup failed for %s: %v",
				node.MonsterEncounterID.String(),
				err,
			)
			return nil, nil, nil, nil
		}
		if encounter != nil {
			name := strings.TrimSpace(encounter.Name)
			if name != "" {
				submissionType := node.SubmissionType
				if strings.TrimSpace(string(submissionType)) == "" {
					submissionType = models.DefaultQuestNodeSubmissionType()
				}
				difficulty := 0
				if len(encounter.Members) > 0 {
					totalLevel := 0
					for _, member := range encounter.Members {
						totalLevel += member.Monster.EffectiveLevel()
					}
					difficulty = int(float64(totalLevel) / float64(len(encounter.Members)))
				}
				return &QuestNodeObjective{
					ID:             encounter.ID,
					Type:           QuestNodeObjectiveTypeMonsterEncounter,
					Prompt:         "Defeat " + name,
					Description:    strings.TrimSpace(encounter.Description),
					ImageURL:       strings.TrimSpace(encounter.ImageURL),
					ThumbnailURL:   strings.TrimSpace(encounter.ThumbnailURL),
					SubmissionType: submissionType,
					Difficulty:     difficulty,
				}, encounter.PointOfInterest, nil, nil
			}
		}
		return nil, nil, nil, nil
	}

	if node.MonsterID != nil {
		monster, err := dbClient.Monster().FindByID(
			ctx,
			*node.MonsterID,
		)
		if err != nil {
			log.Printf(
				"resolveObjectiveText: monster lookup failed for %s: %v",
				node.MonsterID.String(),
				err,
			)
			return nil, nil, nil, nil
		}
		if monster != nil {
			name := strings.TrimSpace(monster.Name)
			if name != "" {
				submissionType := node.SubmissionType
				if strings.TrimSpace(string(submissionType)) == "" {
					submissionType = models.DefaultQuestNodeSubmissionType()
				}
				return &QuestNodeObjective{
					ID:             monster.ID,
					Type:           QuestNodeObjectiveTypeMonster,
					Prompt:         "Defeat " + name,
					Description:    strings.TrimSpace(monster.Description),
					ImageURL:       strings.TrimSpace(monster.ImageURL),
					ThumbnailURL:   strings.TrimSpace(monster.ThumbnailURL),
					SubmissionType: submissionType,
					Difficulty:     monster.EffectiveLevel(),
				}, nil, nil, nil
			}
		}
	}

	return nil, nil, nil, nil
}

func convertPolygonPoints(raw [][2]float64) []QuestNodePolygonPoint {
	if len(raw) == 0 {
		return nil
	}
	points := make([]QuestNodePolygonPoint, 0, len(raw))
	for _, point := range raw {
		points = append(points, QuestNodePolygonPoint{
			Latitude:  point[1],
			Longitude: point[0],
		})
	}
	return points
}
