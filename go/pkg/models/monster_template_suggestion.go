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
	MonsterTemplateSuggestionJobStatusQueued     = "queued"
	MonsterTemplateSuggestionJobStatusInProgress = "in_progress"
	MonsterTemplateSuggestionJobStatusCompleted  = "completed"
	MonsterTemplateSuggestionJobStatusFailed     = "failed"

	MonsterTemplateSuggestionDraftStatusSuggested = "suggested"
	MonsterTemplateSuggestionDraftStatusConverted = "converted"
)

type MonsterTemplateSuggestionPayload struct {
	MonsterType                   string    `json:"monsterType"`
	GenreID                       uuid.UUID `json:"genreId"`
	ZoneKind                      string    `json:"zoneKind"`
	Name                          string    `json:"name"`
	Description                   string    `json:"description"`
	BaseStrength                  int       `json:"baseStrength"`
	BaseDexterity                 int       `json:"baseDexterity"`
	BaseConstitution              int       `json:"baseConstitution"`
	BaseIntelligence              int       `json:"baseIntelligence"`
	BaseWisdom                    int       `json:"baseWisdom"`
	BaseCharisma                  int       `json:"baseCharisma"`
	PhysicalDamageBonusPercent    int       `json:"physicalDamageBonusPercent"`
	PiercingDamageBonusPercent    int       `json:"piercingDamageBonusPercent"`
	SlashingDamageBonusPercent    int       `json:"slashingDamageBonusPercent"`
	BludgeoningDamageBonusPercent int       `json:"bludgeoningDamageBonusPercent"`
	FireDamageBonusPercent        int       `json:"fireDamageBonusPercent"`
	IceDamageBonusPercent         int       `json:"iceDamageBonusPercent"`
	LightningDamageBonusPercent   int       `json:"lightningDamageBonusPercent"`
	PoisonDamageBonusPercent      int       `json:"poisonDamageBonusPercent"`
	ArcaneDamageBonusPercent      int       `json:"arcaneDamageBonusPercent"`
	HolyDamageBonusPercent        int       `json:"holyDamageBonusPercent"`
	ShadowDamageBonusPercent      int       `json:"shadowDamageBonusPercent"`
	PhysicalResistancePercent     int       `json:"physicalResistancePercent"`
	PiercingResistancePercent     int       `json:"piercingResistancePercent"`
	SlashingResistancePercent     int       `json:"slashingResistancePercent"`
	BludgeoningResistancePercent  int       `json:"bludgeoningResistancePercent"`
	FireResistancePercent         int       `json:"fireResistancePercent"`
	IceResistancePercent          int       `json:"iceResistancePercent"`
	LightningResistancePercent    int       `json:"lightningResistancePercent"`
	PoisonResistancePercent       int       `json:"poisonResistancePercent"`
	ArcaneResistancePercent       int       `json:"arcaneResistancePercent"`
	HolyResistancePercent         int       `json:"holyResistancePercent"`
	ShadowResistancePercent       int       `json:"shadowResistancePercent"`
}

type MonsterTemplateSuggestionPayloadValue MonsterTemplateSuggestionPayload

func (p MonsterTemplateSuggestionPayloadValue) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *MonsterTemplateSuggestionPayloadValue) Scan(value interface{}) error {
	if value == nil {
		*p = MonsterTemplateSuggestionPayloadValue{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan MonsterTemplateSuggestionPayloadValue: value is not []byte")
	}
	var decoded MonsterTemplateSuggestionPayloadValue
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*p = decoded
	return nil
}

type MonsterTemplateSuggestionJob struct {
	ID           uuid.UUID           `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt    time.Time           `json:"createdAt"`
	UpdatedAt    time.Time           `json:"updatedAt"`
	Status       string              `json:"status"`
	MonsterType  MonsterTemplateType `json:"monsterType" gorm:"column:monster_type"`
	GenreID      uuid.UUID           `json:"genreId" gorm:"column:genre_id;type:uuid"`
	Genre        *ZoneGenre          `json:"genre,omitempty" gorm:"foreignKey:GenreID"`
	ZoneKind     string              `json:"zoneKind" gorm:"column:zone_kind"`
	YeetIt       bool                `json:"yeetIt" gorm:"column:yeet_it"`
	Source       string              `json:"source"`
	Count        int                 `json:"count"`
	CreatedCount int                 `json:"createdCount" gorm:"column:created_count"`
	ErrorMessage *string             `json:"errorMessage,omitempty" gorm:"column:error_message"`
}

func (MonsterTemplateSuggestionJob) TableName() string {
	return "monster_template_suggestion_jobs"
}

type MonsterTemplateSuggestionDraft struct {
	ID                uuid.UUID                             `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt         time.Time                             `json:"createdAt"`
	UpdatedAt         time.Time                             `json:"updatedAt"`
	JobID             uuid.UUID                             `json:"jobId" gorm:"column:job_id;type:uuid"`
	Status            string                                `json:"status"`
	MonsterType       MonsterTemplateType                   `json:"monsterType" gorm:"column:monster_type"`
	GenreID           uuid.UUID                             `json:"genreId" gorm:"column:genre_id;type:uuid"`
	Genre             *ZoneGenre                            `json:"genre,omitempty" gorm:"foreignKey:GenreID"`
	ZoneKind          string                                `json:"zoneKind" gorm:"column:zone_kind"`
	Name              string                                `json:"name"`
	Description       string                                `json:"description"`
	Payload           MonsterTemplateSuggestionPayloadValue `json:"payload" gorm:"column:payload;type:jsonb"`
	MonsterTemplateID *uuid.UUID                            `json:"monsterTemplateId,omitempty" gorm:"column:monster_template_id;type:uuid"`
	MonsterTemplate   *MonsterTemplate                      `json:"monsterTemplate,omitempty" gorm:"foreignKey:MonsterTemplateID"`
	ConvertedAt       *time.Time                            `json:"convertedAt,omitempty" gorm:"column:converted_at"`
}

func (MonsterTemplateSuggestionDraft) TableName() string {
	return "monster_template_suggestion_drafts"
}

func NormalizeMonsterTemplateSuggestionJobStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case MonsterTemplateSuggestionJobStatusInProgress:
		return MonsterTemplateSuggestionJobStatusInProgress
	case MonsterTemplateSuggestionJobStatusCompleted:
		return MonsterTemplateSuggestionJobStatusCompleted
	case MonsterTemplateSuggestionJobStatusFailed:
		return MonsterTemplateSuggestionJobStatusFailed
	default:
		return MonsterTemplateSuggestionJobStatusQueued
	}
}

func NormalizeMonsterTemplateSuggestionDraftStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case MonsterTemplateSuggestionDraftStatusConverted:
		return MonsterTemplateSuggestionDraftStatusConverted
	default:
		return MonsterTemplateSuggestionDraftStatusSuggested
	}
}
