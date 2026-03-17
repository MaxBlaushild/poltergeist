package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type scenarioTemplateGenerationJobHandle struct {
	db *gorm.DB
}

func (h *scenarioTemplateGenerationJobHandle) Create(ctx context.Context, job *models.ScenarioTemplateGenerationJob) error {
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *scenarioTemplateGenerationJobHandle) Update(ctx context.Context, job *models.ScenarioTemplateGenerationJob) error {
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *scenarioTemplateGenerationJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ScenarioTemplateGenerationJob, error) {
	var job models.ScenarioTemplateGenerationJob
	if err := h.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (h *scenarioTemplateGenerationJobHandle) FindRecent(ctx context.Context, limit int) ([]models.ScenarioTemplateGenerationJob, error) {
	var jobs []models.ScenarioTemplateGenerationJob
	q := h.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
