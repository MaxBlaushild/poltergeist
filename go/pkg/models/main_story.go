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
	MainStorySuggestionJobStatusQueued     = "queued"
	MainStorySuggestionJobStatusInProgress = "in_progress"
	MainStorySuggestionJobStatusCompleted  = "completed"
	MainStorySuggestionJobStatusFailed     = "failed"

	MainStorySuggestionDraftStatusSuggested = "suggested"
	MainStorySuggestionDraftStatusConverted = "converted"

	MainStoryDistrictRunStatusQueued     = "queued"
	MainStoryDistrictRunStatusInProgress = "in_progress"
	MainStoryDistrictRunStatusCompleted  = "completed"
	MainStoryDistrictRunStatusFailed     = "failed"
)

type MainStoryBeatDraft struct {
	OrderIndex                    int                           `json:"orderIndex"`
	Act                           int                           `json:"act"`
	StoryRole                     string                        `json:"storyRole"`
	ChapterTitle                  string                        `json:"chapterTitle"`
	ChapterSummary                string                        `json:"chapterSummary"`
	Purpose                       string                        `json:"purpose"`
	WhatChanges                   string                        `json:"whatChanges"`
	IntroducedCharacterKeys       StringArray                   `json:"introducedCharacterKeys"`
	RequiredCharacterKeys         StringArray                   `json:"requiredCharacterKeys"`
	IntroducedRevealKeys          StringArray                   `json:"introducedRevealKeys"`
	RequiredRevealKeys            StringArray                   `json:"requiredRevealKeys"`
	RequiredZoneTags              StringArray                   `json:"requiredZoneTags"`
	RequiredLocationArchetypeIDs  StringArray                   `json:"requiredLocationArchetypeIds"`
	PreferredContentMix           StringArray                   `json:"preferredContentMix"`
	QuestGiverCharacterKey        string                        `json:"questGiverCharacterKey"`
	QuestGiverCharacterID         *uuid.UUID                    `json:"questGiverCharacterId,omitempty"`
	QuestGiverCharacterName       string                        `json:"questGiverCharacterName,omitempty"`
	Name                          string                        `json:"name"`
	Hook                          string                        `json:"hook"`
	Description                   string                        `json:"description"`
	AcceptanceDialogue            StringArray                   `json:"acceptanceDialogue"`
	RequiredStoryFlags            StringArray                   `json:"requiredStoryFlags"`
	SetStoryFlags                 StringArray                   `json:"setStoryFlags"`
	ClearStoryFlags               StringArray                   `json:"clearStoryFlags"`
	QuestGiverRelationshipEffects CharacterRelationshipState    `json:"questGiverRelationshipEffects"`
	WorldChanges                  []MainStoryWorldChange        `json:"worldChanges"`
	UnlockedScenarios             []MainStoryUnlockedScenario   `json:"unlockedScenarios"`
	UnlockedChallenges            []MainStoryUnlockedChallenge  `json:"unlockedChallenges"`
	UnlockedMonsterEncounters     []MainStoryUnlockedEncounter  `json:"unlockedMonsterEncounters"`
	QuestGiverAfterDescription    string                        `json:"questGiverAfterDescription"`
	QuestGiverAfterDialogue       StringArray                   `json:"questGiverAfterDialogue"`
	CharacterTags                 StringArray                   `json:"characterTags"`
	InternalTags                  StringArray                   `json:"internalTags"`
	DifficultyMode                QuestDifficultyMode           `json:"difficultyMode"`
	Difficulty                    int                           `json:"difficulty"`
	MonsterEncounterTargetLevel   int                           `json:"monsterEncounterTargetLevel"`
	WhyThisScales                 string                        `json:"whyThisScales"`
	Steps                         QuestArchetypeSuggestionSteps `json:"steps"`
	ChallengeTemplateSeeds        StringArray                   `json:"challengeTemplateSeeds"`
	ScenarioTemplateSeeds         StringArray                   `json:"scenarioTemplateSeeds"`
	MonsterTemplateSeeds          StringArray                   `json:"monsterTemplateSeeds"`
	Warnings                      StringArray                   `json:"warnings"`
	QuestArchetypeID              *uuid.UUID                    `json:"questArchetypeId,omitempty"`
	QuestArchetypeName            string                        `json:"questArchetypeName,omitempty"`
}

type MainStoryUnlockedScenario struct {
	Name                string      `json:"name"`
	Prompt              string      `json:"prompt"`
	PointOfInterestHint string      `json:"pointOfInterestHint,omitempty"`
	InternalTags        StringArray `json:"internalTags,omitempty"`
	Difficulty          int         `json:"difficulty,omitempty"`
}

type MainStoryUnlockedChallenge struct {
	Question            string                  `json:"question"`
	Description         string                  `json:"description"`
	PointOfInterestHint string                  `json:"pointOfInterestHint,omitempty"`
	SubmissionType      QuestNodeSubmissionType `json:"submissionType,omitempty"`
	Proficiency         *string                 `json:"proficiency,omitempty"`
	StatTags            StringArray             `json:"statTags,omitempty"`
	Difficulty          int                     `json:"difficulty,omitempty"`
}

