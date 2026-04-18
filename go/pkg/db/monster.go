package db

import (
	"context"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type monsterHandle struct {
	db *gorm.DB
}

type monsterAdminListRow struct {
	ID   uuid.UUID `gorm:"column:id"`
	Name string    `gorm:"column:name"`
}

func (h *monsterHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("Zone").
		Preload("Genre").
		Preload("Template").
		Preload("Template.Genre").
		Preload("Template.Spells").
		Preload("Template.Spells.Spell").
		Preload("Template.Progressions").
		Preload("Template.Progressions.Progression").
		Preload("Template.Progressions.Progression.Members", func(db *gorm.DB) *gorm.DB {
			return db.Order("level_band ASC")
		}).
		Preload("Template.Progressions.Progression.Members.Spell").
		Preload("DominantHandInventoryItem").
		Preload("OffHandInventoryItem").
		Preload("WeaponInventoryItem").
		Preload("ItemRewards").
		Preload("ItemRewards.InventoryItem")
}

func (h *monsterHandle) Create(ctx context.Context, monster *models.Monster) error {
	now := time.Now()
	if monster.ID == uuid.Nil {
		monster.ID = uuid.New()
	}
	if monster.CreatedAt.IsZero() {
		monster.CreatedAt = now
	}
	monster.UpdatedAt = now
	if monster.Level < 1 {
		monster.Level = 1
	}
	if monster.DominantHandInventoryItemID != nil && monster.WeaponInventoryItemID == nil {
		monster.WeaponInventoryItemID = monster.DominantHandInventoryItemID
	}
	if strings.TrimSpace(string(monster.RewardMode)) == "" {
		if monster.RewardExperience > 0 || monster.RewardGold > 0 {
			monster.RewardMode = models.RewardModeExplicit
		} else {
			monster.RewardMode = models.RewardModeRandom
		}
	}
	monster.RewardMode = models.NormalizeRewardMode(string(monster.RewardMode))
	monster.RandomRewardSize = models.NormalizeRandomRewardSize(string(monster.RandomRewardSize))
	if monster.ImageGenerationStatus == "" {
		monster.ImageGenerationStatus = models.MonsterImageGenerationStatusNone
	}
	resolvedGenreID, err := resolveMonsterGenreID(ctx, h.db, monster)
	if err != nil {
		return err
	}
	monster.GenreID = resolvedGenreID
	if err := monster.SetGeometry(monster.Latitude, monster.Longitude); err != nil {
		return err
	}
	return h.db.WithContext(ctx).Create(monster).Error
}

func (h *monsterHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Monster, error) {
	var monster models.Monster
	if err := h.preloadBase(ctx).Where("id = ?", id).First(&monster).Error; err != nil {
		return nil, err
	}
	return &monster, nil
}

func (h *monsterHandle) FindAll(ctx context.Context) ([]models.Monster, error) {
	var monsters []models.Monster
	if err := h.preloadBase(ctx).Order("name ASC").Find(&monsters).Error; err != nil {
		return nil, err
	}
	return monsters, nil
}

func (h *monsterHandle) adminListBaseQuery(
	ctx context.Context,
	params MonsterAdminListParams,
) *gorm.DB {
	query := h.db.WithContext(ctx).
		Model(&models.Monster{}).
		Joins("LEFT JOIN zones ON zones.id = monsters.zone_id").
		Joins("LEFT JOIN monster_templates ON monster_templates.id = monsters.template_id")

	if normalizedQuery := strings.TrimSpace(strings.ToLower(params.Query)); normalizedQuery != "" {
		searchTerm := "%" + normalizedQuery + "%"
		query = query.Where(
			`(
				LOWER(monsters.name) LIKE ?
				OR LOWER(monsters.description) LIKE ?
				OR LOWER(COALESCE(zones.name, '')) LIKE ?
				OR LOWER(COALESCE(monster_templates.name, '')) LIKE ?
			)`,
			searchTerm,
			searchTerm,
			searchTerm,
			searchTerm,
		)
	}

	if normalizedZoneQuery := strings.TrimSpace(strings.ToLower(params.ZoneQuery)); normalizedZoneQuery != "" {
		zoneSearchTerm := "%" + normalizedZoneQuery + "%"
		query = query.Where("LOWER(COALESCE(zones.name, '')) LIKE ?", zoneSearchTerm)
	}

	return query
}

func (h *monsterHandle) ListAdmin(
	ctx context.Context,
	params MonsterAdminListParams,
) (*MonsterAdminListResult, error) {
	var total int64
	if err := h.adminListBaseQuery(ctx, params).
		Distinct("monsters.id").
		Count(&total).Error; err != nil {
		return nil, err
	}

	rows := []monsterAdminListRow{}
	if err := h.adminListBaseQuery(ctx, params).
		Select("monsters.id, monsters.name").
		Distinct().
		Order("monsters.name ASC").
		Limit(params.PageSize).
		Offset((params.Page - 1) * params.PageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}

	monsters := make([]models.Monster, 0, len(ids))
	if len(ids) > 0 {
		loaded := []models.Monster{}
		if err := h.preloadBase(ctx).
			Where("monsters.id IN ?", ids).
			Find(&loaded).Error; err != nil {
			return nil, err
		}
		monstersByID := make(map[uuid.UUID]models.Monster, len(loaded))
		for _, monster := range loaded {
			monstersByID[monster.ID] = monster
		}
		for _, id := range ids {
			monster, ok := monstersByID[id]
			if ok {
				monsters = append(monsters, monster)
			}
		}
	}

	return &MonsterAdminListResult{
		Monsters: monsters,
		Total:    total,
	}, nil
}

func (h *monsterHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]models.Monster, error) {
	var monsters []models.Monster
	if err := h.preloadBase(ctx).
		Where("zone_id = ?", zoneID).
		Order("name ASC").
		Find(&monsters).Error; err != nil {
		return nil, err
	}
	return monsters, nil
}

