package models

import (
	"time"

	"github.com/google/uuid"
)

type SonarSurvey struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	Title        string    `gorm:"unique" json:"title"`
	ReferrerID   uuid.UUID
	Referrer     User
	ProgenitorID uuid.UUID
}
