package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questArchetypeSuggestionDraftHandle struct {
	db *gorm.DB
}

func (h *questArchetypeSuggestionDraftHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).Preload("QuestArchetype").Preload("QuestArchetype.Root")
}

func (h *questArchetypeSuggestionDraftHandle) Create(ctx context.Context, draft *models.QuestArchetypeSuggestionDraft) error {
	if draft != nil {
		draft.Status = models.NormalizeQuestArchetypeSuggestionDraftStatus(draft.Status)
	}
	return h.db.WithContext(ctx).Create(draft).Error
}

func (h *questArchetypeSuggestionDraftHandle) Update(ctx context.Context, draft *models.QuestArchetypeSuggestionDraft) error {
	if draft != nil {
		draft.Status = models.NormalizeQuestArchetypeSuggestionDraftStatus(draft.Status)
	}
	return h.db.WithContext(ctx).Save(draft).Error
}

func (h *questArchetypeSuggestionDraftHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.QuestArchetypeSuggestionDraft, error) {
	var draft models.QuestArchetypeSuggestionDraft
	if err := h.preloadBase(ctx).First(&draft, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &draft, nil
}

func (h *questArchetypeSuggestionDraftHandle) FindByJobID(ctx context.Context, jobID uuid.UUID) ([]models.QuestArchetypeSuggestionDraft, error) {
	var drafts []models.QuestArchetypeSuggestionDraft
	if err := h.preloadBase(ctx).
		Where("job_id = ?", jobID).
		Order("created_at ASC").
		Find(&drafts).Error; err != nil {
		return nil, err
	}
	return drafts, nil
}

func (h *questArchetypeSuggestionDraftHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.QuestArchetypeSuggestionDraft{}, "id = ?", id).Error
}
