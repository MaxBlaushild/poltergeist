package models

import (
	"time"

	"github.com/google/uuid"
)

type SpellProgression struct {
	ID          uuid.UUID               `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt   time.Time               `json:"createdAt"`
	UpdatedAt   time.Time               `json:"updatedAt"`
	Name        string                  `json:"name"`
	AbilityType SpellAbilityType        `json:"abilityType" gorm:"column:ability_type"`
	Members     []SpellProgressionSpell `json:"members,omitempty" gorm:"foreignKey:ProgressionID"`
}

type SpellProgressionSpell struct {
	ID            uuid.UUID        `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt     time.Time        `json:"createdAt"`
	UpdatedAt     time.Time        `json:"updatedAt"`
	ProgressionID uuid.UUID        `json:"progressionId" gorm:"column:progression_id"`
	Progression   SpellProgression `json:"progression,omitempty" gorm:"foreignKey:ProgressionID"`
	SpellID       uuid.UUID        `json:"spellId" gorm:"column:spell_id"`
	Spell         Spell            `json:"-" gorm:"foreignKey:SpellID"`
	LevelBand     int              `json:"levelBand" gorm:"column:level_band"`
}

func (s *SpellProgression) TableName() string {
	return "spell_progressions"
}

func (s *SpellProgressionSpell) TableName() string {
	return "spell_progression_spells"
}
