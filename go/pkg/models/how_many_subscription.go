package models

import (
	"time"

	"github.com/google/uuid"
)

type HowManySubscription struct {
	ID               uuid.UUID `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
	User             User      `json:"user"`
	UserID           uuid.UUID `json:"userId"`
	Subscribed       bool      `gorm:"default:false" json:"subscribed"`
	NumFreeQuestions uint      `gorm:"default:0" json:"numFreeQuestions"`
}
