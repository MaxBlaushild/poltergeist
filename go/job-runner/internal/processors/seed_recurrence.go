package processors

import (
	"fmt"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

const zoneSeedDefaultRecurrenceFrequency = models.QuestRecurrenceWeekly

func zoneSeedDefaultRecurrence(now time.Time) (uuid.UUID, string, time.Time, error) {
	nextAt, ok := models.NextQuestRecurrenceAt(now, zoneSeedDefaultRecurrenceFrequency)
	if !ok {
		return uuid.Nil, "", time.Time{}, fmt.Errorf("unsupported zone seed recurrence frequency %q", zoneSeedDefaultRecurrenceFrequency)
	}

	return uuid.New(), zoneSeedDefaultRecurrenceFrequency, nextAt, nil
}
