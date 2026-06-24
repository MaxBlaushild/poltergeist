package models

import (
	"time"

	"github.com/google/uuid"
)

type TradesARGlassesLead struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
	Email     string    `gorm:"type:text;uniqueIndex;not null" json:"email"`
	Trade     string    `gorm:"type:text;not null;default:''" json:"trade"`
	CrewSize  string    `gorm:"type:text;not null;default:''" json:"crewSize"`
	Source    string    `gorm:"type:text;not null;default:'landing-page'" json:"source"`
	UserAgent string    `gorm:"type:text;not null;default:''" json:"userAgent"`
}

func (TradesARGlassesLead) TableName() string {
	return "trades_ar_glasses_leads"
}
