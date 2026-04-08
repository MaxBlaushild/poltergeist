package models

import (
	"time"

	"github.com/google/uuid"
)

type UserExpositionCompletion struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	UserID       uuid.UUID `json:"userId" gorm:"column:user_id"`
	ExpositionID uuid.UUID `json:"expositionId" gorm:"column:exposition_id"`
}

func (u *UserExpositionCompletion) TableName() string {
	return "user_exposition_completions"
}
