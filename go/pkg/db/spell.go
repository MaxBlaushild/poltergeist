package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type spellHandler struct {
	db *gorm.DB
}

func (h *spellHandler) Create(ctx context.Context, spell *models.Spell) error {
	if spell != nil {
		resolvedGenreID, err := resolveSpellGenreID(ctx, h.db, spell)
		if err != nil {
			return err
		}
		spell.GenreID = resolvedGenreID
	}
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
	if spell.ImageGenerationStatus == "" {
		spell.ImageGenerationStatus = models.SpellImageGenerationStatusNone
	}
	if spell.AbilityType == "" {
		spell.AbilityType = models.SpellAbilityTypeSpell
	}
	if spell.AbilityLevel <= 0 {
		spell.AbilityLevel = 1
	}
	return h.db.WithContext(ctx).Create(spell).Error
}

func (h *spellHandler) FindByID(ctx context.Context, spellID uuid.UUID) (*models.Spell, error) {
	var spell models.Spell
	if err := h.db.WithContext(ctx).
		Preload("Genre").
		Preload("ProgressionLinks").
		Preload("ProgressionLinks.Progression").
		Where("id = ?", spellID).
		First(&spell).Error; err != nil {
		return nil, err
	}
	return &spell, nil
}

func (h *spellHandler) FindAll(ctx context.Context) ([]models.Spell, error) {
	var spells []models.Spell
	if err := h.db.WithContext(ctx).
		Preload("Genre").
		Preload("ProgressionLinks").
		Preload("ProgressionLinks.Progression").
		Order("name ASC").
		Find(&spells).Error; err != nil {
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

func (h *spellHandler) CreateProgression(ctx context.Context, progression *models.SpellProgression) error {
	if progression == nil {
		return nil
	}
	now := time.Now()
	if progression.ID == uuid.Nil {
		progression.ID = uuid.New()
	}
	if progression.CreatedAt.IsZero() {
		progression.CreatedAt = now
	}
	progression.UpdatedAt = now
	if progression.AbilityType == "" {
		progression.AbilityType = models.SpellAbilityTypeSpell
	}
	return h.db.WithContext(ctx).Create(progression).Error
}

func (h *spellHandler) FindProgressionByID(ctx context.Context, progressionID uuid.UUID) (*models.SpellProgression, error) {
	var progression models.SpellProgression
	if err := h.db.WithContext(ctx).
		Preload("Members").
		Preload("Members.Spell").
		Preload("Members.Spell.Genre").
		Where("id = ?", progressionID).
		First(&progression).Error; err != nil {
		return nil, err
	}
	return &progression, nil
}

func (h *spellHandler) FindProgressionBySpellID(ctx context.Context, spellID uuid.UUID) (*models.SpellProgression, error) {
	var link models.SpellProgressionSpell
	if err := h.db.WithContext(ctx).
		Where("spell_id = ?", spellID).
		First(&link).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var progression models.SpellProgression
	if err := h.db.WithContext(ctx).
		Preload("Members").
		Preload("Members.Spell").
		Preload("Members.Spell.Genre").
		Where("id = ?", link.ProgressionID).
		First(&progression).Error; err != nil {
		return nil, err
	}
	return &progression, nil
}

func (h *spellHandler) FindProgressionMembers(ctx context.Context, progressionID uuid.UUID) ([]models.SpellProgressionSpell, error) {
	var members []models.SpellProgressionSpell
	if err := h.db.WithContext(ctx).
		Preload("Spell").
		Preload("Spell.Genre").
		Where("progression_id = ?", progressionID).
		Order("level_band ASC").
		Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (h *spellHandler) UpsertProgressionMember(ctx context.Context, progressionID uuid.UUID, spellID uuid.UUID, levelBand int) error {
	now := time.Now()
	member := models.SpellProgressionSpell{
		ID:            uuid.New(),
		CreatedAt:     now,
		UpdatedAt:     now,
		ProgressionID: progressionID,
		SpellID:       spellID,
		LevelBand:     levelBand,
	}

	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "spell_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"progression_id", "level_band", "updated_at"}),
		}).
		Create(&member).Error
}
