package models

import (
	"time"

	"github.com/google/uuid"
)

type SonarSurvey struct {
	ID                     uuid.UUID               `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt              time.Time               `db:"created_at" json:"createdAt"`
	UpdatedAt              time.Time               `db:"updated_at" json:"updatedAt"`
	Title                  string                  `gorm:"unique" json:"title"`
	UserID                 uuid.UUID               `json:"userId"`
	User                   User                    `json:"user"`
	SonarActivities        []SonarActivity         `gorm:"many2many:sonar_survey_activities;" json:"activities"`
	SonarSurveySubmissions []SonarSurveySubmission `json:"surveySubmissions"`
}
