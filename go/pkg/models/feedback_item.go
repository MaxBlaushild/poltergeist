package models

import (
	"time"

	"github.com/google/uuid"
)

type FeedbackItem struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time  `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"not null" json:"updatedAt"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"userId"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ZoneID    *uuid.UUID `gorm:"type:uuid;index" json:"zoneId,omitempty"`
	Zone      *Zone      `gorm:"foreignKey:ZoneID" json:"zone,omitempty"`
	Route     string     `gorm:"type:text;not null;default:''" json:"route"`
	Message   string     `gorm:"type:text;not null" json:"message"`
}
