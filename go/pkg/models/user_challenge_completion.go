package models

import (
	"time"

	"github.com/google/uuid"
)

type UserChallengeCompletion struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	UserID      uuid.UUID `json:"userId" gorm:"column:user_id"`
	ChallengeID uuid.UUID `json:"challengeId" gorm:"column:challenge_id"`
}

func (u *UserChallengeCompletion) TableName() string {
	return "user_challenge_completions"
}
