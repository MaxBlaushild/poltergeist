package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ExpositionTemplateGenerationStatusQueued     = "queued"
	ExpositionTemplateGenerationStatusInProgress = "in_progress"
	ExpositionTemplateGenerationStatusCompleted  = "completed"
	ExpositionTemplateGenerationStatusFailed     = "failed"
)

type ExpositionTemplateGenerationJob struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	ZoneKind     string    `json:"zoneKind" gorm:"column:zone_kind"`
	Status       string    `json:"status"`
	Count        int       `json:"count"`
	CreatedCount int       `json:"createdCount" gorm:"column:created_count"`
	ErrorMessage *string   `json:"errorMessage,omitempty"`
}

func (ExpositionTemplateGenerationJob) TableName() string {
	return "exposition_template_generation_jobs"
}
