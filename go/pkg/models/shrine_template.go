package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShrineEffectKind string

const (
	ShrineEffectKindStrength            ShrineEffectKind = "strength"
	ShrineEffectKindDexterity           ShrineEffectKind = "dexterity"
	ShrineEffectKindConstitution        ShrineEffectKind = "constitution"
	ShrineEffectKindIntelligence        ShrineEffectKind = "intelligence"
	ShrineEffectKindWisdom              ShrineEffectKind = "wisdom"
	ShrineEffectKindCharisma            ShrineEffectKind = "charisma"
	ShrineEffectKindHealthRegen         ShrineEffectKind = "health_regen"
	ShrineEffectKindManaRegen           ShrineEffectKind = "mana_regen"
	ShrineEffectKindPhysicalDamage      ShrineEffectKind = "physical_damage"
	ShrineEffectKindArcaneDamage        ShrineEffectKind = "arcane_damage"
	ShrineEffectKindHolyDamage          ShrineEffectKind = "holy_damage"
	ShrineEffectKindShadowDamage        ShrineEffectKind = "shadow_damage"
	ShrineEffectKindFireResistance      ShrineEffectKind = "fire_resistance"
	ShrineEffectKindIceResistance       ShrineEffectKind = "ice_resistance"
	ShrineEffectKindLightningResistance ShrineEffectKind = "lightning_resistance"
	ShrineEffectKindPoisonResistance    ShrineEffectKind = "poison_resistance"
	ShrineEffectKindPhysicalResistance  ShrineEffectKind = "physical_resistance"
	ShrineEffectKindAllDamageResistance ShrineEffectKind = "warding"
)

var allShrineEffectKinds = []ShrineEffectKind{
	ShrineEffectKindStrength,
	ShrineEffectKindDexterity,
	ShrineEffectKindConstitution,
	ShrineEffectKindIntelligence,
	ShrineEffectKindWisdom,
	ShrineEffectKindCharisma,
	ShrineEffectKindHealthRegen,
	ShrineEffectKindManaRegen,
	ShrineEffectKindPhysicalDamage,
	ShrineEffectKindArcaneDamage,
	ShrineEffectKindHolyDamage,
	ShrineEffectKindShadowDamage,
	ShrineEffectKindFireResistance,
	ShrineEffectKindIceResistance,
	ShrineEffectKindLightningResistance,
	ShrineEffectKindPoisonResistance,
	ShrineEffectKindPhysicalResistance,
	ShrineEffectKindAllDamageResistance,
}

func AllShrineEffectKinds() []ShrineEffectKind {
	return append([]ShrineEffectKind(nil), allShrineEffectKinds...)
}

func NormalizeShrineEffectKind(raw string) ShrineEffectKind {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	for _, kind := range allShrineEffectKinds {
		if normalized == string(kind) {
			return kind
		}
	}
	return ShrineEffectKindStrength
}

type ShrineTemplate struct {
	ID                uuid.UUID        `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time        `json:"createdAt"`
	UpdatedAt         time.Time        `json:"updatedAt"`
	ZoneKind          string           `json:"zoneKind,omitempty" gorm:"column:zone_kind"`
	Name              string           `json:"name"`
	Description       string           `json:"description"`
	BlessingName      string           `json:"blessingName" gorm:"column:blessing_name"`
	EffectDescription string           `json:"effectDescription" gorm:"column:effect_description"`
	EffectKind        ShrineEffectKind `json:"effectKind" gorm:"column:effect_kind"`
	BaseMagnitude     int              `json:"baseMagnitude" gorm:"column:base_magnitude"`
}

func (ShrineTemplate) TableName() string {
	return "shrine_templates"
}

func (s *ShrineTemplate) BeforeSave(tx *gorm.DB) error {
	s.ZoneKind = NormalizeZoneKind(s.ZoneKind)
	s.Name = strings.TrimSpace(s.Name)
	s.Description = strings.TrimSpace(s.Description)
	s.BlessingName = strings.TrimSpace(s.BlessingName)
	s.EffectDescription = strings.TrimSpace(s.EffectDescription)
	s.EffectKind = NormalizeShrineEffectKind(string(s.EffectKind))
	if s.BaseMagnitude <= 0 {
		s.BaseMagnitude = 1
	}
	return nil
}
