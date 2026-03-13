package models

import (
	"time"

	"github.com/google/uuid"
)

type MonsterStatusEffectType string

const (
	MonsterStatusEffectTypeStatModifier   MonsterStatusEffectType = "stat_modifier"
	MonsterStatusEffectTypeDamageOverTime MonsterStatusEffectType = "damage_over_time"
	MonsterStatusEffectTypeHealthOverTime MonsterStatusEffectType = "health_over_time"
)

type MonsterStatus struct {
	ID              uuid.UUID               `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt       time.Time               `json:"createdAt"`
	UpdatedAt       time.Time               `json:"updatedAt"`
	UserID          uuid.UUID               `json:"userId" gorm:"column:user_id"`
	BattleID        uuid.UUID               `json:"battleId" gorm:"column:battle_id"`
	MonsterID       uuid.UUID               `json:"monsterId" gorm:"column:monster_id"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	Effect          string                  `json:"effect"`
	Positive        bool                    `json:"positive" gorm:"column:positive"`
	EffectType      MonsterStatusEffectType `json:"effectType" gorm:"column:effect_type"`
	DamagePerTick   int                     `json:"damagePerTick" gorm:"column:damage_per_tick"`
	HealthPerTick   int                     `json:"healthPerTick" gorm:"column:health_per_tick"`
	StrengthMod     int                     `json:"strengthMod" gorm:"column:strength_mod"`
	DexterityMod    int                     `json:"dexterityMod" gorm:"column:dexterity_mod"`
	ConstitutionMod int                     `json:"constitutionMod" gorm:"column:constitution_mod"`
	IntelligenceMod int                     `json:"intelligenceMod" gorm:"column:intelligence_mod"`
	WisdomMod       int                     `json:"wisdomMod" gorm:"column:wisdom_mod"`
	CharismaMod     int                     `json:"charismaMod" gorm:"column:charisma_mod"`
	StartedAt       time.Time               `json:"startedAt" gorm:"column:started_at"`
	LastTickAt      *time.Time              `json:"lastTickAt,omitempty" gorm:"column:last_tick_at"`
	ExpiresAt       time.Time               `json:"expiresAt" gorm:"column:expires_at"`
}

func (m *MonsterStatus) TableName() string {
	return "monster_statuses"
}

func (m MonsterStatus) StatModifiers() CharacterStatBonuses {
	return CharacterStatBonuses{
		Strength:     m.StrengthMod,
		Dexterity:    m.DexterityMod,
		Constitution: m.ConstitutionMod,
		Intelligence: m.IntelligenceMod,
		Wisdom:       m.WisdomMod,
		Charisma:     m.CharismaMod,
	}
}
