package models

import (
	"time"

	"github.com/google/uuid"
)

type ScenarioItemChoiceReward struct {
	ID              uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
	ScenarioID      uuid.UUID     `json:"scenarioId" gorm:"column:scenario_id"`
	InventoryItemID int           `json:"inventoryItemId" gorm:"column:inventory_item_id"`
	InventoryItem   InventoryItem `json:"inventoryItem" gorm:"foreignKey:InventoryItemID"`
	Quantity        int           `json:"quantity"`
}

func (s *ScenarioItemChoiceReward) TableName() string {
	return "scenario_item_choice_rewards"
}
