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

type ResetMonsterTemplateProgressionsProcessor struct {
	dbClient    db.DbClient
	redisClient *redis.Client
}

func NewResetMonsterTemplateProgressionsProcessor(
	dbClient db.DbClient,
	redisClient *redis.Client,
) ResetMonsterTemplateProgressionsProcessor {
	log.Println("Initializing ResetMonsterTemplateProgressionsProcessor")
	return ResetMonsterTemplateProgressionsProcessor{
		dbClient:    dbClient,
		redisClient: redisClient,
	}
}

func (p *ResetMonsterTemplateProgressionsProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing reset monster template progressions task: %v", task.Type())

	var payload jobs.ResetMonsterTemplateProgressionsTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	if payload.JobID == uuid.Nil {
		return fmt.Errorf("missing job ID")
	}

	statusKey := jobs.MonsterTemplateProgressionResetStatusKey(payload.JobID)
	now := time.Now().UTC()
	status := jobs.MonsterTemplateProgressionResetStatus{
		JobID:       payload.JobID,
		Status:      jobs.MonsterTemplateProgressionResetStatusInProgress,
		TemplateIDs: payload.MonsterTemplateIDs,
		StartedAt:   &now,
		UpdatedAt:   now,
	}
	p.setStatus(ctx, statusKey, status)

	templates, err := p.dbClient.MonsterTemplate().FindAll(ctx)
	if err != nil {
		p.markFailed(ctx, statusKey, status, err)
		return fmt.Errorf("failed to load monster templates: %w", err)
	}

	candidates, err := loadMonsterTemplateProgressionCandidates(ctx, p.dbClient)
	if err != nil {
		p.markFailed(ctx, statusKey, status, err)
		return fmt.Errorf("failed to load spell progressions: %w", err)
	}
	if len(candidates) == 0 {
		err = fmt.Errorf("no spell or technique progressions are available")
		p.markFailed(ctx, statusKey, status, err)
		return err
	}

	selected := make(map[uuid.UUID]struct{}, len(payload.MonsterTemplateIDs))
	for _, id := range payload.MonsterTemplateIDs {
		selected[id] = struct{}{}
	}

	targets := make([]uuid.UUID, 0, len(templates))
	for i := range templates {
		if len(selected) > 0 {
			if _, ok := selected[templates[i].ID]; !ok {
				continue
			}
		}
		targets = append(targets, templates[i].ID)
	}
	status.TotalCount = len(targets)
	status.UpdatedAt = time.Now().UTC()
	p.setStatus(ctx, statusKey, status)

	templateIndex := make(map[uuid.UUID]int, len(templates))
	for i := range templates {
		templateIndex[templates[i].ID] = i
	}

	for _, templateID := range targets {
		index, ok := templateIndex[templateID]
		if !ok {
			continue
		}
		template := templates[index]
		progressions := chooseProgressionsForMonsterTemplate(&template, candidates)
		if err := p.dbClient.MonsterTemplate().ReplaceProgressions(ctx, template.ID, progressions); err != nil {
			p.markFailed(ctx, statusKey, status, err)
			return fmt.Errorf("failed to reset monster template %s progressions: %w", template.ID, err)
		}
		status.UpdatedCount++
		status.UpdatedAt = time.Now().UTC()
		p.setStatus(ctx, statusKey, status)
	}

	completedAt := time.Now().UTC()
	status.Status = jobs.MonsterTemplateProgressionResetStatusCompleted
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setStatus(ctx, statusKey, status)
	return nil
}

func (p *ResetMonsterTemplateProgressionsProcessor) markFailed(
	ctx context.Context,
	statusKey string,
	status jobs.MonsterTemplateProgressionResetStatus,
	cause error,
) {
	if cause != nil {
		status.Error = cause.Error()
	}
	completedAt := time.Now().UTC()
	status.Status = jobs.MonsterTemplateProgressionResetStatusFailed
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setStatus(ctx, statusKey, status)
}

func (p *ResetMonsterTemplateProgressionsProcessor) setStatus(
	ctx context.Context,
	statusKey string,
	status jobs.MonsterTemplateProgressionResetStatus,
) {
	if p.redisClient == nil || strings.TrimSpace(statusKey) == "" {
		return
	}
	payload, err := json.Marshal(status)
	if err != nil {
		log.Printf("Failed to marshal monster template progression reset status: %v", err)
		return
	}
	if err := p.redisClient.Set(ctx, statusKey, payload, jobs.MonsterTemplateProgressionResetStatusTTL).Err(); err != nil {
		log.Printf("Failed to write monster template progression reset status: %v", err)
	}
}
