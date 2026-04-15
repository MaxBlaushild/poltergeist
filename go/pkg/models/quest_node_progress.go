package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestNodeProgress struct {
	ID                uuid.UUID               `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time               `json:"createdAt"`
	UpdatedAt         time.Time               `json:"updatedAt"`
	QuestAcceptanceID uuid.UUID               `json:"questAcceptanceId" gorm:"type:uuid"`
	QuestNodeID       uuid.UUID               `json:"questNodeId" gorm:"type:uuid"`
	Status            QuestNodeProgressStatus `json:"status" gorm:"type:text;default:'active'"`
	AttemptCount      int                     `json:"attemptCount" gorm:"column:attempt_count;default:0"`
	LastFailedAt      *time.Time              `json:"lastFailedAt"`
	LastFailureReason string                  `json:"lastFailureReason" gorm:"column:last_failure_reason"`
	CompletedAt       *time.Time              `json:"completedAt"`
}

func (q *QuestNodeProgress) TableName() string {
	return "quest_node_progress"
}

func (q *QuestNodeProgress) BeforeSave(tx *gorm.DB) error {
	q.Status = NormalizeQuestNodeProgressStatus(string(q.Status))
	q.LastFailureReason = strings.TrimSpace(q.LastFailureReason)
	if q.Status == QuestNodeProgressStatusCompleted && q.CompletedAt == nil {
		now := time.Now()
		q.CompletedAt = &now
	}
	return nil
}
