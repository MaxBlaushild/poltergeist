package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestNode struct {
	ID                 uuid.UUID               `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt          time.Time               `json:"createdAt"`
	UpdatedAt          time.Time               `json:"updatedAt"`
	QuestID            uuid.UUID               `json:"questId" gorm:"type:uuid"`
	OrderIndex         int                     `json:"orderIndex"`
	ScenarioID         *uuid.UUID              `json:"scenarioId" gorm:"type:uuid"`
	Scenario           *Scenario               `json:"scenario,omitempty" gorm:"foreignKey:ScenarioID"`
	MonsterID          *uuid.UUID              `json:"monsterId" gorm:"type:uuid"`
	Monster            *Monster                `json:"monster,omitempty" gorm:"foreignKey:MonsterID"`
	MonsterEncounterID *uuid.UUID              `json:"monsterEncounterId" gorm:"type:uuid"`
	MonsterEncounter   *MonsterEncounter       `json:"monsterEncounter,omitempty" gorm:"foreignKey:MonsterEncounterID"`
	ChallengeID        *uuid.UUID              `json:"challengeId" gorm:"type:uuid"`
	Challenge          *Challenge              `json:"challenge,omitempty" gorm:"foreignKey:ChallengeID"`
	ExpositionID       *uuid.UUID              `json:"expositionId" gorm:"type:uuid"`
	Exposition         *Exposition             `json:"exposition,omitempty" gorm:"foreignKey:ExpositionID"`
	FetchCharacterID   *uuid.UUID              `json:"fetchCharacterId,omitempty" gorm:"column:fetch_character_id;type:uuid"`
	FetchCharacter     *Character              `json:"fetchCharacter,omitempty" gorm:"foreignKey:FetchCharacterID"`
	FetchRequirements  FetchQuestRequirements  `json:"fetchRequirements" gorm:"column:fetch_requirements_json;type:jsonb;default:'[]'"`
	StoryFlagKey       string                  `json:"storyFlagKey,omitempty" gorm:"column:story_flag_key"`
	SubmissionType     QuestNodeSubmissionType `json:"submissionType" gorm:"type:text;default:photo"`
	Children           []QuestNodeChild        `json:"children" gorm:"foreignKey:QuestNodeID"`
}

type QuestNodeSubmissionType string

const (
	QuestNodeSubmissionTypeText  QuestNodeSubmissionType = "text"
	QuestNodeSubmissionTypePhoto QuestNodeSubmissionType = "photo"
	QuestNodeSubmissionTypeVideo QuestNodeSubmissionType = "video"
)

func (t QuestNodeSubmissionType) IsValid() bool {
	switch t {
	case QuestNodeSubmissionTypeText, QuestNodeSubmissionTypePhoto, QuestNodeSubmissionTypeVideo:
		return true
	default:
		return false
	}
}

func DefaultQuestNodeSubmissionType() QuestNodeSubmissionType {
	return QuestNodeSubmissionTypePhoto
}

func (q *QuestNode) TableName() string {
	return "quest_nodes"
}

func (q *QuestNode) BeforeCreate(tx *gorm.DB) (err error) {
	if q.SubmissionType == "" {
		q.SubmissionType = DefaultQuestNodeSubmissionType()
	}
	q.StoryFlagKey = NormalizeStoryFlagKey(q.StoryFlagKey)
	q.FetchRequirements = NormalizeFetchQuestRequirements(q.FetchRequirements)
	return nil
}

func (q *QuestNode) StoryFlagKeyNormalized() string {
	if q == nil {
		return ""
	}
	return NormalizeStoryFlagKey(q.StoryFlagKey)
}

func (q *QuestNode) IsStoryFlagNode() bool {
	return q != nil && strings.TrimSpace(q.StoryFlagKeyNormalized()) != ""
}

func (q *QuestNode) IsFetchQuestNode() bool {
	return q != nil &&
		q.FetchCharacterID != nil &&
		*q.FetchCharacterID != uuid.Nil &&
		len(q.FetchRequirements) > 0
}
