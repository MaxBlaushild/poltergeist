package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mainStorySuggestionJobHandle struct {
	db *gorm.DB
}

func (h *mainStorySuggestionJobHandle) Create(ctx context.Context, job *models.MainStorySuggestionJob) error {
	if job != nil {
		job.Status = models.NormalizeMainStorySuggestionJobStatus(job.Status)
	}
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *mainStorySuggestionJobHandle) Update(ctx context.Context, job *models.MainStorySuggestionJob) error {
	if job != nil {
		job.Status = models.NormalizeMainStorySuggestionJobStatus(job.Status)
	}
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *mainStorySuggestionJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.MainStorySuggestionJob, error) {
	var job models.MainStorySuggestionJob
	if err := h.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (h *mainStorySuggestionJobHandle) FindRecent(ctx context.Context, limit int) ([]models.MainStorySuggestionJob, error) {
	var jobs []models.MainStorySuggestionJob
	query := h.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
