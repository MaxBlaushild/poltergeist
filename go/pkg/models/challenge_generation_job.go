package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	ChallengeGenerationStatusQueued     = "queued"
	ChallengeGenerationStatusInProgress = "in_progress"
	ChallengeGenerationStatusCompleted  = "completed"
	ChallengeGenerationStatusFailed     = "failed"
)

type ChallengeGenerationJob struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	ZoneID            uuid.UUID  `json:"zoneId" gorm:"type:uuid"`
	PointOfInterestID *uuid.UUID `json:"pointOfInterestId,omitempty" gorm:"column:point_of_interest_id;type:uuid"`
	Status            string     `json:"status"`
	Count             int        `json:"count"`
	CreatedCount      int        `json:"createdCount" gorm:"column:created_count"`
	ErrorMessage      *string    `json:"errorMessage,omitempty"`
}

func (ChallengeGenerationJob) TableName() string {
	return "challenge_generation_jobs"
}
