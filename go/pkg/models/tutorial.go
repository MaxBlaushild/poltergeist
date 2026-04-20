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
	ID                                int                      `gorm:"primaryKey" json:"id"`
	CharacterID                       *uuid.UUID               `json:"characterId"`
	Character                         *Character               `json:"character,omitempty" gorm:"foreignKey:CharacterID"`
	DialogueJSON                      datatypes.JSON           `gorm:"column:dialogue_json;type:jsonb;default:'[]'" json:"-"`
	Dialogue                          DialogueSequence         `gorm:"-" json:"dialogue"`
	LoadoutDialogueJSON               datatypes.JSON           `gorm:"column:loadout_dialogue_json;type:jsonb;default:'[]'" json:"-"`
	LoadoutDialogue                   DialogueSequence         `gorm:"-" json:"loadoutDialogue"`
	LoadoutObjectiveCopy              string                   `gorm:"column:loadout_objective_copy" json:"loadoutObjectiveCopy"`
	BaseQuestArchetypeID              *uuid.UUID               `json:"baseQuestArchetypeId" gorm:"column:base_quest_archetype_id;type:uuid"`
	BaseQuestArchetype                *QuestArchetype          `json:"baseQuestArchetype,omitempty" gorm:"foreignKey:BaseQuestArchetypeID"`
	BaseQuestGiverCharacterID         *uuid.UUID               `json:"baseQuestGiverCharacterId" gorm:"column:base_quest_giver_character_id;type:uuid"`
	BaseQuestGiverCharacter           *Character               `json:"baseQuestGiverCharacter,omitempty" gorm:"foreignKey:BaseQuestGiverCharacterID"`
	BaseQuestGiverCharacterTemplateID *uuid.UUID               `json:"baseQuestGiverCharacterTemplateId" gorm:"column:base_quest_giver_character_template_id;type:uuid"`
	BaseQuestGiverCharacterTemplate   *CharacterTemplate       `json:"baseQuestGiverCharacterTemplate,omitempty" gorm:"foreignKey:BaseQuestGiverCharacterTemplateID"`
	PostMonsterDialogueJSON           datatypes.JSON           `gorm:"column:post_monster_dialogue_json;type:jsonb;default:'[]'" json:"-"`
	PostMonsterDialogue               DialogueSequence         `gorm:"-" json:"postMonsterDialogue"`
	BaseKitDialogueJSON               datatypes.JSON           `gorm:"column:base_kit_dialogue_json;type:jsonb;default:'[]'" json:"-"`
	BaseKitDialogue                   DialogueSequence         `gorm:"-" json:"baseKitDialogue"`
	BaseKitObjectiveCopy              string                   `gorm:"column:base_kit_objective_copy" json:"baseKitObjectiveCopy"`
	PostBasePlacementDialogueJSON     datatypes.JSON           `gorm:"column:post_base_placement_dialogue_json;type:jsonb;default:'[]'" json:"-"`
	PostBasePlacementDialogue         DialogueSequence         `gorm:"-" json:"postBasePlacementDialogue"`
	HearthObjectiveCopy               string                   `gorm:"column:hearth_objective_copy" json:"hearthObjectiveCopy"`
	PostBaseDialogueJSON              datatypes.JSON           `gorm:"column:post_base_dialogue_json;type:jsonb;default:'[]'" json:"-"`
	PostBaseDialogue                  DialogueSequence         `gorm:"-" json:"postBaseDialogue"`
	ScenarioPrompt                    string                   `json:"scenarioPrompt"`
	ScenarioImageURL                  string                   `gorm:"column:scenario_image_url" json:"scenarioImageUrl"`
	ImageGenerationStatus             string                   `gorm:"column:image_generation_status" json:"imageGenerationStatus"`
	ImageGenerationError              *string                  `gorm:"column:image_generation_error" json:"imageGenerationError,omitempty"`
	OptionsJSON                       datatypes.JSON           `gorm:"column:options_json;type:jsonb;default:'[]'" json:"-"`
	Options                           []TutorialScenarioOption `gorm:"-" json:"options"`
	MonsterEncounterID                *uuid.UUID               `gorm:"column:monster_encounter_id;type:uuid" json:"monsterEncounterId"`
	MonsterEncounter                  *MonsterEncounter        `json:"monsterEncounter,omitempty" gorm:"foreignKey:MonsterEncounterID"`
	MonsterRewardExperience           int                      `gorm:"column:monster_reward_experience" json:"monsterRewardExperience"`
	MonsterRewardGold                 int                      `gorm:"column:monster_reward_gold" json:"monsterRewardGold"`
	MonsterItemRewardsJSON            datatypes.JSON           `gorm:"column:monster_item_rewards_json;type:jsonb;default:'[]'" json:"-"`
	MonsterItemRewards                []TutorialItemReward     `gorm:"-" json:"monsterItemRewards"`
	RewardExperience                  int                      `gorm:"column:reward_experience" json:"rewardExperience"`
	RewardGold                        int                      `gorm:"column:reward_gold" json:"rewardGold"`
	ItemRewardsJSON                   datatypes.JSON           `gorm:"column:item_rewards_json;type:jsonb;default:'[]'" json:"-"`
	ItemRewards                       []TutorialItemReward     `gorm:"-" json:"itemRewards"`
	SpellRewardsJSON                  datatypes.JSON           `gorm:"column:spell_rewards_json;type:jsonb;default:'[]'" json:"-"`
	SpellRewards                      []TutorialSpellReward    `gorm:"-" json:"spellRewards"`
	CreatedAt                         time.Time                `json:"createdAt"`
	UpdatedAt                         time.Time                `json:"updatedAt"`
}

