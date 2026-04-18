package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ScenarioTemplateGenerationStatusQueued     = "queued"
	ScenarioTemplateGenerationStatusInProgress = "in_progress"
	ScenarioTemplateGenerationStatusCompleted  = "completed"
	ScenarioTemplateGenerationStatusFailed     = "failed"
)

type ScenarioTemplateGenerationJob struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	GenreID      uuid.UUID  `json:"genreId" gorm:"column:genre_id;type:uuid"`
	Genre        *ZoneGenre `json:"genre,omitempty" gorm:"foreignKey:GenreID"`
	Status       string     `json:"status"`
	Count        int        `json:"count"`
	OpenEnded    bool       `json:"openEnded" gorm:"column:open_ended"`
	CreatedCount int        `json:"createdCount" gorm:"column:created_count"`
	ErrorMessage *string    `json:"errorMessage,omitempty"`
}

func (ScenarioTemplateGenerationJob) TableName() string {
	return "scenario_template_generation_jobs"
}
