package models

import (
	"time"

	"github.com/google/uuid"
)

type SonarSurveySubmissionAnswer struct {
	ID                      uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt               time.Time
	UpdatedAt               time.Time
	SonarSurveyID           uuid.UUID
	SonarSurveySubmissionID uuid.UUID
	SonarActivityID         uuid.UUID
	SonarSurvey             SonarSurvey
	SonarActivity           SonarActivity
	SonarSurveySubmission   SonarSurveySubmission
	Down                    bool
	Notes                   string
}
