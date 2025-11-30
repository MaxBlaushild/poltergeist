package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FeteRoomLinkedListTeam struct {
	ID           uuid.UUID      `gorm:"primary_key" json:"id"`
	CreatedAt    time.Time      `gorm:"not null" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"not null" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	FeteRoomID   uuid.UUID      `gorm:"not null" json:"feteRoomId"`
	FeteRoom     FeteRoom       `gorm:"foreignKey:FeteRoomID"`
	FirstTeamID  uuid.UUID      `gorm:"not null" json:"firstTeamId"`
	FirstTeam    FeteTeam       `gorm:"foreignKey:FirstTeamID"`
	SecondTeamID uuid.UUID      `gorm:"not null" json:"secondTeamId"`
	SecondTeam   FeteTeam       `gorm:"foreignKey:SecondTeamID"`
}
