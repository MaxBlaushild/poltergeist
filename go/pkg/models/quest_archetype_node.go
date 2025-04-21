package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestArchetypeNode struct {
	ID                  uuid.UUID                 `json:"id"`
	CreatedAt           time.Time                 `json:"createdAt"`
	UpdatedAt           time.Time                 `json:"updatedAt"`
	DeletedAt           gorm.DeletedAt            `json:"deletedAt"`
	LocationArchetype   LocationArchetype         `json:"locationArchetype"`
	LocationArchetypeID uuid.UUID                 `json:"locationArchetypeId"`
	Challenges          []QuestArchetypeChallenge `json:"challenges" gorm:"many2many:quest_archetype_node_challenges;"`
}

func (q *QuestArchetypeNode) GetRandomChallenge() (string, error) {
	return q.LocationArchetype.GetRandomChallenge()
}
