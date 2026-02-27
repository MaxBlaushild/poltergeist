package models

import (
	"time"

	"github.com/google/uuid"
)

type ScenarioOption struct {
	ID               uuid.UUID                  `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt        time.Time                  `json:"createdAt"`
	UpdatedAt        time.Time                  `json:"updatedAt"`
	ScenarioID       uuid.UUID                  `json:"scenarioId"`
	OptionText       string                     `json:"optionText" gorm:"column:option_text"`
	SuccessText      string                     `json:"successText" gorm:"column:success_text"`
	FailureText      string                     `json:"failureText" gorm:"column:failure_text"`
	StatTag          string                     `json:"statTag" gorm:"column:stat_tag"`
	Proficiencies    StringArray                `json:"proficiencies" gorm:"type:jsonb"`
	Difficulty       *int                       `json:"difficulty"`
	RewardExperience int                        `json:"rewardExperience" gorm:"column:reward_experience"`
	RewardGold       int                        `json:"rewardGold" gorm:"column:reward_gold"`
	ItemRewards      []ScenarioOptionItemReward `json:"itemRewards" gorm:"foreignKey:ScenarioOptionID"`
}

func (s *ScenarioOption) TableName() string {
	return "scenario_options"
}
