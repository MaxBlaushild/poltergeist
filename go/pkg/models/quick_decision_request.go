package models

import (
	"time"

	"github.com/google/uuid"
)

type QuickDecisionRequest struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
	UserID    uuid.UUID `gorm:"type:uuid;column:user_id;not null" json:"userId"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	Question  string    `gorm:"type:text;not null" json:"question"`
	Option1   string    `gorm:"type:varchar;column:option_1;not null" json:"option1"`
	Option2   string    `gorm:"type:varchar;column:option_2;not null" json:"option2"`
	Option3   *string   `gorm:"type:varchar;column:option_3" json:"option3"`
}

func (QuickDecisionRequest) TableName() string {
	return "quick_decision_requests"
}
