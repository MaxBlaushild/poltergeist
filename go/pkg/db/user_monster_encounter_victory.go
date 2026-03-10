package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userMonsterEncounterVictoryHandle struct {
	db *gorm.DB
}

func (h *userMonsterEncounterVictoryHandle) Upsert(
	ctx context.Context,
	userID uuid.UUID,
	encounterID uuid.UUID,
) error {
	now := time.Now()
	record := models.UserMonsterEncounterVictory{
		ID:                 uuid.New(),
		CreatedAt:          now,
		UpdatedAt:          now,
		UserID:             userID,
		MonsterEncounterID: encounterID,
	}
	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "monster_encounter_id"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"updated_at": now,
			}),
		}).
		Create(&record).Error
}

func (h *userMonsterEncounterVictoryHandle) FindEncounterIDsByUserAndZone(
	ctx context.Context,
	userID uuid.UUID,
	zoneID uuid.UUID,
) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0)
	if err := h.db.WithContext(ctx).
		Table("user_monster_encounter_victories AS umev").
		Select("umev.monster_encounter_id").
		Joins("JOIN monster_encounters me ON me.id = umev.monster_encounter_id").
		Where("umev.user_id = ? AND me.zone_id = ?", userID, zoneID).
		Find(&ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
