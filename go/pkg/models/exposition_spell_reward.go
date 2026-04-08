package models

import (
	"time"

	"github.com/google/uuid"
)

type ExpositionSpellReward struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	ExpositionID uuid.UUID `json:"expositionId"`
	SpellID      uuid.UUID `json:"spellId"`
	Spell        Spell     `json:"spell" gorm:"foreignKey:SpellID"`
}

func (e *ExpositionSpellReward) TableName() string {
	return "exposition_spell_rewards"
}
