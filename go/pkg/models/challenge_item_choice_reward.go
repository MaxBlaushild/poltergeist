package models

import (
	"time"

	"github.com/google/uuid"
)

type ChallengeItemChoiceReward struct {
	ID              uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
	ChallengeID     uuid.UUID     `json:"challengeId" gorm:"column:challenge_id"`
	InventoryItemID int           `json:"inventoryItemId" gorm:"column:inventory_item_id"`
	InventoryItem   InventoryItem `json:"inventoryItem" gorm:"foreignKey:InventoryItemID"`
	Quantity        int           `json:"quantity"`
}

func (c *ChallengeItemChoiceReward) TableName() string {
	return "challenge_item_choice_rewards"
}
