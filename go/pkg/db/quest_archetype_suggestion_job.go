package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questArchetypeSuggestionJobHandle struct {
	db *gorm.DB
}

func (h *questArchetypeSuggestionJobHandle) Create(ctx context.Context, job *models.QuestArchetypeSuggestionJob) error {
	if job != nil {
		job.Status = models.NormalizeQuestArchetypeSuggestionJobStatus(job.Status)
	}
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *questArchetypeSuggestionJobHandle) Update(ctx context.Context, job *models.QuestArchetypeSuggestionJob) error {
	if job != nil {
		job.Status = models.NormalizeQuestArchetypeSuggestionJobStatus(job.Status)
	}
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *questArchetypeSuggestionJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.QuestArchetypeSuggestionJob, error) {
	var job models.QuestArchetypeSuggestionJob
	if err := h.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (h *questArchetypeSuggestionJobHandle) FindRecent(ctx context.Context, limit int) ([]models.QuestArchetypeSuggestionJob, error) {
	var jobs []models.QuestArchetypeSuggestionJob
	query := h.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
