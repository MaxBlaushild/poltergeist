package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type CharacterStoryVariant struct {
	ID                 uuid.UUID        `json:"id"`
	CreatedAt          time.Time        `json:"createdAt"`
	UpdatedAt          time.Time        `json:"updatedAt"`
	Priority           int              `json:"priority"`
	RequiredStoryFlags StringArray      `json:"requiredStoryFlags"`
	Description        string           `json:"description"`
	Dialogue           DialogueSequence `json:"dialogue"`
}

type CharacterStoryVariants []CharacterStoryVariant

func (v CharacterStoryVariants) Value() (driver.Value, error) {
	if v == nil {
		return json.Marshal([]CharacterStoryVariant{})
	}
	return json.Marshal(v)
}

func (v *CharacterStoryVariants) Scan(value interface{}) error {
	if value == nil {
		*v = CharacterStoryVariants{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan CharacterStoryVariants: value is not []byte")
	}
	var decoded []CharacterStoryVariant
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*v = decoded
	return nil
}

type PointOfInterestStoryVariant struct {
	ID                 uuid.UUID   `json:"id"`
	CreatedAt          time.Time   `json:"createdAt"`
	UpdatedAt          time.Time   `json:"updatedAt"`
	Priority           int         `json:"priority"`
	RequiredStoryFlags StringArray `json:"requiredStoryFlags"`
	Description        string      `json:"description"`
	Clue               string      `json:"clue"`
}

type PointOfInterestStoryVariants []PointOfInterestStoryVariant

func (v PointOfInterestStoryVariants) Value() (driver.Value, error) {
	if v == nil {
		return json.Marshal([]PointOfInterestStoryVariant{})
	}
	return json.Marshal(v)
}

func (v *PointOfInterestStoryVariants) Scan(value interface{}) error {
	if value == nil {
		*v = PointOfInterestStoryVariants{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan PointOfInterestStoryVariants: value is not []byte")
	}
	var decoded []PointOfInterestStoryVariant
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*v = decoded
	return nil
}

type UserStoryFlag struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserID    uuid.UUID `json:"userId" gorm:"column:user_id;type:uuid;index"`
	FlagKey   string    `json:"flagKey" gorm:"column:flag_key;index"`
	Value     bool      `json:"value" gorm:"column:value"`
}

func (UserStoryFlag) TableName() string {
	return "user_story_flags"
}
