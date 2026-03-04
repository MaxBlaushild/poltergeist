package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type challengeGenerationJobHandle struct {
	db *gorm.DB
}

func (h *challengeGenerationJobHandle) Create(ctx context.Context, job *models.ChallengeGenerationJob) error {
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *challengeGenerationJobHandle) Update(ctx context.Context, job *models.ChallengeGenerationJob) error {
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *challengeGenerationJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ChallengeGenerationJob, error) {
	var job models.ChallengeGenerationJob
	if err := h.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (h *challengeGenerationJobHandle) FindRecent(ctx context.Context, limit int) ([]models.ChallengeGenerationJob, error) {
	var jobs []models.ChallengeGenerationJob
	q := h.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (h *challengeGenerationJobHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID, limit int) ([]models.ChallengeGenerationJob, error) {
	var jobs []models.ChallengeGenerationJob
	q := h.db.WithContext(ctx).Where("zone_id = ?", zoneID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
