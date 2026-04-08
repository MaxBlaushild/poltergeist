package models

import (
	"time"

	"github.com/google/uuid"
)

type ExpositionItemReward struct {
	ID              uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
	ExpositionID    uuid.UUID     `json:"expositionId"`
	InventoryItemID int           `json:"inventoryItemId"`
	InventoryItem   InventoryItem `json:"inventoryItem" gorm:"foreignKey:InventoryItemID"`
	Quantity        int           `json:"quantity"`
}

func (e *ExpositionItemReward) TableName() string {
	return "exposition_item_rewards"
}
