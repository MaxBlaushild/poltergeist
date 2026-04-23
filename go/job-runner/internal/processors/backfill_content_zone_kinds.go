package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type BackfillContentZoneKindsProcessor struct {
	dbClient    db.DbClient
	deepPriest  deep_priest.DeepPriest
	redisClient *redis.Client
}

func NewBackfillContentZoneKindsProcessor(
	dbClient db.DbClient,
	redisClient *redis.Client,
	deepPriest deep_priest.DeepPriest,
) BackfillContentZoneKindsProcessor {
	log.Println("Initializing BackfillContentZoneKindsProcessor")
	return BackfillContentZoneKindsProcessor{
		dbClient:    dbClient,
		deepPriest:  deepPriest,
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
	if err := p.backfillMonsterTemplateZoneKinds(ctx, &status.Summary); err != nil {
		p.markFailed(ctx, statusKey, status, err)
		return fmt.Errorf("failed to backfill monster template zone kinds: %w", err)
	}
	if err := p.backfillScenarioTemplateZoneKinds(ctx, &status.Summary); err != nil {
		p.markFailed(ctx, statusKey, status, err)
		return fmt.Errorf("failed to backfill scenario template zone kinds: %w", err)
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

func appendZoneKindBackfillSummaryResult(
	summary *jobs.ZoneKindBackfillSummary,
	result jobs.ZoneKindBackfillResult,
) {
	if summary == nil {
		return
	}
	summary.Results = append(summary.Results, result)
	summary.MissingCount += result.MissingCount
	summary.AssignedCount += result.AssignedCount
	summary.AmbiguousCount += result.AmbiguousCount
	summary.SkippedCount += result.SkippedCount
}

func (p *BackfillContentZoneKindsProcessor) backfillMonsterTemplateZoneKinds(
	ctx context.Context,
	summary *jobs.ZoneKindBackfillSummary,
) error {
	templates, err := p.dbClient.MonsterTemplate().FindAll(ctx)
	if err != nil {
		return err
	}
	zoneKinds, err := p.dbClient.ZoneKind().FindAll(ctx)
	if err != nil {
		return err
	}

	missingCount := 0
	assignedCount := 0
	for i := range templates {
		template := &templates[i]
		if strings.TrimSpace(template.ZoneKind) != "" {
			continue
		}
		missingCount++
		profile := scoreMonsterTemplateProfile(ctx, template, zoneKinds, p.deepPriest)
		zoneKind := models.NormalizeZoneKind(profile.ZoneKind)
		if zoneKind == "" {
			continue
		}
		template.ZoneKind = zoneKind
		if err := p.dbClient.MonsterTemplate().Update(ctx, template.ID, template); err != nil {
			return err
		}
		assignedCount++
	}

	appendZoneKindBackfillSummaryResult(summary, jobs.ZoneKindBackfillResult{
		ContentType:   "monster_templates",
		MissingCount:  missingCount,
		AssignedCount: assignedCount,
		SkippedCount:  clampZoneKindBackfillSkippedCount(missingCount - assignedCount),
	})
	return nil
}

func (p *BackfillContentZoneKindsProcessor) backfillScenarioTemplateZoneKinds(
	ctx context.Context,
	summary *jobs.ZoneKindBackfillSummary,
) error {
	templates, err := p.dbClient.ScenarioTemplate().FindAll(ctx)
	if err != nil {
		return err
	}
	zoneKinds, err := p.dbClient.ZoneKind().FindAll(ctx)
	if err != nil {
		return err
	}

	missingCount := 0
	assignedCount := 0
	for i := range templates {
		template := &templates[i]
		if strings.TrimSpace(template.ZoneKind) != "" {
			continue
		}
		missingCount++
		zoneKind := models.NormalizeZoneKind(
			classifyScenarioTemplateZoneKind(ctx, template, zoneKinds, p.deepPriest),
		)
		if zoneKind == "" {
			continue
		}
		template.ZoneKind = zoneKind
		if err := p.dbClient.ScenarioTemplate().Update(ctx, template.ID, template); err != nil {
			return err
		}
		assignedCount++
	}

	appendZoneKindBackfillSummaryResult(summary, jobs.ZoneKindBackfillResult{
		ContentType:   "scenario_templates",
		MissingCount:  missingCount,
		AssignedCount: assignedCount,
		SkippedCount:  clampZoneKindBackfillSkippedCount(missingCount - assignedCount),
	})
	return nil
}

func clampZoneKindBackfillSkippedCount(value int) int {
	if value < 0 {
		return 0
	}
	return value
}
