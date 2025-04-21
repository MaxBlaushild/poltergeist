package models

import (
	"time"

	"github.com/google/uuid"
)

type ZoneQuestArchetype struct {
	ID               uuid.UUID      `json:"id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        *time.Time     `json:"deleted_at"`
	Zone             Zone           `json:"zone"`
	ZoneID           uuid.UUID      `json:"zone_id"`
	QuestArchetype   QuestArchetype `json:"quest_archetype"`
	QuestArchetypeID uuid.UUID      `json:"quest_archetype_id"`
	NumberOfQuests   int            `json:"number_of_quests"`
}

func (zqa *ZoneQuestArchetype) TableName() string {
	return "zone_quest_archetypes"
}
