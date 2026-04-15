package models

import (
	"time"

	"github.com/google/uuid"
)

type UserResourceGathering struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	UserID     uuid.UUID `json:"userId" gorm:"column:user_id;not null"`
	User       User      `json:"user"`
	ResourceID uuid.UUID `json:"resourceId" gorm:"column:resource_id;not null"`
	Resource   Resource  `json:"resource"`
	GatheredAt time.Time `json:"gatheredAt" gorm:"column:gathered_at;not null"`
}

func (UserResourceGathering) TableName() string {
	return "user_resource_gatherings"
}
