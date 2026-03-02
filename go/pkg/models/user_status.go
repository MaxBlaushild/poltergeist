package models

import (
	"time"

	"github.com/google/uuid"
)

type UserStatusEffectType string

const (
	UserStatusEffectTypeStatModifier   UserStatusEffectType = "stat_modifier"
	UserStatusEffectTypeDamageOverTime UserStatusEffectType = "damage_over_time"
)

type UserStatus struct {
	ID              uuid.UUID            `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt       time.Time            `json:"createdAt"`
	UpdatedAt       time.Time            `json:"updatedAt"`
	UserID          uuid.UUID            `json:"userId"`
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	Effect          string               `json:"effect"`
	Positive        bool                 `json:"positive" gorm:"column:positive"`
	EffectType      UserStatusEffectType `json:"effectType" gorm:"column:effect_type"`
	DamagePerTick   int                  `json:"damagePerTick" gorm:"column:damage_per_tick"`
	StrengthMod     int                  `json:"strengthMod" gorm:"column:strength_mod"`
	DexterityMod    int                  `json:"dexterityMod" gorm:"column:dexterity_mod"`
	ConstitutionMod int                  `json:"constitutionMod" gorm:"column:constitution_mod"`
	IntelligenceMod int                  `json:"intelligenceMod" gorm:"column:intelligence_mod"`
	WisdomMod       int                  `json:"wisdomMod" gorm:"column:wisdom_mod"`
	CharismaMod     int                  `json:"charismaMod" gorm:"column:charisma_mod"`
	StartedAt       time.Time            `json:"startedAt" gorm:"column:started_at"`
	LastTickAt      *time.Time           `json:"lastTickAt,omitempty" gorm:"column:last_tick_at"`
	ExpiresAt       time.Time            `json:"expiresAt" gorm:"column:expires_at"`
}

func (u *UserStatus) TableName() string {
	return "user_statuses"
}

func (u UserStatus) StatModifiers() CharacterStatBonuses {
	return CharacterStatBonuses{
		Strength:     u.StrengthMod,
		Dexterity:    u.DexterityMod,
		Constitution: u.ConstitutionMod,
		Intelligence: u.IntelligenceMod,
		Wisdom:       u.WisdomMod,
		Charisma:     u.CharismaMod,
	}
}
