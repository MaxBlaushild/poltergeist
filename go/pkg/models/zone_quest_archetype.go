package models

import (
	"time"

	"github.com/google/uuid"
)

type ZoneQuestArchetype struct {
	ID               uuid.UUID      `json:"id"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        *time.Time     `json:"deletedAt"`
	Zone             Zone           `json:"zone"`
	ZoneID           uuid.UUID      `json:"zoneId"`
	QuestArchetype   QuestArchetype `json:"questArchetype"`
	QuestArchetypeID uuid.UUID      `json:"questArchetypeId"`
	NumberOfQuests   int            `json:"numberOfQuests"`
}

func (zqa *ZoneQuestArchetype) TableName() string {
	return "zone_quest_archetypes"
}
