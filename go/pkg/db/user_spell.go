package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userSpellHandler struct {
	db *gorm.DB
}

func (h *userSpellHandler) GrantToUser(ctx context.Context, userID uuid.UUID, spellID uuid.UUID) error {
	now := time.Now()
	userSpell := &models.UserSpell{
		ID:         uuid.New(),
		CreatedAt:  now,
		UpdatedAt:  now,
		UserID:     userID,
		SpellID:    spellID,
		AcquiredAt: now,
	}
	return h.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "spell_id"}},
		DoNothing: true,
	}).Create(userSpell).Error
}

func (h *userSpellHandler) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserSpell, error) {
	var userSpells []models.UserSpell
	if err := h.db.WithContext(ctx).
		Preload("Spell").
		Where("user_id = ?", userID).
		Order("acquired_at ASC").
		Find(&userSpells).Error; err != nil {
		return nil, err
	}
	return userSpells, nil
}

func (h *userSpellHandler) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserSpell{}).Error
}
