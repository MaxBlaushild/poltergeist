package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestArchetypeChallenge struct {
	ID             uuid.UUID           `json:"id"`
	CreatedAt      time.Time           `json:"createdAt"`
	UpdatedAt      time.Time           `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt      `json:"deletedAt"`
	Reward         *uuid.UUID          `json:"reward"`
	UnlockedNodeID *uuid.UUID          `json:"unlockedNodeId"`
	UnlockedNode   *QuestArchetypeNode `json:"unlockedNode"`
}
