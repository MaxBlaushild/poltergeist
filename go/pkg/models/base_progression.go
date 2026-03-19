package models

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

type BaseResourceKey string

const (
	BaseResourceTimber       BaseResourceKey = "timber"
	BaseResourceStone        BaseResourceKey = "stone"
	BaseResourceIron         BaseResourceKey = "iron"
	BaseResourceHerbs        BaseResourceKey = "herbs"
	BaseResourceArcaneDust   BaseResourceKey = "arcane_dust"
	BaseResourceMonsterParts BaseResourceKey = "monster_parts"
	BaseResourceRelicShards  BaseResourceKey = "relic_shards"
)

func NormalizeBaseResourceKey(raw string) BaseResourceKey {
	switch BaseResourceKey(strings.ToLower(strings.TrimSpace(raw))) {
	case BaseResourceTimber:
		return BaseResourceTimber
	case BaseResourceStone:
		return BaseResourceStone
	case BaseResourceIron:
		return BaseResourceIron
	case BaseResourceHerbs:
		return BaseResourceHerbs
	case BaseResourceArcaneDust:
		return BaseResourceArcaneDust
	case BaseResourceMonsterParts:
		return BaseResourceMonsterParts
	case BaseResourceRelicShards:
		return BaseResourceRelicShards
	default:
		return ""
	}
}

func IsValidBaseResourceKey(raw string) bool {
	return NormalizeBaseResourceKey(raw) != ""
}

type BaseStructureEffectType string

const (
	BaseStructureEffectRestBonus       BaseStructureEffectType = "rest_bonus"
	BaseStructureEffectCraftUnlock     BaseStructureEffectType = "craft_unlock"
	BaseStructureEffectDailyChoiceBuff BaseStructureEffectType = "daily_choice_buff"
	BaseStructureEffectRewardBias      BaseStructureEffectType = "reward_bias"
)

const (
	BaseStructureImageGenerationStatusNone       = "none"
	BaseStructureImageGenerationStatusQueued     = "queued"
	BaseStructureImageGenerationStatusInProgress = "in_progress"
	BaseStructureImageGenerationStatusComplete   = "complete"
	BaseStructureImageGenerationStatusFailed     = "failed"
)

type BaseResourceDelta struct {
	ResourceKey BaseResourceKey `json:"resourceKey"`
	Amount      int             `json:"amount"`
}

type BaseMaterialRewards []BaseResourceDelta

func (r BaseMaterialRewards) Value() (driver.Value, error) {
	if r == nil {
		return json.Marshal([]BaseResourceDelta{})
	}
	return json.Marshal(r)
}

