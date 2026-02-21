package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Party struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	LeaderID  uuid.UUID `json:"leaderId"`
	Leader    User      `gorm:"foreignKey:LeaderID" json:"leader"`
	Members   []User    `gorm:"foreignKey:PartyID" json:"members"`
}

func (p *Party) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
