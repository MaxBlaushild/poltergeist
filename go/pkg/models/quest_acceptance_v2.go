package models

import (
  "time"

  "github.com/google/uuid"
)

type QuestAcceptanceV2 struct {
  ID         uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
  CreatedAt  time.Time  `json:"createdAt"`
  UpdatedAt  time.Time  `json:"updatedAt"`
  UserID     uuid.UUID  `json:"userId" gorm:"type:uuid"`
  QuestID    uuid.UUID  `json:"questId" gorm:"type:uuid"`
  AcceptedAt time.Time  `json:"acceptedAt"`
  TurnedInAt *time.Time `json:"turnedInAt"`
}

func (q *QuestAcceptanceV2) TableName() string {
  return "quest_acceptances_v2"
}
