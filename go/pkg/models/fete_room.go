package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FeteRoom struct {
	ID            uuid.UUID      `gorm:"primary_key" json:"id"`
	CreatedAt     time.Time      `gorm:"not null" json:"createdAt"`
	UpdatedAt     time.Time      `gorm:"not null" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Name          string         `gorm:"not null" json:"name"`
	Open          bool           `gorm:"not null" json:"open"`
	CurrentTeamID uuid.UUID      `gorm:"not null" json:"currentTeamId"`
	HueLightID    *int           `gorm:"column:hue_light_id" json:"hueLightId,omitempty"`
}
