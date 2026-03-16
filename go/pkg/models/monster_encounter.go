package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MonsterEncounterType string

const (
	MonsterEncounterTypeMonster MonsterEncounterType = "monster"
	MonsterEncounterTypeBoss    MonsterEncounterType = "boss"
	MonsterEncounterTypeRaid    MonsterEncounterType = "raid"
)

func NormalizeMonsterEncounterType(raw string) MonsterEncounterType {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(MonsterEncounterTypeBoss):
		return MonsterEncounterTypeBoss
	case string(MonsterEncounterTypeRaid):
		return MonsterEncounterTypeRaid
	default:
		return MonsterEncounterTypeMonster
	}
}

type MonsterEncounterRewardItem struct {
	InventoryItemID int `json:"inventoryItemId"`
	Quantity        int `json:"quantity"`
}

type MonsterEncounterRewardItems []MonsterEncounterRewardItem

func (r MonsterEncounterRewardItems) Value() (driver.Value, error) {
	if r == nil {
		return json.Marshal([]MonsterEncounterRewardItem{})
	}
	return json.Marshal(r)
}

func (r *MonsterEncounterRewardItems) Scan(value interface{}) error {
	if value == nil {
		*r = MonsterEncounterRewardItems{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*r = MonsterEncounterRewardItems{}
		return nil
	}
	if len(bytes) == 0 {
		*r = MonsterEncounterRewardItems{}
		return nil
	}
	return json.Unmarshal(bytes, r)
}

type MonsterEncounter struct {
	ID                          uuid.UUID                   `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt                   time.Time                   `json:"createdAt"`
	UpdatedAt                   time.Time                   `json:"updatedAt"`
	Name                        string                      `json:"name"`
	Description                 string                      `json:"description"`
	ImageURL                    string                      `json:"imageUrl" gorm:"column:image_url"`
	ThumbnailURL                string                      `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	EncounterType               MonsterEncounterType        `json:"encounterType" gorm:"column:encounter_type"`
	OwnerUserID                 *uuid.UUID                  `json:"ownerUserId,omitempty" gorm:"column:owner_user_id;type:uuid"`
	OwnerUser                   *User                       `json:"ownerUser,omitempty" gorm:"foreignKey:OwnerUserID"`
	Ephemeral                   bool                        `json:"ephemeral" gorm:"column:ephemeral"`
	ScaleWithUserLevel          bool                        `json:"scaleWithUserLevel" gorm:"column:scale_with_user_level"`
	RecurringMonsterEncounterID *uuid.UUID                  `json:"recurringMonsterEncounterId,omitempty" gorm:"column:recurring_monster_encounter_id;type:uuid"`
	RecurrenceFrequency         *string                     `json:"recurrenceFrequency,omitempty" gorm:"column:recurrence_frequency"`
	NextRecurrenceAt            *time.Time                  `json:"nextRecurrenceAt,omitempty" gorm:"column:next_recurrence_at"`
	RetiredAt                   *time.Time                  `json:"retiredAt,omitempty" gorm:"column:retired_at"`
	ZoneID                      uuid.UUID                   `json:"zoneId" gorm:"column:zone_id"`
	Zone                        Zone                        `json:"zone"`
	Latitude                    float64                     `json:"latitude"`
	Longitude                   float64                     `json:"longitude"`
	Geometry                    string                      `json:"geometry" gorm:"type:geometry(Point,4326)"`
	RewardMode                  RewardMode                  `json:"rewardMode" gorm:"column:reward_mode"`
	RandomRewardSize            RandomRewardSize            `json:"randomRewardSize" gorm:"column:random_reward_size"`
	RewardExperience            int                         `json:"rewardExperience" gorm:"column:reward_experience"`
	RewardGold                  int                         `json:"rewardGold" gorm:"column:reward_gold"`
	ItemRewards                 MonsterEncounterRewardItems `json:"itemRewards" gorm:"column:item_rewards_json;type:jsonb;default:'[]'"`
	Members                     []MonsterEncounterMember    `json:"members" gorm:"foreignKey:MonsterEncounterID"`
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
	m.EncounterType = NormalizeMonsterEncounterType(string(m.EncounterType))
	m.RewardMode = NormalizeRewardMode(string(m.RewardMode))
	m.RandomRewardSize = NormalizeRandomRewardSize(string(m.RandomRewardSize))
	return m.SetGeometry(m.Latitude, m.Longitude)
}

func (m *MonsterEncounter) SetGeometry(latitude, longitude float64) error {
	m.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", longitude, latitude)
	return nil
}
