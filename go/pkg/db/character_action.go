package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type characterActionHandler struct {
	db *gorm.DB
}

func (h *characterActionHandler) Create(ctx context.Context, characterAction *models.CharacterAction) error {
	return h.db.WithContext(ctx).Create(characterAction).Error
}

func (h *characterActionHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.CharacterAction, error) {
	var characterAction models.CharacterAction
	if err := h.db.WithContext(ctx).Preload("Character").First(&characterAction, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &characterAction, nil
}

func (h *characterActionHandler) FindAll(ctx context.Context) ([]*models.CharacterAction, error) {
	var characterActions []*models.CharacterAction
	if err := h.db.WithContext(ctx).Preload("Character").Find(&characterActions).Error; err != nil {
		return nil, err
	}
	return characterActions, nil
}

func (h *characterActionHandler) FindByCharacterID(ctx context.Context, characterID uuid.UUID) ([]*models.CharacterAction, error) {
	var characterActions []*models.CharacterAction
	if err := h.db.WithContext(ctx).Preload("Character").Where("character_id = ?", characterID).Find(&characterActions).Error; err != nil {
		return nil, err
	}
	return characterActions, nil
}

func (h *characterActionHandler) Update(ctx context.Context, id uuid.UUID, updates *models.CharacterAction) error {
	return h.db.WithContext(ctx).Model(&models.CharacterAction{}).Where("id = ?", id).Updates(updates).Error
}

func (h *characterActionHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.CharacterAction{}, id).Error
}
