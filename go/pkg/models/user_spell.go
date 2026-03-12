package models

import (
	"time"

	"github.com/google/uuid"
)

type UserSpell struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	UserID            uuid.UUID  `json:"userId"`
	SpellID           uuid.UUID  `json:"spellId"`
	Spell             Spell      `json:"spell" gorm:"foreignKey:SpellID"`
	AcquiredAt        time.Time  `json:"acquiredAt" gorm:"column:acquired_at"`
	CooldownExpiresAt *time.Time `json:"cooldownExpiresAt,omitempty" gorm:"column:cooldown_expires_at"`
}

func (u *UserSpell) TableName() string {
	return "user_spells"
}
