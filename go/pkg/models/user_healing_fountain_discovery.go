package models

import (
	"time"

	"github.com/google/uuid"
)

type UserHealingFountainDiscovery struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	UserID            uuid.UUID       `json:"userId" gorm:"column:user_id;not null"`
	User              User            `json:"user"`
	HealingFountainID uuid.UUID       `json:"healingFountainId" gorm:"column:healing_fountain_id;not null"`
	HealingFountain   HealingFountain `json:"healingFountain"`
}

func (UserHealingFountainDiscovery) TableName() string {
	return "user_healing_fountain_discoveries"
}
