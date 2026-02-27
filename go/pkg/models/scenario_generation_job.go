package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ScenarioGenerationStatusQueued     = "queued"
	ScenarioGenerationStatusInProgress = "in_progress"
	ScenarioGenerationStatusCompleted  = "completed"
	ScenarioGenerationStatusFailed     = "failed"
)

type ScenarioGenerationJob struct {
	ID                  uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
	ZoneID              uuid.UUID  `json:"zoneId" gorm:"type:uuid"`
	Status              string     `json:"status"`
	OpenEnded           bool       `json:"openEnded" gorm:"column:open_ended"`
	Latitude            *float64   `json:"latitude,omitempty"`
	Longitude           *float64   `json:"longitude,omitempty"`
	GeneratedScenarioID *uuid.UUID `json:"generatedScenarioId,omitempty" gorm:"column:generated_scenario_id;type:uuid"`
	ErrorMessage        *string    `json:"errorMessage,omitempty"`
}

func (ScenarioGenerationJob) TableName() string {
	return "scenario_generation_jobs"
}
