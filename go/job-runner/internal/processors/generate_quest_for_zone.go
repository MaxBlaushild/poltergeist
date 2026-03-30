package processors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/dungeonmaster"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
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
	attemptNumber, maxRetry := questGenerationRetryState(ctx)
	log.Printf(
		"Processing generate quest for zone task type=%s attempt=%d max_retry=%d",
		task.Type(),
		attemptNumber,
		maxRetry,
	)

	var payload jobs.GenerateQuestForZoneTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf(
		"Generating quest for zone zone_id=%s quest_archetype_id=%s quest_generation_job_id=%v quest_giver_character_id=%v attempt=%d/%d",
		payload.ZoneID,
		payload.QuestArchetypeID,
		payload.QuestGenerationJobID,
		payload.QuestGiverCharacterID,
		attemptNumber,
		maxRetry,
	)
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
	releaseReservation := false
	if questGenerationJobID != nil {
		started, startErr := p.dbClient.QuestGenerationJob().TryStart(ctx, *questGenerationJobID)
		if startErr != nil {
			log.Printf("Failed to reserve quest generation job slot: %v", startErr)
			return fmt.Errorf("failed to reserve quest generation job slot: %w", startErr)
		}
		if !started {
			log.Printf(
				"Skipping duplicate or excess quest generation task zone_id=%s quest_archetype_id=%s quest_generation_job_id=%s",
				zoneID,
				questArchetypeID,
				*questGenerationJobID,
			)
			return nil
		}
		releaseReservation = true
	}

	quest, err := p.dungeonmaster.GenerateQuest(ctx, zone, questArchetypeID, questGiverCharacterID)
	if err != nil {
		attemptNumber, maxRetry := questGenerationRetryState(ctx)
		nonRetriable := isNonRetriableQuestGenerationError(err)
		log.Printf(
			"Failed to generate quest zone_id=%s zone_name=%q quest_archetype_id=%s quest_generation_job_id=%v attempt=%d/%d non_retriable=%t error=%v",
			zoneID,
			zone.Name,
			questArchetypeID,
			questGenerationJobID,
			attemptNumber,
			maxRetry,
			nonRetriable,
			err,
		)
		if questGenerationJobID != nil {
			shouldRecord := shouldRecordFailure(ctx)
			if nonRetriable {
				shouldRecord = true
			}
			if shouldRecord {
				releaseReservation = false
				if recordErr := p.dbClient.QuestGenerationJob().RecordFailure(ctx, *questGenerationJobID, err.Error()); recordErr != nil {
					log.Printf("Failed to record quest generation failure: %v", recordErr)
				}
			} else if releaseReservation {
				if releaseErr := p.dbClient.QuestGenerationJob().ReleaseReservation(ctx, *questGenerationJobID); releaseErr != nil {
					log.Printf("Failed to release quest generation job reservation: %v", releaseErr)
				}
				releaseReservation = false
			}
		}
		if nonRetriable {
			return fmt.Errorf("non-retriable error: %v: %w", err, asynq.SkipRetry)
		}
		return fmt.Errorf("failed to generate quest: %w", err)
	}
	log.Printf("Successfully generated quest for zone %v", zone.Name)

	if questGenerationJobID != nil && quest != nil {
		if err := p.dbClient.QuestGenerationJob().RecordSuccess(ctx, *questGenerationJobID, quest.ID); err != nil {
			log.Printf("Failed to record quest generation success: %v", err)
		} else {
			releaseReservation = false
		}
	}

	if questGenerationJobID != nil && releaseReservation {
		if err := p.dbClient.QuestGenerationJob().ReleaseReservation(ctx, *questGenerationJobID); err != nil {
			log.Printf("Failed to release quest generation job reservation after generation: %v", err)
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

func isNonRetriableQuestGenerationError(err error) bool {
	if err == nil {
		return false
	}
	return isBadRequestError(err) ||
		isDatabaseSchemaError(err) ||
		dungeonmaster.IsNonRetriableQuestGenerationError(err) ||
		errors.Is(err, gorm.ErrRecordNotFound)
}

func isDatabaseSchemaError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "sqlstate 42703") || strings.Contains(message, "sqlstate 42p01") {
		return true
	}
	if strings.Contains(message, "column") && strings.Contains(message, "does not exist") {
		return true
	}
	if strings.Contains(message, "relation") && strings.Contains(message, "does not exist") {
		return true
	}
	return false
}

func questGenerationRetryState(ctx context.Context) (int, int) {
	maxRetry, hasMax := asynq.GetMaxRetry(ctx)
	retryCount, hasRetry := asynq.GetRetryCount(ctx)
	if !hasMax || !hasRetry {
		return 1, 0
	}
	return retryCount + 1, maxRetry + 1
}