func (r *BaseMaterialRewards) Scan(value interface{}) error {
	if value == nil {
		*r = BaseMaterialRewards{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*r = BaseMaterialRewards{}
		return nil
	}
	if len(bytes) == 0 {
		*r = BaseMaterialRewards{}
		return nil
	}
	return json.Unmarshal(bytes, r)
}

type BaseResourceBalance struct {
	UserID      uuid.UUID       `json:"userId" gorm:"column:user_id;type:uuid;primaryKey"`
	ResourceKey BaseResourceKey `json:"resourceKey" gorm:"column:resource_key;primaryKey"`
	Amount      int             `json:"amount"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

func (BaseResourceBalance) TableName() string {
	return "base_resource_balances"
}

type BaseResourceLedger struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID      uuid.UUID       `json:"userId" gorm:"column:user_id;type:uuid"`
	ResourceKey BaseResourceKey `json:"resourceKey" gorm:"column:resource_key"`
	Delta       int             `json:"delta"`
	SourceType  string          `json:"sourceType" gorm:"column:source_type"`
	SourceID    *uuid.UUID      `json:"sourceId,omitempty" gorm:"column:source_id;type:uuid"`
	Notes       *string         `json:"notes,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
}

func (BaseResourceLedger) TableName() string {
	return "base_resource_ledger"
}

type BaseStructureDefinition struct {
	ID                 uuid.UUID                  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt          time.Time                  `json:"createdAt"`
	UpdatedAt          time.Time                  `json:"updatedAt"`
	Key                string                     `json:"key"`
	Name               string                     `json:"name"`
	Description        string                     `json:"description"`
	Category           string                     `json:"category"`
	MaxLevel           int                        `json:"maxLevel" gorm:"column:max_level"`
	SortOrder          int                        `json:"sortOrder" gorm:"column:sort_order"`
	ImageURL           string                     `json:"imageUrl" gorm:"column:image_url"`
	ImagePrompt        string                     `json:"imagePrompt" gorm:"column:image_prompt"`
	TopDownImagePrompt string                     `json:"topDownImagePrompt" gorm:"column:top_down_image_prompt"`
	EffectType         BaseStructureEffectType    `json:"effectType" gorm:"column:effect_type"`
	EffectConfig       MetadataJSONB              `json:"effectConfig" gorm:"column:effect_config;type:jsonb"`
	PrereqConfig       MetadataJSONB              `json:"prereqConfig" gorm:"column:prereq_config;type:jsonb"`
	Active             bool                       `json:"active"`
	LevelCosts         []BaseStructureLevelCost   `json:"levelCosts,omitempty" gorm:"foreignKey:StructureDefinitionID"`
	LevelVisuals       []BaseStructureLevelVisual `json:"levelVisuals,omitempty" gorm:"foreignKey:StructureDefinitionID"`
}

func (BaseStructureDefinition) TableName() string {
	return "base_structure_definitions"
}

type BaseStructureLevelCost struct {
	ID                    uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt             time.Time       `json:"createdAt"`
	UpdatedAt             time.Time       `json:"updatedAt"`
	StructureDefinitionID uuid.UUID       `json:"structureDefinitionId" gorm:"column:structure_definition_id;type:uuid"`
	Level                 int             `json:"level"`
	ResourceKey           BaseResourceKey `json:"resourceKey" gorm:"column:resource_key"`
	Amount                int             `json:"amount"`
}

func (BaseStructureLevelCost) TableName() string {
	return "base_structure_level_costs"
}

type BaseStructureLevelVisual struct {
	ID                           uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt                    time.Time `json:"createdAt"`
	UpdatedAt                    time.Time `json:"updatedAt"`
	StructureDefinitionID        uuid.UUID `json:"structureDefinitionId" gorm:"column:structure_definition_id;type:uuid"`
	Level                        int       `json:"level"`
	ImageURL                     string    `json:"imageUrl" gorm:"column:image_url"`
	ThumbnailURL                 string    `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	ImageGenerationStatus        string    `json:"imageGenerationStatus" gorm:"column:image_generation_status"`
	ImageGenerationError         *string   `json:"imageGenerationError,omitempty" gorm:"column:image_generation_error"`
	TopDownImageURL              string    `json:"topDownImageUrl" gorm:"column:top_down_image_url"`
	TopDownThumbnailURL          string    `json:"topDownThumbnailUrl" gorm:"column:top_down_thumbnail_url"`
	TopDownImageGenerationStatus string    `json:"topDownImageGenerationStatus" gorm:"column:top_down_image_generation_status"`
	TopDownImageGenerationError  *string   `json:"topDownImageGenerationError,omitempty" gorm:"column:top_down_image_generation_error"`
}

func (BaseStructureLevelVisual) TableName() string {
	return "base_structure_level_visuals"
}

type BaseGridPosition struct {
	GridX int `json:"gridX"`
	GridY int `json:"gridY"`
}

type UserBaseStructure struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	BaseID       uuid.UUID `json:"baseId" gorm:"column:base_id;type:uuid"`
	UserID       uuid.UUID `json:"userId" gorm:"column:user_id;type:uuid"`
	StructureKey string    `json:"structureKey" gorm:"column:structure_key"`
	Level        int       `json:"level"`
	GridX        int       `json:"gridX" gorm:"column:grid_x"`
	GridY        int       `json:"gridY" gorm:"column:grid_y"`
}

func (UserBaseStructure) TableName() string {
	return "user_base_structures"
}

type UserBaseDailyState struct {
	ID        uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
	UserID    uuid.UUID     `json:"userId" gorm:"column:user_id;type:uuid"`
	StateKey  string        `json:"stateKey" gorm:"column:state_key"`
	StateJSON MetadataJSONB `json:"state" gorm:"column:state_json;type:jsonb"`
	ResetsOn  time.Time     `json:"resetsOn" gorm:"column:resets_on;type:date"`
}

func (UserBaseDailyState) TableName() string {
	return "user_base_daily_state"
}