const (
	TutorialImageGenerationStatusNone       = "none"
	TutorialImageGenerationStatusQueued     = "queued"
	TutorialImageGenerationStatusInProgress = "in_progress"
	TutorialImageGenerationStatusComplete   = "complete"
	TutorialImageGenerationStatusFailed     = "failed"
)

const (
	TutorialStageWelcome             = "welcome"
	TutorialStageScenario            = "scenario"
	TutorialStageLoadout             = "loadout"
	TutorialStageMonster             = "monster"
	TutorialStagePostMonsterDialogue = "post_monster_dialogue"
	TutorialStageBaseKit             = "base_kit"
	TutorialStagePostBasePlacement   = "post_base_placement_dialogue"
	TutorialStageHearth              = "hearth"
	TutorialStagePostBaseDialogue    = "post_base_dialogue"
	TutorialStageCompleted           = "completed"
)

func (TutorialConfig) TableName() string {
	return "tutorial_configs"
}

func (c *TutorialConfig) BeforeSave(tx *gorm.DB) error {
	if c.Dialogue == nil {
		c.Dialogue = DialogueSequence{}
	}
	if c.LoadoutDialogue == nil {
		c.LoadoutDialogue = DialogueSequence{}
	}
	if c.PostMonsterDialogue == nil {
		c.PostMonsterDialogue = DialogueSequence{}
	}
	if c.BaseKitDialogue == nil {
		c.BaseKitDialogue = DialogueSequence{}
	}
	if c.PostBasePlacementDialogue == nil {
		c.PostBasePlacementDialogue = DialogueSequence{}
	}
	if c.PostBaseDialogue == nil {
		c.PostBaseDialogue = DialogueSequence{}
	}
	if c.Options == nil {
		c.Options = []TutorialScenarioOption{}
	}
	if c.MonsterItemRewards == nil {
		c.MonsterItemRewards = []TutorialItemReward{}
	}
	if c.ItemRewards == nil {
		c.ItemRewards = []TutorialItemReward{}
	}
	if c.SpellRewards == nil {
		c.SpellRewards = []TutorialSpellReward{}
	}

	c.Dialogue = normalizeDialogueSequence(c.Dialogue)
	c.LoadoutDialogue = normalizeDialogueSequence(c.LoadoutDialogue)
	c.PostMonsterDialogue = normalizeDialogueSequence(c.PostMonsterDialogue)
	c.BaseKitDialogue = normalizeDialogueSequence(c.BaseKitDialogue)
	c.PostBasePlacementDialogue = normalizeDialogueSequence(c.PostBasePlacementDialogue)
	c.PostBaseDialogue = normalizeDialogueSequence(c.PostBaseDialogue)

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

	c.MonsterItemRewards = normalizeTutorialItemRewards(c.MonsterItemRewards)
	c.ItemRewards = normalizeTutorialItemRewards(c.ItemRewards)
	c.SpellRewards = normalizeTutorialSpellRewards(c.SpellRewards)

	if err := assignTutorialJSON(&c.DialogueJSON, c.Dialogue); err != nil {
		return err
	}
	if err := assignTutorialJSON(&c.LoadoutDialogueJSON, c.LoadoutDialogue); err != nil {
		return err
	}
	if err := assignTutorialJSON(&c.PostMonsterDialogueJSON, c.PostMonsterDialogue); err != nil {
		return err
	}
	if err := assignTutorialJSON(&c.BaseKitDialogueJSON, c.BaseKitDialogue); err != nil {
		return err
	}
	if err := assignTutorialJSON(&c.PostBasePlacementDialogueJSON, c.PostBasePlacementDialogue); err != nil {
		return err
	}
	if err := assignTutorialJSON(&c.PostBaseDialogueJSON, c.PostBaseDialogue); err != nil {
		return err
	}
	if err := assignTutorialJSON(&c.OptionsJSON, c.Options); err != nil {
		return err
	}
	if err := assignTutorialJSON(&c.MonsterItemRewardsJSON, c.MonsterItemRewards); err != nil {
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
	c.LoadoutObjectiveCopy = strings.TrimSpace(c.LoadoutObjectiveCopy)
	c.BaseKitObjectiveCopy = strings.TrimSpace(c.BaseKitObjectiveCopy)
	c.HearthObjectiveCopy = strings.TrimSpace(c.HearthObjectiveCopy)
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
	if c.MonsterRewardExperience < 0 {
		c.MonsterRewardExperience = 0
	}
	if c.MonsterRewardGold < 0 {
		c.MonsterRewardGold = 0
	}
	return nil
}

func (c *TutorialConfig) AfterFind(tx *gorm.DB) error {
	if err := parseTutorialDialogueSequenceJSON(c.DialogueJSON, &c.Dialogue); err != nil {
		return err
	}
	if err := parseTutorialDialogueSequenceJSON(c.LoadoutDialogueJSON, &c.LoadoutDialogue); err != nil {
		return err
	}
	if err := parseTutorialDialogueSequenceJSON(c.PostMonsterDialogueJSON, &c.PostMonsterDialogue); err != nil {
		return err
	}
	if err := parseTutorialDialogueSequenceJSON(c.BaseKitDialogueJSON, &c.BaseKitDialogue); err != nil {
		return err
	}
	if err := parseTutorialDialogueSequenceJSON(c.PostBasePlacementDialogueJSON, &c.PostBasePlacementDialogue); err != nil {
		return err
	}
	if err := parseTutorialDialogueSequenceJSON(c.PostBaseDialogueJSON, &c.PostBaseDialogue); err != nil {
		return err
	}
	if err := parseTutorialJSON(c.OptionsJSON, &c.Options); err != nil {
		return err
	}
	if err := parseTutorialJSON(c.MonsterItemRewardsJSON, &c.MonsterItemRewards); err != nil {
		return err
	}
	if err := parseTutorialJSON(c.ItemRewardsJSON, &c.ItemRewards); err != nil {
		return err
	}
	if err := parseTutorialJSON(c.SpellRewardsJSON, &c.SpellRewards); err != nil {
		return err
	}
	if c.Dialogue == nil {
		c.Dialogue = DialogueSequence{}
	}
	if c.LoadoutDialogue == nil {
		c.LoadoutDialogue = DialogueSequence{}
	}
	if c.PostMonsterDialogue == nil {
		c.PostMonsterDialogue = DialogueSequence{}
	}
	if c.BaseKitDialogue == nil {
		c.BaseKitDialogue = DialogueSequence{}
	}
	if c.PostBasePlacementDialogue == nil {
		c.PostBasePlacementDialogue = DialogueSequence{}
	}
	if c.PostBaseDialogue == nil {
		c.PostBaseDialogue = DialogueSequence{}
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
	if c.MonsterItemRewards == nil {
		c.MonsterItemRewards = []TutorialItemReward{}
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
	UserID                     uuid.UUID         `gorm:"primaryKey" json:"userId"`
	User                       User              `json:"user"`
	Stage                      string            `gorm:"column:stage" json:"stage"`
	TutorialScenarioID         *uuid.UUID        `gorm:"column:tutorial_scenario_id;type:uuid" json:"tutorialScenarioId"`
	TutorialScenario           *Scenario         `json:"tutorialScenario,omitempty" gorm:"foreignKey:TutorialScenarioID"`
	SelectedScenarioOptionID   *uuid.UUID        `gorm:"column:selected_scenario_option_id;type:uuid" json:"selectedScenarioOptionId"`
	RequiredEquipItemIDsJSON   datatypes.JSON    `gorm:"column:required_equip_item_ids_json;type:jsonb;default:'[]'" json:"-"`
	RequiredEquipItemIDs       []int             `gorm:"-" json:"requiredEquipItemIds"`
	CompletedEquipItemIDsJSON  datatypes.JSON    `gorm:"column:completed_equip_item_ids_json;type:jsonb;default:'[]'" json:"-"`
	CompletedEquipItemIDs      []int             `gorm:"-" json:"completedEquipItemIds"`
	RequiredUseItemIDsJSON     datatypes.JSON    `gorm:"column:required_use_item_ids_json;type:jsonb;default:'[]'" json:"-"`
	RequiredUseItemIDs         []int             `gorm:"-" json:"requiredUseItemIds"`
	CompletedUseItemIDsJSON    datatypes.JSON    `gorm:"column:completed_use_item_ids_json;type:jsonb;default:'[]'" json:"-"`
	CompletedUseItemIDs        []int             `gorm:"-" json:"completedUseItemIds"`
	TutorialMonsterEncounterID *uuid.UUID        `gorm:"column:tutorial_monster_encounter_id;type:uuid" json:"tutorialMonsterEncounterId"`
	TutorialMonsterEncounter   *MonsterEncounter `json:"tutorialMonsterEncounter,omitempty" gorm:"foreignKey:TutorialMonsterEncounterID"`
	ActivatedAt                *time.Time        `json:"activatedAt"`
	CompletedAt                *time.Time        `json:"completedAt"`
	CreatedAt                  time.Time         `json:"createdAt"`
	UpdatedAt                  time.Time         `json:"updatedAt"`
}

func (UserTutorialState) TableName() string {
	return "user_tutorial_states"
}

func (s *UserTutorialState) BeforeSave(tx *gorm.DB) error {
	s.Stage = normalizeTutorialStage(s.Stage)
	s.RequiredEquipItemIDs = normalizeTutorialIntList(s.RequiredEquipItemIDs)
	s.CompletedEquipItemIDs = normalizeTutorialIntList(s.CompletedEquipItemIDs)
	s.RequiredUseItemIDs = normalizeTutorialIntList(s.RequiredUseItemIDs)
	s.CompletedUseItemIDs = normalizeTutorialIntList(s.CompletedUseItemIDs)
	if err := assignTutorialJSON(&s.RequiredEquipItemIDsJSON, s.RequiredEquipItemIDs); err != nil {
		return err
	}
	if err := assignTutorialJSON(&s.CompletedEquipItemIDsJSON, s.CompletedEquipItemIDs); err != nil {
		return err
	}
	if err := assignTutorialJSON(&s.RequiredUseItemIDsJSON, s.RequiredUseItemIDs); err != nil {
		return err
	}
	if err := assignTutorialJSON(&s.CompletedUseItemIDsJSON, s.CompletedUseItemIDs); err != nil {
		return err
	}
	return nil
}

func (s *UserTutorialState) AfterFind(tx *gorm.DB) error {
	if err := parseTutorialJSON(s.RequiredEquipItemIDsJSON, &s.RequiredEquipItemIDs); err != nil {
		return err
	}
	if err := parseTutorialJSON(s.CompletedEquipItemIDsJSON, &s.CompletedEquipItemIDs); err != nil {
		return err
	}
	if err := parseTutorialJSON(s.RequiredUseItemIDsJSON, &s.RequiredUseItemIDs); err != nil {
		return err
	}
	if err := parseTutorialJSON(s.CompletedUseItemIDsJSON, &s.CompletedUseItemIDs); err != nil {
		return err
	}
	s.Stage = normalizeTutorialStage(s.Stage)
	s.RequiredEquipItemIDs = normalizeTutorialIntList(s.RequiredEquipItemIDs)
	s.CompletedEquipItemIDs = normalizeTutorialIntList(s.CompletedEquipItemIDs)
	s.RequiredUseItemIDs = normalizeTutorialIntList(s.RequiredUseItemIDs)
	s.CompletedUseItemIDs = normalizeTutorialIntList(s.CompletedUseItemIDs)
	return nil
}

func (s *UserTutorialState) HasOutstandingLoadoutRequirements() bool {
	if s == nil {
		return false
	}
	return len(tutorialIncompleteItems(s.RequiredEquipItemIDs, s.CompletedEquipItemIDs)) > 0 ||
		len(tutorialIncompleteItems(s.RequiredUseItemIDs, s.CompletedUseItemIDs)) > 0
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

func parseTutorialDialogueSequenceJSON(raw datatypes.JSON, target *DialogueSequence) error {
	if len(raw) == 0 {
		*target = DialogueSequence{}
		return nil
	}

	var messages []DialogueMessage
	if err := json.Unmarshal(raw, &messages); err == nil {
		*target = normalizeDialogueSequence(messages)
		return nil
	}

	var legacy []string
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return err
	}
	*target = DialogueSequenceFromStringLines(legacy)
	return nil
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

func normalizeTutorialStage(input string) string {
	switch strings.TrimSpace(strings.ToLower(input)) {
	case TutorialStageScenario:
		return TutorialStageScenario
	case TutorialStageLoadout:
		return TutorialStageLoadout
	case TutorialStageMonster:
		return TutorialStageMonster
	case TutorialStagePostMonsterDialogue:
		return TutorialStagePostMonsterDialogue
	case TutorialStageBaseKit:
		return TutorialStageBaseKit
	case TutorialStagePostBasePlacement:
		return TutorialStagePostBasePlacement
	case TutorialStageHearth:
		return TutorialStageHearth
	case TutorialStagePostBaseDialogue:
		return TutorialStagePostBaseDialogue
	case TutorialStageCompleted:
		return TutorialStageCompleted
	default:
		return TutorialStageWelcome
	}
}

func normalizeTutorialIntList(input []int) []int {
	if len(input) == 0 {
		return []int{}
	}
	seen := map[int]struct{}{}
	values := make([]int, 0, len(input))
	for _, value := range input {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values
}

func tutorialIncompleteItems(required []int, completed []int) []int {
	if len(required) == 0 {
		return []int{}
	}
	done := map[int]struct{}{}
	for _, value := range completed {
		done[value] = struct{}{}
	}
	remaining := make([]int, 0, len(required))
	for _, value := range required {
		if _, ok := done[value]; ok {
			continue
		}
		remaining = append(remaining, value)
	}
	return remaining
}
