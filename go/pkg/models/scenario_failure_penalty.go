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
	Name                          string `json:"name"`
	Description                   string `json:"description"`
	Effect                        string `json:"effect"`
	EffectType                    string `json:"effectType"`
	Positive                      bool   `json:"positive"`
	DamagePerTick                 int    `json:"damagePerTick"`
	HealthPerTick                 int    `json:"healthPerTick"`
	ManaPerTick                   int    `json:"manaPerTick"`
	DurationSeconds               int    `json:"durationSeconds"`
	StrengthMod                   int    `json:"strengthMod"`
	DexterityMod                  int    `json:"dexterityMod"`
	ConstitutionMod               int    `json:"constitutionMod"`
	IntelligenceMod               int    `json:"intelligenceMod"`
	WisdomMod                     int    `json:"wisdomMod"`
	CharismaMod                   int    `json:"charismaMod"`
	PhysicalDamageBonusPercent    int    `json:"physicalDamageBonusPercent"`
	PiercingDamageBonusPercent    int    `json:"piercingDamageBonusPercent"`
	SlashingDamageBonusPercent    int    `json:"slashingDamageBonusPercent"`
	BludgeoningDamageBonusPercent int    `json:"bludgeoningDamageBonusPercent"`
	FireDamageBonusPercent        int    `json:"fireDamageBonusPercent"`
	IceDamageBonusPercent         int    `json:"iceDamageBonusPercent"`
	LightningDamageBonusPercent   int    `json:"lightningDamageBonusPercent"`
	PoisonDamageBonusPercent      int    `json:"poisonDamageBonusPercent"`
	ArcaneDamageBonusPercent      int    `json:"arcaneDamageBonusPercent"`
	HolyDamageBonusPercent        int    `json:"holyDamageBonusPercent"`
	ShadowDamageBonusPercent      int    `json:"shadowDamageBonusPercent"`
	PhysicalResistancePercent     int    `json:"physicalResistancePercent"`
	PiercingResistancePercent     int    `json:"piercingResistancePercent"`
	SlashingResistancePercent     int    `json:"slashingResistancePercent"`
	BludgeoningResistancePercent  int    `json:"bludgeoningResistancePercent"`
	FireResistancePercent         int    `json:"fireResistancePercent"`
	IceResistancePercent          int    `json:"iceResistancePercent"`
	LightningResistancePercent    int    `json:"lightningResistancePercent"`
	PoisonResistancePercent       int    `json:"poisonResistancePercent"`
	ArcaneResistancePercent       int    `json:"arcaneResistancePercent"`
	HolyResistancePercent         int    `json:"holyResistancePercent"`
	ShadowResistancePercent       int    `json:"shadowResistancePercent"`
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
