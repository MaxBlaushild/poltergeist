package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Scenario struct {
	ID                        uuid.UUID                      `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                 time.Time                      `json:"createdAt"`
	UpdatedAt                 time.Time                      `json:"updatedAt"`
	ZoneID                    uuid.UUID                      `json:"zoneId"`
	Zone                      Zone                           `json:"zone"`
	Latitude                  float64                        `json:"latitude"`
	Longitude                 float64                        `json:"longitude"`
	Geometry                  string                         `json:"geometry" gorm:"type:geometry(Point,4326)"`
	Prompt                    string                         `json:"prompt"`
	ImageURL                  string                         `json:"imageUrl" gorm:"column:image_url"`
	ThumbnailURL              string                         `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	Difficulty                int                            `json:"difficulty"`
	RewardExperience          int                            `json:"rewardExperience" gorm:"column:reward_experience"`
	RewardGold                int                            `json:"rewardGold" gorm:"column:reward_gold"`
	OpenEnded                 bool                           `json:"openEnded" gorm:"column:open_ended"`
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
	Options                   []ScenarioOption               `json:"options" gorm:"foreignKey:ScenarioID"`
	ItemRewards               []ScenarioItemReward           `json:"itemRewards" gorm:"foreignKey:ScenarioID"`
	SpellRewards              []ScenarioSpellReward          `json:"spellRewards" gorm:"foreignKey:ScenarioID"`
}

func (s *Scenario) TableName() string {
	return "scenarios"
}

func (s *Scenario) BeforeSave(tx *gorm.DB) error {
	if s.Latitude != 0 || s.Longitude != 0 {
		if err := s.SetGeometry(s.Latitude, s.Longitude); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scenario) SetGeometry(latitude, longitude float64) error {
	s.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", longitude, latitude)
	return nil
}
