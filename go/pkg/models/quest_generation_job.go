package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	QuestGenerationStatusQueued     = "queued"
	QuestGenerationStatusInProgress = "in_progress"
	QuestGenerationStatusCompleted  = "completed"
	QuestGenerationStatusFailed     = "failed"
)

type QuestGenerationJob struct {
	ID                    uuid.UUID   `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt             time.Time   `json:"createdAt"`
	UpdatedAt             time.Time   `json:"updatedAt"`
	ZoneQuestArchetypeID  uuid.UUID   `json:"zoneQuestArchetypeId" gorm:"type:uuid"`
	ZoneID                uuid.UUID   `json:"zoneId" gorm:"type:uuid"`
	QuestArchetypeID      uuid.UUID   `json:"questArchetypeId" gorm:"type:uuid"`
	QuestGiverCharacterID *uuid.UUID  `json:"questGiverCharacterId,omitempty" gorm:"type:uuid"`
	Status                string      `json:"status"`
	TotalCount            int         `json:"totalCount"`
	CompletedCount        int         `json:"completedCount"`
	FailedCount           int         `json:"failedCount"`
	ErrorMessage          *string     `json:"errorMessage"`
	QuestIDs              StringArray `json:"questIds" gorm:"type:jsonb"`
	Quests                []Quest     `json:"quests,omitempty" gorm:"-"`
}

func (QuestGenerationJob) TableName() string {
	return "quest_generation_jobs"
}
