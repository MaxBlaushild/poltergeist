package models

import (
	"time"

	"github.com/google/uuid"
)

type MonsterBattleState string

const (
	MonsterBattleStatePendingPartyResponses MonsterBattleState = "pending_party_responses"
	MonsterBattleStateActive                MonsterBattleState = "active"
)

type MonsterBattle struct {
	ID                   uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	UserID               uuid.UUID  `json:"userId" gorm:"column:user_id"`
	MonsterID            uuid.UUID  `json:"monsterId" gorm:"column:monster_id"`
	State                string     `json:"state" gorm:"column:state"`
	TurnIndex            int        `json:"turnIndex" gorm:"column:turn_index"`
	StartedAt            time.Time  `json:"startedAt" gorm:"column:started_at"`
	LastActivityAt       time.Time  `json:"lastActivityAt" gorm:"column:last_activity_at"`
	MonsterHealthDeficit int        `json:"monsterHealthDeficit" gorm:"column:monster_health_deficit"`
	EndedAt              *time.Time `json:"endedAt,omitempty" gorm:"column:ended_at"`
}

func (m *MonsterBattle) TableName() string {
	return "monster_battles"
}
