package models

import (
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
	MonsterID          *uuid.UUID              `json:"monsterId" gorm:"type:uuid"`
	MonsterEncounterID *uuid.UUID              `json:"monsterEncounterId" gorm:"type:uuid"`
	ChallengeID        *uuid.UUID              `json:"challengeId" gorm:"type:uuid"`
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
	return nil
}
