package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type inventoryItemHandler struct {
	db *gorm.DB
}

func (h *inventoryItemHandler) FindAll(ctx context.Context) ([]models.InventoryItem, error) {
	var items []models.InventoryItem
	result := h.db.WithContext(ctx).Order("id").Find(&items)
	return items, result.Error
}

func (h *inventoryItemHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.InventoryItem, error) {
	var item models.InventoryItem
	result := h.db.WithContext(ctx).Where("id = ?", id).First(&item)
	if result.Error != nil {
		return nil, result.Error
	}
	return &item, nil
}

func (h *inventoryItemHandler) Create(ctx context.Context, item *models.InventoryItem) error {
	return h.db.WithContext(ctx).Create(item).Error
}

func (h *inventoryItemHandler) Update(ctx context.Context, item *models.InventoryItem) error {
	return h.db.WithContext(ctx).Save(item).Error
}

func (h *inventoryItemHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.InventoryItem{}, id).Error
}
