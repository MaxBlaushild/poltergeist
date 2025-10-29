package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// These are active state changes that immediately effect the user experience

type ActivityType string

const (
	ActivityTypeLevelUp            ActivityType = "level_up"
	ActivityTypeChallengeCompleted ActivityType = "challenge_completed"
	ActivityTypeQuestCompleted     ActivityType = "quest_completed"
	ActivityTypeItemReceived       ActivityType = "item_received"
	ActivityTypeReputationUp       ActivityType = "reputation_up"
)

type LevelUpActivity struct {
	NewLevel int `json:"newLevel"`
}

type ItemAwarded struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"imageUrl"`
}

type POIInfo struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	ImageURL string    `json:"imageURL"`
}

type ChallengeCompletedActivity struct {
	ChallengeID       uuid.UUID     `json:"challengeId"`
	Successful        bool          `json:"successful"`
	Reason            string        `json:"reason"`
	SubmitterID       *uuid.UUID    `json:"submitterId,omitempty"`
	ExperienceAwarded int           `json:"experienceAwarded"`
	ReputationAwarded int           `json:"reputationAwarded"`
	ItemsAwarded      []ItemAwarded `json:"itemsAwarded"`
	GoldAwarded       int           `json:"goldAwarded"`
	QuestID           uuid.UUID     `json:"questId"`
	QuestName         string        `json:"questName"`
	QuestCompleted    bool          `json:"questCompleted"`
	CurrentPOI        POIInfo       `json:"currentPOI"`
	NextPOI           *POIInfo      `json:"nextPOI,omitempty"`
	ZoneID            uuid.UUID     `json:"zoneId"`
	ZoneName          string        `json:"zoneName"`
}

type QuestCompletedActivity struct {
	QuestID     uuid.UUID `json:"questId"`
	GoldAwarded int       `json:"goldAwarded"`
}

type ItemReceivedActivity struct {
	ItemID   int    `json:"itemId"`
	ItemName string `json:"itemName"`
}

type ReputationUpActivity struct {
	NewLevel int       `json:"newLevel"`
	ZoneName string    `json:"zoneName"`
	ZoneID   uuid.UUID `json:"zoneId"`
}

type Activity struct {
	ID           uuid.UUID      `json:"id"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	UserID       uuid.UUID      `json:"userId"`
	ActivityType ActivityType   `json:"activityType"`
	Data         datatypes.JSON `json:"data"`
	Seen         bool           `json:"seen"`
}
