package models

import (
	"time"

	"github.com/google/uuid"
)

type QuestItemReward struct {
	ID              uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
	QuestID         uuid.UUID     `json:"questId"`
	Quest           Quest         `json:"quest" gorm:"foreignKey:QuestID"`
	InventoryItemID int           `json:"inventoryItemId"`
	InventoryItem   InventoryItem `json:"inventoryItem" gorm:"foreignKey:InventoryItemID"`
	Quantity        int           `json:"quantity"`
}

func (q *QuestItemReward) TableName() string {
	return "quest_item_rewards"
}
