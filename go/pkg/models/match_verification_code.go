package models

import (
	"time"

	"github.com/google/uuid"
)

type MatchVerificationCode struct {
	ID                 uuid.UUID        `db:"id"`
	MatchID            uuid.UUID        `db:"match_id"`
	CreatedAt          time.Time        `db:"created_at"`
	UpdatedAt          time.Time        `db:"updated_at"`
	VerificationCodeID uuid.UUID        `db:"verification_code_id"`
	VerificationCode   VerificationCode `db:"verification_code"`
}
