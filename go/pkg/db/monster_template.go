package db

import (
	"context"
	"strings"
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
		Preload("Genre").
		Preload("Spells").
		Preload("Spells.Spell").
		Preload("Progressions").
		Preload("Progressions.Progression").
		Preload("Progressions.Progression.Members", func(db *gorm.DB) *gorm.DB {
			return db.Order("level_band ASC")
		}).
		Preload("Progressions.Progression.Members.Spell")
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
	template.MonsterType = models.NormalizeMonsterTemplateType(string(template.MonsterType))
	template.ZoneKind = models.NormalizeZoneKind(template.ZoneKind)
	resolvedGenreID, err := resolveMonsterTemplateGenreID(ctx, h.db, template)
	if err != nil {
		return err
	}
	template.GenreID = resolvedGenreID
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

func (h *monsterTemplateHandle) FindAllActive(ctx context.Context) ([]models.MonsterTemplate, error) {
	var templates []models.MonsterTemplate
	if err := h.preloadBase(ctx).
		Where("archived = ?", false).
		Order("name ASC").
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

type monsterTemplateAdminListRow struct {
	ID   uuid.UUID `gorm:"column:id"`
	Name string    `gorm:"column:name"`
}

func (h *monsterTemplateHandle) adminListBaseQuery(
	ctx context.Context,
	params MonsterTemplateAdminListParams,
) *gorm.DB {
	query := h.db.WithContext(ctx).Model(&models.MonsterTemplate{})

	if normalizedQuery := strings.TrimSpace(strings.ToLower(params.Query)); normalizedQuery != "" {
		searchTerm := "%" + normalizedQuery + "%"
		query = query.Where(
			"(LOWER(monster_templates.name) LIKE ? OR LOWER(monster_templates.description) LIKE ?)",
			searchTerm,
			searchTerm,
		)
	}

	if normalizedZoneQuery := strings.TrimSpace(strings.ToLower(params.ZoneQuery)); normalizedZoneQuery != "" {
		zoneSearchTerm := "%" + normalizedZoneQuery + "%"
		query = query.
			Joins("JOIN monsters ON monsters.template_id = monster_templates.id").
			Joins("JOIN zones ON zones.id = monsters.zone_id").
			Where("LOWER(zones.name) LIKE ?", zoneSearchTerm)
	}

	if params.GenreID != nil && *params.GenreID != uuid.Nil {
		query = query.Where("monster_templates.genre_id = ?", *params.GenreID)
	}

	switch normalizedType := strings.TrimSpace(strings.ToLower(params.MonsterType)); normalizedType {
	case "", "all":
	default:
		query = query.Where(
			"monster_templates.monster_type = ?",
			models.NormalizeMonsterTemplateType(normalizedType),
		)
	}

	return query
}

func (h *monsterTemplateHandle) ListAdmin(
	ctx context.Context,
	params MonsterTemplateAdminListParams,
) (*MonsterTemplateAdminListResult, error) {
	var total int64
	countQuery := h.adminListBaseQuery(ctx, params)
	if params.Archived != nil {
		countQuery = countQuery.Where("monster_templates.archived = ?", *params.Archived)
	}
	if err := countQuery.Distinct("monster_templates.id").Count(&total).Error; err != nil {
		return nil, err
	}

	var activeCount int64
	if err := h.adminListBaseQuery(ctx, params).
		Where("monster_templates.archived = ?", false).
		Distinct("monster_templates.id").
		Count(&activeCount).Error; err != nil {
		return nil, err
	}

	var archivedCount int64
	if err := h.adminListBaseQuery(ctx, params).
		Where("monster_templates.archived = ?", true).
		Distinct("monster_templates.id").
		Count(&archivedCount).Error; err != nil {
		return nil, err
	}

	rows := []monsterTemplateAdminListRow{}
	listQuery := h.adminListBaseQuery(ctx, params)
	if params.Archived != nil {
		listQuery = listQuery.Where("monster_templates.archived = ?", *params.Archived)
	}
	if err := listQuery.
		Select("monster_templates.id, monster_templates.name").
		Distinct().
		Order("monster_templates.name ASC").
		Limit(params.PageSize).
		Offset((params.Page - 1) * params.PageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}

	templates := make([]models.MonsterTemplate, 0, len(ids))
	if len(ids) > 0 {
		loaded := []models.MonsterTemplate{}
		if err := h.preloadBase(ctx).
			Where("monster_templates.id IN ?", ids).
			Find(&loaded).Error; err != nil {
			return nil, err
		}
		templatesByID := make(map[uuid.UUID]models.MonsterTemplate, len(loaded))
		for _, template := range loaded {
			templatesByID[template.ID] = template
		}
		for _, id := range ids {
			template, ok := templatesByID[id]
			if ok {
				templates = append(templates, template)
			}
		}
	}

	return &MonsterTemplateAdminListResult{
		Templates:     templates,
		Total:         total,
		ActiveCount:   activeCount,
		ArchivedCount: archivedCount,
	}, nil
}

func (h *monsterTemplateHandle) Update(ctx context.Context, id uuid.UUID, updates *models.MonsterTemplate) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	updates.MonsterType = models.NormalizeMonsterTemplateType(string(updates.MonsterType))
	updates.ZoneKind = models.NormalizeZoneKind(updates.ZoneKind)
	resolvedGenreID, err := resolveMonsterTemplateGenreID(ctx, h.db, updates)
	if err != nil {
		return err
	}
	updates.GenreID = resolvedGenreID
	payload := map[string]interface{}{
		"archived":                         updates.Archived,
		"monster_type":                     updates.MonsterType,
		"zone_kind":                        updates.ZoneKind,
		"genre_id":                         updates.GenreID,
		"name":                             updates.Name,
		"description":                      updates.Description,
		"image_url":                        updates.ImageURL,
		"thumbnail_url":                    updates.ThumbnailURL,
		"base_strength":                    updates.BaseStrength,
		"base_dexterity":                   updates.BaseDexterity,
		"base_constitution":                updates.BaseConstitution,
		"base_intelligence":                updates.BaseIntelligence,
		"base_wisdom":                      updates.BaseWisdom,
		"base_charisma":                    updates.BaseCharisma,
		"physical_damage_bonus_percent":    updates.PhysicalDamageBonusPercent,
		"piercing_damage_bonus_percent":    updates.PiercingDamageBonusPercent,
		"slashing_damage_bonus_percent":    updates.SlashingDamageBonusPercent,
		"bludgeoning_damage_bonus_percent": updates.BludgeoningDamageBonusPercent,
		"fire_damage_bonus_percent":        updates.FireDamageBonusPercent,
		"ice_damage_bonus_percent":         updates.IceDamageBonusPercent,
		"lightning_damage_bonus_percent":   updates.LightningDamageBonusPercent,
		"poison_damage_bonus_percent":      updates.PoisonDamageBonusPercent,
		"arcane_damage_bonus_percent":      updates.ArcaneDamageBonusPercent,
		"holy_damage_bonus_percent":        updates.HolyDamageBonusPercent,
		"shadow_damage_bonus_percent":      updates.ShadowDamageBonusPercent,
		"physical_resistance_percent":      updates.PhysicalResistancePercent,
		"piercing_resistance_percent":      updates.PiercingResistancePercent,
		"slashing_resistance_percent":      updates.SlashingResistancePercent,
		"bludgeoning_resistance_percent":   updates.BludgeoningResistancePercent,
		"fire_resistance_percent":          updates.FireResistancePercent,
		"ice_resistance_percent":           updates.IceResistancePercent,
		"lightning_resistance_percent":     updates.LightningResistancePercent,
		"poison_resistance_percent":        updates.PoisonResistancePercent,
		"arcane_resistance_percent":        updates.ArcaneResistancePercent,
		"holy_resistance_percent":          updates.HolyResistancePercent,
		"shadow_resistance_percent":        updates.ShadowResistancePercent,
		"image_generation_status":          updates.ImageGenerationStatus,
		"image_generation_error":           updates.ImageGenerationError,
		"updated_at":                       updates.UpdatedAt,
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

func (h *monsterTemplateHandle) ReplaceProgressions(
	ctx context.Context,
	templateID uuid.UUID,
	progressions []models.MonsterTemplateProgression,
) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("monster_template_id = ?", templateID).Delete(&models.MonsterTemplateProgression{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, progression := range progressions {
			progression.ID = uuid.New()
			progression.MonsterTemplateID = templateID
			progression.CreatedAt = now
			progression.UpdatedAt = now
			if err := tx.Create(&progression).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
