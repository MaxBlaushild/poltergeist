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

func (h *characterHandler) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("Genre").
		Preload("PointOfInterest").
		Preload("Locations")
}

func (h *characterHandler) Create(ctx context.Context, character *models.Character) error {
	if character != nil {
		resolvedGenreID, err := resolveCharacterGenreID(ctx, h.db, character)
		if err != nil {
			return err
		}
		character.GenreID = resolvedGenreID
	}
	return h.db.WithContext(ctx).Create(character).Error
}

func (h *characterHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.Character, error) {
	var character models.Character
	if err := h.preloadBase(ctx).
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
	if err := h.preloadBase(ctx).
		Find(&characters).Error; err != nil {
		return nil, err
	}
	return characters, nil
}

func (h *characterHandler) FindByPointOfInterestID(ctx context.Context, pointOfInterestID uuid.UUID) ([]*models.Character, error) {
	var characters []*models.Character
	if err := h.preloadBase(ctx).
		Where("point_of_interest_id = ?", pointOfInterestID).
		Find(&characters).Error; err != nil {
		return nil, err
	}
	return characters, nil
}

func (h *characterHandler) Update(ctx context.Context, id uuid.UUID, updates *models.Character) error {
	return h.db.WithContext(ctx).Model(&models.Character{}).Where("id = ?", id).Updates(updates).Error
}

func (h *characterHandler) UpdateFields(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return h.db.WithContext(ctx).Model(&models.Character{}).Where("id = ?", id).Updates(updates).Error
}

func (h *characterHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Character{}, id).Error
}
