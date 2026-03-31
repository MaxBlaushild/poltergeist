package models

import (
	"time"

	"github.com/google/uuid"
)

type UserStatusEffectType string

const (
	UserStatusEffectTypeStatModifier   UserStatusEffectType = "stat_modifier"
	UserStatusEffectTypeDamageOverTime UserStatusEffectType = "damage_over_time"
	UserStatusEffectTypeHealthOverTime UserStatusEffectType = "health_over_time"
	UserStatusEffectTypeManaOverTime   UserStatusEffectType = "mana_over_time"
)

type UserStatus struct {
	ID                            uuid.UUID            `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt                     time.Time            `json:"createdAt"`
	UpdatedAt                     time.Time            `json:"updatedAt"`
	UserID                        uuid.UUID            `json:"userId"`
	Name                          string               `json:"name"`
	Description                   string               `json:"description"`
	Effect                        string               `json:"effect"`
	Positive                      bool                 `json:"positive" gorm:"column:positive"`
	EffectType                    UserStatusEffectType `json:"effectType" gorm:"column:effect_type"`
	DamagePerTick                 int                  `json:"damagePerTick" gorm:"column:damage_per_tick"`
	HealthPerTick                 int                  `json:"healthPerTick" gorm:"column:health_per_tick"`
	ManaPerTick                   int                  `json:"manaPerTick" gorm:"column:mana_per_tick"`
	StrengthMod                   int                  `json:"strengthMod" gorm:"column:strength_mod"`
	DexterityMod                  int                  `json:"dexterityMod" gorm:"column:dexterity_mod"`
	ConstitutionMod               int                  `json:"constitutionMod" gorm:"column:constitution_mod"`
	IntelligenceMod               int                  `json:"intelligenceMod" gorm:"column:intelligence_mod"`
	WisdomMod                     int                  `json:"wisdomMod" gorm:"column:wisdom_mod"`
	CharismaMod                   int                  `json:"charismaMod" gorm:"column:charisma_mod"`
	PhysicalDamageBonusPercent    int                  `json:"physicalDamageBonusPercent" gorm:"column:physical_damage_bonus_percent"`
	PiercingDamageBonusPercent    int                  `json:"piercingDamageBonusPercent" gorm:"column:piercing_damage_bonus_percent"`
	SlashingDamageBonusPercent    int                  `json:"slashingDamageBonusPercent" gorm:"column:slashing_damage_bonus_percent"`
	BludgeoningDamageBonusPercent int                  `json:"bludgeoningDamageBonusPercent" gorm:"column:bludgeoning_damage_bonus_percent"`
	FireDamageBonusPercent        int                  `json:"fireDamageBonusPercent" gorm:"column:fire_damage_bonus_percent"`
	IceDamageBonusPercent         int                  `json:"iceDamageBonusPercent" gorm:"column:ice_damage_bonus_percent"`
	LightningDamageBonusPercent   int                  `json:"lightningDamageBonusPercent" gorm:"column:lightning_damage_bonus_percent"`
	PoisonDamageBonusPercent      int                  `json:"poisonDamageBonusPercent" gorm:"column:poison_damage_bonus_percent"`
	ArcaneDamageBonusPercent      int                  `json:"arcaneDamageBonusPercent" gorm:"column:arcane_damage_bonus_percent"`
	HolyDamageBonusPercent        int                  `json:"holyDamageBonusPercent" gorm:"column:holy_damage_bonus_percent"`
	ShadowDamageBonusPercent      int                  `json:"shadowDamageBonusPercent" gorm:"column:shadow_damage_bonus_percent"`
	PhysicalResistancePercent     int                  `json:"physicalResistancePercent" gorm:"column:physical_resistance_percent"`
	PiercingResistancePercent     int                  `json:"piercingResistancePercent" gorm:"column:piercing_resistance_percent"`
	SlashingResistancePercent     int                  `json:"slashingResistancePercent" gorm:"column:slashing_resistance_percent"`
	BludgeoningResistancePercent  int                  `json:"bludgeoningResistancePercent" gorm:"column:bludgeoning_resistance_percent"`
	FireResistancePercent         int                  `json:"fireResistancePercent" gorm:"column:fire_resistance_percent"`
	IceResistancePercent          int                  `json:"iceResistancePercent" gorm:"column:ice_resistance_percent"`
	LightningResistancePercent    int                  `json:"lightningResistancePercent" gorm:"column:lightning_resistance_percent"`
	PoisonResistancePercent       int                  `json:"poisonResistancePercent" gorm:"column:poison_resistance_percent"`
	ArcaneResistancePercent       int                  `json:"arcaneResistancePercent" gorm:"column:arcane_resistance_percent"`
	HolyResistancePercent         int                  `json:"holyResistancePercent" gorm:"column:holy_resistance_percent"`
	ShadowResistancePercent       int                  `json:"shadowResistancePercent" gorm:"column:shadow_resistance_percent"`
	StartedAt                     time.Time            `json:"startedAt" gorm:"column:started_at"`
	LastTickAt                    *time.Time           `json:"lastTickAt,omitempty" gorm:"column:last_tick_at"`
	ExpiresAt                     time.Time            `json:"expiresAt" gorm:"column:expires_at"`
}

func (u *UserStatus) TableName() string {
	return "user_statuses"
}

func (u UserStatus) StatModifiers() CharacterStatBonuses {
	return CharacterStatBonuses{
		Strength:                      u.StrengthMod,
		Dexterity:                     u.DexterityMod,
		Constitution:                  u.ConstitutionMod,
		Intelligence:                  u.IntelligenceMod,
		Wisdom:                        u.WisdomMod,
		Charisma:                      u.CharismaMod,
		PhysicalDamageBonusPercent:    u.PhysicalDamageBonusPercent,
		PiercingDamageBonusPercent:    u.PiercingDamageBonusPercent,
		SlashingDamageBonusPercent:    u.SlashingDamageBonusPercent,
		BludgeoningDamageBonusPercent: u.BludgeoningDamageBonusPercent,
		FireDamageBonusPercent:        u.FireDamageBonusPercent,
		IceDamageBonusPercent:         u.IceDamageBonusPercent,
		LightningDamageBonusPercent:   u.LightningDamageBonusPercent,
		PoisonDamageBonusPercent:      u.PoisonDamageBonusPercent,
		ArcaneDamageBonusPercent:      u.ArcaneDamageBonusPercent,
		HolyDamageBonusPercent:        u.HolyDamageBonusPercent,
		ShadowDamageBonusPercent:      u.ShadowDamageBonusPercent,
		PhysicalResistancePercent:     u.PhysicalResistancePercent,
		PiercingResistancePercent:     u.PiercingResistancePercent,
		SlashingResistancePercent:     u.SlashingResistancePercent,
		BludgeoningResistancePercent:  u.BludgeoningResistancePercent,
		FireResistancePercent:         u.FireResistancePercent,
		IceResistancePercent:          u.IceResistancePercent,
		LightningResistancePercent:    u.LightningResistancePercent,
		PoisonResistancePercent:       u.PoisonResistancePercent,
		ArcaneResistancePercent:       u.ArcaneResistancePercent,
		HolyResistancePercent:         u.HolyResistancePercent,
		ShadowResistancePercent:       u.ShadowResistancePercent,
	}
}
