package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ChallengeTemplateGenerationStatusQueued     = "queued"
	ChallengeTemplateGenerationStatusInProgress = "in_progress"
	ChallengeTemplateGenerationStatusCompleted  = "completed"
	ChallengeTemplateGenerationStatusFailed     = "failed"
)

type ChallengeTemplateGenerationJob struct {
	ID                  uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
	LocationArchetypeID uuid.UUID `json:"locationArchetypeId" gorm:"column:location_archetype_id;type:uuid"`
	Status              string    `json:"status"`
	Count               int       `json:"count"`
	CreatedCount        int       `json:"createdCount" gorm:"column:created_count"`
	ErrorMessage        *string   `json:"errorMessage,omitempty"`
}

func (ChallengeTemplateGenerationJob) TableName() string {
	return "challenge_template_generation_jobs"
}
