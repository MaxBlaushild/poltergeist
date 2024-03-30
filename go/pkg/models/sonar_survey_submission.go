package models

import (
	"time"

	"github.com/google/uuid"
)

type SonarSurveySubmission struct {
	ID                           uuid.UUID                     `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt                    time.Time                     `db:"created_at" json:"createdAt"`
	UpdatedAt                    time.Time                     `db:"updated_at" json:"updatedAt"`
	UserID                       uuid.UUID                     `json:"userId"`
	User                         User                          `json:"user"`
	SonarSurveyID                uuid.UUID                     `json:"surveyId"`
	SonarSurvey                  SonarSurvey                   `json:"survey"`
	SonarSurveySubmissionAnswers []SonarSurveySubmissionAnswer `json:"answers"`
}
