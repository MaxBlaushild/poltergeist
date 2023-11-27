package models

import (
	"time"

	"github.com/google/uuid"
)

type SonarSurveySubmission struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
	Title         string    `gorm:"unique" json:"title"`
	UserID        uuid.UUID
	User          User
	SonarSurveyID uuid.UUID
	SonarSurvey   SonarSurvey
}
