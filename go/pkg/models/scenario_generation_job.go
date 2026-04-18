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
	GenreID             uuid.UUID  `json:"genreId" gorm:"column:genre_id;type:uuid"`
	Genre               *ZoneGenre `json:"genre,omitempty" gorm:"foreignKey:GenreID"`
	Status              string     `json:"status"`
	OpenEnded           bool       `json:"openEnded" gorm:"column:open_ended"`
	ScaleWithUserLevel  bool       `json:"scaleWithUserLevel" gorm:"column:scale_with_user_level"`
	Latitude            *float64   `json:"latitude,omitempty"`
	Longitude           *float64   `json:"longitude,omitempty"`
	RecurringScenarioID *uuid.UUID `json:"recurringScenarioId,omitempty" gorm:"column:recurring_scenario_id;type:uuid"`
	RecurrenceFrequency *string    `json:"recurrenceFrequency,omitempty" gorm:"column:recurrence_frequency"`
	NextRecurrenceAt    *time.Time `json:"nextRecurrenceAt,omitempty" gorm:"column:next_recurrence_at"`
	GeneratedScenarioID *uuid.UUID `json:"generatedScenarioId,omitempty" gorm:"column:generated_scenario_id;type:uuid"`
	ErrorMessage        *string    `json:"errorMessage,omitempty"`
}

func (ScenarioGenerationJob) TableName() string {
	return "scenario_generation_jobs"
}