type MainStoryUnlockedEncounter struct {
	Name                 string               `json:"name"`
	Description          string               `json:"description"`
	PointOfInterestHint  string               `json:"pointOfInterestHint,omitempty"`
	EncounterType        MonsterEncounterType `json:"encounterType,omitempty"`
	MonsterCount         int                  `json:"monsterCount,omitempty"`
	EncounterTone        StringArray          `json:"encounterTone,omitempty"`
	MonsterTemplateHints StringArray          `json:"monsterTemplateHints,omitempty"`
	TargetLevel          int                  `json:"targetLevel,omitempty"`
}

type MainStoryBeatDrafts []MainStoryBeatDraft

func (s MainStoryBeatDrafts) Value() (driver.Value, error) {
	if s == nil {
		return json.Marshal([]MainStoryBeatDraft{})
	}
	return json.Marshal(s)
}

func (s *MainStoryBeatDrafts) Scan(value interface{}) error {
	if value == nil {
		*s = MainStoryBeatDrafts{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan MainStoryBeatDrafts: value is not []byte")
	}
	var decoded []MainStoryBeatDraft
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*s = decoded
	return nil
}

type MainStoryTemplate struct {
	ID                uuid.UUID           `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time           `json:"createdAt"`
	UpdatedAt         time.Time           `json:"updatedAt"`
	Name              string              `json:"name"`
	Premise           string              `json:"premise"`
	DistrictFit       string              `json:"districtFit" gorm:"column:district_fit"`
	Tone              string              `json:"tone"`
	ThemeTags         StringArray         `json:"themeTags" gorm:"column:theme_tags;type:jsonb"`
	InternalTags      StringArray         `json:"internalTags" gorm:"column:internal_tags;type:jsonb"`
	FactionKeys       StringArray         `json:"factionKeys" gorm:"column:faction_keys;type:jsonb"`
	CharacterKeys     StringArray         `json:"characterKeys" gorm:"column:character_keys;type:jsonb"`
	RevealKeys        StringArray         `json:"revealKeys" gorm:"column:reveal_keys;type:jsonb"`
	ClimaxSummary     string              `json:"climaxSummary" gorm:"column:climax_summary"`
	ResolutionSummary string              `json:"resolutionSummary" gorm:"column:resolution_summary"`
	WhyItWorks        string              `json:"whyItWorks" gorm:"column:why_it_works"`
	Beats             MainStoryBeatDrafts `json:"beats" gorm:"type:jsonb"`
}

func (MainStoryTemplate) TableName() string {
	return "main_story_templates"
}

type MainStoryDistrictBeatRun struct {
	OrderIndex              int        `json:"orderIndex"`
	Act                     int        `json:"act"`
	ChapterTitle            string     `json:"chapterTitle"`
	StoryRole               string     `json:"storyRole"`
	Status                  string     `json:"status"`
	ZoneID                  *uuid.UUID `json:"zoneId,omitempty"`
	ZoneName                string     `json:"zoneName,omitempty"`
	PointOfInterestID       *uuid.UUID `json:"pointOfInterestId,omitempty"`
	PointOfInterestName     string     `json:"pointOfInterestName,omitempty"`
	QuestID                 *uuid.UUID `json:"questId,omitempty"`
	QuestName               string     `json:"questName,omitempty"`
	QuestArchetypeID        *uuid.UUID `json:"questArchetypeId,omitempty"`
	QuestArchetypeName      string     `json:"questArchetypeName,omitempty"`
	QuestGiverCharacterID   *uuid.UUID `json:"questGiverCharacterId,omitempty"`
	QuestGiverCharacterName string     `json:"questGiverCharacterName,omitempty"`
	ErrorMessage            string     `json:"errorMessage,omitempty"`
}

type MainStoryDistrictBeatRuns []MainStoryDistrictBeatRun

func (s MainStoryDistrictBeatRuns) Value() (driver.Value, error) {
	if s == nil {
		return json.Marshal([]MainStoryDistrictBeatRun{})
	}
	return json.Marshal(s)
}

func (s *MainStoryDistrictBeatRuns) Scan(value interface{}) error {
	if value == nil {
		*s = MainStoryDistrictBeatRuns{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan MainStoryDistrictBeatRuns: value is not []byte")
	}
	var decoded []MainStoryDistrictBeatRun
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*s = decoded
	return nil
}

type MainStoryDistrictRun struct {
	ID                    uuid.UUID                 `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt             time.Time                 `json:"createdAt"`
	UpdatedAt             time.Time                 `json:"updatedAt"`
	MainStoryTemplateID   uuid.UUID                 `json:"mainStoryTemplateId" gorm:"column:main_story_template_id;type:uuid"`
	DistrictID            uuid.UUID                 `json:"districtId" gorm:"column:district_id;type:uuid"`
	ZoneID                *uuid.UUID                `json:"zoneId,omitempty" gorm:"column:zone_id;type:uuid"`
	Status                string                    `json:"status"`
	BeatRuns              MainStoryDistrictBeatRuns `json:"beatRuns" gorm:"column:beat_runs;type:jsonb"`
	GeneratedCharacterIDs StringArray               `json:"generatedCharacterIds" gorm:"column:generated_character_ids;type:jsonb"`
	ErrorMessage          *string                   `json:"errorMessage,omitempty" gorm:"column:error_message"`
}

func (MainStoryDistrictRun) TableName() string {
	return "main_story_district_runs"
}

type MainStorySuggestionJob struct {
	ID                           uuid.UUID   `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                    time.Time   `json:"createdAt"`
	UpdatedAt                    time.Time   `json:"updatedAt"`
	Status                       string      `json:"status"`
	Count                        int         `json:"count"`
	QuestCount                   int         `json:"questCount" gorm:"column:quest_count"`
	ThemePrompt                  string      `json:"themePrompt" gorm:"column:theme_prompt"`
	DistrictFit                  string      `json:"districtFit" gorm:"column:district_fit"`
	Tone                         string      `json:"tone"`
	FamilyTags                   StringArray `json:"familyTags" gorm:"column:family_tags;type:jsonb"`
	CharacterTags                StringArray `json:"characterTags" gorm:"column:character_tags;type:jsonb"`
	InternalTags                 StringArray `json:"internalTags" gorm:"column:internal_tags;type:jsonb"`
	RequiredLocationArchetypeIDs StringArray `json:"requiredLocationArchetypeIds" gorm:"column:required_location_archetype_ids;type:jsonb"`
	RequiredLocationMetadataTags StringArray `json:"requiredLocationMetadataTags" gorm:"column:required_location_metadata_tags;type:jsonb"`
	CreatedCount                 int         `json:"createdCount" gorm:"column:created_count"`
	ErrorMessage                 *string     `json:"errorMessage,omitempty"`
}

func (MainStorySuggestionJob) TableName() string {
	return "main_story_suggestion_jobs"
}

type MainStorySuggestionDraft struct {
	ID                  uuid.UUID           `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt           time.Time           `json:"createdAt"`
	UpdatedAt           time.Time           `json:"updatedAt"`
	JobID               uuid.UUID           `json:"jobId" gorm:"column:job_id;type:uuid"`
	Status              string              `json:"status"`
	Name                string              `json:"name"`
	Premise             string              `json:"premise"`
	DistrictFit         string              `json:"districtFit" gorm:"column:district_fit"`
	Tone                string              `json:"tone"`
	ThemeTags           StringArray         `json:"themeTags" gorm:"column:theme_tags;type:jsonb"`
	InternalTags        StringArray         `json:"internalTags" gorm:"column:internal_tags;type:jsonb"`
	FactionKeys         StringArray         `json:"factionKeys" gorm:"column:faction_keys;type:jsonb"`
	CharacterKeys       StringArray         `json:"characterKeys" gorm:"column:character_keys;type:jsonb"`
	RevealKeys          StringArray         `json:"revealKeys" gorm:"column:reveal_keys;type:jsonb"`
	ClimaxSummary       string              `json:"climaxSummary" gorm:"column:climax_summary"`
	ResolutionSummary   string              `json:"resolutionSummary" gorm:"column:resolution_summary"`
	WhyItWorks          string              `json:"whyItWorks" gorm:"column:why_it_works"`
	Beats               MainStoryBeatDrafts `json:"beats" gorm:"type:jsonb"`
	Warnings            StringArray         `json:"warnings" gorm:"type:jsonb"`
	MainStoryTemplateID *uuid.UUID          `json:"mainStoryTemplateId,omitempty" gorm:"column:main_story_template_id;type:uuid"`
	MainStoryTemplate   *MainStoryTemplate  `json:"mainStoryTemplate,omitempty" gorm:"foreignKey:MainStoryTemplateID"`
	ConvertedAt         *time.Time          `json:"convertedAt,omitempty" gorm:"column:converted_at"`
}

func (MainStorySuggestionDraft) TableName() string {
	return "main_story_suggestion_drafts"
}

func NormalizeMainStorySuggestionJobStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case MainStorySuggestionJobStatusInProgress:
		return MainStorySuggestionJobStatusInProgress
	case MainStorySuggestionJobStatusCompleted:
		return MainStorySuggestionJobStatusCompleted
	case MainStorySuggestionJobStatusFailed:
		return MainStorySuggestionJobStatusFailed
	default:
		return MainStorySuggestionJobStatusQueued
	}
}

func NormalizeMainStorySuggestionDraftStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case MainStorySuggestionDraftStatusConverted:
		return MainStorySuggestionDraftStatusConverted
	default:
		return MainStorySuggestionDraftStatusSuggested
	}
}

func NormalizeMainStoryDistrictRunStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case MainStoryDistrictRunStatusQueued:
		return MainStoryDistrictRunStatusQueued
	case MainStoryDistrictRunStatusCompleted:
		return MainStoryDistrictRunStatusCompleted
	case MainStoryDistrictRunStatusFailed:
		return MainStoryDistrictRunStatusFailed
	default:
		return MainStoryDistrictRunStatusInProgress
	}
}
