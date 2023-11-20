package models

import (
	"time"

	"github.com/google/uuid"
)

type QuestionSet struct {
	ID        uuid.UUID  `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	Questions []Question `json:"questions"`
}
