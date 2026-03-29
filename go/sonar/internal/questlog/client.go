package questlog

import (
	"context"
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
	QuestNodeObjectiveTypeScenario         QuestNodeObjectiveType = "scenario"
	QuestNodeObjectiveTypeMonsterEncounter QuestNodeObjectiveType = "monster_encounter"
	QuestNodeObjectiveTypeMonster          QuestNodeObjectiveType = "monster"
)

type QuestNodeObjective struct {
	ID              uuid.UUID                      `json:"id"`
	Type            QuestNodeObjectiveType         `json:"type"`
	Prompt          string                         `json:"prompt"`
	Description     string                         `json:"description,omitempty"`
	ImageURL        string                         `json:"imageUrl,omitempty"`
	ThumbnailURL    string                         `json:"thumbnailUrl,omitempty"`
	Reward          int                            `json:"reward"`
	InventoryItemID *int                           `json:"inventoryItemId"`
	SubmissionType  models.QuestNodeSubmissionType `json:"submissionType"`
	Difficulty      int                            `json:"difficulty"`
	StatTags        []string                       `json:"statTags,omitempty"`
	Proficiency     *string                        `json:"proficiency,omitempty"`
}

type QuestNode struct {
	ID                 uuid.UUID                      `json:"id"`
	OrderIndex         int                            `json:"orderIndex"`
	ObjectiveText      string                         `json:"objectiveText,omitempty"`
	Objective          *QuestNodeObjective            `json:"objective,omitempty"`
	PointOfInterest    *models.PointOfInterest        `json:"pointOfInterest,omitempty"`
	Polygon            []QuestNodePolygonPoint        `json:"polygon,omitempty"`
	ScenarioID         *uuid.UUID                     `json:"scenarioId,omitempty"`
	MonsterID          *uuid.UUID                     `json:"monsterId,omitempty"`
	MonsterEncounterID *uuid.UUID                     `json:"monsterEncounterId,omitempty"`
	ChallengeID        *uuid.UUID                     `json:"challengeId,omitempty"`
	SubmissionType     models.QuestNodeSubmissionType `json:"submissionType"`
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
	ID                    uuid.UUID               `json:"id"`
	Name                  string                  `json:"name"`
	Description           string                  `json:"description"`
	AcceptanceDialogue    []string                `json:"acceptanceDialogue,omitempty"`
	ImageUrl              string                  `json:"imageUrl"`
	RewardMode            models.RewardMode       `json:"rewardMode"`
	RandomRewardSize      models.RandomRewardSize `json:"randomRewardSize"`
	Gold                  int                     `json:"gold"`
	ItemRewards           []QuestItemReward       `json:"itemRewards"`
	SpellRewards          []QuestSpellReward      `json:"spellRewards"`
	QuestGiverCharacterID *uuid.UUID              `json:"questGiverCharacterId,omitempty"`
	RecurringQuestID      *uuid.UUID              `json:"recurringQuestId,omitempty"`
	IsAccepted            bool                    `json:"isAccepted"`
	TurnedInAt            *time.Time              `json:"turnedInAt,omitempty"`
	CompletionCount       int                     `json:"completionCount,omitempty"`
	ReadyToTurnIn         bool                    `json:"readyToTurnIn"`
	CurrentNode           *QuestNode              `json:"currentNode,omitempty"`
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

func NewClient(dbClient db.DbClient) QuestlogClient {
	log.Printf("Creating new QuestlogClient")
	return &questlogClient{dbClient: dbClient}
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
			ID:                    quest.ID,
			Name:                  quest.Name,
			Description:           quest.Description,
			AcceptanceDialogue:    []string(quest.AcceptanceDialogue),
			ImageUrl:              quest.ImageURL,
			RewardMode:            quest.RewardMode,
			RandomRewardSize:      quest.RandomRewardSize,
			Gold:                  quest.Gold,
			ItemRewards:           itemRewards,
			SpellRewards:          spellRewards,
			QuestGiverCharacterID: quest.QuestGiverCharacterID,
			RecurringQuestID:      &seriesIDCopy,
			IsAccepted:            accepted,
			TurnedInAt:            acceptance.TurnedInAt,
			ReadyToTurnIn:         accepted && acceptance.TurnedInAt == nil && allCompleted,
			CurrentNode:           currentNode,
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
	objectiveText := ""
	if objective != nil {
		objectiveText = strings.TrimSpace(objective.Prompt)
	}

	return &QuestNode{
		ID:                 node.ID,
		OrderIndex:         node.OrderIndex,
		ObjectiveText:      objectiveText,
		Objective:          objective,
		PointOfInterest:    pointOfInterest,
		Polygon:            polygon,
		ScenarioID:         node.ScenarioID,
		MonsterID:          node.MonsterID,
		MonsterEncounterID: node.MonsterEncounterID,
		ChallengeID:        node.ChallengeID,
		SubmissionType:     submissionType,
	}, nil
}

func buildQuestNodeObjective(
	ctx context.Context,
	dbClient db.DbClient,
	node models.QuestNode,
) (*QuestNodeObjective, *models.PointOfInterest, []QuestNodePolygonPoint, error) {
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
