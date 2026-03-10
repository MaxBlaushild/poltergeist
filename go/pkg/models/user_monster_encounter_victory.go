package models

import (
	"time"

	"github.com/google/uuid"
)

type UserMonsterEncounterVictory struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	UserID             uuid.UUID `json:"userId" gorm:"column:user_id"`
	MonsterEncounterID uuid.UUID `json:"monsterEncounterId" gorm:"column:monster_encounter_id"`
}

func (UserMonsterEncounterVictory) TableName() string {
	return "user_monster_encounter_victories"
}