func (h *monsterHandle) FindByZoneIDExcludingQuestNodes(ctx context.Context, zoneID uuid.UUID) ([]models.Monster, error) {
	var monsters []models.Monster
	if err := h.preloadBase(ctx).
		Where("zone_id = ?", zoneID).
		Where(
			`NOT EXISTS (
				SELECT 1
				FROM quest_nodes qn
				WHERE qn.monster_id = monsters.id
					OR qn.monster_encounter_id IN (
						SELECT mem.monster_encounter_id
						FROM monster_encounter_members mem
						WHERE mem.monster_id = monsters.id
					)
			)`,
		).
		Order("name ASC").
		Find(&monsters).Error; err != nil {
		return nil, err
	}
	return monsters, nil
}

func (h *monsterHandle) CountByTemplateID(ctx context.Context, templateID uuid.UUID) (int64, error) {
	var count int64
	if err := h.db.WithContext(ctx).Model(&models.Monster{}).Where("template_id = ?", templateID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (h *monsterHandle) Update(ctx context.Context, id uuid.UUID, updates *models.Monster) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	if updates.Level < 1 {
		updates.Level = 1
	}
	updates.RewardMode = models.NormalizeRewardMode(string(updates.RewardMode))
	updates.RandomRewardSize = models.NormalizeRandomRewardSize(string(updates.RandomRewardSize))
	resolvedGenreID, err := resolveMonsterGenreID(ctx, h.db, updates)
	if err != nil {
		return err
	}
	updates.GenreID = resolvedGenreID
	if err := updates.SetGeometry(updates.Latitude, updates.Longitude); err != nil {
		return err
	}

	payload := map[string]interface{}{
		"name":                            updates.Name,
		"description":                     updates.Description,
		"image_url":                       updates.ImageURL,
		"thumbnail_url":                   updates.ThumbnailURL,
		"zone_id":                         updates.ZoneID,
		"latitude":                        updates.Latitude,
		"longitude":                       updates.Longitude,
		"geometry":                        updates.Geometry,
		"template_id":                     updates.TemplateID,
		"genre_id":                        updates.GenreID,
		"dominant_hand_inventory_item_id": updates.DominantHandInventoryItemID,
		"off_hand_inventory_item_id":      updates.OffHandInventoryItemID,
		"weapon_inventory_item_id":        updates.WeaponInventoryItemID,
		"level":                           updates.Level,
		"reward_mode":                     updates.RewardMode,
		"random_reward_size":              updates.RandomRewardSize,
		"reward_experience":               updates.RewardExperience,
		"reward_gold":                     updates.RewardGold,
		"image_generation_status":         updates.ImageGenerationStatus,
		"image_generation_error":          updates.ImageGenerationError,
		"updated_at":                      updates.UpdatedAt,
	}
	if updates.DominantHandInventoryItemID != nil {
		payload["weapon_inventory_item_id"] = updates.DominantHandInventoryItemID
	}
	return h.db.WithContext(ctx).Model(&models.Monster{}).Where("id = ?", id).Updates(payload).Error
}

func (h *monsterHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Monster{}, "id = ?", id).Error
}

func (h *monsterHandle) ReplaceItemRewards(ctx context.Context, monsterID uuid.UUID, rewards []models.MonsterItemReward) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("monster_id = ?", monsterID).Delete(&models.MonsterItemReward{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, reward := range rewards {
			reward.ID = uuid.New()
			reward.MonsterID = monsterID
			reward.CreatedAt = now
			reward.UpdatedAt = now
			if err := tx.Create(&reward).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
