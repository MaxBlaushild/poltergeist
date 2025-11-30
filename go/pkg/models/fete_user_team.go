package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FeteUserTeam struct {
	ID        uuid.UUID      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time      `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"not null" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	UserID    uuid.UUID      `gorm:"not null" json:"userId"`
	User      User           `gorm:"foreignKey:UserID"`
	TeamID    uuid.UUID      `gorm:"not null" json:"teamId"`
	Team      Team           `gorm:"foreignKey:TeamID"`
}
