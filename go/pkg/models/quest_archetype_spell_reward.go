package models

import (
	"time"

	"github.com/google/uuid"
)

type QuestArchetypeSpellReward struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	QuestArchetypeID uuid.UUID      `json:"questArchetypeId"`
	QuestArchetype   QuestArchetype `json:"questArchetype" gorm:"foreignKey:QuestArchetypeID"`
	SpellID          uuid.UUID      `json:"spellId"`
	Spell            Spell          `json:"spell" gorm:"foreignKey:SpellID"`
}

func (q *QuestArchetypeSpellReward) TableName() string {
	return "quest_archetype_spell_rewards"
}
