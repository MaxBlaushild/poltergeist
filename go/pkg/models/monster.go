package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// SpeedModifiers represents different movement speeds for a monster
type SpeedModifiers map[string]int

// Scan implements the sql.Scanner interface for SpeedModifiers
func (s *SpeedModifiers) Scan(value interface{}) error {
	if value == nil {
		*s = make(SpeedModifiers)
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("cannot scan SpeedModifiers from non-string type")
	}
}

// Value implements the driver.Valuer interface for SpeedModifiers
func (s SpeedModifiers) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// SkillProficiencies represents skill proficiency bonuses
type SkillProficiencies map[string]int

// Scan implements the sql.Scanner interface for SkillProficiencies
func (s *SkillProficiencies) Scan(value interface{}) error {
	if value == nil {
		*s = make(SkillProficiencies)
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("cannot scan SkillProficiencies from non-string type")
	}
}

// Value implements the driver.Valuer interface for SkillProficiencies
func (s SkillProficiencies) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// MonsterAbility represents a special ability, action, or reaction
type MonsterAbility struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	AttackBonus     *int   `json:"attack_bonus,omitempty"`
	Damage          string `json:"damage,omitempty"`
	DamageType      string `json:"damage_type,omitempty"`
	AdditionalDamage string `json:"additional_damage,omitempty"`
	SaveDC          *int   `json:"save_dc,omitempty"`
	SaveAbility     string `json:"save_ability,omitempty"`
	Area            string `json:"area,omitempty"`
	Special         string `json:"special,omitempty"`
	Recharge        string `json:"recharge,omitempty"` // e.g., "5-6" for recharge on 5 or 6
}

// MonsterAbilities represents a slice of MonsterAbility
type MonsterAbilities []MonsterAbility

// Scan implements the sql.Scanner interface for MonsterAbilities
func (m *MonsterAbilities) Scan(value interface{}) error {
	if value == nil {
		*m = []MonsterAbility{}
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	default:
		return errors.New("cannot scan MonsterAbilities from non-string type")
	}
}

// Value implements the driver.Valuer interface for MonsterAbilities
func (m MonsterAbilities) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

type Monster struct {
	ID        uuid.UUID `db:"id" gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`

	// Basic Information
	Name      string  `json:"name" gorm:"unique;not null"`
	Size      string  `json:"size" gorm:"not null;default:'Medium'"`
	Type      string  `json:"type" gorm:"not null"`
	Subtype   *string `json:"subtype"`
	Alignment string  `json:"alignment" gorm:"not null;default:'unaligned'"`

	// Core Stats
	ArmorClass     int            `json:"armorClass" gorm:"not null;default:10"`
	HitPoints      int            `json:"hitPoints" gorm:"not null;default:1"`
	HitDice        *string        `json:"hitDice"`
	Speed          int            `json:"speed" gorm:"not null;default:30"`
	SpeedModifiers SpeedModifiers `json:"speedModifiers" gorm:"type:jsonb"`

	// Ability Scores
	Strength     int `json:"strength" gorm:"not null;default:10"`
	Dexterity    int `json:"dexterity" gorm:"not null;default:10"`
	Constitution int `json:"constitution" gorm:"not null;default:10"`
	Intelligence int `json:"intelligence" gorm:"not null;default:10"`
	Wisdom       int `json:"wisdom" gorm:"not null;default:10"`
	Charisma     int `json:"charisma" gorm:"not null;default:10"`

	// Derived Stats
	ProficiencyBonus  int     `json:"proficiencyBonus" gorm:"not null;default:2"`
	ChallengeRating   float64 `json:"challengeRating" gorm:"type:decimal(4,2);not null;default:0"`
	ExperiencePoints  int     `json:"experiencePoints" gorm:"not null;default:0"`

	// Skills and Saves
	SavingThrowProficiencies pq.StringArray     `json:"savingThrowProficiencies" gorm:"type:text[]"`
	SkillProficiencies      SkillProficiencies `json:"skillProficiencies" gorm:"type:jsonb"`

	// Resistances and Immunities
	DamageVulnerabilities pq.StringArray `json:"damageVulnerabilities" gorm:"type:text[]"`
	DamageResistances     pq.StringArray `json:"damageResistances" gorm:"type:text[]"`
	DamageImmunities      pq.StringArray `json:"damageImmunities" gorm:"type:text[]"`
	ConditionImmunities   pq.StringArray `json:"conditionImmunities" gorm:"type:text[]"`

	// Senses
	Blindsight        int `json:"blindsight" gorm:"default:0"`
	Darkvision        int `json:"darkvision" gorm:"default:0"`
	Tremorsense       int `json:"tremorsense" gorm:"default:0"`
	Truesight         int `json:"truesight" gorm:"default:0"`
	PassivePerception int `json:"passivePerception" gorm:"not null;default:10"`

	// Languages
	Languages pq.StringArray `json:"languages" gorm:"type:text[]"`

	// Special Abilities, Actions, etc.
	SpecialAbilities         MonsterAbilities `json:"specialAbilities" gorm:"type:jsonb"`
	Actions                  MonsterAbilities `json:"actions" gorm:"type:jsonb"`
	LegendaryActions         MonsterAbilities `json:"legendaryActions" gorm:"type:jsonb"`
	LegendaryActionsPerTurn  int              `json:"legendaryActionsPerTurn" gorm:"default:0"`
	Reactions                MonsterAbilities `json:"reactions" gorm:"type:jsonb"`

	// Visual and Flavor
	ImageURL    *string `json:"imageUrl"`
	Description *string `json:"description"`
	FlavorText  *string `json:"flavorText"`
	Environment *string `json:"environment"`

	// Meta
	Source string `json:"source" gorm:"default:'Custom'"`
	Active bool   `json:"active" gorm:"default:true"`
}