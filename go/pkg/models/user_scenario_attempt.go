package models

import (
	"time"

	"github.com/google/uuid"
)

type UserScenarioAttempt struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	UserID            uuid.UUID       `json:"userId"`
	User              User            `json:"user"`
	ScenarioID        uuid.UUID       `json:"scenarioId"`
	Scenario          Scenario        `json:"scenario"`
	ScenarioOptionID  *uuid.UUID      `json:"scenarioOptionId"`
	ScenarioOption    *ScenarioOption `json:"scenarioOption"`
	FreeformResponse  *string         `json:"freeformResponse"`
	Roll              int             `json:"roll"`
	StatTag           string          `json:"statTag"`
	StatValue         int             `json:"statValue"`
	ProficienciesUsed StringArray     `json:"proficienciesUsed" gorm:"column:proficiencies_used;type:jsonb"`
	ProficiencyBonus  int             `json:"proficiencyBonus"`
	ResponseScore     int             `json:"responseScore"`
	CreativityBonus   int             `json:"creativityBonus"`
	Threshold         int             `json:"threshold"`
	TotalScore        int             `json:"totalScore"`
	Successful        bool            `json:"successful"`
	Reasoning         *string         `json:"reasoning"`
	RewardExperience  int             `json:"rewardExperience"`
	RewardGold        int             `json:"rewardGold"`
}

func (u *UserScenarioAttempt) TableName() string {
	return "user_scenario_attempts"
}
