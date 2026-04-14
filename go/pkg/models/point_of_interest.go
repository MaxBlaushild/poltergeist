package models

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PointOfInterest struct {
	ID                         uuid.UUID                    `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                  time.Time                    `json:"createdAt"`
	UpdatedAt                  time.Time                    `json:"updatedAt"`
	Name                       string                       `json:"name"`
	OriginalName               string                       `json:"originalName"`
	Clue                       string                       `json:"clue"`
	Lat                        string                       `json:"lat"`
	Lng                        string                       `json:"lng"`
	ImageUrl                   string                       `json:"imageURL"`
	ThumbnailURL               string                       `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	ImageGenerationStatus      string                       `json:"imageGenerationStatus" gorm:"column:image_generation_status"`
	ImageGenerationError       *string                      `json:"imageGenerationError,omitempty" gorm:"column:image_generation_error"`
	Description                string                       `json:"description"`
	StoryVariants              PointOfInterestStoryVariants `json:"storyVariants" gorm:"column:story_variants;type:jsonb;default:'[]'"`
	PointOfInterestChallenges  []PointOfInterestChallenge   `json:"pointOfInterestChallenges"`
	Characters                 []Character                  `json:"characters" gorm:"foreignKey:PointOfInterestID"`
	Geometry                   string                       `json:"geometry" gorm:"type:geometry(Point,4326)"`
	Tags                       []Tag                        `json:"tags" gorm:"many2many:tag_entities;joinForeignKey:point_of_interest_id;joinReferences:tag_id"`
	RewardMode                 RewardMode                   `json:"rewardMode" gorm:"column:reward_mode"`
	RandomRewardSize           RandomRewardSize             `json:"randomRewardSize" gorm:"column:random_reward_size"`
	RewardExperience           int                          `json:"rewardExperience" gorm:"column:reward_experience"`
	RewardGold                 int                          `json:"rewardGold" gorm:"column:reward_gold"`
	MaterialRewards            BaseMaterialRewards          `json:"materialRewards" gorm:"column:material_rewards_json;type:jsonb;default:'[]'"`
	ItemRewards                []PointOfInterestItemReward  `json:"itemRewards" gorm:"foreignKey:PointOfInterestID"`
	SpellRewards               []PointOfInterestSpellReward `json:"spellRewards" gorm:"foreignKey:PointOfInterestID"`
	GoogleMapsPlaceID          *string                      `json:"googleMapsPlaceId"`
	GoogleMapsPlaceName        *string                      `json:"googleMapsPlaceName"`
	LastUsedInQuestAt          *time.Time                   `json:"lastUsedInQuestAt,omitempty"`
	UnlockTier                 *int                         `json:"unlockTier" gorm:"column:unlock_tier"`
	HasAvailableQuest          bool                         `json:"hasAvailableQuest" gorm:"-"`
	HasAvailableMainStoryQuest bool                         `json:"hasAvailableMainStoryQuest" gorm:"-"`
}

func (p *PointOfInterest) TableName() string {
	return "points_of_interest"
}

const (
	PointOfInterestImageGenerationStatusNone       = "none"
	PointOfInterestImageGenerationStatusQueued     = "queued"
	PointOfInterestImageGenerationStatusInProgress = "in_progress"
	PointOfInterestImageGenerationStatusComplete   = "complete"
	PointOfInterestImageGenerationStatusFailed     = "failed"
)

func (p *PointOfInterest) BeforeSave(tx *gorm.DB) error {
	if p.MaterialRewards == nil {
		p.MaterialRewards = BaseMaterialRewards{}
	}
	if p.Lat != "" && p.Lng != "" {
		if err := p.SetGeometry(p.Lat, p.Lng); err != nil {
			return err
		}
	}
	return nil
}

func (p *PointOfInterest) SetGeometry(lat string, lng string) error {
	floatLat, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		return err
	}
	floatLng, err := strconv.ParseFloat(lng, 64)
	if err != nil {
		return err
	}

	// Create WKT (Well-Known Text) format: 'SRID=4326;POINT(lng lat)'
	p.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", floatLng, floatLat)
	return nil
}
