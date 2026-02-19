package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type outfitProfileGenerationHandle struct {
	db *gorm.DB
}

func (h *outfitProfileGenerationHandle) Create(ctx context.Context, gen *models.OutfitProfileGeneration) error {
	return h.db.WithContext(ctx).Create(gen).Error
}

func (h *outfitProfileGenerationHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.OutfitProfileGeneration, error) {
	var gen models.OutfitProfileGeneration
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&gen).Error; err != nil {
		return nil, err
	}
	return &gen, nil
}

func (h *outfitProfileGenerationHandle) FindByOwnedInventoryItemID(ctx context.Context, ownedItemID uuid.UUID) (*models.OutfitProfileGeneration, error) {
	var gen models.OutfitProfileGeneration
	if err := h.db.WithContext(ctx).
		Where("owned_inventory_item_id = ?", ownedItemID).
		Order("created_at desc").
		First(&gen).Error; err != nil {
		return nil, err
	}
	return &gen, nil
}

func (h *outfitProfileGenerationHandle) Update(ctx context.Context, id uuid.UUID, updates *models.OutfitProfileGeneration) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Model(&models.OutfitProfileGeneration{}).Where("id = ?", id).Updates(updates).Error
}
