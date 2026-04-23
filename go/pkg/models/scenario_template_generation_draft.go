package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	ScenarioTemplateGenerationDraftStatusSuggested = "suggested"
	ScenarioTemplateGenerationDraftStatusConverted = "converted"
)

type ScenarioTemplateGenerationDraftPayload struct {
	GenreID                   uuid.UUID                      `json:"genreId"`
	ZoneKind                  string                         `json:"zoneKind,omitempty"`
	Prompt                    string                         `json:"prompt"`
	ImageURL                  string                         `json:"imageUrl"`
	ThumbnailURL              string                         `json:"thumbnailUrl"`
	ScaleWithUserLevel        bool                           `json:"scaleWithUserLevel"`
	RewardMode                RewardMode                     `json:"rewardMode"`
	RandomRewardSize          RandomRewardSize               `json:"randomRewardSize"`
	Difficulty                int                            `json:"difficulty"`
	RewardExperience          int                            `json:"rewardExperience"`
	RewardGold                int                            `json:"rewardGold"`
	OpenEnded                 bool                           `json:"openEnded"`
	SuccessHandoffText        string                         `json:"successHandoffText"`
	FailureHandoffText        string                         `json:"failureHandoffText"`
	FailurePenaltyMode        ScenarioFailurePenaltyMode     `json:"failurePenaltyMode"`
	FailureHealthDrainType    ScenarioFailureDrainType       `json:"failureHealthDrainType"`
	FailureHealthDrainValue   int                            `json:"failureHealthDrainValue"`
	FailureManaDrainType      ScenarioFailureDrainType       `json:"failureManaDrainType"`
	FailureManaDrainValue     int                            `json:"failureManaDrainValue"`
	FailureStatuses           ScenarioFailureStatusTemplates `json:"failureStatuses"`
	SuccessRewardMode         ScenarioSuccessRewardMode      `json:"successRewardMode"`
	SuccessHealthRestoreType  ScenarioFailureDrainType       `json:"successHealthRestoreType"`
	SuccessHealthRestoreValue int                            `json:"successHealthRestoreValue"`
	SuccessManaRestoreType    ScenarioFailureDrainType       `json:"successManaRestoreType"`
	SuccessManaRestoreValue   int                            `json:"successManaRestoreValue"`
	SuccessStatuses           ScenarioFailureStatusTemplates `json:"successStatuses"`
	Options                   ScenarioTemplateOptions        `json:"options"`
	ItemRewards               ScenarioTemplateRewards        `json:"itemRewards"`
	ItemChoiceRewards         ScenarioTemplateRewards        `json:"itemChoiceRewards"`
	SpellRewards              ScenarioTemplateSpellRewards   `json:"spellRewards"`
}

type ScenarioTemplateGenerationDraftPayloadValue ScenarioTemplateGenerationDraftPayload

