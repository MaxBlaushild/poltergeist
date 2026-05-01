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
	QuestArchetypeSuggestionJobStatusQueued     = "queued"
	QuestArchetypeSuggestionJobStatusInProgress = "in_progress"
	QuestArchetypeSuggestionJobStatusCompleted  = "completed"
	QuestArchetypeSuggestionJobStatusFailed     = "failed"

	QuestArchetypeSuggestionDraftStatusSuggested = "suggested"
	QuestArchetypeSuggestionDraftStatusConverted = "converted"
)

var questArchetypeSuggestionKnownFamilySlugs = []string{
	"investigation",
	"delivery",
	"negotiation",
	"pursuit",
	"containment",
	"omen_chasing",
	"ritual_interruption",
	"survival",
	"rescue",
	"combat_finale",
}

var questArchetypeSuggestionKnownFamilySlugSet = func() map[string]struct{} {
	out := make(map[string]struct{}, len(questArchetypeSuggestionKnownFamilySlugs))
	for _, slug := range questArchetypeSuggestionKnownFamilySlugs {
		out[slug] = struct{}{}
	}
	return out
}()

type QuestArchetypeSuggestionStep struct {
	Source                  string                  `json:"source"`
	Content                 string                  `json:"content"`
	LocationConcept         string                  `json:"locationConcept"`
	LocationArchetypeName   string                  `json:"locationArchetypeName"`
	LocationArchetypeID     *uuid.UUID              `json:"locationArchetypeId,omitempty"`
	LocationMetadataTags    []string                `json:"locationMetadataTags"`
	DistanceMeters          *int                    `json:"distanceMeters,omitempty"`
	TemplateConcept         string                  `json:"templateConcept"`
	PotentialContent        []string                `json:"potentialContent"`
	ChallengeQuestion       string                  `json:"challengeQuestion,omitempty"`
	ChallengeDescription    string                  `json:"challengeDescription,omitempty"`
	ChallengeSubmissionType QuestNodeSubmissionType `json:"challengeSubmissionType,omitempty"`
	ChallengeProficiency    *string                 `json:"challengeProficiency,omitempty"`
	ChallengeStatTags       []string                `json:"challengeStatTags,omitempty"`
	ScenarioPrompt          string                  `json:"scenarioPrompt,omitempty"`
	ScenarioOpenEnded       bool                    `json:"scenarioOpenEnded"`
	ScenarioBeats           []string                `json:"scenarioBeats,omitempty"`
	MonsterTemplateNames    []string                `json:"monsterTemplateNames,omitempty"`
	MonsterTemplateIDs      []string                `json:"monsterTemplateIds,omitempty"`
	EncounterTone           []string                `json:"encounterTone,omitempty"`
}

type QuestArchetypeSuggestionSteps []QuestArchetypeSuggestionStep

func (s QuestArchetypeSuggestionSteps) Value() (driver.Value, error) {
	if s == nil {
		return json.Marshal([]QuestArchetypeSuggestionStep{})
	}
	return json.Marshal(s)
}

