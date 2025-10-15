package models

import (
	"time"

	"github.com/google/uuid"
)

type Party struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	LeaderID  uuid.UUID `json:"leaderId"`
	Leader    User      `gorm:"foreignKey:LeaderID" json:"leader"`
	Members   []User    `gorm:"foreignKey:PartyID" json:"members"`
}
