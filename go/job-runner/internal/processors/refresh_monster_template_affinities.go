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
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type RefreshMonsterTemplateAffinitiesProcessor struct {
	dbClient    db.DbClient
	deepPriest  deep_priest.DeepPriest
	redisClient *redis.Client
}

func NewRefreshMonsterTemplateAffinitiesProcessor(
	dbClient db.DbClient,
	redisClient *redis.Client,
	deepPriest deep_priest.DeepPriest,
) RefreshMonsterTemplateAffinitiesProcessor {
	log.Println("Initializing RefreshMonsterTemplateAffinitiesProcessor")
	return RefreshMonsterTemplateAffinitiesProcessor{
		dbClient:    dbClient,
		deepPriest:  deepPriest,
		redisClient: redisClient,
	}
}

func (p *RefreshMonsterTemplateAffinitiesProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing refresh monster template affinities task: %v", task.Type())

	var payload jobs.RefreshMonsterTemplateAffinitiesTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	if payload.JobID == uuid.Nil {
		return fmt.Errorf("missing job ID")
	}

	statusKey := jobs.MonsterTemplateAffinityRefreshStatusKey(payload.JobID)
	now := time.Now().UTC()
	status := jobs.MonsterTemplateAffinityRefreshStatus{
		JobID:       payload.JobID,
		Status:      jobs.MonsterTemplateAffinityRefreshStatusInProgress,
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
		applyAffinityBonusesToMonsterTemplate(
			&template,
			scoreMonsterTemplateAffinities(ctx, &template, p.deepPriest),
		)
		if err := p.dbClient.MonsterTemplate().Update(ctx, template.ID, &template); err != nil {
			p.markFailed(ctx, statusKey, status, err)
			return fmt.Errorf("failed to update monster template %s affinities: %w", template.ID, err)
		}
		status.UpdatedCount++
		status.UpdatedAt = time.Now().UTC()
		p.setStatus(ctx, statusKey, status)
	}

	completedAt := time.Now().UTC()
	status.Status = jobs.MonsterTemplateAffinityRefreshStatusCompleted
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setStatus(ctx, statusKey, status)
	return nil
}

func (p *RefreshMonsterTemplateAffinitiesProcessor) markFailed(
	ctx context.Context,
	statusKey string,
	status jobs.MonsterTemplateAffinityRefreshStatus,
	cause error,
) {
	if cause != nil {
		status.Error = cause.Error()
	}
	completedAt := time.Now().UTC()
	status.Status = jobs.MonsterTemplateAffinityRefreshStatusFailed
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setStatus(ctx, statusKey, status)
}

func (p *RefreshMonsterTemplateAffinitiesProcessor) setStatus(
	ctx context.Context,
	statusKey string,
	status jobs.MonsterTemplateAffinityRefreshStatus,
) {
	if p.redisClient == nil || strings.TrimSpace(statusKey) == "" {
		return
	}
	payload, err := json.Marshal(status)
	if err != nil {
		log.Printf("Failed to marshal monster template affinity refresh status: %v", err)
		return
	}
	if err := p.redisClient.Set(ctx, statusKey, payload, jobs.MonsterTemplateAffinityRefreshStatusTTL).Err(); err != nil {
		log.Printf("Failed to write monster template affinity refresh status: %v", err)
	}
}
