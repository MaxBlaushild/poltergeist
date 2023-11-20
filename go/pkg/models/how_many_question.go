package models

import (
	"time"

	"github.com/google/uuid"
)

type HowManyQuestion struct {
	ID          uuid.UUID `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Text        string    `json:"text"`
	HowMany     int       `json:"howMany"`
	Explanation string    `json:"explanation"`
	Valid       bool      `json:"valid"`
	Done        bool      `json:"done"`
}
