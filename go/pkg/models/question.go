package models

import (
	"time"

	"github.com/google/uuid"
)

type Question struct {
	ID            uuid.UUID   `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt     time.Time   `db:"created_at"`
	UpdatedAt     time.Time   `db:"updated_at"`
	CategoryID    uuid.UUID   `json:"categoryId"`
	Category      Category    `json:"category"`
	QuestionSetID uuid.UUID   `json:"questionSetId"`
	QuestionSet   QuestionSet `json:"questionSet"`
	Prompt        string      `json:"prompt"`
	Answer        string      `json:"answer"`
}
