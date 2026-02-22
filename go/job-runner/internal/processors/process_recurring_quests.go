package processors

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type ProcessRecurringQuestsProcessor struct {
	dbClient db.DbClient
}

func NewProcessRecurringQuestsProcessor(dbClient db.DbClient) ProcessRecurringQuestsProcessor {
	log.Println("Initializing ProcessRecurringQuestsProcessor")
	return ProcessRecurringQuestsProcessor{dbClient: dbClient}
}

func (p *ProcessRecurringQuestsProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing recurring quests task: %v", task.Type())

	now := time.Now()
	due, err := p.dbClient.Quest().FindDueRecurring(ctx, now, 50)
	if err != nil {
		log.Printf("Failed to find recurring quests: %v", err)
		return err
	}

	if len(due) == 0 {
		log.Println("No recurring quests due")
		return nil
	}

	log.Printf("Found %d recurring quests due", len(due))
	for _, quest := range due {
		if err := p.processQuest(ctx, quest.ID, now); err != nil {
			log.Printf("Failed to process recurring quest %s: %v", quest.ID, err)
		}
	}

	return nil
}

func (p *ProcessRecurringQuestsProcessor) processQuest(ctx context.Context, questID uuid.UUID, now time.Time) error {
	quest, err := p.dbClient.Quest().FindByID(ctx, questID)
	if err != nil {
		return err
	}
	if quest == nil {
		return nil
	}
	if quest.RecurrenceFrequency == nil || quest.NextRecurrenceAt == nil {
		return nil
	}
	if quest.NextRecurrenceAt.After(now) {
		return nil
	}

	frequency := models.NormalizeQuestRecurrenceFrequency(*quest.RecurrenceFrequency)
	if !models.IsValidQuestRecurrenceFrequency(frequency) {
		log.Printf("Recurring quest %s has invalid frequency %q", quest.ID, frequency)
		return nil
	}

	recurringID := quest.RecurringQuestID
	if recurringID == nil {
		newID := uuid.New()
		recurringID = &newID
	}

	nextAt, ok := advanceQuestRecurrence(now, quest.NextRecurrenceAt, frequency)
	if !ok {
		log.Printf("Recurring quest %s has unsupported frequency %q", quest.ID, frequency)
		return nil
	}

	newQuest, err := p.cloneQuest(ctx, quest, *recurringID, frequency, nextAt, now)
	if err != nil {
		return err
	}

	if quest.QuestGiverCharacterID != nil {
		if err := p.removeQuestActionForCharacter(ctx, quest.ID, *quest.QuestGiverCharacterID); err != nil {
			log.Printf("Failed to remove quest action for quest %s: %v", quest.ID, err)
		}
		if err := p.ensureQuestActionForCharacter(ctx, newQuest.ID, *quest.QuestGiverCharacterID); err != nil {
			log.Printf("Failed to attach quest action for quest %s: %v", newQuest.ID, err)
		}
	}

	quest.RecurringQuestID = recurringID
	quest.RecurrenceFrequency = nil
	quest.NextRecurrenceAt = nil
	quest.UpdatedAt = now
	if err := p.dbClient.Quest().Update(ctx, quest.ID, quest); err != nil {
		log.Printf("Failed to clear recurrence fields for quest %s: %v", quest.ID, err)
	}

	log.Printf("Recurring quest %s recreated as %s", quest.ID, newQuest.ID)
	return nil
}

func advanceQuestRecurrence(now time.Time, scheduled *time.Time, frequency string) (time.Time, bool) {
	base := now
	if scheduled != nil && !scheduled.IsZero() {
		base = *scheduled
	}
	next, ok := models.NextQuestRecurrenceAt(base, frequency)
	if !ok {
		return time.Time{}, false
	}
	for !next.After(now) {
		next, ok = models.NextQuestRecurrenceAt(next, frequency)
		if !ok {
			return time.Time{}, false
		}
	}
	return next, true
}

func (p *ProcessRecurringQuestsProcessor) cloneQuest(
	ctx context.Context,
	source *models.Quest,
	recurringID uuid.UUID,
	frequency string,
	nextAt time.Time,
	now time.Time,
) (*models.Quest, error) {
	newQuest := &models.Quest{
		ID:                    uuid.New(),
		CreatedAt:             now,
		UpdatedAt:             now,
		Name:                  source.Name,
		Description:           source.Description,
		AcceptanceDialogue:    source.AcceptanceDialogue,
		ImageURL:              source.ImageURL,
		ZoneID:                source.ZoneID,
		QuestArchetypeID:      source.QuestArchetypeID,
		QuestGiverCharacterID: source.QuestGiverCharacterID,
		RecurringQuestID:      &recurringID,
		RecurrenceFrequency:   &frequency,
		NextRecurrenceAt:      &nextAt,
		Gold:                  source.Gold,
	}

	if err := p.dbClient.Quest().Create(ctx, newQuest); err != nil {
		return nil, err
	}

	if err := p.copyQuestRewards(ctx, newQuest.ID, source.ItemRewards, now); err != nil {
		return nil, err
	}

	if err := p.copyQuestNodes(ctx, newQuest.ID, source.Nodes, now); err != nil {
		return nil, err
	}

	return newQuest, nil
}

