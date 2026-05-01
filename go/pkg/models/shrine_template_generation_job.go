package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ShrineTemplateGenerationStatusQueued     = "queued"
	ShrineTemplateGenerationStatusInProgress = "in_progress"
	ShrineTemplateGenerationStatusCompleted  = "completed"
	ShrineTemplateGenerationStatusFailed     = "failed"
)

type ShrineTemplateGenerationJob struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	ZoneKind     string    `json:"zoneKind,omitempty" gorm:"column:zone_kind"`
	Status       string    `json:"status"`
	Count        int       `json:"count"`
	CreatedCount int       `json:"createdCount" gorm:"column:created_count"`
	ErrorMessage *string   `json:"errorMessage,omitempty"`
}

func (ShrineTemplateGenerationJob) TableName() string {
	return "shrine_template_generation_jobs"
}
