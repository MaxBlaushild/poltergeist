package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterestSpellReward struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	PointOfInterestID uuid.UUID `json:"pointOfInterestId"`
	SpellID           uuid.UUID `json:"spellId"`
	Spell             Spell     `json:"spell" gorm:"foreignKey:SpellID"`
}

func (p *PointOfInterestSpellReward) TableName() string {
	return "point_of_interest_spell_rewards"
}
