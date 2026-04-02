package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	StoryWorldChangeTypeMoveCharacter = "move_character"
	StoryWorldChangeTypeShowPOIText   = "show_poi_text"
)

func NormalizeStoryWorldChangeType(value string) string {
	switch value {
	case StoryWorldChangeTypeMoveCharacter:
		return StoryWorldChangeTypeMoveCharacter
	case StoryWorldChangeTypeShowPOIText:
		return StoryWorldChangeTypeShowPOIText
	default:
		return ""
	}
}

type MainStoryWorldChange struct {
	Type                string      `json:"type"`
	TargetKey           string      `json:"targetKey"`
	CharacterKey        string      `json:"characterKey,omitempty"`
	PointOfInterestHint string      `json:"pointOfInterestHint,omitempty"`
	DestinationHint     string      `json:"destinationHint,omitempty"`
	ZoneTags            StringArray `json:"zoneTags,omitempty"`
	Description         string      `json:"description,omitempty"`
	Clue                string      `json:"clue,omitempty"`
}

type StoryWorldChange struct {
	ID                           uuid.UUID   `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                    time.Time   `json:"createdAt"`
	UpdatedAt                    time.Time   `json:"updatedAt"`
	MainStoryTemplateID          uuid.UUID   `json:"mainStoryTemplateId" gorm:"column:main_story_template_id;type:uuid;index"`
	QuestArchetypeID             *uuid.UUID  `json:"questArchetypeId,omitempty" gorm:"column:quest_archetype_id;type:uuid"`
	BeatOrder                    int         `json:"beatOrder" gorm:"column:beat_order"`
	Priority                     int         `json:"priority"`
	EffectType                   string      `json:"effectType" gorm:"column:effect_type"`
	TargetKey                    string      `json:"targetKey" gorm:"column:target_key"`
	RequiredStoryFlags           StringArray `json:"requiredStoryFlags" gorm:"column:required_story_flags;type:jsonb;default:'[]'"`
	CharacterID                  *uuid.UUID  `json:"characterId,omitempty" gorm:"column:character_id;type:uuid"`
	PointOfInterestID            *uuid.UUID  `json:"pointOfInterestId,omitempty" gorm:"column:point_of_interest_id;type:uuid"`
	DestinationPointOfInterestID *uuid.UUID  `json:"destinationPointOfInterestId,omitempty" gorm:"column:destination_point_of_interest_id;type:uuid"`
	Description                  string      `json:"description"`
	Clue                         string      `json:"clue"`
}

func (StoryWorldChange) TableName() string {
	return "story_world_changes"
}
