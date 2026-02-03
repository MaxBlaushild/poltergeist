package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRecentPostTag struct {
	UserID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"userId"`
	Tag          string    `gorm:"type:text;primaryKey" json:"tag"`
	LastPostedAt time.Time `gorm:"not null" json:"lastPostedAt"`
}

func (UserRecentPostTag) TableName() string {
	return "user_recent_post_tags"
}
