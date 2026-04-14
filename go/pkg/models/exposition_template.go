package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExpositionTemplateItemReward struct {
	InventoryItemID int `json:"inventoryItemId"`
	Quantity        int `json:"quantity"`
}

type ExpositionTemplateItemRewards []ExpositionTemplateItemReward

func (r ExpositionTemplateItemRewards) Value() (driver.Value, error) {
	if r == nil {
		return json.Marshal([]ExpositionTemplateItemReward{})
	}
	return json.Marshal(r)
}

func (r *ExpositionTemplateItemRewards) Scan(value interface{}) error {
	if value == nil {
		*r = ExpositionTemplateItemRewards{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan ExpositionTemplateItemRewards: value is not []byte")
	}
	var decoded []ExpositionTemplateItemReward
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*r = decoded
	return nil
}

type ExpositionTemplateSpellReward struct {
	SpellID uuid.UUID `json:"spellId"`
}

type ExpositionTemplateSpellRewards []ExpositionTemplateSpellReward

func (r ExpositionTemplateSpellRewards) Value() (driver.Value, error) {
	if r == nil {
		return json.Marshal([]ExpositionTemplateSpellReward{})
	}
	return json.Marshal(r)
}

func (r *ExpositionTemplateSpellRewards) Scan(value interface{}) error {
	if value == nil {
		*r = ExpositionTemplateSpellRewards{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan ExpositionTemplateSpellRewards: value is not []byte")
	}
	var decoded []ExpositionTemplateSpellReward
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*r = decoded
	return nil
}

type ExpositionTemplate struct {
	ID                 uuid.UUID                      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt          time.Time                      `json:"createdAt"`
	UpdatedAt          time.Time                      `json:"updatedAt"`
	Title              string                         `json:"title"`
	Description        string                         `json:"description"`
	Dialogue           DialogueSequence               `json:"dialogue" gorm:"type:jsonb;default:'[]'"`
	RequiredStoryFlags StringArray                    `json:"requiredStoryFlags" gorm:"column:required_story_flags;type:jsonb;default:'[]'"`
	ImageURL           string                         `json:"imageUrl" gorm:"column:image_url"`
	ThumbnailURL       string                         `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	RewardMode         RewardMode                     `json:"rewardMode" gorm:"column:reward_mode"`
	RandomRewardSize   RandomRewardSize               `json:"randomRewardSize" gorm:"column:random_reward_size"`
	RewardExperience   int                            `json:"rewardExperience" gorm:"column:reward_experience"`
	RewardGold         int                            `json:"rewardGold" gorm:"column:reward_gold"`
	MaterialRewards    BaseMaterialRewards            `json:"materialRewards" gorm:"column:material_rewards_json;type:jsonb;default:'[]'"`
	ItemRewards        ExpositionTemplateItemRewards  `json:"itemRewards" gorm:"column:item_rewards_json;type:jsonb;default:'[]'"`
	SpellRewards       ExpositionTemplateSpellRewards `json:"spellRewards" gorm:"column:spell_rewards_json;type:jsonb;default:'[]'"`
}

func (ExpositionTemplate) TableName() string {
	return "exposition_templates"
}

func (e *ExpositionTemplate) BeforeSave(tx *gorm.DB) error {
	e.Title = strings.TrimSpace(e.Title)
	e.Description = strings.TrimSpace(e.Description)
	e.ImageURL = strings.TrimSpace(e.ImageURL)
	e.ThumbnailURL = strings.TrimSpace(e.ThumbnailURL)
	e.RewardMode = NormalizeRewardMode(string(e.RewardMode))
	e.RandomRewardSize = NormalizeRandomRewardSize(string(e.RandomRewardSize))
	if e.RequiredStoryFlags == nil {
		e.RequiredStoryFlags = StringArray{}
	}
	if e.Dialogue == nil {
		e.Dialogue = DialogueSequence{}
	}
	if e.MaterialRewards == nil {
		e.MaterialRewards = BaseMaterialRewards{}
	}
	if e.ItemRewards == nil {
		e.ItemRewards = ExpositionTemplateItemRewards{}
	}
	if e.SpellRewards == nil {
		e.SpellRewards = ExpositionTemplateSpellRewards{}
	}
	if e.RewardExperience < 0 {
		e.RewardExperience = 0
	}
	if e.RewardGold < 0 {
		e.RewardGold = 0
	}
	return nil
}
