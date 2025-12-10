package models

import (
	"time"

	"github.com/google/uuid"
)

type QuickDecisionRequest struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"userId"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	Question  string    `gorm:"type:text;not null" json:"question"`
	Option1   string    `gorm:"type:varchar;not null" json:"option1"`
	Option2   string    `gorm:"type:varchar;not null" json:"option2"`
	Option3   *string   `gorm:"type:varchar" json:"option3"`
}

func (QuickDecisionRequest) TableName() string {
	return "quick_decision_requests"
}
