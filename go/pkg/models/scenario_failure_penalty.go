package models

import (
	"database/sql/driver"
	"encoding/json"
)

type ScenarioFailurePenaltyMode string

const (
	ScenarioFailurePenaltyModeShared     ScenarioFailurePenaltyMode = "shared"
	ScenarioFailurePenaltyModeIndividual ScenarioFailurePenaltyMode = "individual"
)

type ScenarioSuccessRewardMode string

const (
	ScenarioSuccessRewardModeShared     ScenarioSuccessRewardMode = "shared"
	ScenarioSuccessRewardModeIndividual ScenarioSuccessRewardMode = "individual"
)

type ScenarioFailureDrainType string

const (
	ScenarioFailureDrainTypeNone    ScenarioFailureDrainType = "none"
	ScenarioFailureDrainTypeFlat    ScenarioFailureDrainType = "flat"
	ScenarioFailureDrainTypePercent ScenarioFailureDrainType = "percent"
)

type ScenarioFailureStatusTemplate struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	Effect          string `json:"effect"`
	EffectType      string `json:"effectType"`
	Positive        bool   `json:"positive"`
	DamagePerTick   int    `json:"damagePerTick"`
	DurationSeconds int    `json:"durationSeconds"`
	StrengthMod     int    `json:"strengthMod"`
	DexterityMod    int    `json:"dexterityMod"`
	ConstitutionMod int    `json:"constitutionMod"`
	IntelligenceMod int    `json:"intelligenceMod"`
	WisdomMod       int    `json:"wisdomMod"`
	CharismaMod     int    `json:"charismaMod"`
}

type ScenarioFailureStatusTemplates []ScenarioFailureStatusTemplate

func (s ScenarioFailureStatusTemplates) Value() (driver.Value, error) {
	if s == nil {
		return json.Marshal([]ScenarioFailureStatusTemplate{})
	}
	return json.Marshal(s)
}

func (s *ScenarioFailureStatusTemplates) Scan(value interface{}) error {
	if value == nil {
		*s = ScenarioFailureStatusTemplates{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*s = ScenarioFailureStatusTemplates{}
		return nil
	}

	return json.Unmarshal(bytes, s)
}
