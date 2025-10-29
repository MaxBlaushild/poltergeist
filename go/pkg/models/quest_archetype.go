package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestArchetype struct {
	ID          uuid.UUID          `json:"id"`
	Name        string             `json:"name"`
	DefaultGold int                `json:"defaultGold"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt     `json:"deletedAt"`
	Root        QuestArchetypeNode `json:"root"`
	RootID      uuid.UUID          `json:"rootId"`
}
