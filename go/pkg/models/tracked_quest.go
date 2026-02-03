package models

import (
  "time"

  "github.com/google/uuid"
  "gorm.io/gorm"
)

type TrackedQuest struct {
  ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
  CreatedAt time.Time      `json:"createdAt"`
  UpdatedAt time.Time      `json:"updatedAt"`
  DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`
  QuestID   uuid.UUID      `json:"questId" gorm:"type:uuid"`
  Quest     Quest          `json:"quest" gorm:"foreignKey:QuestID"`
  UserID    uuid.UUID      `json:"userId" gorm:"type:uuid"`
  User      User           `json:"user" gorm:"foreignKey:UserID"`
}

func (TrackedQuest) TableName() string {
  return "tracked_quests"
}
