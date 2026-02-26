package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Scenario struct {
	ID               uuid.UUID            `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt        time.Time            `json:"createdAt"`
	UpdatedAt        time.Time            `json:"updatedAt"`
	ZoneID           uuid.UUID            `json:"zoneId"`
	Zone             Zone                 `json:"zone"`
	Latitude         float64              `json:"latitude"`
	Longitude        float64              `json:"longitude"`
	Geometry         string               `json:"geometry" gorm:"type:geometry(Point,4326)"`
	Prompt           string               `json:"prompt"`
	ImageURL         string               `json:"imageUrl" gorm:"column:image_url"`
	ThumbnailURL     string               `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	Difficulty       int                  `json:"difficulty"`
	RewardExperience int                  `json:"rewardExperience" gorm:"column:reward_experience"`
	RewardGold       int                  `json:"rewardGold" gorm:"column:reward_gold"`
	OpenEnded        bool                 `json:"openEnded" gorm:"column:open_ended"`
	Options          []ScenarioOption     `json:"options" gorm:"foreignKey:ScenarioID"`
	ItemRewards      []ScenarioItemReward `json:"itemRewards" gorm:"foreignKey:ScenarioID"`
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
