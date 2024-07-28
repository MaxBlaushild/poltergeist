package models

import (
	"time"

	"github.com/google/uuid"
)

type Match struct {
	ID                uuid.UUID          `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time          `db:"created_at"`
	UpdatedAt         time.Time          `db:"updated_at"`
	CreatorID         uuid.UUID          `db:"creator_id"`
	Creator           User               `db:"creator" gorm:"foreignKey:CreatorID"`
	VerificationCodes []VerificationCode `gorm:"many2many:match_verification_codes;"`
	StartedAt         *time.Time         `db:"started_at"`
	EndedAt           *time.Time         `db:"ended_at"`
}