func (s *QuestArchetypeSuggestionSteps) Scan(value interface{}) error {
	if value == nil {
		*s = QuestArchetypeSuggestionSteps{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan QuestArchetypeSuggestionSteps: value is not []byte")
	}
	var decoded []QuestArchetypeSuggestionStep
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*s = decoded
	return nil
}

type QuestArchetypeSuggestionNodeOutcome struct {
	Outcome     string `json:"outcome"`
	NextNodeKey string `json:"nextNodeKey,omitempty"`
}

type QuestArchetypeSuggestionNodeOutcomes []QuestArchetypeSuggestionNodeOutcome

type QuestArchetypeSuggestionNode struct {
	NodeKey                 string                               `json:"nodeKey"`
	Source                  string                               `json:"source"`
	Content                 string                               `json:"content"`
	LocationConcept         string                               `json:"locationConcept"`
	LocationArchetypeName   string                               `json:"locationArchetypeName"`
	LocationArchetypeID     *uuid.UUID                           `json:"locationArchetypeId,omitempty"`
	LocationMetadataTags    []string                             `json:"locationMetadataTags"`
	DistanceMeters          *int                                 `json:"distanceMeters,omitempty"`
	TemplateConcept         string                               `json:"templateConcept"`
	PotentialContent        []string                             `json:"potentialContent"`
	ChallengeQuestion       string                               `json:"challengeQuestion,omitempty"`
	ChallengeDescription    string                               `json:"challengeDescription,omitempty"`
	ChallengeSubmissionType QuestNodeSubmissionType              `json:"challengeSubmissionType,omitempty"`
	ChallengeProficiency    *string                              `json:"challengeProficiency,omitempty"`
	ChallengeStatTags       []string                             `json:"challengeStatTags,omitempty"`
	ScenarioPrompt          string                               `json:"scenarioPrompt,omitempty"`
	ScenarioOpenEnded       bool                                 `json:"scenarioOpenEnded"`
	ScenarioBeats           []string                             `json:"scenarioBeats,omitempty"`
	MonsterTemplateNames    []string                             `json:"monsterTemplateNames,omitempty"`
	MonsterTemplateIDs      []string                             `json:"monsterTemplateIds,omitempty"`
	EncounterTone           []string                             `json:"encounterTone,omitempty"`
	Outcomes                QuestArchetypeSuggestionNodeOutcomes `json:"outcomes,omitempty"`
}

type QuestArchetypeSuggestionNodes []QuestArchetypeSuggestionNode

func (n QuestArchetypeSuggestionNodes) Value() (driver.Value, error) {
	if n == nil {
		return json.Marshal([]QuestArchetypeSuggestionNode{})
	}
	return json.Marshal(n)
}

func (n *QuestArchetypeSuggestionNodes) Scan(value interface{}) error {
	if value == nil {
		*n = QuestArchetypeSuggestionNodes{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan QuestArchetypeSuggestionNodes: value is not []byte")
	}
	var decoded []QuestArchetypeSuggestionNode
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*n = decoded
	return nil
}

type QuestArchetypeSuggestionFamilyMixTargets map[string]int

func (m QuestArchetypeSuggestionFamilyMixTargets) Value() (driver.Value, error) {
	if m == nil {
		return json.Marshal(map[string]int{})
	}
	return json.Marshal(m)
}

func (m *QuestArchetypeSuggestionFamilyMixTargets) Scan(value interface{}) error {
	if value == nil {
		*m = QuestArchetypeSuggestionFamilyMixTargets{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan QuestArchetypeSuggestionFamilyMixTargets: value is not []byte")
	}
	var decoded map[string]int
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*m = NormalizeQuestArchetypeSuggestionFamilyMixTargets(decoded)
	return nil
}

func QuestArchetypeSuggestionKnownFamilySlugs() []string {
	out := make([]string, len(questArchetypeSuggestionKnownFamilySlugs))
	copy(out, questArchetypeSuggestionKnownFamilySlugs)
	return out
}

func NormalizeQuestArchetypeSuggestionFamilySlug(raw string) string {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		return ""
	}
	var builder strings.Builder
	lastUnderscore := false
	for _, r := range trimmed {
		isAlphaNum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlphaNum {
			builder.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore && builder.Len() > 0 {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}
	normalized := strings.Trim(builder.String(), "_")
	if _, ok := questArchetypeSuggestionKnownFamilySlugSet[normalized]; !ok {
		return ""
	}
	return normalized
}

func NormalizeQuestArchetypeSuggestionFamilyMixTargets(
	raw map[string]int,
) QuestArchetypeSuggestionFamilyMixTargets {
	out := QuestArchetypeSuggestionFamilyMixTargets{}
	for key, count := range raw {
		slug := NormalizeQuestArchetypeSuggestionFamilySlug(key)
		if slug == "" || count <= 0 {
			continue
		}
		out[slug] += count
	}
	return out
}

type QuestArchetypeSuggestionJob struct {
	ID                           uuid.UUID                                `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                    time.Time                                `json:"createdAt"`
	UpdatedAt                    time.Time                                `json:"updatedAt"`
	Status                       string                                   `json:"status"`
	Count                        int                                      `json:"count"`
	ZoneKind                     string                                   `json:"zoneKind" gorm:"column:zone_kind"`
	ThemePrompt                  string                                   `json:"themePrompt" gorm:"column:theme_prompt"`
	FamilyTags                   StringArray                              `json:"familyTags" gorm:"column:family_tags;type:jsonb"`
	FamilyMixTargets             QuestArchetypeSuggestionFamilyMixTargets `json:"familyMixTargets" gorm:"column:family_mix_targets;type:jsonb"`
	CharacterTags                StringArray                              `json:"characterTags" gorm:"column:character_tags;type:jsonb"`
	InternalTags                 StringArray                              `json:"internalTags" gorm:"column:internal_tags;type:jsonb"`
	RequiredLocationArchetypeIDs StringArray                              `json:"requiredLocationArchetypeIds" gorm:"column:required_location_archetype_ids;type:jsonb"`
	RequiredLocationMetadataTags StringArray                              `json:"requiredLocationMetadataTags" gorm:"column:required_location_metadata_tags;type:jsonb"`
	CreatedCount                 int                                      `json:"createdCount" gorm:"column:created_count"`
	ErrorMessage                 *string                                  `json:"errorMessage,omitempty"`
}

func (QuestArchetypeSuggestionJob) TableName() string {
	return "quest_archetype_suggestion_jobs"
}

type QuestArchetypeSuggestionDraft struct {
	ID                          uuid.UUID                     `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                   time.Time                     `json:"createdAt"`
	UpdatedAt                   time.Time                     `json:"updatedAt"`
	JobID                       uuid.UUID                     `json:"jobId" gorm:"column:job_id;type:uuid"`
	Status                      string                        `json:"status"`
	Name                        string                        `json:"name"`
	Hook                        string                        `json:"hook"`
	Description                 string                        `json:"description"`
	ZoneKind                    string                        `json:"zoneKind" gorm:"column:zone_kind"`
	AcceptanceDialogue          StringArray                   `json:"acceptanceDialogue" gorm:"column:acceptance_dialogue;type:jsonb"`
	CharacterTags               StringArray                   `json:"characterTags" gorm:"column:character_tags;type:jsonb"`
	InternalTags                StringArray                   `json:"internalTags" gorm:"column:internal_tags;type:jsonb"`
	DifficultyMode              QuestDifficultyMode           `json:"difficultyMode" gorm:"column:difficulty_mode"`
	Difficulty                  int                           `json:"difficulty"`
	MonsterEncounterTargetLevel int                           `json:"monsterEncounterTargetLevel" gorm:"column:monster_encounter_target_level"`
	WhyThisScales               string                        `json:"whyThisScales" gorm:"column:why_this_scales"`
	Steps                       QuestArchetypeSuggestionSteps `json:"steps" gorm:"type:jsonb"`
	Nodes                       QuestArchetypeSuggestionNodes `json:"nodes" gorm:"column:nodes;type:jsonb"`
	ChallengeTemplateSeeds      StringArray                   `json:"challengeTemplateSeeds" gorm:"column:challenge_template_seeds;type:jsonb"`
	ScenarioTemplateSeeds       StringArray                   `json:"scenarioTemplateSeeds" gorm:"column:scenario_template_seeds;type:jsonb"`
	MonsterTemplateSeeds        StringArray                   `json:"monsterTemplateSeeds" gorm:"column:monster_template_seeds;type:jsonb"`
	Warnings                    StringArray                   `json:"warnings" gorm:"type:jsonb"`
	QuestArchetypeID            *uuid.UUID                    `json:"questArchetypeId,omitempty" gorm:"column:quest_archetype_id;type:uuid"`
	QuestArchetype              *QuestArchetype               `json:"questArchetype,omitempty" gorm:"foreignKey:QuestArchetypeID"`
	ConvertedAt                 *time.Time                    `json:"convertedAt,omitempty" gorm:"column:converted_at"`
}

func (QuestArchetypeSuggestionDraft) TableName() string {
	return "quest_archetype_suggestion_drafts"
}

func NormalizeQuestArchetypeSuggestionJobStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case QuestArchetypeSuggestionJobStatusInProgress:
		return QuestArchetypeSuggestionJobStatusInProgress
	case QuestArchetypeSuggestionJobStatusCompleted:
		return QuestArchetypeSuggestionJobStatusCompleted
	case QuestArchetypeSuggestionJobStatusFailed:
		return QuestArchetypeSuggestionJobStatusFailed
	default:
		return QuestArchetypeSuggestionJobStatusQueued
	}
}

func NormalizeQuestArchetypeSuggestionDraftStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case QuestArchetypeSuggestionDraftStatusConverted:
		return QuestArchetypeSuggestionDraftStatusConverted
	default:
		return QuestArchetypeSuggestionDraftStatusSuggested
	}
}
