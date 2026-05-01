package models

import (
	"time"

	"github.com/google/uuid"
)

type UserShrineUse struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserID    uuid.UUID `json:"userId" gorm:"not null"`
	User      User      `json:"user"`
	ShrineID  uuid.UUID `json:"shrineId" gorm:"column:shrine_id;not null"`
	Shrine    Shrine    `json:"shrine"`
	UsedAt    time.Time `json:"usedAt" gorm:"column:used_at;not null"`
}

func (UserShrineUse) TableName() string {
	return "user_shrine_uses"
}
