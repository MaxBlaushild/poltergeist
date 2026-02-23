package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type GenerateQuestForZoneProcessor struct {
	dbClient      db.DbClient
	dungeonmaster dungeonmaster.Client
}

func NewGenerateQuestForZoneProcessor(dbClient db.DbClient, dungeonmaster dungeonmaster.Client) GenerateQuestForZoneProcessor {
	log.Println("Initializing GenerateQuestForZoneProcessor")
	return GenerateQuestForZoneProcessor{
		dbClient:      dbClient,
		dungeonmaster: dungeonmaster,
	}
}

func (p *GenerateQuestForZoneProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate quest for zone task: %v", task.Type())

	var payload jobs.GenerateQuestForZoneTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Generating quest for zone ID: %v with quest archetype ID: %v", payload.ZoneID, payload.QuestArchetypeID)
	return p.generateQuestForZone(ctx, payload.ZoneID, payload.QuestArchetypeID, payload.QuestGiverCharacterID, payload.QuestGenerationJobID)
}

func (p *GenerateQuestForZoneProcessor) generateQuestForZone(ctx context.Context, zoneID uuid.UUID, questArchetypeID uuid.UUID, questGiverCharacterID *uuid.UUID, questGenerationJobID *uuid.UUID) error {
	log.Printf("Finding zone with ID: %v", zoneID)
	zone, err := p.dbClient.Zone().FindByID(ctx, zoneID)
	if err != nil {
		log.Printf("Failed to find zone: %v", err)
		return fmt.Errorf("failed to find zone: %w", err)
	}

	log.Printf("Found zone: %v, generating quest...", zone.Name)
	if questGenerationJobID != nil {
		if err := p.dbClient.QuestGenerationJob().MarkInProgress(ctx, *questGenerationJobID); err != nil {
			log.Printf("Failed to mark quest generation job in progress: %v", err)
		}
	}

	quest, err := p.dungeonmaster.GenerateQuest(ctx, zone, questArchetypeID, questGiverCharacterID)
	if err != nil {
		log.Printf("Failed to generate quest: %v", err)
		if questGenerationJobID != nil {
			shouldRecord := shouldRecordFailure(ctx)
			if isBadRequestError(err) {
				shouldRecord = true
			}
			if shouldRecord {
				if recordErr := p.dbClient.QuestGenerationJob().RecordFailure(ctx, *questGenerationJobID, err.Error()); recordErr != nil {
					log.Printf("Failed to record quest generation failure: %v", recordErr)
				}
			}
		}
		if isBadRequestError(err) {
			return fmt.Errorf("non-retriable error: %v: %w", err, asynq.SkipRetry)
		}
		return fmt.Errorf("failed to generate quest: %w", err)
	}
	log.Printf("Successfully generated quest for zone %v", zone.Name)

	if questGenerationJobID != nil && quest != nil {
		if err := p.dbClient.QuestGenerationJob().RecordSuccess(ctx, *questGenerationJobID, quest.ID); err != nil {
			log.Printf("Failed to record quest generation success: %v", err)
		}
	}

	return nil
}

func shouldRecordFailure(ctx context.Context) bool {
	maxRetry, hasMax := asynq.GetMaxRetry(ctx)
	retryCount, hasRetry := asynq.GetRetryCount(ctx)
	if !hasMax || !hasRetry {
		return true
	}
	return retryCount >= maxRetry
}

func isBadRequestError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "status 400") || strings.Contains(message, "400 bad request")
}
