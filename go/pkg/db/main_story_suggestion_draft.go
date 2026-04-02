package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mainStorySuggestionDraftHandle struct {
	db *gorm.DB
}

func (h *mainStorySuggestionDraftHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).Preload("MainStoryTemplate")
}

func (h *mainStorySuggestionDraftHandle) Create(ctx context.Context, draft *models.MainStorySuggestionDraft) error {
	if draft != nil {
		draft.Status = models.NormalizeMainStorySuggestionDraftStatus(draft.Status)
	}
	return h.db.WithContext(ctx).Create(draft).Error
}

func (h *mainStorySuggestionDraftHandle) Update(ctx context.Context, draft *models.MainStorySuggestionDraft) error {
	if draft != nil {
		draft.Status = models.NormalizeMainStorySuggestionDraftStatus(draft.Status)
	}
	return h.db.WithContext(ctx).Save(draft).Error
}

func (h *mainStorySuggestionDraftHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.MainStorySuggestionDraft, error) {
	var draft models.MainStorySuggestionDraft
	if err := h.preloadBase(ctx).First(&draft, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &draft, nil
}

func (h *mainStorySuggestionDraftHandle) FindByJobID(ctx context.Context, jobID uuid.UUID) ([]models.MainStorySuggestionDraft, error) {
	var drafts []models.MainStorySuggestionDraft
	if err := h.preloadBase(ctx).
		Where("job_id = ?", jobID).
		Order("created_at ASC").
		Find(&drafts).Error; err != nil {
		return nil, err
	}
	return drafts, nil
}

func (h *mainStorySuggestionDraftHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.MainStorySuggestionDraft{}, "id = ?", id).Error
}
