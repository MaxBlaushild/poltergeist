package models

import (
	"time"

	"github.com/google/uuid"
)

type MatchVerificationCode struct {
	ID                 uuid.UUID        `json:"id" db:"id"`
	MatchID            uuid.UUID        `json:"matchId" db:"match_id"`
	CreatedAt          time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time        `json:"updatedAt" db:"updated_at"`
	VerificationCodeID uuid.UUID        `json:"verificationCodeId" db:"verification_code_id"`
	VerificationCode   VerificationCode `json:"verificationCode" db:"verification_code"`
}
