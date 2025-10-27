package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type movementPatternHandler struct {
	db *gorm.DB
}

func (h *movementPatternHandler) Create(ctx context.Context, movementPattern *models.MovementPattern) error {
	return h.db.WithContext(ctx).Create(movementPattern).Error
}

func (h *movementPatternHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.MovementPattern, error) {
	var movementPattern models.MovementPattern
	if err := h.db.WithContext(ctx).First(&movementPattern, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &movementPattern, nil
}

func (h *movementPatternHandler) FindAll(ctx context.Context) ([]*models.MovementPattern, error) {
	var movementPatterns []*models.MovementPattern
	if err := h.db.WithContext(ctx).Find(&movementPatterns).Error; err != nil {
		return nil, err
	}
	return movementPatterns, nil
}

func (h *movementPatternHandler) Update(ctx context.Context, id uuid.UUID, updates *models.MovementPattern) error {
	return h.db.WithContext(ctx).Model(&models.MovementPattern{}).Where("id = ?", id).Updates(updates).Error
}

func (h *movementPatternHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.MovementPattern{}, id).Error
}

func (h *movementPatternHandler) FindByType(ctx context.Context, patternType models.MovementPatternType) ([]*models.MovementPattern, error) {
	var movementPatterns []*models.MovementPattern
	if err := h.db.WithContext(ctx).Where("movement_pattern_type = ?", patternType).Find(&movementPatterns).Error; err != nil {
		return nil, err
	}
	return movementPatterns, nil
}

func (h *movementPatternHandler) FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]*models.MovementPattern, error) {
	var movementPatterns []*models.MovementPattern
	if err := h.db.WithContext(ctx).Where("zone_id = ?", zoneID).Find(&movementPatterns).Error; err != nil {
		return nil, err
	}
	return movementPatterns, nil
}
