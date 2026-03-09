package models

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TutorialScenarioOption struct {
	OptionText       string                `json:"optionText"`
	StatTag          string                `json:"statTag"`
	Difficulty       int                   `json:"difficulty"`
	RewardExperience int                   `json:"rewardExperience"`
	RewardGold       int                   `json:"rewardGold"`
	ItemRewards      []TutorialItemReward  `json:"itemRewards"`
	SpellRewards     []TutorialSpellReward `json:"spellRewards"`
}

type TutorialItemReward struct {
	InventoryItemID int `json:"inventoryItemId"`
	Quantity        int `json:"quantity"`
}

type TutorialSpellReward struct {
	SpellID string `json:"spellId"`
}

type TutorialConfig struct {
	ID                    int                      `gorm:"primaryKey" json:"id"`
	CharacterID           *uuid.UUID               `json:"characterId"`
	Character             *Character               `json:"character,omitempty" gorm:"foreignKey:CharacterID"`
	DialogueJSON          datatypes.JSON           `gorm:"column:dialogue_json;type:jsonb;default:'[]'" json:"-"`
	Dialogue              []string                 `gorm:"-" json:"dialogue"`
	ScenarioPrompt        string                   `json:"scenarioPrompt"`
	ScenarioImageURL      string                   `gorm:"column:scenario_image_url" json:"scenarioImageUrl"`
	ImageGenerationStatus string                   `gorm:"column:image_generation_status" json:"imageGenerationStatus"`
	ImageGenerationError  *string                  `gorm:"column:image_generation_error" json:"imageGenerationError,omitempty"`
	OptionsJSON           datatypes.JSON           `gorm:"column:options_json;type:jsonb;default:'[]'" json:"-"`
	Options               []TutorialScenarioOption `gorm:"-" json:"options"`
	RewardExperience      int                      `gorm:"column:reward_experience" json:"rewardExperience"`
	RewardGold            int                      `gorm:"column:reward_gold" json:"rewardGold"`
	ItemRewardsJSON       datatypes.JSON           `gorm:"column:item_rewards_json;type:jsonb;default:'[]'" json:"-"`
	ItemRewards           []TutorialItemReward     `gorm:"-" json:"itemRewards"`
	SpellRewardsJSON      datatypes.JSON           `gorm:"column:spell_rewards_json;type:jsonb;default:'[]'" json:"-"`
	SpellRewards          []TutorialSpellReward    `gorm:"-" json:"spellRewards"`
	CreatedAt             time.Time                `json:"createdAt"`
	UpdatedAt             time.Time                `json:"updatedAt"`
}

const (
	TutorialImageGenerationStatusNone       = "none"
	TutorialImageGenerationStatusQueued     = "queued"
	TutorialImageGenerationStatusInProgress = "in_progress"
	TutorialImageGenerationStatusComplete   = "complete"
	TutorialImageGenerationStatusFailed     = "failed"
)

func (TutorialConfig) TableName() string {
	return "tutorial_configs"
}

func (c *TutorialConfig) BeforeSave(tx *gorm.DB) error {
	if c.Dialogue == nil {
		c.Dialogue = []string{}
	}
	if c.Options == nil {
		c.Options = []TutorialScenarioOption{}
	}
	if c.ItemRewards == nil {
		c.ItemRewards = []TutorialItemReward{}
	}
	if c.SpellRewards == nil {
		c.SpellRewards = []TutorialSpellReward{}
	}

	dialogue := make([]string, 0, len(c.Dialogue))
	for _, line := range c.Dialogue {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		dialogue = append(dialogue, trimmed)
	}
	c.Dialogue = dialogue

	options := make([]TutorialScenarioOption, 0, len(c.Options))
	for _, option := range c.Options {
		text := strings.TrimSpace(option.OptionText)
		statTag := strings.ToLower(strings.TrimSpace(option.StatTag))
		if text == "" || statTag == "" {
			continue
		}
		itemRewards := normalizeTutorialItemRewards(option.ItemRewards)
		spellRewards := normalizeTutorialSpellRewards(option.SpellRewards)
		if option.RewardExperience < 0 {
			option.RewardExperience = 0
		}
		if option.RewardGold < 0 {
			option.RewardGold = 0
		}
		if option.Difficulty < 0 {
			option.Difficulty = 0
		}
		options = append(options, TutorialScenarioOption{
			OptionText:       text,
			StatTag:          statTag,
			Difficulty:       option.Difficulty,
			RewardExperience: option.RewardExperience,
			RewardGold:       option.RewardGold,
			ItemRewards:      itemRewards,
			SpellRewards:     spellRewards,
		})
	}
	c.Options = options

	c.ItemRewards = normalizeTutorialItemRewards(c.ItemRewards)
	c.SpellRewards = normalizeTutorialSpellRewards(c.SpellRewards)

	if err := assignTutorialJSON(&c.DialogueJSON, c.Dialogue); err != nil {
		return err
	}
	if err := assignTutorialJSON(&c.OptionsJSON, c.Options); err != nil {
		return err
	}
	if err := assignTutorialJSON(&c.ItemRewardsJSON, c.ItemRewards); err != nil {
		return err
	}
	if err := assignTutorialJSON(&c.SpellRewardsJSON, c.SpellRewards); err != nil {
		return err
	}
	c.ScenarioPrompt = strings.TrimSpace(c.ScenarioPrompt)
	c.ScenarioImageURL = strings.TrimSpace(c.ScenarioImageURL)
	c.ImageGenerationStatus = strings.TrimSpace(c.ImageGenerationStatus)
	if c.ImageGenerationStatus == "" {
		c.ImageGenerationStatus = TutorialImageGenerationStatusNone
	}
	if c.ImageGenerationError != nil {
		trimmed := strings.TrimSpace(*c.ImageGenerationError)
		if trimmed == "" {
			c.ImageGenerationError = nil
		} else {
			c.ImageGenerationError = &trimmed
		}
	}
	if c.RewardExperience < 0 {
		c.RewardExperience = 0
	}
	if c.RewardGold < 0 {
		c.RewardGold = 0
	}
	return nil
}

