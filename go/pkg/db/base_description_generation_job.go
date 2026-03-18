package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type baseDescriptionGenerationJobHandle struct {
	db *gorm.DB
}

func (h *baseDescriptionGenerationJobHandle) Create(ctx context.Context, job *models.BaseDescriptionGenerationJob) error {
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *baseDescriptionGenerationJobHandle) Update(ctx context.Context, job *models.BaseDescriptionGenerationJob) error {
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *baseDescriptionGenerationJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.BaseDescriptionGenerationJob, error) {
	var job models.BaseDescriptionGenerationJob
	if err := h.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (h *baseDescriptionGenerationJobHandle) FindRecent(ctx context.Context, limit int) ([]models.BaseDescriptionGenerationJob, error) {
	var jobs []models.BaseDescriptionGenerationJob
	q := h.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (h *baseDescriptionGenerationJobHandle) FindByBaseID(ctx context.Context, baseID uuid.UUID, limit int) ([]models.BaseDescriptionGenerationJob, error) {
	var jobs []models.BaseDescriptionGenerationJob
	q := h.db.WithContext(ctx).Where("base_id = ?", baseID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
