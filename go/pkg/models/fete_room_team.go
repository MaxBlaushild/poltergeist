package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FeteRoomTeam struct {
	ID         uuid.UUID      `gorm:"primary_key" json:"id"`
	CreatedAt  time.Time      `gorm:"not null" json:"createdAt"`
	UpdatedAt  time.Time      `gorm:"not null" json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	FeteRoomID uuid.UUID      `gorm:"not null" json:"feteRoomId"`
	FeteRoom   FeteRoom       `gorm:"foreignKey:FeteRoomID"`
	TeamID     uuid.UUID      `gorm:"not null" json:"teamId"`
	Team       FeteTeam       `gorm:"foreignKey:TeamID"`
}
