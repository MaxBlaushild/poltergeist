package models

import (
	"time"

	"github.com/google/uuid"
)

type UserProficiency struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	UserID      uuid.UUID `json:"userId"`
	Proficiency string    `json:"proficiency"`
	Level       int       `json:"level"`
}

func (u *UserProficiency) TableName() string {
	return "user_proficiencies"
}
