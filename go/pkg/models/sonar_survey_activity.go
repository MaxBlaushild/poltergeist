package models

import (
	"time"

	"github.com/google/uuid"
)

type SonarSurveyActivity struct {
	ID              uuid.UUID     `json:"id"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
	SonarSurveyID   uuid.UUID     `json:"surveyId"`
	SonarActivityID uuid.UUID     `json:"activityId"`
	SonarSurvey     SonarSurvey   `json:"survey"`
	SonarActivity   SonarActivity `json:"activity"`
}
