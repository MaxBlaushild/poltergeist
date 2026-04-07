package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ActionType string
type DialogueEffect string

const (
	ActionTypeTalk      ActionType = "talk"
	ActionTypeShop      ActionType = "shop"
	ActionTypeGiveQuest ActionType = "giveQuest"
	// Future: ActionTypeGift, ActionTypeTrade, etc.

	DialogueEffectNone       DialogueEffect = ""
	DialogueEffectAngry      DialogueEffect = "angry"
	DialogueEffectSurprised  DialogueEffect = "surprised"
	DialogueEffectWhisper    DialogueEffect = "whisper"
	DialogueEffectShout      DialogueEffect = "shout"
	DialogueEffectMysterious DialogueEffect = "mysterious"
	DialogueEffectDetermined DialogueEffect = "determined"
)

type DialogueMessage struct {
	Speaker string         `json:"speaker"` // "character" or "user"
	Text    string         `json:"text"`
	Order   int            `json:"order"`
	Effect  DialogueEffect `json:"effect,omitempty"`
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
		var legacyMessages []string
		if legacyErr := json.Unmarshal(bytes, &legacyMessages); legacyErr != nil {
			return err
		}
		*ds = DialogueSequenceFromStringLines(legacyMessages)
		return nil
	}

	*ds = normalizeDialogueSequence(messages)
	return nil
}

// Value implements the driver.Valuer interface for writing to database
func (ds DialogueSequence) Value() (driver.Value, error) {
	return json.Marshal(normalizeDialogueSequence(ds))
}

func NormalizeDialogueEffect(value string) DialogueEffect {
	switch DialogueEffect(strings.ToLower(strings.TrimSpace(value))) {
	case DialogueEffectAngry:
		return DialogueEffectAngry
	case DialogueEffectSurprised:
		return DialogueEffectSurprised
	case DialogueEffectWhisper:
		return DialogueEffectWhisper
	case DialogueEffectShout:
		return DialogueEffectShout
	case DialogueEffectMysterious:
		return DialogueEffectMysterious
	case DialogueEffectDetermined:
		return DialogueEffectDetermined
	default:
		return DialogueEffectNone
	}
}

func normalizeDialogueMessage(message DialogueMessage, order int) DialogueMessage {
	speaker := strings.TrimSpace(message.Speaker)
	if speaker == "" {
		speaker = "character"
	}
	return DialogueMessage{
		Speaker: speaker,
		Text:    strings.TrimSpace(message.Text),
		Order:   order,
		Effect:  NormalizeDialogueEffect(string(message.Effect)),
	}
}

func normalizeDialogueSequence(messages []DialogueMessage) DialogueSequence {
	normalized := make(DialogueSequence, 0, len(messages))
	for _, message := range messages {
		next := normalizeDialogueMessage(message, len(normalized))
		if next.Text == "" {
			continue
		}
		normalized = append(normalized, next)
	}
	if normalized == nil {
		return DialogueSequence{}
	}
	return normalized
}

func DialogueSequenceFromStringLines(lines []string) DialogueSequence {
	messages := make([]DialogueMessage, 0, len(lines))
	for _, line := range lines {
		messages = append(messages, DialogueMessage{
			Speaker: "character",
			Text:    line,
			Order:   len(messages),
		})
	}
	return normalizeDialogueSequence(messages)
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
