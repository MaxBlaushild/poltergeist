package models

import (
	"time"

	"github.com/google/uuid"
)

type UserTreasureChestOpening struct {
	ID              uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
	UserID          uuid.UUID     `json:"userId" gorm:"not null"`
	User            User          `json:"user"`
	TreasureChestID uuid.UUID     `json:"treasureChestId" gorm:"not null"`
	TreasureChest   TreasureChest `json:"treasureChest"`
}

func (UserTreasureChestOpening) TableName() string {
	return "user_treasure_chest_openings"
}
