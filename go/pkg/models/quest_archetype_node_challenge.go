package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestArchetypeNodeChallenge struct {
	ID                        uuid.UUID               `json:"id"`
	CreatedAt                 time.Time               `json:"createdAt"`
	UpdatedAt                 time.Time               `json:"updatedAt"`
	QuestArchetypeChallengeID uuid.UUID               `json:"questArchtypeChallengeId"`
	QuestArchetypeChallenge   QuestArchetypeChallenge `json:"questArchtypeChallenge"`
	QuestArchetypeNodeID      uuid.UUID               `json:"questArchtypeNodeId"`
	QuestArchetypeNode        QuestArchetypeNode      `json:"questArchtypeNode"`
}

func (q *QuestArchetypeNodeChallenge) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	if q.CreatedAt.IsZero() {
		q.CreatedAt = now
	}
	q.UpdatedAt = now
	return nil
}
