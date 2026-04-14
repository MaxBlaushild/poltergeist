package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CharacterTemplate struct {
	ID                    uuid.UUID              `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt             time.Time              `json:"createdAt"`
	UpdatedAt             time.Time              `json:"updatedAt"`
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	InternalTags          StringArray            `json:"internalTags" gorm:"column:internal_tags;type:jsonb"`
	MapIconURL            string                 `json:"mapIconUrl" gorm:"column:map_icon_url"`
	DialogueImageURL      string                 `json:"dialogueImageUrl" gorm:"column:dialogue_image_url"`
	ThumbnailURL          string                 `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	StoryVariants         CharacterStoryVariants `json:"storyVariants" gorm:"column:story_variants;type:jsonb;default:'[]'"`
	ImageGenerationStatus string                 `json:"imageGenerationStatus" gorm:"column:image_generation_status"`
	ImageGenerationError  *string                `json:"imageGenerationError,omitempty" gorm:"column:image_generation_error"`
}

func (CharacterTemplate) TableName() string {
	return "character_templates"
}

func (c *CharacterTemplate) BeforeSave(tx *gorm.DB) error {
	c.Name = strings.TrimSpace(c.Name)
	c.Description = strings.TrimSpace(c.Description)
	c.MapIconURL = strings.TrimSpace(c.MapIconURL)
	c.DialogueImageURL = strings.TrimSpace(c.DialogueImageURL)
	c.ThumbnailURL = strings.TrimSpace(c.ThumbnailURL)
	c.ImageGenerationStatus = strings.TrimSpace(c.ImageGenerationStatus)
	if c.InternalTags == nil {
		c.InternalTags = StringArray{}
	}
	if c.StoryVariants == nil {
		c.StoryVariants = CharacterStoryVariants{}
	}
	if c.ImageGenerationError != nil {
		trimmed := strings.TrimSpace(*c.ImageGenerationError)
		if trimmed == "" {
			c.ImageGenerationError = nil
		} else {
			c.ImageGenerationError = &trimmed
		}
	}
	if c.ImageGenerationStatus == "" {
		if c.DialogueImageURL != "" {
			c.ImageGenerationStatus = CharacterImageGenerationStatusComplete
		} else {
			c.ImageGenerationStatus = CharacterImageGenerationStatusNone
		}
	}
	return nil
}
