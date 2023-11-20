package models

import (
	"time"

	"github.com/google/uuid"
)

type UserSubmission struct {
	ID            uuid.UUID    `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt     time.Time    `db:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at"`
	User          User         `json:"user"`
	UserID        uuid.UUID    `json:"userId"`
	QuestionSet   QuestionSet  `json:"questionSet"`
	QuestionSetID uuid.UUID    `json:"questionSetId"`
	UserAnswers   []UserAnswer `json:"userAnswers"`
}
