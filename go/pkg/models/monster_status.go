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
	ID                            uuid.UUID               `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt                     time.Time               `json:"createdAt"`
	UpdatedAt                     time.Time               `json:"updatedAt"`
	UserID                        uuid.UUID               `json:"userId" gorm:"column:user_id"`
	BattleID                      uuid.UUID               `json:"battleId" gorm:"column:battle_id"`
	MonsterID                     uuid.UUID               `json:"monsterId" gorm:"column:monster_id"`
	Name                          string                  `json:"name"`
	Description                   string                  `json:"description"`
	Effect                        string                  `json:"effect"`
	Positive                      bool                    `json:"positive" gorm:"column:positive"`
	EffectType                    MonsterStatusEffectType `json:"effectType" gorm:"column:effect_type"`
	DamagePerTick                 int                     `json:"damagePerTick" gorm:"column:damage_per_tick"`
	HealthPerTick                 int                     `json:"healthPerTick" gorm:"column:health_per_tick"`
	StrengthMod                   int                     `json:"strengthMod" gorm:"column:strength_mod"`
	DexterityMod                  int                     `json:"dexterityMod" gorm:"column:dexterity_mod"`
	ConstitutionMod               int                     `json:"constitutionMod" gorm:"column:constitution_mod"`
	IntelligenceMod               int                     `json:"intelligenceMod" gorm:"column:intelligence_mod"`
	WisdomMod                     int                     `json:"wisdomMod" gorm:"column:wisdom_mod"`
	CharismaMod                   int                     `json:"charismaMod" gorm:"column:charisma_mod"`
	PhysicalDamageBonusPercent    int                     `json:"physicalDamageBonusPercent" gorm:"column:physical_damage_bonus_percent"`
	PiercingDamageBonusPercent    int                     `json:"piercingDamageBonusPercent" gorm:"column:piercing_damage_bonus_percent"`
	SlashingDamageBonusPercent    int                     `json:"slashingDamageBonusPercent" gorm:"column:slashing_damage_bonus_percent"`
	BludgeoningDamageBonusPercent int                     `json:"bludgeoningDamageBonusPercent" gorm:"column:bludgeoning_damage_bonus_percent"`
	FireDamageBonusPercent        int                     `json:"fireDamageBonusPercent" gorm:"column:fire_damage_bonus_percent"`
	IceDamageBonusPercent         int                     `json:"iceDamageBonusPercent" gorm:"column:ice_damage_bonus_percent"`
	LightningDamageBonusPercent   int                     `json:"lightningDamageBonusPercent" gorm:"column:lightning_damage_bonus_percent"`
	PoisonDamageBonusPercent      int                     `json:"poisonDamageBonusPercent" gorm:"column:poison_damage_bonus_percent"`
	ArcaneDamageBonusPercent      int                     `json:"arcaneDamageBonusPercent" gorm:"column:arcane_damage_bonus_percent"`
	HolyDamageBonusPercent        int                     `json:"holyDamageBonusPercent" gorm:"column:holy_damage_bonus_percent"`
	ShadowDamageBonusPercent      int                     `json:"shadowDamageBonusPercent" gorm:"column:shadow_damage_bonus_percent"`
	PhysicalResistancePercent     int                     `json:"physicalResistancePercent" gorm:"column:physical_resistance_percent"`
	PiercingResistancePercent     int                     `json:"piercingResistancePercent" gorm:"column:piercing_resistance_percent"`
	SlashingResistancePercent     int                     `json:"slashingResistancePercent" gorm:"column:slashing_resistance_percent"`
	BludgeoningResistancePercent  int                     `json:"bludgeoningResistancePercent" gorm:"column:bludgeoning_resistance_percent"`
	FireResistancePercent         int                     `json:"fireResistancePercent" gorm:"column:fire_resistance_percent"`
	IceResistancePercent          int                     `json:"iceResistancePercent" gorm:"column:ice_resistance_percent"`
	LightningResistancePercent    int                     `json:"lightningResistancePercent" gorm:"column:lightning_resistance_percent"`
	PoisonResistancePercent       int                     `json:"poisonResistancePercent" gorm:"column:poison_resistance_percent"`
	ArcaneResistancePercent       int                     `json:"arcaneResistancePercent" gorm:"column:arcane_resistance_percent"`
	HolyResistancePercent         int                     `json:"holyResistancePercent" gorm:"column:holy_resistance_percent"`
	ShadowResistancePercent       int                     `json:"shadowResistancePercent" gorm:"column:shadow_resistance_percent"`
	StartedAt                     time.Time               `json:"startedAt" gorm:"column:started_at"`
	LastTickAt                    *time.Time              `json:"lastTickAt,omitempty" gorm:"column:last_tick_at"`
	ExpiresAt                     time.Time               `json:"expiresAt" gorm:"column:expires_at"`
}

func (m *MonsterStatus) TableName() string {
	return "monster_statuses"
}

func (m MonsterStatus) StatModifiers() CharacterStatBonuses {
	return CharacterStatBonuses{
		Strength:                      m.StrengthMod,
		Dexterity:                     m.DexterityMod,
		Constitution:                  m.ConstitutionMod,
		Intelligence:                  m.IntelligenceMod,
		Wisdom:                        m.WisdomMod,
		Charisma:                      m.CharismaMod,
		PhysicalDamageBonusPercent:    m.PhysicalDamageBonusPercent,
		PiercingDamageBonusPercent:    m.PiercingDamageBonusPercent,
		SlashingDamageBonusPercent:    m.SlashingDamageBonusPercent,
		BludgeoningDamageBonusPercent: m.BludgeoningDamageBonusPercent,
		FireDamageBonusPercent:        m.FireDamageBonusPercent,
		IceDamageBonusPercent:         m.IceDamageBonusPercent,
		LightningDamageBonusPercent:   m.LightningDamageBonusPercent,
		PoisonDamageBonusPercent:      m.PoisonDamageBonusPercent,
		ArcaneDamageBonusPercent:      m.ArcaneDamageBonusPercent,
		HolyDamageBonusPercent:        m.HolyDamageBonusPercent,
		ShadowDamageBonusPercent:      m.ShadowDamageBonusPercent,
		PhysicalResistancePercent:     m.PhysicalResistancePercent,
		PiercingResistancePercent:     m.PiercingResistancePercent,
		SlashingResistancePercent:     m.SlashingResistancePercent,
		BludgeoningResistancePercent:  m.BludgeoningResistancePercent,
		FireResistancePercent:         m.FireResistancePercent,
		IceResistancePercent:          m.IceResistancePercent,
		LightningResistancePercent:    m.LightningResistancePercent,
		PoisonResistancePercent:       m.PoisonResistancePercent,
		ArcaneResistancePercent:       m.ArcaneResistancePercent,
		HolyResistancePercent:         m.HolyResistancePercent,
		ShadowResistancePercent:       m.ShadowResistancePercent,
	}
}
