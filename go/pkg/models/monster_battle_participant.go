package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MonsterBattleItemAwards []ItemAwarded

func (a MonsterBattleItemAwards) Value() (driver.Value, error) {
	if a == nil {
		return json.Marshal([]ItemAwarded{})
	}
	return json.Marshal(a)
}

func (a *MonsterBattleItemAwards) Scan(value interface{}) error {
	if value == nil {
		*a = MonsterBattleItemAwards{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*a = MonsterBattleItemAwards{}
		return nil
	}
	if len(bytes) == 0 {
		*a = MonsterBattleItemAwards{}
		return nil
	}

	return json.Unmarshal(bytes, a)
}

type MonsterBattleParticipant struct {
	ID               uuid.UUID               `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt        time.Time               `json:"createdAt"`
	UpdatedAt        time.Time               `json:"updatedAt"`
	BattleID         uuid.UUID               `json:"battleId" gorm:"column:battle_id"`
	UserID           uuid.UUID               `json:"userId" gorm:"column:user_id"`
	IsInitiator      bool                    `json:"isInitiator" gorm:"column:is_initiator"`
	JoinedAt         time.Time               `json:"joinedAt" gorm:"column:joined_at"`
	RewardExperience int                     `json:"rewardExperience" gorm:"column:reward_experience;default:0"`
	RewardGold       int                     `json:"rewardGold" gorm:"column:reward_gold;default:0"`
	ItemsAwarded     MonsterBattleItemAwards `json:"itemsAwarded" gorm:"column:items_awarded;type:jsonb;default:'[]'"`
	BaseResourcesAwarded BaseMaterialRewards `json:"baseResourcesAwarded" gorm:"column:base_resources_awarded;type:jsonb;default:'[]'"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (m *MonsterBattleParticipant) TableName() string {
	return "monster_battle_participants"
}
