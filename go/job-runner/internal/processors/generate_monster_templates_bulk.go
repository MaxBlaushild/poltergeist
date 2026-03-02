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
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// GenerateMonsterTemplatesBulkProcessor creates DnD-inspired monster templates in the background.
type GenerateMonsterTemplatesBulkProcessor struct {
	dbClient    db.DbClient
	redisClient *redis.Client
}

func NewGenerateMonsterTemplatesBulkProcessor(dbClient db.DbClient, redisClient *redis.Client) GenerateMonsterTemplatesBulkProcessor {
	log.Println("Initializing GenerateMonsterTemplatesBulkProcessor")
	return GenerateMonsterTemplatesBulkProcessor{
		dbClient:    dbClient,
		redisClient: redisClient,
	}
}

func (p *GenerateMonsterTemplatesBulkProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate monster templates bulk task: %v", task.Type())

	var payload jobs.GenerateMonsterTemplatesBulkTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Failed to unmarshal generate monster templates bulk payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if payload.JobID == uuid.Nil {
		return fmt.Errorf("missing job ID")
	}

	statusKey := jobs.MonsterTemplateBulkStatusKey(payload.JobID)
	now := time.Now().UTC()
	status := jobs.MonsterTemplateBulkStatus{
		JobID:        payload.JobID,
		Status:       jobs.MonsterTemplateBulkStatusInProgress,
		Source:       strings.TrimSpace(payload.Source),
		TotalCount:   payload.TotalCount,
		CreatedCount: 0,
		StartedAt:    &now,
		UpdatedAt:    now,
	}
	if status.TotalCount <= 0 {
		status.TotalCount = len(payload.Templates)
	}
	if status.Source == "" {
		status.Source = "dnd_inspired"
	}
	p.setStatus(ctx, statusKey, status)

	if len(payload.Templates) == 0 {
		err := fmt.Errorf("no monster templates provided for bulk generation")
		p.markFailed(ctx, statusKey, status, err)
		return err
	}

	for index, spec := range payload.Templates {
		emptyError := ""
		template := &models.MonsterTemplate{
			Name:                  strings.TrimSpace(spec.Name),
			Description:           strings.TrimSpace(spec.Description),
			BaseStrength:          spec.BaseStrength,
			BaseDexterity:         spec.BaseDexterity,
			BaseConstitution:      spec.BaseConstitution,
			BaseIntelligence:      spec.BaseIntelligence,
			BaseWisdom:            spec.BaseWisdom,
			BaseCharisma:          spec.BaseCharisma,
			ImageGenerationStatus: models.MonsterTemplateImageGenerationStatusNone,
			ImageGenerationError:  &emptyError,
		}
		if template.Name == "" {
			template.Name = fmt.Sprintf("Monster Template %d", index+1)
		}

		if err := p.dbClient.MonsterTemplate().Create(ctx, template); err != nil {
			p.markFailed(ctx, statusKey, status, err)
			return fmt.Errorf("failed to create monster template %d/%d: %w", index+1, len(payload.Templates), err)
		}

		status.CreatedCount = index + 1
		status.UpdatedAt = time.Now().UTC()
		p.setStatus(ctx, statusKey, status)
	}

	completedAt := time.Now().UTC()
	status.Status = jobs.MonsterTemplateBulkStatusCompleted
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setStatus(ctx, statusKey, status)

	return nil
}

func (p *GenerateMonsterTemplatesBulkProcessor) markFailed(ctx context.Context, statusKey string, status jobs.MonsterTemplateBulkStatus, cause error) {
	if cause != nil {
		status.Error = cause.Error()
	}
	completedAt := time.Now().UTC()
	status.Status = jobs.MonsterTemplateBulkStatusFailed
	status.CompletedAt = &completedAt
	status.UpdatedAt = completedAt
	p.setStatus(ctx, statusKey, status)
}

func (p *GenerateMonsterTemplatesBulkProcessor) setStatus(ctx context.Context, statusKey string, status jobs.MonsterTemplateBulkStatus) {
	if p.redisClient == nil || strings.TrimSpace(statusKey) == "" {
		return
	}
	payload, err := json.Marshal(status)
	if err != nil {
		log.Printf("Failed to marshal monster template bulk status: %v", err)
		return
	}
	if err := p.redisClient.Set(ctx, statusKey, payload, jobs.MonsterTemplateBulkStatusTTL).Err(); err != nil {
		log.Printf("Failed to write monster template bulk status: %v", err)
	}
}
