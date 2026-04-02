package models

import (
	"time"

	"github.com/google/uuid"
)

type Character struct {
	ID                         uuid.UUID           `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                  time.Time           `json:"createdAt"`
	UpdatedAt                  time.Time           `json:"updatedAt"`
	Name                       string              `json:"name"`
	Description                string              `json:"description"`
	InternalTags               StringArray         `json:"internalTags" gorm:"column:internal_tags;type:jsonb"`
	MapIconURL                 string              `json:"mapIconUrl"`
	DialogueImageURL           string              `json:"dialogueImageUrl"`
	ThumbnailURL               string              `json:"thumbnailUrl"`
	ImageGenerationStatus      string              `json:"imageGenerationStatus" gorm:"column:image_generation_status"`
	ImageGenerationError       *string             `json:"imageGenerationError,omitempty" gorm:"column:image_generation_error"`
	Locations                  []CharacterLocation `json:"locations" gorm:"foreignKey:CharacterID"`
	PointOfInterestID          *uuid.UUID          `json:"pointOfInterestId,omitempty" gorm:"type:uuid"`
	PointOfInterest            *PointOfInterest    `json:"pointOfInterest,omitempty" gorm:"foreignKey:PointOfInterestID"`
	HasAvailableQuest          bool                `json:"hasAvailableQuest" gorm:"-"`
	HasAvailableMainStoryQuest bool                `json:"hasAvailableMainStoryQuest" gorm:"-"`
}

func (n *Character) TableName() string {
	return "characters"
}

const (
	CharacterImageGenerationStatusNone       = "none"
	CharacterImageGenerationStatusQueued     = "queued"
	CharacterImageGenerationStatusInProgress = "in_progress"
	CharacterImageGenerationStatusComplete   = "complete"
	CharacterImageGenerationStatusFailed     = "failed"
)
