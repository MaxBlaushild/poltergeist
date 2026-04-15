package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type ZoneSeedResolvedCounts struct {
	PlaceCount           int `json:"placeCount"`
	MonsterCount         int `json:"monsterCount"`
	BossEncounterCount   int `json:"bossEncounterCount"`
	RaidEncounterCount   int `json:"raidEncounterCount"`
	InputEncounterCount  int `json:"inputEncounterCount"`
	OptionEncounterCount int `json:"optionEncounterCount"`
	TreasureChestCount   int `json:"treasureChestCount"`
	HealingFountainCount int `json:"healingFountainCount"`
}

func (c ZoneSeedResolvedCounts) HasContent() bool {
	return c.PlaceCount != 0 ||
		c.MonsterCount != 0 ||
		c.BossEncounterCount != 0 ||
		c.RaidEncounterCount != 0 ||
		c.InputEncounterCount != 0 ||
		c.OptionEncounterCount != 0 ||
		c.TreasureChestCount != 0 ||
		c.HealingFountainCount != 0
}

type ZoneSeedAutoAudit struct {
	ZoneAreaSquareFeet float64                `json:"zoneAreaSquareFeet,omitempty"`
	ZoneAreaAcres      float64                `json:"zoneAreaAcres,omitempty"`
	RecommendedCounts  ZoneSeedResolvedCounts `json:"recommendedCounts,omitempty"`
	FinalCounts        ZoneSeedResolvedCounts `json:"finalCounts,omitempty"`
	Warnings           StringArray            `json:"warnings,omitempty"`
}

func (a ZoneSeedAutoAudit) HasData() bool {
	return a.ZoneAreaSquareFeet > 0 ||
		a.ZoneAreaAcres > 0 ||
		a.RecommendedCounts.HasContent() ||
		a.FinalCounts.HasContent() ||
		len(a.Warnings) > 0
}

func (a ZoneSeedAutoAudit) Value() (driver.Value, error) {
	if !a.HasData() {
		return json.Marshal(map[string]any{})
	}
	return json.Marshal(a)
}

func (a *ZoneSeedAutoAudit) Scan(value interface{}) error {
	if value == nil {
		*a = ZoneSeedAutoAudit{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to scan ZoneSeedAutoAudit: unsupported type")
	}

	if len(bytes) == 0 {
		*a = ZoneSeedAutoAudit{}
		return nil
	}

	return json.Unmarshal(bytes, a)
}
