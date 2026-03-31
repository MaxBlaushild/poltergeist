package models

import (
	"time"

	"github.com/google/uuid"
)

type ZoneAdminSummary struct {
	ID                      uuid.UUID   `json:"id" gorm:"column:id"`
	CreatedAt               time.Time   `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt               time.Time   `json:"updatedAt" gorm:"column:updated_at"`
	Name                    string      `json:"name" gorm:"column:name"`
	Description             string      `json:"description" gorm:"column:description"`
	InternalTags            StringArray `json:"internalTags" gorm:"column:internal_tags;type:jsonb"`
	Latitude                float64     `json:"latitude" gorm:"column:latitude"`
	Longitude               float64     `json:"longitude" gorm:"column:longitude"`
	ZoneImportID            *uuid.UUID  `json:"zoneImportId" gorm:"column:zone_import_id"`
	ImportMetroName         *string     `json:"importMetroName,omitempty" gorm:"column:import_metro_name"`
	BoundaryPointCount      int         `json:"boundaryPointCount" gorm:"column:boundary_point_count"`
	PointOfInterestCount    int         `json:"pointOfInterestCount" gorm:"column:point_of_interest_count"`
	QuestCount              int         `json:"questCount" gorm:"column:quest_count"`
	ZoneQuestArchetypeCount int         `json:"zoneQuestArchetypeCount" gorm:"column:zone_quest_archetype_count"`
	ChallengeCount          int         `json:"challengeCount" gorm:"column:challenge_count"`
	ScenarioCount           int         `json:"scenarioCount" gorm:"column:scenario_count"`
	MonsterCount            int         `json:"monsterCount" gorm:"column:monster_count"`
	MonsterEncounterCount   int         `json:"monsterEncounterCount" gorm:"column:monster_encounter_count"`
	TreasureChestCount      int         `json:"treasureChestCount" gorm:"column:treasure_chest_count"`
	HealingFountainCount    int         `json:"healingFountainCount" gorm:"column:healing_fountain_count"`
}
