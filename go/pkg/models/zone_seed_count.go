package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type ZoneSeedCountAudit struct {
	TargetCounts               ZoneSeedResolvedCounts `json:"targetCounts,omitempty"`
	ExistingCounts             ZoneSeedResolvedCounts `json:"existingCounts,omitempty"`
	QueuedCounts               ZoneSeedResolvedCounts `json:"queuedCounts,omitempty"`
	RemainingRequiredPlaceTags StringArray            `json:"remainingRequiredPlaceTags,omitempty"`
	Warnings                   StringArray            `json:"warnings,omitempty"`
}

func (a ZoneSeedCountAudit) HasData() bool {
	return a.TargetCounts.HasContent() ||
		a.ExistingCounts.HasContent() ||
		a.QueuedCounts.HasContent() ||
		len(a.RemainingRequiredPlaceTags) > 0 ||
		len(a.Warnings) > 0
}

func (a ZoneSeedCountAudit) Value() (driver.Value, error) {
	if !a.HasData() {
		return json.Marshal(map[string]any{})
	}
	return json.Marshal(a)
}

func (a *ZoneSeedCountAudit) Scan(value interface{}) error {
	if value == nil {
		*a = ZoneSeedCountAudit{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to scan ZoneSeedCountAudit: unsupported type")
	}

	if len(bytes) == 0 {
		*a = ZoneSeedCountAudit{}
		return nil
	}

	return json.Unmarshal(bytes, a)
}
