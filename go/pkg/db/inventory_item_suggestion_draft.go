package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type inventoryItemSuggestionDraftHandle struct {
	db *gorm.DB
}

func (h *inventoryItemSuggestionDraftHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("InventoryItem").
		Preload("InventoryItem.Genre")
}

func (h *inventoryItemSuggestionDraftHandle) Create(ctx context.Context, draft *models.InventoryItemSuggestionDraft) error {
	if draft != nil {
		draft.Status = models.NormalizeInventoryItemSuggestionDraftStatus(draft.Status)
	}
	return h.db.WithContext(ctx).Create(draft).Error
}

func (h *inventoryItemSuggestionDraftHandle) Update(ctx context.Context, draft *models.InventoryItemSuggestionDraft) error {
	if draft != nil {
		draft.Status = models.NormalizeInventoryItemSuggestionDraftStatus(draft.Status)
	}
	return h.db.WithContext(ctx).Save(draft).Error
}

func (h *inventoryItemSuggestionDraftHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.InventoryItemSuggestionDraft, error) {
	var draft models.InventoryItemSuggestionDraft
	if err := h.preloadBase(ctx).First(&draft, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &draft, nil
}

func (h *inventoryItemSuggestionDraftHandle) FindByJobID(ctx context.Context, jobID uuid.UUID) ([]models.InventoryItemSuggestionDraft, error) {
	var drafts []models.InventoryItemSuggestionDraft
	if err := h.preloadBase(ctx).
		Where("job_id = ?", jobID).
		Order("created_at ASC").
		Find(&drafts).Error; err != nil {
		return nil, err
	}
	return drafts, nil
}

func (h *inventoryItemSuggestionDraftHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.InventoryItemSuggestionDraft{}, "id = ?", id).Error
}
