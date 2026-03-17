package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type challengeTemplateGenerationJobHandle struct {
	db *gorm.DB
}

func (h *challengeTemplateGenerationJobHandle) Create(ctx context.Context, job *models.ChallengeTemplateGenerationJob) error {
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *challengeTemplateGenerationJobHandle) Update(ctx context.Context, job *models.ChallengeTemplateGenerationJob) error {
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *challengeTemplateGenerationJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ChallengeTemplateGenerationJob, error) {
	var job models.ChallengeTemplateGenerationJob
	if err := h.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (h *challengeTemplateGenerationJobHandle) FindRecent(ctx context.Context, limit int) ([]models.ChallengeTemplateGenerationJob, error) {
	var jobs []models.ChallengeTemplateGenerationJob
	q := h.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (h *challengeTemplateGenerationJobHandle) FindByLocationArchetypeID(ctx context.Context, locationArchetypeID uuid.UUID, limit int) ([]models.ChallengeTemplateGenerationJob, error) {
	var jobs []models.ChallengeTemplateGenerationJob
	q := h.db.WithContext(ctx).Where("location_archetype_id = ?", locationArchetypeID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
