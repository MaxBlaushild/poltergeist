package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type zoneQuestArchetypeHandle struct {
	db *gorm.DB
}

func (h *zoneQuestArchetypeHandle) Create(ctx context.Context, zoneQuestArchetype *models.ZoneQuestArchetype) error {
	return h.db.Create(zoneQuestArchetype).Error
}

func (h *zoneQuestArchetypeHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]*models.ZoneQuestArchetype, error) {
	var zoneQuestArchetypes []*models.ZoneQuestArchetype
	if err := h.db.Where("zone_id = ?", zoneID).Find(&zoneQuestArchetypes).Error; err != nil {
		return nil, err
	}
	return zoneQuestArchetypes, nil
}

func (h *zoneQuestArchetypeHandle) FindByZoneIDAndQuestArchetypeID(ctx context.Context, zoneID uuid.UUID, questArchetypeID uuid.UUID) (*models.ZoneQuestArchetype, error) {
	var zoneQuestArchetype *models.ZoneQuestArchetype
	if err := h.db.Where("zone_id = ? AND quest_archetype_id = ?", zoneID, questArchetypeID).First(&zoneQuestArchetype).Error; err != nil {
		return nil, err
	}
	return zoneQuestArchetype, nil
}

func (h *zoneQuestArchetypeHandle) Delete(ctx context.Context, zoneQuestArchetypeID uuid.UUID) error {
	return h.db.Where("id = ?", zoneQuestArchetypeID).Delete(&models.ZoneQuestArchetype{}).Error
}

func (h *zoneQuestArchetypeHandle) DeleteByZoneIDAndQuestArchetypeID(ctx context.Context, zoneID uuid.UUID, questArchetypeID uuid.UUID) error {
	return h.db.Where("zone_id = ? AND quest_archetype_id = ?", zoneID, questArchetypeID).Delete(&models.ZoneQuestArchetype{}).Error
}

func (h *zoneQuestArchetypeHandle) DeleteByZoneID(ctx context.Context, zoneID uuid.UUID) error {
	return h.db.Where("zone_id = ?", zoneID).Delete(&models.ZoneQuestArchetype{}).Error
}

func (h *zoneQuestArchetypeHandle) DeleteByQuestArchetypeID(ctx context.Context, questArchetypeID uuid.UUID) error {
	return h.db.Where("quest_archetype_id = ?", questArchetypeID).Delete(&models.ZoneQuestArchetype{}).Error
}

func (h *zoneQuestArchetypeHandle) DeleteAll(ctx context.Context) error {
	return h.db.Delete(&models.ZoneQuestArchetype{}).Error
}

func (h *zoneQuestArchetypeHandle) FindAll(ctx context.Context) ([]*models.ZoneQuestArchetype, error) {
	var zoneQuestArchetypes []*models.ZoneQuestArchetype
	if err := h.db.Preload("QuestArchetype").Preload("Zone").Find(&zoneQuestArchetypes).Error; err != nil {
		return nil, err
	}
	return zoneQuestArchetypes, nil
}
