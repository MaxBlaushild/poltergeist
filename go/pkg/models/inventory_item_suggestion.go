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
	InventoryItemSuggestionJobStatusQueued     = "queued"
	InventoryItemSuggestionJobStatusInProgress = "in_progress"
	InventoryItemSuggestionJobStatusCompleted  = "completed"
	InventoryItemSuggestionJobStatusFailed     = "failed"

	InventoryItemSuggestionDraftStatusSuggested = "suggested"
	InventoryItemSuggestionDraftStatusConverted = "converted"
)

type InventoryItemSuggestionPayload struct {
	Category  string        `json:"category"`
	WhyItFits string        `json:"whyItFits"`
	Item      InventoryItem `json:"item"`
}

type InventoryItemSuggestionPayloadValue InventoryItemSuggestionPayload

func (p InventoryItemSuggestionPayloadValue) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *InventoryItemSuggestionPayloadValue) Scan(value interface{}) error {
	if value == nil {
		*p = InventoryItemSuggestionPayloadValue{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan InventoryItemSuggestionPayloadValue: value is not []byte")
	}
	var decoded InventoryItemSuggestionPayloadValue
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*p = decoded
	return nil
}

type InventoryItemSuggestionJob struct {
	ID           uuid.UUID   `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
	GenreID      uuid.UUID   `json:"genreId" gorm:"column:genre_id;type:uuid"`
	Genre        *ZoneGenre  `json:"genre,omitempty" gorm:"foreignKey:GenreID"`
	Status       string      `json:"status"`
	Count        int         `json:"count"`
	ThemePrompt  string      `json:"themePrompt" gorm:"column:theme_prompt"`
	Categories   StringArray `json:"categories" gorm:"type:jsonb"`
	RarityTiers  StringArray `json:"rarityTiers" gorm:"column:rarity_tiers;type:jsonb"`
	EquipSlots   StringArray `json:"equipSlots" gorm:"column:equip_slots;type:jsonb"`
	StatTags     StringArray `json:"statTags" gorm:"column:stat_tags;type:jsonb"`
	BenefitTags  StringArray `json:"benefitTags" gorm:"column:benefit_tags;type:jsonb"`
	StatusNames  StringArray `json:"statusNames" gorm:"column:status_names;type:jsonb"`
	InternalTags StringArray `json:"internalTags" gorm:"column:internal_tags;type:jsonb"`
	MinItemLevel int         `json:"minItemLevel" gorm:"column:min_item_level"`
	MaxItemLevel int         `json:"maxItemLevel" gorm:"column:max_item_level"`
	CreatedCount int         `json:"createdCount" gorm:"column:created_count"`
	ErrorMessage *string     `json:"errorMessage,omitempty" gorm:"column:error_message"`
}

func (InventoryItemSuggestionJob) TableName() string {
	return "inventory_item_suggestion_jobs"
}

type InventoryItemSuggestionDraft struct {
	ID              uuid.UUID                           `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt       time.Time                           `json:"createdAt"`
	UpdatedAt       time.Time                           `json:"updatedAt"`
	JobID           uuid.UUID                           `json:"jobId" gorm:"column:job_id;type:uuid"`
	Status          string                              `json:"status"`
	Name            string                              `json:"name"`
	Category        string                              `json:"category"`
	RarityTier      string                              `json:"rarityTier" gorm:"column:rarity_tier"`
	ItemLevel       int                                 `json:"itemLevel" gorm:"column:item_level"`
	EquipSlot       *string                             `json:"equipSlot,omitempty" gorm:"column:equip_slot"`
	WhyItFits       string                              `json:"whyItFits" gorm:"column:why_it_fits"`
	InternalTags    StringArray                         `json:"internalTags" gorm:"column:internal_tags;type:jsonb"`
	Warnings        StringArray                         `json:"warnings" gorm:"type:jsonb"`
	Payload         InventoryItemSuggestionPayloadValue `json:"payload" gorm:"column:payload;type:jsonb"`
	InventoryItemID *int                                `json:"inventoryItemId,omitempty" gorm:"column:inventory_item_id"`
	InventoryItem   *InventoryItem                      `json:"inventoryItem,omitempty" gorm:"foreignKey:InventoryItemID"`
	ConvertedAt     *time.Time                          `json:"convertedAt,omitempty" gorm:"column:converted_at"`
}

func (InventoryItemSuggestionDraft) TableName() string {
	return "inventory_item_suggestion_drafts"
}

func NormalizeInventoryItemSuggestionJobStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case InventoryItemSuggestionJobStatusInProgress:
		return InventoryItemSuggestionJobStatusInProgress
	case InventoryItemSuggestionJobStatusCompleted:
		return InventoryItemSuggestionJobStatusCompleted
	case InventoryItemSuggestionJobStatusFailed:
		return InventoryItemSuggestionJobStatusFailed
	default:
		return InventoryItemSuggestionJobStatusQueued
	}
}

func NormalizeInventoryItemSuggestionDraftStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case InventoryItemSuggestionDraftStatusConverted:
		return InventoryItemSuggestionDraftStatusConverted
	default:
		return InventoryItemSuggestionDraftStatusSuggested
	}
}
