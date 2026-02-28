package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SpellEffectType string

const (
	SpellEffectTypeDealDamage             SpellEffectType = "deal_damage"
	SpellEffectTypeRestoreLifePartyMember SpellEffectType = "restore_life_party_member"
	SpellEffectTypeRestoreLifeAllParty    SpellEffectType = "restore_life_all_party_members"
	SpellEffectTypeApplyBeneficialStatus  SpellEffectType = "apply_beneficial_statuses"
	SpellEffectTypeRemoveDetrimental      SpellEffectType = "remove_detrimental_statuses"
)

type SpellEffect struct {
	Type             SpellEffectType                `json:"type"`
	Amount           int                            `json:"amount,omitempty"`
	StatusesToApply  ScenarioFailureStatusTemplates `json:"statusesToApply,omitempty"`
	StatusesToRemove StringArray                    `json:"statusesToRemove,omitempty"`
	EffectData       map[string]interface{}         `json:"effectData,omitempty"`
}

type SpellEffects []SpellEffect

func (s SpellEffects) Value() (driver.Value, error) {
	if s == nil {
		return json.Marshal([]SpellEffect{})
	}
	return json.Marshal(s)
}

func (s *SpellEffects) Scan(value interface{}) error {
	if value == nil {
		*s = SpellEffects{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*s = SpellEffects{}
		return nil
	}

	return json.Unmarshal(bytes, s)
}

type Spell struct {
	ID                    uuid.UUID    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt             time.Time    `json:"createdAt"`
	UpdatedAt             time.Time    `json:"updatedAt"`
	Name                  string       `json:"name"`
	Description           string       `json:"description"`
	IconURL               string       `json:"iconUrl" gorm:"column:icon_url"`
	ImageGenerationStatus string       `json:"imageGenerationStatus" gorm:"column:image_generation_status"`
	ImageGenerationError  *string      `json:"imageGenerationError,omitempty" gorm:"column:image_generation_error"`
	EffectText            string       `json:"effectText" gorm:"column:effect_text"`
	SchoolOfMagic         string       `json:"schoolOfMagic" gorm:"column:school_of_magic"`
	ManaCost              int          `json:"manaCost" gorm:"column:mana_cost"`
	Effects               SpellEffects `json:"effects" gorm:"column:effects;type:jsonb"`
}

func (s *Spell) TableName() string {
	return "spells"
}

const (
	SpellImageGenerationStatusNone       = "none"
	SpellImageGenerationStatusQueued     = "queued"
	SpellImageGenerationStatusInProgress = "in_progress"
	SpellImageGenerationStatusComplete   = "complete"
	SpellImageGenerationStatusFailed     = "failed"
)
