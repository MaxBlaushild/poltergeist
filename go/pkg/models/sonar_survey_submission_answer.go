package models

import (
	"time"

	"github.com/google/uuid"
)

type SonarSurveySubmissionAnswer struct {
	ID                      uuid.UUID             `gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt               time.Time             `json:"createdAt"`
	UpdatedAt               time.Time             `json:"updatedAt"`
	SonarSurveyID           uuid.UUID             `json:"surveyId"`
	SonarSurveySubmissionID uuid.UUID             `json:"submissionId"`
	SonarActivityID         uuid.UUID             `json:"activityId"`
	SonarSurvey             SonarSurvey           `json:"survey"`
	SonarActivity           SonarActivity         `json:"activity"`
	SonarSurveySubmission   SonarSurveySubmission `json:"submission"`
	Down                    bool                  `json:"down"`
	Notes                   string                `json:"notes"`
}
