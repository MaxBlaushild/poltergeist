package models

import (
	"time"

	"github.com/google/uuid"
)

type QuestArchetypeItemReward struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	QuestArchetypeID uuid.UUID      `json:"questArchetypeId"`
	QuestArchetype   QuestArchetype `json:"questArchetype" gorm:"foreignKey:QuestArchetypeID"`
	InventoryItemID  int            `json:"inventoryItemId"`
	InventoryItem    InventoryItem  `json:"inventoryItem" gorm:"foreignKey:InventoryItemID"`
	Quantity         int            `json:"quantity"`
}

func (q *QuestArchetypeItemReward) TableName() string {
	return "quest_archetype_item_rewards"
}
