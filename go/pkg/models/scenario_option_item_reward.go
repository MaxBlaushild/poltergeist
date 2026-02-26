package models

import (
	"time"

	"github.com/google/uuid"
)

type ScenarioOptionItemReward struct {
	ID               uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt        time.Time     `json:"createdAt"`
	UpdatedAt        time.Time     `json:"updatedAt"`
	ScenarioOptionID uuid.UUID     `json:"scenarioOptionId" gorm:"column:scenario_option_id"`
	InventoryItemID  int           `json:"inventoryItemId"`
	InventoryItem    InventoryItem `json:"inventoryItem" gorm:"foreignKey:InventoryItemID"`
	Quantity         int           `json:"quantity"`
}

func (s *ScenarioOptionItemReward) TableName() string {
	return "scenario_option_item_rewards"
}
