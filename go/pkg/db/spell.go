package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type spellHandler struct {
	db *gorm.DB
}

func (h *spellHandler) Create(ctx context.Context, spell *models.Spell) error {
	now := time.Now()
	if spell.ID == uuid.Nil {
		spell.ID = uuid.New()
	}
	if spell.CreatedAt.IsZero() {
		spell.CreatedAt = now
	}
	spell.UpdatedAt = now
	if spell.Effects == nil {
		spell.Effects = models.SpellEffects{}
	}
	return h.db.WithContext(ctx).Create(spell).Error
}

func (h *spellHandler) FindByID(ctx context.Context, spellID uuid.UUID) (*models.Spell, error) {
	var spell models.Spell
	if err := h.db.WithContext(ctx).Where("id = ?", spellID).First(&spell).Error; err != nil {
		return nil, err
	}
	return &spell, nil
}

func (h *spellHandler) FindAll(ctx context.Context) ([]models.Spell, error) {
	var spells []models.Spell
	if err := h.db.WithContext(ctx).Order("name ASC").Find(&spells).Error; err != nil {
		return nil, err
	}
	return spells, nil
}

func (h *spellHandler) Update(ctx context.Context, spellID uuid.UUID, updates map[string]interface{}) error {
	if updates == nil {
		return nil
	}
	updates["updated_at"] = time.Now()
	return h.db.WithContext(ctx).Model(&models.Spell{}).Where("id = ?", spellID).Updates(updates).Error
}

func (h *spellHandler) Delete(ctx context.Context, spellID uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Spell{}, "id = ?", spellID).Error
}
