package models

import (
  "time"

  "github.com/google/uuid"
)

type QuestNodeProgress struct {
  ID                uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
  CreatedAt         time.Time  `json:"createdAt"`
  UpdatedAt         time.Time  `json:"updatedAt"`
  QuestAcceptanceID uuid.UUID  `json:"questAcceptanceId" gorm:"type:uuid"`
  QuestNodeID       uuid.UUID  `json:"questNodeId" gorm:"type:uuid"`
  CompletedAt       *time.Time `json:"completedAt"`
}

func (q *QuestNodeProgress) TableName() string {
  return "quest_node_progress"
}
