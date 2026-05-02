package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RewardProfile struct {
	ID                        uuid.UUID   `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                 time.Time   `json:"createdAt"`
	UpdatedAt                 time.Time   `json:"updatedAt"`
	Slug                      string      `json:"slug" gorm:"column:slug"`
	Name                      string      `json:"name"`
	Description               string      `json:"description"`
	Active                    bool        `json:"active"`
	PreferredItemTags         StringArray `json:"preferredItemTags" gorm:"column:preferred_item_tags;type:jsonb"`
	PreferredMaterialKeys     StringArray `json:"preferredMaterialKeys" gorm:"column:preferred_material_keys;type:jsonb"`
	PreferredDamageAffinities StringArray `json:"preferredDamageAffinities" gorm:"column:preferred_damage_affinities;type:jsonb"`
	PreferredResourceTypeIDs  StringArray `json:"preferredResourceTypeIds" gorm:"column:preferred_resource_type_ids;type:jsonb"`
	PreferEquipment           bool        `json:"preferEquipment" gorm:"column:prefer_equipment"`
	PreferUtility             bool        `json:"preferUtility" gorm:"column:prefer_utility"`
	PreferKnowledge           bool        `json:"preferKnowledge" gorm:"column:prefer_knowledge"`
	PreferNonEquipment        bool        `json:"preferNonEquipment" gorm:"column:prefer_non_equipment"`
}

func (RewardProfile) TableName() string {
	return "reward_profiles"
}

func NormalizeRewardProfileSlug(raw string) string {
	return NormalizeZoneKind(raw)
}

func normalizeRewardProfileMaterialKeys(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[BaseResourceKey]struct{}{}
	for _, raw := range values {
		resourceKey := NormalizeBaseResourceKey(raw)
		if resourceKey == "" {
			continue
		}
		if _, exists := seen[resourceKey]; exists {
			continue
		}
		seen[resourceKey] = struct{}{}
		normalized = append(normalized, string(resourceKey))
	}
	return normalized
}

func normalizeRewardProfileResourceTypeIDs(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[uuid.UUID]struct{}{}
	for _, raw := range values {
		id, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil || id == uuid.Nil {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id.String())
	}
	return normalized
}

func (r *RewardProfile) BeforeSave(tx *gorm.DB) error {
	r.Name = strings.TrimSpace(r.Name)
	r.Description = strings.TrimSpace(r.Description)
	r.Slug = NormalizeRewardProfileSlug(r.Slug)
	if r.Slug == "" {
		r.Slug = NormalizeRewardProfileSlug(r.Name)
	}
	r.PreferredItemTags = StringArray(NormalizeTagList([]string(r.PreferredItemTags)))
	r.PreferredMaterialKeys = StringArray(normalizeRewardProfileMaterialKeys([]string(r.PreferredMaterialKeys)))
	r.PreferredDamageAffinities = StringArray(NormalizeTagList([]string(r.PreferredDamageAffinities)))
	r.PreferredResourceTypeIDs = StringArray(normalizeRewardProfileResourceTypeIDs([]string(r.PreferredResourceTypeIDs)))
	return nil
}
