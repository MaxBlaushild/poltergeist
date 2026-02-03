package models

import (
  "time"

  "github.com/google/uuid"
)

type QuestNodeChild struct {
  ID                   uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
  CreatedAt            time.Time  `json:"createdAt"`
  UpdatedAt            time.Time  `json:"updatedAt"`
  QuestNodeID          uuid.UUID  `json:"questNodeId" gorm:"type:uuid"`
  NextQuestNodeID      uuid.UUID  `json:"nextQuestNodeId" gorm:"type:uuid"`
  QuestNodeChallengeID *uuid.UUID `json:"questNodeChallengeId" gorm:"type:uuid"`
}

func (q *QuestNodeChild) TableName() string {
  return "quest_node_children"
}
