package models

import (
	"time"

	"github.com/google/uuid"
)

type ScenarioOptionSpellReward struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	ScenarioOptionID uuid.UUID `json:"scenarioOptionId" gorm:"column:scenario_option_id"`
	SpellID          uuid.UUID `json:"spellId"`
	Spell            Spell     `json:"spell" gorm:"foreignKey:SpellID"`
}

func (s *ScenarioOptionSpellReward) TableName() string {
	return "scenario_option_spell_rewards"
}
