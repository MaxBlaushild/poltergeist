package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type shrineTemplateGenerationJobHandle struct {
	db *gorm.DB
}

func (h *shrineTemplateGenerationJobHandle) Create(ctx context.Context, job *models.ShrineTemplateGenerationJob) error {
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *shrineTemplateGenerationJobHandle) Update(ctx context.Context, job *models.ShrineTemplateGenerationJob) error {
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *shrineTemplateGenerationJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ShrineTemplateGenerationJob, error) {
	var job models.ShrineTemplateGenerationJob
	if err := h.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (h *shrineTemplateGenerationJobHandle) FindRecent(ctx context.Context, limit int) ([]models.ShrineTemplateGenerationJob, error) {
	var jobs []models.ShrineTemplateGenerationJob
	q := h.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
