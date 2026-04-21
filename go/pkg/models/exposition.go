package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Exposition struct {
	ID                 uuid.UUID               `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt          time.Time               `json:"createdAt"`
	UpdatedAt          time.Time               `json:"updatedAt"`
	ZoneID             uuid.UUID               `json:"zoneId" gorm:"column:zone_id"`
	ZoneKind           string                  `json:"zoneKind,omitempty" gorm:"column:zone_kind"`
	Zone               Zone                    `json:"zone"`
	PointOfInterestID  *uuid.UUID              `json:"pointOfInterestId,omitempty" gorm:"column:point_of_interest_id;type:uuid"`
	PointOfInterest    *PointOfInterest        `json:"pointOfInterest,omitempty" gorm:"foreignKey:PointOfInterestID"`
	Latitude           float64                 `json:"latitude"`
	Longitude          float64                 `json:"longitude"`
	Geometry           string                  `json:"geometry" gorm:"type:geometry(Point,4326)"`
	Title              string                  `json:"title"`
	Description        string                  `json:"description"`
	Dialogue           DialogueSequence        `json:"dialogue" gorm:"type:jsonb"`
	RequiredStoryFlags StringArray             `json:"requiredStoryFlags" gorm:"column:required_story_flags;type:jsonb;default:'[]'"`
	ImageURL           string                  `json:"imageUrl" gorm:"column:image_url"`
	ThumbnailURL       string                  `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	RewardMode         RewardMode              `json:"rewardMode" gorm:"column:reward_mode"`
	RandomRewardSize   RandomRewardSize        `json:"randomRewardSize" gorm:"column:random_reward_size"`
	RewardExperience   int                     `json:"rewardExperience" gorm:"column:reward_experience"`
	RewardGold         int                     `json:"rewardGold" gorm:"column:reward_gold"`
	MaterialRewards    BaseMaterialRewards     `json:"materialRewards" gorm:"column:material_rewards_json;type:jsonb;default:'[]'"`
	ItemRewards        []ExpositionItemReward  `json:"itemRewards" gorm:"foreignKey:ExpositionID"`
	SpellRewards       []ExpositionSpellReward `json:"spellRewards" gorm:"foreignKey:ExpositionID"`
}

func (e *Exposition) TableName() string {
	return "expositions"
}

func (e *Exposition) BeforeSave(tx *gorm.DB) error {
	if e.RequiredStoryFlags == nil {
		e.RequiredStoryFlags = StringArray{}
	}
	if e.MaterialRewards == nil {
		e.MaterialRewards = BaseMaterialRewards{}
	}
	if e.Dialogue == nil {
		e.Dialogue = DialogueSequence{}
	}
	if e.Latitude != 0 || e.Longitude != 0 {
		if err := e.SetGeometry(e.Latitude, e.Longitude); err != nil {
			return err
		}
	}
	return nil
}

func (e *Exposition) SetGeometry(latitude, longitude float64) error {
	e.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", longitude, latitude)
	return nil
}
