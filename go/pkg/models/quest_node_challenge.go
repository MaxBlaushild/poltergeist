package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	QuestNodeChallengeShuffleStatusIdle       = "idle"
	QuestNodeChallengeShuffleStatusQueued     = "queued"
	QuestNodeChallengeShuffleStatusInProgress = "in_progress"
	QuestNodeChallengeShuffleStatusCompleted  = "completed"
	QuestNodeChallengeShuffleStatusFailed     = "failed"
)

type QuestNodeChallenge struct {
	ID                     uuid.UUID               `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt              time.Time               `json:"createdAt"`
	UpdatedAt              time.Time               `json:"updatedAt"`
	QuestNodeID            uuid.UUID               `json:"questNodeId" gorm:"type:uuid"`
	Tier                   int                     `json:"tier"`
	Question               string                  `json:"question"`
	Reward                 int                     `json:"reward"`
	InventoryItemID        *int                    `json:"inventoryItemId"`
	SubmissionType         QuestNodeSubmissionType `json:"submissionType" gorm:"type:text;default:photo"`
	Difficulty             int                     `json:"difficulty" gorm:"default:0"`
	StatTags               StringArray             `json:"statTags,omitempty" gorm:"type:jsonb"`
	Proficiency            *string                 `json:"proficiency,omitempty"`
	ChallengeShuffleStatus string                  `json:"challengeShuffleStatus,omitempty" gorm:"type:text;default:'idle'"`
	ChallengeShuffleError  *string                 `json:"challengeShuffleError,omitempty"`
}

func (q *QuestNodeChallenge) TableName() string {
	return "quest_node_challenges"
}
