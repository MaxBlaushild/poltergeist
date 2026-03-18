package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	BaseDescriptionGenerationStatusQueued     = "queued"
	BaseDescriptionGenerationStatusInProgress = "in_progress"
	BaseDescriptionGenerationStatusCompleted  = "completed"
	BaseDescriptionGenerationStatusFailed     = "failed"
)

type BaseDescriptionGenerationJob struct {
	ID                   uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
	BaseID               uuid.UUID `json:"baseId" gorm:"type:uuid"`
	Status               string    `json:"status"`
	GeneratedDescription *string   `json:"generatedDescription,omitempty" gorm:"column:generated_description"`
	GeneratedImageURL    *string   `json:"generatedImageUrl,omitempty" gorm:"column:generated_image_url"`
	ErrorMessage         *string   `json:"errorMessage,omitempty"`
}

func (BaseDescriptionGenerationJob) TableName() string {
	return "base_description_generation_jobs"
}
