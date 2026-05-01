package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/hibiken/asynq"
)

type GenerateShrineTemplatesProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
}

func NewGenerateShrineTemplatesProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
) GenerateShrineTemplatesProcessor {
	log.Println("Initializing GenerateShrineTemplatesProcessor")
	return GenerateShrineTemplatesProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
	}
}

func (p *GenerateShrineTemplatesProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate shrine templates task: %v", task.Type())

	var payload jobs.GenerateShrineTemplatesTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := p.dbClient.ShrineTemplateGenerationJob().FindByID(ctx, payload.JobID)
	if err != nil {
		return err
	}
	if job == nil {
		log.Printf("Shrine template generation job %s not found", payload.JobID)
		return nil
	}

	job.Status = models.ShrineTemplateGenerationStatusInProgress
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ShrineTemplateGenerationJob().Update(ctx, job); err != nil {
		return err
	}

	created, err := generateShrineTemplatesForZoneKind(
		ctx,
		p.dbClient,
		p.deepPriestClient,
		job.ZoneKind,
		job.Count,
	)
	if err != nil {
		return p.failShrineTemplateGenerationJob(ctx, job, err)
	}

	job.Status = models.ShrineTemplateGenerationStatusCompleted
	job.CreatedCount = len(created)
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	return p.dbClient.ShrineTemplateGenerationJob().Update(ctx, job)
}

func (p *GenerateShrineTemplatesProcessor) failShrineTemplateGenerationJob(
	ctx context.Context,
	job *models.ShrineTemplateGenerationJob,
	err error,
) error {
	msg := err.Error()
	job.Status = models.ShrineTemplateGenerationStatusFailed
	job.ErrorMessage = &msg
	job.UpdatedAt = time.Now()
	if updateErr := p.dbClient.ShrineTemplateGenerationJob().Update(ctx, job); updateErr != nil {
		log.Printf("Failed to mark shrine template generation job %s as failed: %v", job.ID, updateErr)
	}
	return err
}
