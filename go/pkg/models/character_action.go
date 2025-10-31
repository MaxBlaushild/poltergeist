package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ActionType string

const (
	ActionTypeTalk ActionType = "talk"
	// Future: ActionTypeGift, ActionTypeTrade, etc.
)

type DialogueMessage struct {
	Speaker string `json:"speaker"` // "character" or "user"
	Text    string `json:"text"`
	Order   int    `json:"order"`
}

// DialogueSequence is a custom type for []DialogueMessage that implements sql.Scanner and driver.Valuer
type DialogueSequence []DialogueMessage

// Scan implements the sql.Scanner interface for reading from database
func (ds *DialogueSequence) Scan(value interface{}) error {
	if value == nil {
		*ds = []DialogueMessage{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan DialogueSequence: value is not []byte")
	}

	var messages []DialogueMessage
	if err := json.Unmarshal(bytes, &messages); err != nil {
		return err
	}

	*ds = messages
	return nil
}

// Value implements the driver.Valuer interface for writing to database
func (ds DialogueSequence) Value() (driver.Value, error) {
	if ds == nil {
		return json.Marshal([]DialogueMessage{})
	}
	return json.Marshal(ds)
}

// MetadataJSONB is a generic type for JSONB metadata fields
type MetadataJSONB map[string]interface{}

// Scan implements the sql.Scanner interface for reading from database
func (m *MetadataJSONB) Scan(value interface{}) error {
	if value == nil {
		*m = MetadataJSONB{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan MetadataJSONB: value is not []byte")
	}

	if err := json.Unmarshal(bytes, m); err != nil {
		return err
	}

	return nil
}

// Value implements the driver.Valuer interface for writing to database
func (m MetadataJSONB) Value() (driver.Value, error) {
	if m == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(m)
}

type CharacterAction struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	CharacterID uuid.UUID        `json:"characterId" gorm:"type:uuid"`
	Character   *Character       `json:"character,omitempty"`
	ActionType  ActionType       `json:"actionType"`
	Dialogue    DialogueSequence `json:"dialogue" gorm:"type:jsonb"`
	Metadata    MetadataJSONB    `json:"metadata" gorm:"type:jsonb"`
}

func (ca *CharacterAction) TableName() string {
	return "character_actions"
}
