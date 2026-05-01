package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type expositionTemplateGenerationJobHandle struct {
	db *gorm.DB
}

func (h *expositionTemplateGenerationJobHandle) Create(ctx context.Context, job *models.ExpositionTemplateGenerationJob) error {
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *expositionTemplateGenerationJobHandle) Update(ctx context.Context, job *models.ExpositionTemplateGenerationJob) error {
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *expositionTemplateGenerationJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ExpositionTemplateGenerationJob, error) {
	var job models.ExpositionTemplateGenerationJob
	if err := h.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (h *expositionTemplateGenerationJobHandle) FindRecent(ctx context.Context, limit int) ([]models.ExpositionTemplateGenerationJob, error) {
	var jobs []models.ExpositionTemplateGenerationJob
	q := h.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
