package models

import (
	"time"

	"github.com/google/uuid"
)

type TreasureChestItem struct {
	ID              uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
	TreasureChestID uuid.UUID     `json:"treasureChestId"`
	TreasureChest   TreasureChest `json:"treasureChest" gorm:"foreignKey:TreasureChestID"`
	InventoryItemID int           `json:"inventoryItemId"`
	InventoryItem   InventoryItem `json:"inventoryItem" gorm:"foreignKey:InventoryItemID"`
	Quantity        int           `json:"quantity"`
}

func (t *TreasureChestItem) TableName() string {
	return "treasure_chest_items"
}
