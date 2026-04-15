package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Character struct {
	ID                         uuid.UUID                   `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                  time.Time                   `json:"createdAt"`
	UpdatedAt                  time.Time                   `json:"updatedAt"`
	Name                       string                      `json:"name"`
	Description                string                      `json:"description"`
	InternalTags               StringArray                 `json:"internalTags" gorm:"column:internal_tags;type:jsonb"`
	MapIconURL                 string                      `json:"mapIconUrl"`
	DialogueImageURL           string                      `json:"dialogueImageUrl"`
	ThumbnailURL               string                      `json:"thumbnailUrl"`
	OwnerUserID                *uuid.UUID                  `json:"ownerUserId,omitempty" gorm:"column:owner_user_id;type:uuid"`
	Ephemeral                  bool                        `json:"ephemeral" gorm:"column:ephemeral"`
	StoryVariants              CharacterStoryVariants      `json:"storyVariants" gorm:"column:story_variants;type:jsonb;default:'[]'"`
	Relationship               *CharacterRelationshipState `json:"relationship,omitempty" gorm:"-"`
	ImageGenerationStatus      string                      `json:"imageGenerationStatus" gorm:"column:image_generation_status"`
	ImageGenerationError       *string                     `json:"imageGenerationError,omitempty" gorm:"column:image_generation_error"`
	Locations                  []CharacterLocation         `json:"locations" gorm:"foreignKey:CharacterID"`
	PointOfInterestID          *uuid.UUID                  `json:"pointOfInterestId,omitempty" gorm:"type:uuid"`
	PointOfInterest            *PointOfInterest            `json:"pointOfInterest,omitempty" gorm:"foreignKey:PointOfInterestID"`
	HasAvailableQuest          bool                        `json:"hasAvailableQuest" gorm:"-"`
	HasAvailableMainStoryQuest bool                        `json:"hasAvailableMainStoryQuest" gorm:"-"`
}

const CharacterInternalTagGeneratedFetchQuest = "generated_fetch_quest_character"
const CharacterInternalTagGeneratedPOILocal = "generated_poi_local"

func CharacterHasInternalTag(character *Character, tag string) bool {
	if character == nil || tag == "" {
		return false
	}
	for _, existing := range character.InternalTags {
		if existing == tag {
			return true
		}
	}
	return false
}

func (n *Character) TableName() string {
	return "characters"
}

func (n *Character) BeforeSave(tx *gorm.DB) error {
	if n.InternalTags == nil {
		n.InternalTags = StringArray{}
	}
	if n.StoryVariants == nil {
		n.StoryVariants = CharacterStoryVariants{}
	}
	return nil
}

const (
	CharacterImageGenerationStatusNone       = "none"
	CharacterImageGenerationStatusQueued     = "queued"
	CharacterImageGenerationStatusInProgress = "in_progress"
	CharacterImageGenerationStatusComplete   = "complete"
	CharacterImageGenerationStatusFailed     = "failed"
)
