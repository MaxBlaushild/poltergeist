package models

import (
	"time"

	"github.com/google/uuid"
)

type Crystal struct {
	ID               uuid.UUID `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
	Name             string    `json:"name"`
	Clue             string    `json:"clue"`
	CaptureChallenge string    `json:"captureChallenge"`
	AttuneChallenge  string    `json:"attuneChallenge"`
	Captured         bool      `json:"captured"`
	Attuned          bool      `json:"attuned"`
	Lat              string    `json:"lat"`
	Lng              string    `json:"lng"`
	CaptureTeamID    uuid.UUID `json:"captureTeamId"`
}
