package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	QuestRecurrenceDaily   = "daily"
	QuestRecurrenceWeekly  = "weekly"
	QuestRecurrenceMonthly = "monthly"
)

func NormalizeQuestRecurrenceFrequency(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func IsValidQuestRecurrenceFrequency(value string) bool {
	switch NormalizeQuestRecurrenceFrequency(value) {
	case QuestRecurrenceDaily, QuestRecurrenceWeekly, QuestRecurrenceMonthly:
		return true
	default:
		return false
	}
}

func NextQuestRecurrenceAt(base time.Time, frequency string) (time.Time, bool) {
	switch NormalizeQuestRecurrenceFrequency(frequency) {
	case QuestRecurrenceDaily:
		return base.AddDate(0, 0, 1), true
	case QuestRecurrenceWeekly:
		return base.AddDate(0, 0, 7), true
	case QuestRecurrenceMonthly:
		return base.AddDate(0, 1, 0), true
	default:
		return time.Time{}, false
	}
}

type Quest struct {
	ID                    uuid.UUID         `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt             time.Time         `json:"createdAt"`
	UpdatedAt             time.Time         `json:"updatedAt"`
	Name                  string            `json:"name"`
	Description           string            `json:"description"`
	AcceptanceDialogue    StringArray       `json:"acceptanceDialogue,omitempty" gorm:"type:jsonb"`
	ImageURL              string            `json:"imageUrl"`
	ZoneID                *uuid.UUID        `json:"zoneId" gorm:"type:uuid"`
	QuestArchetypeID      *uuid.UUID        `json:"questArchetypeId" gorm:"type:uuid"`
	QuestGiverCharacterID *uuid.UUID        `json:"questGiverCharacterId" gorm:"type:uuid"`
	RecurringQuestID      *uuid.UUID        `json:"recurringQuestId,omitempty" gorm:"type:uuid"`
	RecurrenceFrequency   *string           `json:"recurrenceFrequency,omitempty"`
	NextRecurrenceAt      *time.Time        `json:"nextRecurrenceAt,omitempty"`
	Gold                  int               `json:"gold"`
	ItemRewards           []QuestItemReward `json:"itemRewards" gorm:"foreignKey:QuestID"`
	Nodes                 []QuestNode       `json:"nodes" gorm:"foreignKey:QuestID"`
}

func (q *Quest) TableName() string {
	return "quests"
}
