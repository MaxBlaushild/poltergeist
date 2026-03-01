package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type monsterTemplateHandle struct {
	db *gorm.DB
}

func (h *monsterTemplateHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("Spells").
		Preload("Spells.Spell")
}

func (h *monsterTemplateHandle) Create(ctx context.Context, template *models.MonsterTemplate) error {
	now := time.Now()
	if template.ID == uuid.Nil {
		template.ID = uuid.New()
	}
	if template.CreatedAt.IsZero() {
		template.CreatedAt = now
	}
	template.UpdatedAt = now
	return h.db.WithContext(ctx).Create(template).Error
}

func (h *monsterTemplateHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.MonsterTemplate, error) {
	var template models.MonsterTemplate
	if err := h.preloadBase(ctx).Where("id = ?", id).First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (h *monsterTemplateHandle) FindAll(ctx context.Context) ([]models.MonsterTemplate, error) {
	var templates []models.MonsterTemplate
	if err := h.preloadBase(ctx).Order("name ASC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (h *monsterTemplateHandle) Update(ctx context.Context, id uuid.UUID, updates *models.MonsterTemplate) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	payload := map[string]interface{}{
		"name":                    updates.Name,
		"description":             updates.Description,
		"image_url":               updates.ImageURL,
		"thumbnail_url":           updates.ThumbnailURL,
		"base_strength":           updates.BaseStrength,
		"base_dexterity":          updates.BaseDexterity,
		"base_constitution":       updates.BaseConstitution,
		"base_intelligence":       updates.BaseIntelligence,
		"base_wisdom":             updates.BaseWisdom,
		"base_charisma":           updates.BaseCharisma,
		"image_generation_status": updates.ImageGenerationStatus,
		"image_generation_error":  updates.ImageGenerationError,
		"updated_at":              updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).Model(&models.MonsterTemplate{}).Where("id = ?", id).Updates(payload).Error
}

func (h *monsterTemplateHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.MonsterTemplate{}, "id = ?", id).Error
}

func (h *monsterTemplateHandle) ReplaceSpells(ctx context.Context, templateID uuid.UUID, spells []models.MonsterTemplateSpell) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("monster_template_id = ?", templateID).Delete(&models.MonsterTemplateSpell{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, spell := range spells {
			spell.ID = uuid.New()
			spell.MonsterTemplateID = templateID
			spell.CreatedAt = now
			spell.UpdatedAt = now
			if err := tx.Create(&spell).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
