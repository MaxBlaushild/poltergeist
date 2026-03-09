package models

import (
	"time"

	"github.com/google/uuid"
)

type UserScenarioItemChoicePending struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	UserID           uuid.UUID  `json:"userId" gorm:"column:user_id"`
	ScenarioID       uuid.UUID  `json:"scenarioId" gorm:"column:scenario_id"`
	ScenarioOptionID *uuid.UUID `json:"scenarioOptionId,omitempty" gorm:"column:scenario_option_id;type:uuid"`
}

func (u *UserScenarioItemChoicePending) TableName() string {
	return "user_scenario_item_choice_pendings"
}
