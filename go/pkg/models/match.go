package models

import (
	"time"

	"github.com/google/uuid"
)

type Match struct {
	ID            uuid.UUID   `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt     time.Time   `db:"created_at"`
	UpdatedAt     time.Time   `db:"updated_at"`
	QuestionSetID uint        `json:"questionSetId"`
	QuestionSet   QuestionSet `json:"questionSet"`
	HomeID        uuid.UUID   `json:"homeId"`
	Home          User        `json:"home"`
	AwayID        uuid.UUID   `json:"awayId"`
	Away          User        `json:"away"`
	WinnerID      *uuid.UUID  `json:"winnerId"`
	Winner        User        `json:"winner"`
}
