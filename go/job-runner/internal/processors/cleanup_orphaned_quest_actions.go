package processors

import (
	"context"
	"fmt"
	"log"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type CleanupOrphanedQuestActionsProcessor struct {
	dbClient db.DbClient
}

func NewCleanupOrphanedQuestActionsProcessor(dbClient db.DbClient) CleanupOrphanedQuestActionsProcessor {
	log.Println("Initializing CleanupOrphanedQuestActionsProcessor")
	return CleanupOrphanedQuestActionsProcessor{dbClient: dbClient}
}

func (p *CleanupOrphanedQuestActionsProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing cleanup orphaned quest actions task: %v", task.Type())

	actions, err := p.dbClient.CharacterAction().FindGiveQuestActions(ctx)
	if err != nil {
		log.Printf("Failed to load giveQuest actions: %v", err)
		return fmt.Errorf("failed to load giveQuest actions: %w", err)
	}

	if len(actions) == 0 {
		log.Println("No giveQuest actions to inspect")
		return nil
	}

	questIDSet := map[uuid.UUID]struct{}{}
	actionQuestIDs := map[uuid.UUID]uuid.UUID{}
	orphanedActionIDs := make([]uuid.UUID, 0)
	missingMetadataCount := 0
	invalidQuestIDCount := 0

	for _, action := range actions {
		if action == nil {
			continue
		}
		questIDStr := extractActionQuestID(action.Metadata)
		if questIDStr == "" {
			missingMetadataCount++
			orphanedActionIDs = append(orphanedActionIDs, action.ID)
			continue
		}
		questID, err := uuid.Parse(questIDStr)
		if err != nil || questID == uuid.Nil {
			invalidQuestIDCount++
			orphanedActionIDs = append(orphanedActionIDs, action.ID)
			continue
		}
		actionQuestIDs[action.ID] = questID
		questIDSet[questID] = struct{}{}
	}

	missingQuestCount := 0
	if len(questIDSet) > 0 {
		questIDs := make([]uuid.UUID, 0, len(questIDSet))
		for questID := range questIDSet {
			questIDs = append(questIDs, questID)
		}

		quests, err := p.dbClient.Quest().FindByIDs(ctx, questIDs)
		if err != nil {
			log.Printf("Failed to load quests for giveQuest actions: %v", err)
			return fmt.Errorf("failed to load quests for giveQuest actions: %w", err)
		}

		existingQuestIDs := map[uuid.UUID]struct{}{}
		for _, quest := range quests {
			existingQuestIDs[quest.ID] = struct{}{}
		}

		for actionID, questID := range actionQuestIDs {
			if _, ok := existingQuestIDs[questID]; !ok {
				missingQuestCount++
				orphanedActionIDs = append(orphanedActionIDs, actionID)
			}
		}
	}

	if len(orphanedActionIDs) == 0 {
		log.Println("No orphaned quest actions found")
		return nil
	}

	deletedCount := 0
	for _, actionID := range orphanedActionIDs {
		if err := p.dbClient.CharacterAction().Delete(ctx, actionID); err != nil {
			log.Printf("Failed to delete orphaned character action %s: %v", actionID, err)
			continue
		}
		deletedCount++
	}

	log.Printf(
		"Removed %d orphaned quest actions (missing metadata=%d, invalid quest IDs=%d, missing quests=%d)",
		deletedCount,
		missingMetadataCount,
		invalidQuestIDCount,
		missingQuestCount,
	)

	return nil
}

func extractActionQuestID(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}
	keys := []string{"questId", "pointOfInterestGroupId"}
	for _, key := range keys {
		if val, ok := metadata[key]; ok {
			switch v := val.(type) {
			case string:
				return v
			case uuid.UUID:
				return v.String()
			default:
				return fmt.Sprintf("%v", v)
			}
		}
	}
	return ""
}
