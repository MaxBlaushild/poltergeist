package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestArchetypeChallenge struct {
	ID                  uuid.UUID           `json:"id"`
	CreatedAt           time.Time           `json:"createdAt"`
	UpdatedAt           time.Time           `json:"updatedAt"`
	DeletedAt           gorm.DeletedAt      `json:"deletedAt"`
	ChallengeTemplateID *uuid.UUID          `json:"challengeTemplateId,omitempty" gorm:"column:challenge_template_id;type:uuid"`
	ChallengeTemplate   *ChallengeTemplate  `json:"challengeTemplate,omitempty" gorm:"foreignKey:ChallengeTemplateID"`
	Reward              int                 `json:"reward"`
	InventoryItemID     *int                `json:"inventoryItemId,omitempty"`
	Proficiency         *string             `json:"proficiency,omitempty"`
	Difficulty          int                 `json:"difficulty"`
	UnlockedNodeID      *uuid.UUID          `json:"unlockedNodeId"`
	UnlockedNode        *QuestArchetypeNode `json:"unlockedNode"`
}
