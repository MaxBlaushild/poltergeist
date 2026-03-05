package server

import (
	"fmt"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func applyStandaloneRecurrenceForCreate(
	raw *string,
	now time.Time,
	recurringID **uuid.UUID,
	recurrenceFrequency **string,
	nextRecurrenceAt **time.Time,
) error {
	if raw == nil {
		return nil
	}
	recurrence := models.NormalizeQuestRecurrenceFrequency(*raw)
	if recurrence == "" {
		return nil
	}
	if !models.IsValidQuestRecurrenceFrequency(recurrence) {
		return fmt.Errorf("invalid recurrence frequency")
	}
	nextAt, ok := models.NextQuestRecurrenceAt(now, recurrence)
	if !ok {
		return fmt.Errorf("invalid recurrence frequency")
	}
	recurrenceCopy := recurrence
	nextAtCopy := nextAt
	newRecurringID := uuid.New()
	*recurringID = &newRecurringID
	*recurrenceFrequency = &recurrenceCopy
	*nextRecurrenceAt = &nextAtCopy
	return nil
}

func applyStandaloneRecurrenceForUpdate(
	raw *string,
	now time.Time,
	recurringID **uuid.UUID,
	recurrenceFrequency **string,
	nextRecurrenceAt **time.Time,
) error {
	if raw == nil {
		return nil
	}
	recurrence := models.NormalizeQuestRecurrenceFrequency(*raw)
	if recurrence == "" {
		*recurrenceFrequency = nil
		*nextRecurrenceAt = nil
		return nil
	}
	if !models.IsValidQuestRecurrenceFrequency(recurrence) {
		return fmt.Errorf("invalid recurrence frequency")
	}
	if *recurringID == nil {
		newRecurringID := uuid.New()
		*recurringID = &newRecurringID
	}
	nextAt, ok := models.NextQuestRecurrenceAt(now, recurrence)
	if !ok {
		return fmt.Errorf("invalid recurrence frequency")
	}
	recurrenceCopy := recurrence
	nextAtCopy := nextAt
	*recurrenceFrequency = &recurrenceCopy
	*nextRecurrenceAt = &nextAtCopy
	return nil
}
