package models

import (
	"time"

	"github.com/google/uuid"
)

type UserAnswer struct {
	ID               uuid.UUID      `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
	User             User           `json:"user"`
	UserID           uuid.UUID      `json:"userId"`
	Question         Question       `json:"question"`
	QuestionID       uuid.UUID      `json:"questionId"`
	Answer           string         `json:"answer"`
	Correct          bool           `json:"correct"`
	UserSubmission   UserSubmission `json:"userSubmission"`
	UserSubmissionID uuid.UUID      `json:"userSubmissionId"`
	Points           uint           `json:"points"`
}
