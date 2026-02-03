package models

import (
  "time"

  "github.com/google/uuid"
)

type QuestNodeChallenge struct {
  ID            uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
  CreatedAt     time.Time `json:"createdAt"`
  UpdatedAt     time.Time `json:"updatedAt"`
  QuestNodeID   uuid.UUID `json:"questNodeId" gorm:"type:uuid"`
  Tier          int       `json:"tier"`
  Question      string    `json:"question"`
  Reward        int       `json:"reward"`
  InventoryItemID *int    `json:"inventoryItemId"`
}

func (q *QuestNodeChallenge) TableName() string {
  return "quest_node_challenges"
}
