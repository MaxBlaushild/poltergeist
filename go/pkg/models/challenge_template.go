package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ChallengeTemplateItemChoiceReward struct {
	InventoryItemID int `json:"inventoryItemId"`
	Quantity        int `json:"quantity"`
}

type ChallengeTemplateItemChoiceRewards []ChallengeTemplateItemChoiceReward

func (r ChallengeTemplateItemChoiceRewards) Value() (driver.Value, error) {
	if r == nil {
		return json.Marshal([]ChallengeTemplateItemChoiceReward{})
	}
	return json.Marshal(r)
}

func (r *ChallengeTemplateItemChoiceRewards) Scan(value interface{}) error {
	if value == nil {
		*r = ChallengeTemplateItemChoiceRewards{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan ChallengeTemplateItemChoiceRewards: value is not []byte")
	}
	var decoded []ChallengeTemplateItemChoiceReward
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*r = decoded
	return nil
}

type ChallengeTemplate struct {
	ID                  uuid.UUID                          `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt           time.Time                          `json:"createdAt"`
	UpdatedAt           time.Time                          `json:"updatedAt"`
	LocationArchetypeID uuid.UUID                          `json:"locationArchetypeId" gorm:"column:location_archetype_id;type:uuid"`
	LocationArchetype   *LocationArchetype                 `json:"locationArchetype,omitempty" gorm:"foreignKey:LocationArchetypeID"`
	Question            string                             `json:"question"`
	Description         string                             `json:"description"`
	ImageURL            string                             `json:"imageUrl" gorm:"column:image_url"`
	ThumbnailURL        string                             `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	ScaleWithUserLevel  bool                               `json:"scaleWithUserLevel" gorm:"column:scale_with_user_level"`
	RewardMode          RewardMode                         `json:"rewardMode" gorm:"column:reward_mode"`
	RandomRewardSize    RandomRewardSize                   `json:"randomRewardSize" gorm:"column:random_reward_size"`
	RewardExperience    int                                `json:"rewardExperience" gorm:"column:reward_experience"`
	Reward              int                                `json:"reward"`
	InventoryItemID     *int                               `json:"inventoryItemId" gorm:"column:inventory_item_id"`
	ItemChoiceRewards   ChallengeTemplateItemChoiceRewards `json:"itemChoiceRewards" gorm:"column:item_choice_rewards;type:jsonb"`
	SubmissionType      QuestNodeSubmissionType            `json:"submissionType" gorm:"type:text;default:photo"`
	Difficulty          int                                `json:"difficulty" gorm:"default:0"`
	StatTags            StringArray                        `json:"statTags,omitempty" gorm:"type:jsonb"`
	Proficiency         *string                            `json:"proficiency,omitempty"`
}

func (ChallengeTemplate) TableName() string {
	return "challenge_templates"
}
