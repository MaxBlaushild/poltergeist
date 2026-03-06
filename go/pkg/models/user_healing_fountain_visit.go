package models

import (
	"time"

	"github.com/google/uuid"
)

type UserHealingFountainVisit struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	UserID            uuid.UUID       `json:"userId" gorm:"not null"`
	User              User            `json:"user"`
	HealingFountainID uuid.UUID       `json:"healingFountainId" gorm:"column:healing_fountain_id;not null"`
	HealingFountain   HealingFountain `json:"healingFountain"`
	VisitedAt         time.Time       `json:"visitedAt" gorm:"column:visited_at;not null"`
}

func (UserHealingFountainVisit) TableName() string {
	return "user_healing_fountain_visits"
}
