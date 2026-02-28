package models

import (
	"time"

	"github.com/google/uuid"
)

type QuestSpellReward struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	QuestID   uuid.UUID `json:"questId"`
	Quest     Quest     `json:"quest" gorm:"foreignKey:QuestID"`
	SpellID   uuid.UUID `json:"spellId"`
	Spell     Spell     `json:"spell" gorm:"foreignKey:SpellID"`
}

func (q *QuestSpellReward) TableName() string {
	return "quest_spell_rewards"
}
