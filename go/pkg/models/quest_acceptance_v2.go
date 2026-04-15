package models

import (
	"time"

	"github.com/google/uuid"
)

type QuestAcceptanceV2 struct {
	ID                    uuid.UUID          `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt             time.Time          `json:"createdAt"`
	UpdatedAt             time.Time          `json:"updatedAt"`
	UserID                uuid.UUID          `json:"userId" gorm:"type:uuid"`
	QuestID               uuid.UUID          `json:"questId" gorm:"type:uuid"`
	CurrentQuestNodeID    *uuid.UUID         `json:"currentQuestNodeId" gorm:"column:current_quest_node_id;type:uuid"`
	AcceptedAt            time.Time          `json:"acceptedAt"`
	ObjectivesCompletedAt *time.Time         `json:"objectivesCompletedAt" gorm:"column:objectives_completed_at"`
	ClosedAt              *time.Time         `json:"closedAt" gorm:"column:closed_at"`
	ClosureMethod         QuestClosureMethod `json:"closureMethod" gorm:"column:closure_method;default:'in_person'"`
	DebriefPending        bool               `json:"debriefPending" gorm:"column:debrief_pending;default:false"`
	DebriefedAt           *time.Time         `json:"debriefedAt" gorm:"column:debriefed_at"`
	TurnedInAt            *time.Time         `json:"turnedInAt"`
}

func (q *QuestAcceptanceV2) TableName() string {
	return "quest_acceptances_v2"
}
