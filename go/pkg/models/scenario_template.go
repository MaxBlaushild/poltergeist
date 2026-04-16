package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ScenarioTemplateReward struct {
	InventoryItemID int `json:"inventoryItemId"`
	Quantity        int `json:"quantity"`
}

type ScenarioTemplateRewards []ScenarioTemplateReward

func (r ScenarioTemplateRewards) Value() (driver.Value, error) {
	if r == nil {
		return json.Marshal([]ScenarioTemplateReward{})
	}
	return json.Marshal(r)
}

func (r *ScenarioTemplateRewards) Scan(value interface{}) error {
	if value == nil {
		*r = ScenarioTemplateRewards{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan ScenarioTemplateRewards: value is not []byte")
	}
	var decoded []ScenarioTemplateReward
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*r = decoded
	return nil
}

type ScenarioTemplateSpellReward struct {
	SpellID uuid.UUID `json:"spellId"`
}

type ScenarioTemplateSpellRewards []ScenarioTemplateSpellReward

func (r ScenarioTemplateSpellRewards) Value() (driver.Value, error) {
	if r == nil {
		return json.Marshal([]ScenarioTemplateSpellReward{})
	}
	return json.Marshal(r)
}

func (r *ScenarioTemplateSpellRewards) Scan(value interface{}) error {
	if value == nil {
		*r = ScenarioTemplateSpellRewards{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan ScenarioTemplateSpellRewards: value is not []byte")
	}
	var decoded []ScenarioTemplateSpellReward
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*r = decoded
	return nil
}

type ScenarioTemplateOption struct {
	OptionText                string                         `json:"optionText"`
	SuccessText               string                         `json:"successText"`
	FailureText               string                         `json:"failureText"`
	SuccessHandoffText        string                         `json:"successHandoffText"`
	FailureHandoffText        string                         `json:"failureHandoffText"`
	StatTag                   string                         `json:"statTag"`
	Proficiencies             StringArray                    `json:"proficiencies"`
	Difficulty                *int                           `json:"difficulty"`
	RewardExperience          int                            `json:"rewardExperience"`
	RewardGold                int                            `json:"rewardGold"`
	FailureHealthDrainType    ScenarioFailureDrainType       `json:"failureHealthDrainType"`
	FailureHealthDrainValue   int                            `json:"failureHealthDrainValue"`
	FailureManaDrainType      ScenarioFailureDrainType       `json:"failureManaDrainType"`
	FailureManaDrainValue     int                            `json:"failureManaDrainValue"`
	FailureStatuses           ScenarioFailureStatusTemplates `json:"failureStatuses"`
	SuccessHealthRestoreType  ScenarioFailureDrainType       `json:"successHealthRestoreType"`
	SuccessHealthRestoreValue int                            `json:"successHealthRestoreValue"`
	SuccessManaRestoreType    ScenarioFailureDrainType       `json:"successManaRestoreType"`
	SuccessManaRestoreValue   int                            `json:"successManaRestoreValue"`
	SuccessStatuses           ScenarioFailureStatusTemplates `json:"successStatuses"`
	ItemRewards               ScenarioTemplateRewards        `json:"itemRewards"`
	ItemChoiceRewards         ScenarioTemplateRewards        `json:"itemChoiceRewards"`
	SpellRewards              ScenarioTemplateSpellRewards   `json:"spellRewards"`
}

type ScenarioTemplateOptions []ScenarioTemplateOption

func (o ScenarioTemplateOptions) Value() (driver.Value, error) {
	if o == nil {
		return json.Marshal([]ScenarioTemplateOption{})
	}
	return json.Marshal(o)
}

func (o *ScenarioTemplateOptions) Scan(value interface{}) error {
	if value == nil {
		*o = ScenarioTemplateOptions{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan ScenarioTemplateOptions: value is not []byte")
	}
	var decoded []ScenarioTemplateOption
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*o = decoded
	return nil
}

type ScenarioTemplate struct {
	ID                        uuid.UUID                      `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                 time.Time                      `json:"createdAt"`
	UpdatedAt                 time.Time                      `json:"updatedAt"`
	Prompt                    string                         `json:"prompt"`
	ImageURL                  string                         `json:"imageUrl" gorm:"column:image_url"`
	ThumbnailURL              string                         `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	ScaleWithUserLevel        bool                           `json:"scaleWithUserLevel" gorm:"column:scale_with_user_level"`
	RewardMode                RewardMode                     `json:"rewardMode" gorm:"column:reward_mode"`
	RandomRewardSize          RandomRewardSize               `json:"randomRewardSize" gorm:"column:random_reward_size"`
	Difficulty                int                            `json:"difficulty"`
	RewardExperience          int                            `json:"rewardExperience" gorm:"column:reward_experience"`
	RewardGold                int                            `json:"rewardGold" gorm:"column:reward_gold"`
	OpenEnded                 bool                           `json:"openEnded" gorm:"column:open_ended"`
	SuccessHandoffText        string                         `json:"successHandoffText" gorm:"column:success_handoff_text"`
	FailureHandoffText        string                         `json:"failureHandoffText" gorm:"column:failure_handoff_text"`
	FailurePenaltyMode        ScenarioFailurePenaltyMode     `json:"failurePenaltyMode" gorm:"column:failure_penalty_mode"`
	FailureHealthDrainType    ScenarioFailureDrainType       `json:"failureHealthDrainType" gorm:"column:failure_health_drain_type"`
	FailureHealthDrainValue   int                            `json:"failureHealthDrainValue" gorm:"column:failure_health_drain_value"`
	FailureManaDrainType      ScenarioFailureDrainType       `json:"failureManaDrainType" gorm:"column:failure_mana_drain_type"`
	FailureManaDrainValue     int                            `json:"failureManaDrainValue" gorm:"column:failure_mana_drain_value"`
	FailureStatuses           ScenarioFailureStatusTemplates `json:"failureStatuses" gorm:"column:failure_statuses;type:jsonb"`
	SuccessRewardMode         ScenarioSuccessRewardMode      `json:"successRewardMode" gorm:"column:success_reward_mode"`
	SuccessHealthRestoreType  ScenarioFailureDrainType       `json:"successHealthRestoreType" gorm:"column:success_health_restore_type"`
	SuccessHealthRestoreValue int                            `json:"successHealthRestoreValue" gorm:"column:success_health_restore_value"`
	SuccessManaRestoreType    ScenarioFailureDrainType       `json:"successManaRestoreType" gorm:"column:success_mana_restore_type"`
	SuccessManaRestoreValue   int                            `json:"successManaRestoreValue" gorm:"column:success_mana_restore_value"`
	SuccessStatuses           ScenarioFailureStatusTemplates `json:"successStatuses" gorm:"column:success_statuses;type:jsonb"`
	Options                   ScenarioTemplateOptions        `json:"options" gorm:"type:jsonb"`
	ItemRewards               ScenarioTemplateRewards        `json:"itemRewards" gorm:"column:item_rewards;type:jsonb"`
	ItemChoiceRewards         ScenarioTemplateRewards        `json:"itemChoiceRewards" gorm:"column:item_choice_rewards;type:jsonb"`
	SpellRewards              ScenarioTemplateSpellRewards   `json:"spellRewards" gorm:"column:spell_rewards;type:jsonb"`
}

func (ScenarioTemplate) TableName() string {
	return "scenario_templates"
}
