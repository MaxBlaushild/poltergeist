package models

import (
	"time"

	"github.com/google/uuid"
)

type CrystalUnlocking struct {
	ID        uuid.UUID `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	TeamID    uuid.UUID
	Team      Team
	CrystalID uuid.UUID
	Crystal   Crystal
}
