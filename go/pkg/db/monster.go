package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type monsterHandle struct {
	db *gorm.DB
}

func (h *monsterHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("Zone").
		Preload("Template").
		Preload("Template.Spells").
		Preload("Template.Spells.Spell").
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
	if monster.ImageGenerationStatus == "" {
		monster.ImageGenerationStatus = models.MonsterImageGenerationStatusNone
	}
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
		"dominant_hand_inventory_item_id": updates.DominantHandInventoryItemID,
		"off_hand_inventory_item_id":      updates.OffHandInventoryItemID,
		"weapon_inventory_item_id":        updates.WeaponInventoryItemID,
		"level":                           updates.Level,
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
