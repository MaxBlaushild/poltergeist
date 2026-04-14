package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterestItemReward struct {
	ID                uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
	PointOfInterestID uuid.UUID     `json:"pointOfInterestId"`
	InventoryItemID   int           `json:"inventoryItemId"`
	InventoryItem     InventoryItem `json:"inventoryItem" gorm:"foreignKey:InventoryItemID"`
	Quantity          int           `json:"quantity"`
}

func (p *PointOfInterestItemReward) TableName() string {
	return "point_of_interest_item_rewards"
}
