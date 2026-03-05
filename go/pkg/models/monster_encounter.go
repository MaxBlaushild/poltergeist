package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MonsterEncounter struct {
	ID                          uuid.UUID                `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt                   time.Time                `json:"createdAt"`
	UpdatedAt                   time.Time                `json:"updatedAt"`
	Name                        string                   `json:"name"`
	Description                 string                   `json:"description"`
	ImageURL                    string                   `json:"imageUrl" gorm:"column:image_url"`
	ThumbnailURL                string                   `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	ScaleWithUserLevel          bool                     `json:"scaleWithUserLevel" gorm:"column:scale_with_user_level"`
	RecurringMonsterEncounterID *uuid.UUID               `json:"recurringMonsterEncounterId,omitempty" gorm:"column:recurring_monster_encounter_id;type:uuid"`
	RecurrenceFrequency         *string                  `json:"recurrenceFrequency,omitempty" gorm:"column:recurrence_frequency"`
	NextRecurrenceAt            *time.Time               `json:"nextRecurrenceAt,omitempty" gorm:"column:next_recurrence_at"`
	ZoneID                      uuid.UUID                `json:"zoneId" gorm:"column:zone_id"`
	Zone                        Zone                     `json:"zone"`
	Latitude                    float64                  `json:"latitude"`
	Longitude                   float64                  `json:"longitude"`
	Geometry                    string                   `json:"geometry" gorm:"type:geometry(Point,4326)"`
	Members                     []MonsterEncounterMember `json:"members" gorm:"foreignKey:MonsterEncounterID"`
}

type MonsterEncounterMember struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	MonsterEncounterID uuid.UUID `json:"monsterEncounterId" gorm:"column:monster_encounter_id"`
	MonsterID          uuid.UUID `json:"monsterId" gorm:"column:monster_id"`
	Monster            Monster   `json:"monster" gorm:"foreignKey:MonsterID"`
	Slot               int       `json:"slot"`
}

func (m *MonsterEncounter) TableName() string {
	return "monster_encounters"
}

func (m *MonsterEncounterMember) TableName() string {
	return "monster_encounter_members"
}

func (m *MonsterEncounter) BeforeSave(tx *gorm.DB) error {
	return m.SetGeometry(m.Latitude, m.Longitude)
}

func (m *MonsterEncounter) SetGeometry(latitude, longitude float64) error {
	m.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", longitude, latitude)
	return nil
}
