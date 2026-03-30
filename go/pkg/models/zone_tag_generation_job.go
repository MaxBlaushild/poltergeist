package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ZoneTagGenerationStatusQueued     = "queued"
	ZoneTagGenerationStatusInProgress = "in_progress"
	ZoneTagGenerationStatusCompleted  = "completed"
	ZoneTagGenerationStatusFailed     = "failed"
)

type ZoneTagGenerationJob struct {
	ID               uuid.UUID   `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt        time.Time   `json:"createdAt"`
	UpdatedAt        time.Time   `json:"updatedAt"`
	ZoneID           uuid.UUID   `json:"zoneId" gorm:"type:uuid"`
	Status           string      `json:"status"`
	ContextSnapshot  string      `json:"contextSnapshot" gorm:"column:context_snapshot"`
	GeneratedSummary *string     `json:"generatedSummary,omitempty" gorm:"column:generated_summary"`
	SelectedTags     StringArray `json:"selectedTags" gorm:"column:selected_tags;type:jsonb"`
	ErrorMessage     *string     `json:"errorMessage,omitempty"`
}

func (ZoneTagGenerationJob) TableName() string {
	return "zone_tag_generation_jobs"
}