func (c *TutorialConfig) AfterFind(tx *gorm.DB) error {
	if err := parseTutorialJSON(c.DialogueJSON, &c.Dialogue); err != nil {
		return err
	}
	if err := parseTutorialJSON(c.OptionsJSON, &c.Options); err != nil {
		return err
	}
	if err := parseTutorialJSON(c.ItemRewardsJSON, &c.ItemRewards); err != nil {
		return err
	}
	if err := parseTutorialJSON(c.SpellRewardsJSON, &c.SpellRewards); err != nil {
		return err
	}
	if c.Dialogue == nil {
		c.Dialogue = []string{}
	}
	if c.Options == nil {
		c.Options = []TutorialScenarioOption{}
	}
	for i := range c.Options {
		if c.Options[i].ItemRewards == nil {
			c.Options[i].ItemRewards = []TutorialItemReward{}
		}
		if c.Options[i].SpellRewards == nil {
			c.Options[i].SpellRewards = []TutorialSpellReward{}
		}
	}
	if c.ItemRewards == nil {
		c.ItemRewards = []TutorialItemReward{}
	}
	if c.SpellRewards == nil {
		c.SpellRewards = []TutorialSpellReward{}
	}
	return nil
}

func (c *TutorialConfig) IsConfigured() bool {
	return c != nil &&
		c.CharacterID != nil &&
		len(c.Dialogue) > 0 &&
		strings.TrimSpace(c.ScenarioPrompt) != "" &&
		len(c.Options) > 0
}

type UserTutorialState struct {
	UserID             uuid.UUID  `gorm:"primaryKey" json:"userId"`
	User               User       `json:"user"`
	TutorialScenarioID *uuid.UUID `gorm:"column:tutorial_scenario_id;type:uuid" json:"tutorialScenarioId"`
	TutorialScenario   *Scenario  `json:"tutorialScenario,omitempty" gorm:"foreignKey:TutorialScenarioID"`
	ActivatedAt        *time.Time `json:"activatedAt"`
	CompletedAt        *time.Time `json:"completedAt"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

func (UserTutorialState) TableName() string {
	return "user_tutorial_states"
}

func assignTutorialJSON(target *datatypes.JSON, value interface{}) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	*target = datatypes.JSON(raw)
	return nil
}

func parseTutorialJSON[T any](raw datatypes.JSON, target *[]T) error {
	if len(raw) == 0 {
		*target = []T{}
		return nil
	}
	return json.Unmarshal(raw, target)
}

func normalizeTutorialItemRewards(input []TutorialItemReward) []TutorialItemReward {
	rewards := make([]TutorialItemReward, 0, len(input))
	for _, reward := range input {
		if reward.InventoryItemID <= 0 || reward.Quantity <= 0 {
			continue
		}
		rewards = append(rewards, reward)
	}
	return rewards
}

func normalizeTutorialSpellRewards(input []TutorialSpellReward) []TutorialSpellReward {
	rewards := make([]TutorialSpellReward, 0, len(input))
	for _, reward := range input {
		spellID := strings.TrimSpace(reward.SpellID)
		if spellID == "" {
			continue
		}
		rewards = append(rewards, TutorialSpellReward{SpellID: spellID})
	}
	return rewards
}
