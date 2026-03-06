package models

import (
	"time"

	"github.com/google/uuid"
)

type MonsterBattleParticipant struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	BattleID    uuid.UUID `json:"battleId" gorm:"column:battle_id"`
	UserID      uuid.UUID `json:"userId" gorm:"column:user_id"`
	IsInitiator bool      `json:"isInitiator" gorm:"column:is_initiator"`
	JoinedAt    time.Time `json:"joinedAt" gorm:"column:joined_at"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (m *MonsterBattleParticipant) TableName() string {
	return "monster_battle_participants"
}
