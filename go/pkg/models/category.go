package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID  `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	Title     string     `gorm:"unique" json:"title"`
	Questions []Question `json:"questions"`
}
