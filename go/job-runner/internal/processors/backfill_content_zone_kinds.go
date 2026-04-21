package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type BackfillContentZoneKindsProcessor struct {
	dbClient    db.DbClient
	redisClient *redis.Client
}

func NewBackfillContentZoneKindsProcessor(
	dbClient db.DbClient,
	redisClient *redis.Client,
) BackfillContentZoneKindsProcessor {
	log.Println("Initializing BackfillContentZoneKindsProcessor")
	return BackfillContentZoneKindsProcessor{
		dbClient:    dbClient,
		redisClient: redisClient,
	}
}

func (p *BackfillContentZoneKindsProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing content zone kind backfill task: %v", task.Type())

	var payload jobs.BackfillContentZoneKindsTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	if payload.JobID == uuid.Nil {
		return fmt.Errorf("missing job ID")
	}

	statusKey := jobs.ZoneKindBackfillStatusKey(payload.JobID)
	now := time.Now().UTC()
	status := jobs.ZoneKindBackfillStatus{
		JobID:     payload.JobID,
		Status:    jobs.ZoneKindBackfillStatusInProgress,
		Summary:   jobs.ZoneKindBackfillSummary{Results: []jobs.ZoneKindBackfillResult{}},
		StartedAt: &now,
		UpdatedAt: now,
	}
	p.setStatus(ctx, statusKey, status)

	summary, err := p.dbClient.ZoneKind().BackfillMissingContentKinds(ctx)
	if err != nil {
		p.markFailed(ctx, statusKey, status, err)
		return fmt.Errorf("failed to backfill content zone kinds: %w", err)
	}
	if summary != nil {
		status.Summary = *summary
	}

	completedAt := time.Now().UTC()
	status.Status = jobs.ZoneKindBackfillStatusCompleted
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setStatus(ctx, statusKey, status)
	return nil
}

func (p *BackfillContentZoneKindsProcessor) markFailed(
	ctx context.Context,
	statusKey string,
	status jobs.ZoneKindBackfillStatus,
	cause error,
) {
	if cause != nil {
		status.Error = cause.Error()
	}
	completedAt := time.Now().UTC()
	status.Status = jobs.ZoneKindBackfillStatusFailed
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setStatus(ctx, statusKey, status)
}

func (p *BackfillContentZoneKindsProcessor) setStatus(
	ctx context.Context,
	statusKey string,
	status jobs.ZoneKindBackfillStatus,
) {
	if p.redisClient == nil || strings.TrimSpace(statusKey) == "" {
		return
	}
	payload, err := json.Marshal(status)
	if err != nil {
		log.Printf("Failed to marshal zone kind backfill status: %v", err)
		return
	}
	if err := p.redisClient.Set(ctx, statusKey, payload, jobs.ZoneKindBackfillStatusTTL).Err(); err != nil {
		log.Printf("Failed to write zone kind backfill status: %v", err)
	}
}
