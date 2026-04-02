package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DistrictSeedJobStatusQueued     = "queued"
	DistrictSeedJobStatusInProgress = "in_progress"
	DistrictSeedJobStatusCompleted  = "completed"
	DistrictSeedJobStatusFailed     = "failed"
)

var DistrictSeedJobStatuses = []string{
	DistrictSeedJobStatusQueued,
	DistrictSeedJobStatusInProgress,
	DistrictSeedJobStatusCompleted,
	DistrictSeedJobStatusFailed,
}

func IsValidDistrictSeedJobStatus(status string) bool {
	for _, candidate := range DistrictSeedJobStatuses {
		if candidate == status {
			return true
		}
	}
	return false
}

const (
	DistrictSeedResultStatusQueued    = "queued"
	DistrictSeedResultStatusCompleted = "completed"
	DistrictSeedResultStatusFailed    = "failed"
)

type DistrictSeedJob struct {
	ID                uuid.UUID                `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time                `json:"createdAt"`
	UpdatedAt         time.Time                `json:"updatedAt"`
	DistrictID        uuid.UUID                `json:"districtId" gorm:"type:uuid"`
	Status            string                   `json:"status"`
	ErrorMessage      *string                  `json:"errorMessage,omitempty"`
	QuestArchetypeIDs StringArray              `json:"questArchetypeIds,omitempty" gorm:"type:jsonb"`
	ZoneSeedSettings  DistrictZoneSeedSettings `json:"zoneSeedSettings" gorm:"type:jsonb"`
	ZoneSeedJobIDs    StringArray              `json:"zoneSeedJobIds,omitempty" gorm:"type:jsonb"`
	Results           DistrictSeedResults      `json:"results" gorm:"type:jsonb"`
}

func (DistrictSeedJob) TableName() string {
	return "district_seed_jobs"
}

type DistrictSeedResult struct {
	QuestArchetypeID        string  `json:"questArchetypeId"`
	QuestArchetypeName      string  `json:"questArchetypeName,omitempty"`
	Status                  string  `json:"status"`
	ErrorMessage            *string `json:"errorMessage,omitempty"`
	ZoneID                  *string `json:"zoneId,omitempty"`
	ZoneName                string  `json:"zoneName,omitempty"`
	MatchCount              int     `json:"matchCount"`
	QuestID                 *string `json:"questId,omitempty"`
	QuestGiverCharacterID   *string `json:"questGiverCharacterId,omitempty"`
	QuestGiverCharacterName string  `json:"questGiverCharacterName,omitempty"`
	GeneratedCharacterID    *string `json:"generatedCharacterId,omitempty"`
	GeneratedCharacterName  string  `json:"generatedCharacterName,omitempty"`
}

type DistrictZoneSeedSettings struct {
	PlaceCount           int         `json:"placeCount"`
	MonsterCount         int         `json:"monsterCount"`
	BossEncounterCount   int         `json:"bossEncounterCount"`
	RaidEncounterCount   int         `json:"raidEncounterCount"`
	InputEncounterCount  int         `json:"inputEncounterCount"`
	OptionEncounterCount int         `json:"optionEncounterCount"`
	TreasureChestCount   int         `json:"treasureChestCount"`
	HealingFountainCount int         `json:"healingFountainCount"`
	RequiredPlaceTags    StringArray `json:"requiredPlaceTags,omitempty"`
	ShopkeeperItemTags   StringArray `json:"shopkeeperItemTags,omitempty"`
}

func (s DistrictZoneSeedSettings) HasContent() bool {
	if s.PlaceCount > 0 ||
		s.MonsterCount > 0 ||
		s.BossEncounterCount > 0 ||
		s.RaidEncounterCount > 0 ||
		s.InputEncounterCount > 0 ||
		s.OptionEncounterCount > 0 ||
		s.TreasureChestCount > 0 ||
		s.HealingFountainCount > 0 {
		return true
	}
	for _, tag := range []string(s.RequiredPlaceTags) {
		if strings.TrimSpace(tag) != "" {
			return true
		}
	}
	for _, tag := range []string(s.ShopkeeperItemTags) {
		if strings.TrimSpace(tag) != "" {
			return true
		}
	}
	return false
}

func (s DistrictZoneSeedSettings) Value() (driver.Value, error) {
	if !s.HasContent() {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(s)
}

func (s *DistrictZoneSeedSettings) Scan(value interface{}) error {
	if value == nil {
		*s = DistrictZoneSeedSettings{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to scan DistrictZoneSeedSettings: unsupported type")
	}

	if len(bytes) == 0 {
		*s = DistrictZoneSeedSettings{}
		return nil
	}

	return json.Unmarshal(bytes, s)
}

type DistrictSeedResults []DistrictSeedResult

func (r DistrictSeedResults) Value() (driver.Value, error) {
	if r == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(r)
}

func (r *DistrictSeedResults) Scan(value interface{}) error {
	if value == nil {
		*r = DistrictSeedResults{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to scan DistrictSeedResults: unsupported type")
	}

	if len(bytes) == 0 {
		*r = DistrictSeedResults{}
		return nil
	}

	return json.Unmarshal(bytes, r)
}
