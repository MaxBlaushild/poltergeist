package models

import (
	"time"

	"github.com/google/uuid"
)

type ScenarioOption struct {
	ID                        uuid.UUID                        `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                 time.Time                        `json:"createdAt"`
	UpdatedAt                 time.Time                        `json:"updatedAt"`
	ScenarioID                uuid.UUID                        `json:"scenarioId"`
	OptionText                string                           `json:"optionText" gorm:"column:option_text"`
	SuccessText               string                           `json:"successText" gorm:"column:success_text"`
	FailureText               string                           `json:"failureText" gorm:"column:failure_text"`
	StatTag                   string                           `json:"statTag" gorm:"column:stat_tag"`
	Proficiencies             StringArray                      `json:"proficiencies" gorm:"type:jsonb"`
	Difficulty                *int                             `json:"difficulty"`
	RewardExperience          int                              `json:"rewardExperience" gorm:"column:reward_experience"`
	RewardGold                int                              `json:"rewardGold" gorm:"column:reward_gold"`
	FailureHealthDrainType    ScenarioFailureDrainType         `json:"failureHealthDrainType" gorm:"column:failure_health_drain_type"`
	FailureHealthDrainValue   int                              `json:"failureHealthDrainValue" gorm:"column:failure_health_drain_value"`
	FailureManaDrainType      ScenarioFailureDrainType         `json:"failureManaDrainType" gorm:"column:failure_mana_drain_type"`
	FailureManaDrainValue     int                              `json:"failureManaDrainValue" gorm:"column:failure_mana_drain_value"`
	FailureStatuses           ScenarioFailureStatusTemplates   `json:"failureStatuses" gorm:"column:failure_statuses;type:jsonb"`
	SuccessHealthRestoreType  ScenarioFailureDrainType         `json:"successHealthRestoreType" gorm:"column:success_health_restore_type"`
	SuccessHealthRestoreValue int                              `json:"successHealthRestoreValue" gorm:"column:success_health_restore_value"`
	SuccessManaRestoreType    ScenarioFailureDrainType         `json:"successManaRestoreType" gorm:"column:success_mana_restore_type"`
	SuccessManaRestoreValue   int                              `json:"successManaRestoreValue" gorm:"column:success_mana_restore_value"`
	SuccessStatuses           ScenarioFailureStatusTemplates   `json:"successStatuses" gorm:"column:success_statuses;type:jsonb"`
	ItemRewards               []ScenarioOptionItemReward       `json:"itemRewards" gorm:"foreignKey:ScenarioOptionID"`
	ItemChoiceRewards         []ScenarioOptionItemChoiceReward `json:"itemChoiceRewards" gorm:"foreignKey:ScenarioOptionID"`
	SpellRewards              []ScenarioOptionSpellReward      `json:"spellRewards" gorm:"foreignKey:ScenarioOptionID"`
}

func (s *ScenarioOption) TableName() string {
	return "scenario_options"
}
