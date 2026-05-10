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
	ActionTypeTalk              ActionType = "talk"
	ActionTypeShop              ActionType = "shop"
	ActionTypeGiveQuest         ActionType = "giveQuest"
	ActionTypeReceiveQuestItems ActionType = "receiveQuestItems"
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
	Speaker     string         `json:"speaker"` // "character" or "user"
	Text        string         `json:"text"`
	Order       int            `json:"order"`
	Effect      DialogueEffect `json:"effect,omitempty"`
	CharacterID *uuid.UUID     `json:"characterId,omitempty"`
	SpeakerName string         `json:"speakerName,omitempty"`
	PortraitURL string         `json:"portraitUrl,omitempty"`
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
	var characterID *uuid.UUID
	if speaker == "character" && message.CharacterID != nil && *message.CharacterID != uuid.Nil {
		resolved := *message.CharacterID
		characterID = &resolved
	}
	speakerName := ""
	portraitURL := ""
	if speaker == "character" {
		speakerName = strings.TrimSpace(message.SpeakerName)
		portraitURL = strings.TrimSpace(message.PortraitURL)
	}
	return DialogueMessage{
		Speaker:     speaker,
		Text:        strings.TrimSpace(message.Text),
		Order:       order,
		Effect:      NormalizeDialogueEffect(string(message.Effect)),
		CharacterID: characterID,
		SpeakerName: speakerName,
		PortraitURL: portraitURL,
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
	return DialogueSequenceFromSpeakerNameLines(lines, "")
}

func DialogueSequenceFromSpeakerNameLines(lines []string, speakerName string) DialogueSequence {
	return DialogueSequenceFromSpeakerIdentityLines(lines, speakerName, "")
}

func DialogueSequenceFromSpeakerIdentityLines(
	lines []string,
	speakerName string,
	portraitURL string,
) DialogueSequence {
	trimmedSpeakerName := strings.TrimSpace(speakerName)
	trimmedPortraitURL := strings.TrimSpace(portraitURL)
	messages := make([]DialogueMessage, 0, len(lines))
	for _, line := range lines {
		resolvedSpeakerName := trimmedSpeakerName
		resolvedText := strings.TrimSpace(line)
		if overrideSpeakerName, overrideText, ok := parseDialogueSpeakerOverride(line); ok {
			resolvedSpeakerName = overrideSpeakerName
			resolvedText = overrideText
		}
		messages = append(messages, DialogueMessage{
			Speaker:     "character",
			Text:        resolvedText,
			Order:       len(messages),
			SpeakerName: resolvedSpeakerName,
			PortraitURL: trimmedPortraitURL,
		})
	}
	return normalizeDialogueSequence(messages)
}

func parseDialogueSpeakerOverride(line string) (string, string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return "", "", false
	}
	colonIndex := strings.IndexRune(trimmed, ':')
	if colonIndex <= 0 || colonIndex >= len(trimmed)-1 {
		return "", trimmed, false
	}

	speakerName := strings.TrimSpace(trimmed[:colonIndex])
	text := strings.TrimSpace(trimmed[colonIndex+1:])
	if speakerName == "" || text == "" {
		return "", trimmed, false
	}
	if len([]rune(speakerName)) > 40 || len(strings.Fields(speakerName)) > 5 {
		return "", trimmed, false
	}

	for _, char := range speakerName {
		switch {
		case (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9'):
			continue
		case char == ' ' || char == '\'' || char == '-' || char == '&':
			continue
		default:
			return "", trimmed, false
		}
	}

	lower := strings.ToLower(speakerName)
	for _, prefix := range []string{"at ", "in ", "on ", "from ", "with ", "when "} {
		if strings.HasPrefix(lower, prefix) {
			return "", trimmed, false
		}
	}

	return speakerName, text, true
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