func (p ScenarioTemplateGenerationDraftPayloadValue) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *ScenarioTemplateGenerationDraftPayloadValue) Scan(value interface{}) error {
	if value == nil {
		*p = ScenarioTemplateGenerationDraftPayloadValue{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan ScenarioTemplateGenerationDraftPayloadValue: value is not []byte")
	}
	var decoded ScenarioTemplateGenerationDraftPayloadValue
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*p = decoded
	return nil
}

type ScenarioTemplateGenerationDraft struct {
	ID                 uuid.UUID                                   `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt          time.Time                                   `json:"createdAt"`
	UpdatedAt          time.Time                                   `json:"updatedAt"`
	JobID              uuid.UUID                                   `json:"jobId" gorm:"column:job_id;type:uuid"`
	Status             string                                      `json:"status"`
	GenreID            uuid.UUID                                   `json:"genreId" gorm:"column:genre_id;type:uuid"`
	Genre              *ZoneGenre                                  `json:"genre,omitempty" gorm:"foreignKey:GenreID"`
	ZoneKind           string                                      `json:"zoneKind,omitempty" gorm:"column:zone_kind"`
	Prompt             string                                      `json:"prompt"`
	OpenEnded          bool                                        `json:"openEnded" gorm:"column:open_ended"`
	Difficulty         int                                         `json:"difficulty"`
	Payload            ScenarioTemplateGenerationDraftPayloadValue `json:"payload" gorm:"column:payload;type:jsonb"`
	ScenarioTemplateID *uuid.UUID                                  `json:"scenarioTemplateId,omitempty" gorm:"column:scenario_template_id;type:uuid"`
	ScenarioTemplate   *ScenarioTemplate                           `json:"scenarioTemplate,omitempty" gorm:"foreignKey:ScenarioTemplateID"`
	ConvertedAt        *time.Time                                  `json:"convertedAt,omitempty" gorm:"column:converted_at"`
}

func (ScenarioTemplateGenerationDraft) TableName() string {
	return "scenario_template_generation_drafts"
}

func NormalizeScenarioTemplateGenerationDraftStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case ScenarioTemplateGenerationDraftStatusConverted:
		return ScenarioTemplateGenerationDraftStatusConverted
	default:
		return ScenarioTemplateGenerationDraftStatusSuggested
	}
}

func ScenarioTemplateGenerationDraftPayloadFromTemplate(
	template *ScenarioTemplate,
) ScenarioTemplateGenerationDraftPayloadValue {
	if template == nil {
		return ScenarioTemplateGenerationDraftPayloadValue{}
	}
	return ScenarioTemplateGenerationDraftPayloadValue{
		GenreID:                   template.GenreID,
		ZoneKind:                  NormalizeZoneKind(template.ZoneKind),
		Prompt:                    strings.TrimSpace(template.Prompt),
		ImageURL:                  strings.TrimSpace(template.ImageURL),
		ThumbnailURL:              strings.TrimSpace(template.ThumbnailURL),
		ScaleWithUserLevel:        template.ScaleWithUserLevel,
		RewardMode:                template.RewardMode,
		RandomRewardSize:          template.RandomRewardSize,
		Difficulty:                template.Difficulty,
		RewardExperience:          template.RewardExperience,
		RewardGold:                template.RewardGold,
		OpenEnded:                 template.OpenEnded,
		SuccessHandoffText:        strings.TrimSpace(template.SuccessHandoffText),
		FailureHandoffText:        strings.TrimSpace(template.FailureHandoffText),
		FailurePenaltyMode:        template.FailurePenaltyMode,
		FailureHealthDrainType:    template.FailureHealthDrainType,
		FailureHealthDrainValue:   template.FailureHealthDrainValue,
		FailureManaDrainType:      template.FailureManaDrainType,
		FailureManaDrainValue:     template.FailureManaDrainValue,
		FailureStatuses:           template.FailureStatuses,
		SuccessRewardMode:         template.SuccessRewardMode,
		SuccessHealthRestoreType:  template.SuccessHealthRestoreType,
		SuccessHealthRestoreValue: template.SuccessHealthRestoreValue,
		SuccessManaRestoreType:    template.SuccessManaRestoreType,
		SuccessManaRestoreValue:   template.SuccessManaRestoreValue,
		SuccessStatuses:           template.SuccessStatuses,
		Options:                   template.Options,
		ItemRewards:               template.ItemRewards,
		ItemChoiceRewards:         template.ItemChoiceRewards,
		SpellRewards:              template.SpellRewards,
	}
}

func ScenarioTemplateFromGenerationDraftPayload(
	payload ScenarioTemplateGenerationDraftPayloadValue,
) *ScenarioTemplate {
	return &ScenarioTemplate{
		GenreID:                   payload.GenreID,
		ZoneKind:                  NormalizeZoneKind(payload.ZoneKind),
		Prompt:                    strings.TrimSpace(payload.Prompt),
		ImageURL:                  strings.TrimSpace(payload.ImageURL),
		ThumbnailURL:              strings.TrimSpace(payload.ThumbnailURL),
		ScaleWithUserLevel:        payload.ScaleWithUserLevel,
		RewardMode:                NormalizeRewardMode(string(payload.RewardMode)),
		RandomRewardSize:          NormalizeRandomRewardSize(string(payload.RandomRewardSize)),
		Difficulty:                payload.Difficulty,
		RewardExperience:          payload.RewardExperience,
		RewardGold:                payload.RewardGold,
		OpenEnded:                 payload.OpenEnded,
		SuccessHandoffText:        strings.TrimSpace(payload.SuccessHandoffText),
		FailureHandoffText:        strings.TrimSpace(payload.FailureHandoffText),
		FailurePenaltyMode:        payload.FailurePenaltyMode,
		FailureHealthDrainType:    payload.FailureHealthDrainType,
		FailureHealthDrainValue:   payload.FailureHealthDrainValue,
		FailureManaDrainType:      payload.FailureManaDrainType,
		FailureManaDrainValue:     payload.FailureManaDrainValue,
		FailureStatuses:           payload.FailureStatuses,
		SuccessRewardMode:         payload.SuccessRewardMode,
		SuccessHealthRestoreType:  payload.SuccessHealthRestoreType,
		SuccessHealthRestoreValue: payload.SuccessHealthRestoreValue,
		SuccessManaRestoreType:    payload.SuccessManaRestoreType,
		SuccessManaRestoreValue:   payload.SuccessManaRestoreValue,
		SuccessStatuses:           payload.SuccessStatuses,
		Options:                   payload.Options,
		ItemRewards:               payload.ItemRewards,
		ItemChoiceRewards:         payload.ItemChoiceRewards,
		SpellRewards:              payload.SpellRewards,
	}
}
