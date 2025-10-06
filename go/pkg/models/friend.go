package models

import (
	"time"

	"github.com/google/uuid"
)

type Friend struct {
	ID           uuid.UUID `gorm:"primaryKey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	FirstUserID  uuid.UUID
	SecondUserID uuid.UUID
	FirstUser    *User
	SecondUser   *User
}
