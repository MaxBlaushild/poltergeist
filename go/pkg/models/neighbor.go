package models

import (
	"time"

	"github.com/google/uuid"
)

type Neighbor struct {
	ID           uuid.UUID `db:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	CrystalOneID uuid.UUID `json:"crystalOneId" binding:"required"`
	CrystalOne   Crystal   `json:"crystalOne"`
	CrystalTwoID uuid.UUID `json:"crystalTwoId" binding:"required"`
	CrystalTwo   Crystal   `json:"crystalTwo"`
}