func (p *ProcessRecurringQuestsProcessor) copyQuestRewards(ctx context.Context, questID uuid.UUID, rewards []models.QuestItemReward, now time.Time) error {
	if len(rewards) == 0 {
		return nil
	}
	newRewards := make([]models.QuestItemReward, 0, len(rewards))
	for _, reward := range rewards {
		if reward.InventoryItemID == 0 || reward.Quantity <= 0 {
			continue
		}
		newRewards = append(newRewards, models.QuestItemReward{
			ID:              uuid.New(),
			CreatedAt:       now,
			UpdatedAt:       now,
			QuestID:         questID,
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	if len(newRewards) == 0 {
		return nil
	}
	return p.dbClient.QuestItemReward().ReplaceForQuest(ctx, questID, newRewards)
}

func (p *ProcessRecurringQuestsProcessor) copyQuestNodes(ctx context.Context, questID uuid.UUID, nodes []models.QuestNode, now time.Time) error {
	if len(nodes) == 0 {
		return nil
	}

	nodesCopy := append([]models.QuestNode(nil), nodes...)
	sort.Slice(nodesCopy, func(i, j int) bool {
		return nodesCopy[i].OrderIndex < nodesCopy[j].OrderIndex
	})

	nodeIDMap := map[uuid.UUID]uuid.UUID{}
	challengeIDMap := map[uuid.UUID]uuid.UUID{}

	for _, node := range nodesCopy {
		newNodeID := uuid.New()
		nodeIDMap[node.ID] = newNodeID

		submissionType := node.SubmissionType
		if submissionType == "" {
			submissionType = models.DefaultQuestNodeSubmissionType()
		}

		newNode := &models.QuestNode{
			ID:                newNodeID,
			CreatedAt:         now,
			UpdatedAt:         now,
			QuestID:           questID,
			OrderIndex:        node.OrderIndex,
			PointOfInterestID: node.PointOfInterestID,
			SubmissionType:    submissionType,
		}
		if len(node.PolygonPoints) > 0 {
			newNode.SetPolygonFromPoints(node.PolygonPoints)
		} else if node.Polygon != "" {
			newNode.Polygon = node.Polygon
		}

		if err := p.dbClient.QuestNode().Create(ctx, newNode); err != nil {
			return err
		}

		for _, challenge := range node.Challenges {
			newChallengeID := uuid.New()
			challengeIDMap[challenge.ID] = newChallengeID
			ch := &models.QuestNodeChallenge{
				ID:              newChallengeID,
				CreatedAt:       now,
				UpdatedAt:       now,
				QuestNodeID:     newNodeID,
				Tier:            challenge.Tier,
				Question:        challenge.Question,
				Reward:          challenge.Reward,
				InventoryItemID: challenge.InventoryItemID,
				Difficulty:      challenge.Difficulty,
				StatTags:        challenge.StatTags,
				Proficiency:     challenge.Proficiency,
			}
			if err := p.dbClient.QuestNodeChallenge().Create(ctx, ch); err != nil {
				return err
			}
		}
	}

	for _, node := range nodesCopy {
		for _, child := range node.Children {
			newQuestNodeID, ok := nodeIDMap[child.QuestNodeID]
			if !ok {
				log.Printf("Missing quest node mapping for %s", child.QuestNodeID)
				continue
			}
			nextNodeID, ok := nodeIDMap[child.NextQuestNodeID]
			if !ok {
				log.Printf("Missing quest node mapping for %s", child.NextQuestNodeID)
				continue
			}
			var newChallengeID *uuid.UUID
			if child.QuestNodeChallengeID != nil {
				if mapped, ok := challengeIDMap[*child.QuestNodeChallengeID]; ok {
					mappedCopy := mapped
					newChallengeID = &mappedCopy
				}
			}

			newChild := &models.QuestNodeChild{
				ID:                   uuid.New(),
				CreatedAt:            now,
				UpdatedAt:            now,
				QuestNodeID:          newQuestNodeID,
				NextQuestNodeID:      nextNodeID,
				QuestNodeChallengeID: newChallengeID,
			}
			if err := p.dbClient.QuestNodeChild().Create(ctx, newChild); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *ProcessRecurringQuestsProcessor) ensureQuestActionForCharacter(ctx context.Context, questID uuid.UUID, characterID uuid.UUID) error {
	actions, err := p.dbClient.CharacterAction().FindByCharacterID(ctx, characterID)
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
	return p.dbClient.CharacterAction().Create(ctx, action)
}

func (p *ProcessRecurringQuestsProcessor) removeQuestActionForCharacter(ctx context.Context, questID uuid.UUID, characterID uuid.UUID) error {
	actions, err := p.dbClient.CharacterAction().FindByCharacterID(ctx, characterID)
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
			_ = p.dbClient.CharacterAction().Delete(ctx, action.ID)
		}
	}
	return nil
}
