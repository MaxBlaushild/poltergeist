package models

import (
	"time"

	"github.com/google/uuid"
)

type ScenarioSpellReward struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	ScenarioID uuid.UUID `json:"scenarioId"`
	SpellID    uuid.UUID `json:"spellId"`
	Spell      Spell     `json:"spell" gorm:"foreignKey:SpellID"`
}

func (s *ScenarioSpellReward) TableName() string {
	return "scenario_spell_rewards"
}
