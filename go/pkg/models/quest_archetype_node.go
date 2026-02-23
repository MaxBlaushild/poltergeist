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
	Difficulty          int                       `json:"difficulty" gorm:"default:0"`
}

func (q *QuestArchetypeNode) GetRandomChallenge() (LocationArchetypeChallenge, error) {
	return q.LocationArchetype.GetRandomChallengeByDifficulty(q.Difficulty)
}
