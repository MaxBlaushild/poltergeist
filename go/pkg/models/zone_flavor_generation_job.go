package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ZoneFlavorGenerationStatusQueued     = "queued"
	ZoneFlavorGenerationStatusInProgress = "in_progress"
	ZoneFlavorGenerationStatusCompleted  = "completed"
	ZoneFlavorGenerationStatusFailed     = "failed"
)

type ZoneFlavorGenerationJob struct {
	ID                   uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
	ZoneID               uuid.UUID `json:"zoneId" gorm:"type:uuid"`
	Status               string    `json:"status"`
	GeneratedDescription *string   `json:"generatedDescription,omitempty" gorm:"column:generated_description"`
	ErrorMessage         *string   `json:"errorMessage,omitempty"`
}

func (ZoneFlavorGenerationJob) TableName() string {
	return "zone_flavor_generation_jobs"
}
