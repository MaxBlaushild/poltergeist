package models

import (
	"time"

	"github.com/google/uuid"
)

type Party struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
