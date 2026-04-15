package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestNodeChild struct {
	ID              uuid.UUID                  `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt       time.Time                  `json:"createdAt"`
	UpdatedAt       time.Time                  `json:"updatedAt"`
	QuestNodeID     uuid.UUID                  `json:"questNodeId" gorm:"type:uuid"`
	NextQuestNodeID uuid.UUID                  `json:"nextQuestNodeId" gorm:"type:uuid"`
	Outcome         QuestNodeTransitionOutcome `json:"outcome" gorm:"type:text;default:'success'"`
}

func (q *QuestNodeChild) TableName() string {
	return "quest_node_children"
}

func (q *QuestNodeChild) BeforeSave(tx *gorm.DB) error {
	q.Outcome = NormalizeQuestNodeTransitionOutcome(string(q.Outcome))
	return nil
}

func (q *QuestNodeChild) TransitionOutcome() QuestNodeTransitionOutcome {
	if q == nil {
		return QuestNodeTransitionOutcomeSuccess
	}
	return NormalizeQuestNodeTransitionOutcome(string(q.Outcome))
}
