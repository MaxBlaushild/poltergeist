package models

import (
	"time"

	"github.com/google/uuid"
)

type ScenarioItemReward struct {
	ID              uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
	ScenarioID      uuid.UUID     `json:"scenarioId"`
	InventoryItemID int           `json:"inventoryItemId"`
	InventoryItem   InventoryItem `json:"inventoryItem" gorm:"foreignKey:InventoryItemID"`
	Quantity        int           `json:"quantity"`
}

func (s *ScenarioItemReward) TableName() string {
	return "scenario_item_rewards"
}
