package models

import (
	"time"

	"github.com/google/uuid"
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
