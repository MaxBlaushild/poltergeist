package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type inventoryItemSuggestionJobHandle struct {
	db *gorm.DB
}

func (h *inventoryItemSuggestionJobHandle) Create(ctx context.Context, job *models.InventoryItemSuggestionJob) error {
	if job != nil {
		job.Status = models.NormalizeInventoryItemSuggestionJobStatus(job.Status)
	}
	return h.db.WithContext(ctx).Create(job).Error
}

func (h *inventoryItemSuggestionJobHandle) Update(ctx context.Context, job *models.InventoryItemSuggestionJob) error {
	if job != nil {
		job.Status = models.NormalizeInventoryItemSuggestionJobStatus(job.Status)
	}
	return h.db.WithContext(ctx).Save(job).Error
}

func (h *inventoryItemSuggestionJobHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.InventoryItemSuggestionJob, error) {
	var job models.InventoryItemSuggestionJob
	if err := h.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (h *inventoryItemSuggestionJobHandle) FindRecent(ctx context.Context, limit int) ([]models.InventoryItemSuggestionJob, error) {
	var jobs []models.InventoryItemSuggestionJob
	query := h.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
