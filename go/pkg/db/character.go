package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type characterHandler struct {
	db *gorm.DB
}

func (h *characterHandler) Create(ctx context.Context, character *models.Character) error {
	return h.db.WithContext(ctx).Create(character).Error
}

func (h *characterHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.Character, error) {
	var character models.Character
	if err := h.db.WithContext(ctx).
		Preload("MovementPattern").
		Preload("PointOfInterest").
		Preload("Locations").
		First(&character, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &character, nil
}

func (h *characterHandler) FindAll(ctx context.Context) ([]*models.Character, error) {
	var characters []*models.Character
	if err := h.db.WithContext(ctx).
		Preload("MovementPattern").
		Preload("PointOfInterest").
		Preload("Locations").
		Find(&characters).Error; err != nil {
		return nil, err
	}
	return characters, nil
}

func (h *characterHandler) Update(ctx context.Context, id uuid.UUID, updates *models.Character) error {
	return h.db.WithContext(ctx).Model(&models.Character{}).Where("id = ?", id).Updates(updates).Error
}

func (h *characterHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Character{}, id).Error
}

func (h *characterHandler) FindByMovementPatternType(ctx context.Context, patternType models.MovementPatternType) ([]*models.Character, error) {
	var characters []*models.Character
	if err := h.db.WithContext(ctx).
		Preload("MovementPattern").
		Preload("PointOfInterest").
		Preload("Locations").
		Joins("JOIN movement_patterns ON characters.movement_pattern_id = movement_patterns.id").
		Where("movement_patterns.movement_pattern_type = ?", patternType).
		Find(&characters).Error; err != nil {
		return nil, err
	}
	return characters, nil
}
