package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type monsterEncounterHandle struct {
	db *gorm.DB
}

func (h *monsterEncounterHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("Zone").
		Preload("Members", func(db *gorm.DB) *gorm.DB {
			return db.Order("slot ASC").Order("created_at ASC")
		}).
		Preload("Members.Monster").
		Preload("Members.Monster.Zone").
		Preload("Members.Monster.Template").
		Preload("Members.Monster.Template.Spells").
		Preload("Members.Monster.Template.Spells.Spell").
		Preload("Members.Monster.DominantHandInventoryItem").
		Preload("Members.Monster.OffHandInventoryItem").
		Preload("Members.Monster.WeaponInventoryItem").
		Preload("Members.Monster.ItemRewards").
		Preload("Members.Monster.ItemRewards.InventoryItem")
}

func (h *monsterEncounterHandle) Create(ctx context.Context, encounter *models.MonsterEncounter) error {
	now := time.Now()
	if encounter.ID == uuid.Nil {
		encounter.ID = uuid.New()
	}
	if encounter.CreatedAt.IsZero() {
		encounter.CreatedAt = now
	}
	encounter.UpdatedAt = now
	if err := encounter.SetGeometry(encounter.Latitude, encounter.Longitude); err != nil {
		return err
	}
	return h.db.WithContext(ctx).Create(encounter).Error
}

func (h *monsterEncounterHandle) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.MonsterEncounter, error) {
	var encounter models.MonsterEncounter
	if err := h.preloadBase(ctx).Where("id = ?", id).First(&encounter).Error; err != nil {
		return nil, err
	}
	return &encounter, nil
}

func (h *monsterEncounterHandle) FindAll(ctx context.Context) ([]models.MonsterEncounter, error) {
	var encounters []models.MonsterEncounter
	if err := h.preloadBase(ctx).Order("name ASC").Find(&encounters).Error; err != nil {
		return nil, err
	}
	return encounters, nil
}

func (h *monsterEncounterHandle) FindByZoneID(
	ctx context.Context,
	zoneID uuid.UUID,
) ([]models.MonsterEncounter, error) {
	var encounters []models.MonsterEncounter
	if err := h.preloadBase(ctx).
		Where("zone_id = ?", zoneID).
		Order("name ASC").
		Find(&encounters).Error; err != nil {
		return nil, err
	}
	return encounters, nil
}

func (h *monsterEncounterHandle) FindByZoneIDExcludingQuestNodes(
	ctx context.Context,
	zoneID uuid.UUID,
) ([]models.MonsterEncounter, error) {
	var encounters []models.MonsterEncounter
	if err := h.preloadBase(ctx).
		Where("zone_id = ?", zoneID).
		Where("NOT EXISTS (SELECT 1 FROM quest_nodes qn WHERE qn.monster_encounter_id = monster_encounters.id)").
		Order("name ASC").
		Find(&encounters).Error; err != nil {
		return nil, err
	}
	return encounters, nil
}

func (h *monsterEncounterHandle) FindFirstByMonsterID(
	ctx context.Context,
	monsterID uuid.UUID,
) (*models.MonsterEncounter, error) {
	var encounter models.MonsterEncounter
	err := h.preloadBase(ctx).
		Joins("JOIN monster_encounter_members mem ON mem.monster_encounter_id = monster_encounters.id").
		Where("mem.monster_id = ?", monsterID).
		Order("monster_encounters.created_at ASC").
		First(&encounter).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &encounter, nil
}

func (h *monsterEncounterHandle) Update(
	ctx context.Context,
	id uuid.UUID,
	updates *models.MonsterEncounter,
) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	if err := updates.SetGeometry(updates.Latitude, updates.Longitude); err != nil {
		return err
	}

	payload := map[string]interface{}{
		"name":          updates.Name,
		"description":   updates.Description,
		"image_url":     updates.ImageURL,
		"thumbnail_url": updates.ThumbnailURL,
		"zone_id":       updates.ZoneID,
		"latitude":      updates.Latitude,
		"longitude":     updates.Longitude,
		"geometry":      updates.Geometry,
		"updated_at":    updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).
		Model(&models.MonsterEncounter{}).
		Where("id = ?", id).
		Updates(payload).Error
}

func (h *monsterEncounterHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.MonsterEncounter{}, "id = ?", id).Error
}

func (h *monsterEncounterHandle) ReplaceMembers(
	ctx context.Context,
	encounterID uuid.UUID,
	members []models.MonsterEncounterMember,
) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("monster_encounter_id = ?", encounterID).
			Delete(&models.MonsterEncounterMember{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, member := range members {
			member.ID = uuid.New()
			member.CreatedAt = now
			member.UpdatedAt = now
			member.MonsterEncounterID = encounterID
			if err := tx.Create(&member).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
