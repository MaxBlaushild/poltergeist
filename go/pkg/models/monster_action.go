package models

import (
	"time"

	"github.com/google/uuid"
)

type MonsterAction struct {
	ID        uuid.UUID `db:"id" gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`

	MonsterID uuid.UUID `db:"monster_id" gorm:"not null" json:"monsterId"`

	// Action classification
	ActionType string `db:"action_type" gorm:"not null" json:"actionType"` // 'action', 'special_ability', 'legendary_action', 'reaction'
	OrderIndex int    `db:"order_index" gorm:"default:0" json:"orderIndex"` // For ordering actions within each type

	// Basic action info
	Name        string `db:"name" gorm:"not null" json:"name"`
	Description string `db:"description" gorm:"not null" json:"description"`

	// Attack mechanics
	AttackBonus           *int    `db:"attack_bonus" json:"attackBonus,omitempty"`           // +4 to hit
	DamageDice            *string `db:"damage_dice" json:"damageDice,omitempty"`             // "1d6+2"
	DamageType            *string `db:"damage_type" json:"damageType,omitempty"`             // "slashing", "fire", etc.
	AdditionalDamageDice  *string `db:"additional_damage_dice" json:"additionalDamageDice,omitempty"`  // "4d6"
	AdditionalDamageType  *string `db:"additional_damage_type" json:"additionalDamageType,omitempty"`  // "fire"

	// Save mechanics
	SaveDC                *int    `db:"save_dc" json:"saveDC,omitempty"`                     // DC 15
	SaveAbility           *string `db:"save_ability" json:"saveAbility,omitempty"`          // "Dexterity"
	SaveEffectHalfDamage  bool    `db:"save_effect_half_damage" gorm:"default:false" json:"saveEffectHalfDamage"` // true if half damage on save

	// Range and area
	RangeReach *int    `db:"range_reach" json:"rangeReach,omitempty"` // 5 ft reach, 80 ft range
	RangeLong  *int    `db:"range_long" json:"rangeLong,omitempty"`   // long range for ranged attacks
	AreaType   *string `db:"area_type" json:"areaType,omitempty"`     // "cone", "sphere", "line", "cube"
	AreaSize   *int    `db:"area_size" json:"areaSize,omitempty"`     // 60 (for 60-foot cone)

	// Special mechanics
	Recharge       *string `db:"recharge" json:"recharge,omitempty"`         // "5-6", "short rest", "long rest"
	UsesPerDay     *int    `db:"uses_per_day" json:"usesPerDay,omitempty"`   // 3/Day
	SpecialEffects *string `db:"special_effects" json:"specialEffects,omitempty"` // Additional effects like "target is knocked prone"

	// Legendary action cost
	LegendaryCost int `db:"legendary_cost" gorm:"default:1" json:"legendaryCost"` // How many legendary actions this costs

	Active bool `db:"active" gorm:"default:true" json:"active"`

	// Relationship
	Monster Monster `gorm:"foreignKey:MonsterID" json:"monster,omitempty"`
}

// ActionType constants for better type safety
const (
	ActionTypeAction         = "action"
	ActionTypeSpecialAbility = "special_ability"
	ActionTypeLegendaryAction = "legendary_action"
	ActionTypeReaction       = "reaction"
)

// Common damage types
var DamageTypes = []string{
	"acid", "bludgeoning", "cold", "fire", "force", "lightning", "necrotic",
	"piercing", "poison", "psychic", "radiant", "slashing", "thunder",
}

// Common area types for spells and abilities
var AreaTypes = []string{
	"cone", "cube", "cylinder", "line", "sphere",
}

// Common save abilities
var SaveAbilities = []string{
	"Strength", "Dexterity", "Constitution", "Intelligence", "Wisdom", "Charisma",
}

// GetFormattedDamage returns a formatted string for damage display
func (ma *MonsterAction) GetFormattedDamage() string {
	var damage string
	
	if ma.DamageDice != nil {
		damage = *ma.DamageDice
		if ma.DamageType != nil {
			damage += " " + *ma.DamageType
		}
	}
	
	if ma.AdditionalDamageDice != nil {
		if damage != "" {
			damage += " plus "
		}
		damage += *ma.AdditionalDamageDice
		if ma.AdditionalDamageType != nil {
			damage += " " + *ma.AdditionalDamageType
		}
	}
	
	return damage
}

// GetFormattedRange returns a formatted string for range display
func (ma *MonsterAction) GetFormattedRange() string {
	if ma.RangeReach == nil {
		return ""
	}
	
	rangeStr := ""
	if *ma.RangeReach <= 10 {
		rangeStr = "reach " + string(rune(*ma.RangeReach)) + " ft."
	} else {
		rangeStr = "range " + string(rune(*ma.RangeReach))
		if ma.RangeLong != nil {
			rangeStr += "/" + string(rune(*ma.RangeLong))
		}
		rangeStr += " ft."
	}
	
	return rangeStr
}

// GetFormattedArea returns a formatted string for area effect display
func (ma *MonsterAction) GetFormattedArea() string {
	if ma.AreaType == nil || ma.AreaSize == nil {
		return ""
	}
	
	return string(rune(*ma.AreaSize)) + "-foot " + *ma.AreaType
}

// IsAttack returns true if this action is an attack
func (ma *MonsterAction) IsAttack() bool {
	return ma.AttackBonus != nil
}

// IsSave returns true if this action requires a saving throw
func (ma *MonsterAction) IsSave() bool {
	return ma.SaveDC != nil && ma.SaveAbility != nil
}

// HasRecharge returns true if this action has a recharge mechanic
func (ma *MonsterAction) HasRecharge() bool {
	return ma.Recharge != nil && *ma.Recharge != ""
}

// HasLimitedUses returns true if this action has limited daily uses
func (ma *MonsterAction) HasLimitedUses() bool {
	return ma.UsesPerDay != nil && *ma.UsesPerDay > 0
}