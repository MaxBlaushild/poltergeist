package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	ZoneSeedStatusQueued           = "queued"
	ZoneSeedStatusInProgress       = "in_progress"
	ZoneSeedStatusAwaitingApproval = "awaiting_approval"
	ZoneSeedStatusApproved         = "approved"
	ZoneSeedStatusApplying         = "applying"
	ZoneSeedStatusApplied          = "applied"
	ZoneSeedStatusFailed           = "failed"
)

type ZoneSeedJob struct {
	ID             uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
	ZoneID         uuid.UUID     `json:"zoneId" gorm:"type:uuid"`
	Status         string        `json:"status"`
	ErrorMessage   *string       `json:"errorMessage,omitempty"`
	PlaceCount     int           `json:"placeCount"`
	CharacterCount int           `json:"characterCount"`
	QuestCount     int           `json:"questCount"`
	Draft          ZoneSeedDraft `json:"draft" gorm:"type:jsonb"`
}

func (ZoneSeedJob) TableName() string {
	return "zone_seed_jobs"
}

type ZoneSeedDraft struct {
	FantasyName      string                         `json:"fantasyName,omitempty"`
	ZoneDescription  string                         `json:"zoneDescription,omitempty"`
	PointsOfInterest []ZoneSeedPointOfInterestDraft `json:"pointsOfInterest,omitempty"`
	Characters       []ZoneSeedCharacterDraft       `json:"characters,omitempty"`
	Quests           []ZoneSeedQuestDraft           `json:"quests,omitempty"`
}

type ZoneSeedPointOfInterestDraft struct {
	DraftID          uuid.UUID `json:"draftId"`
	PlaceID          string    `json:"placeId"`
	Name             string    `json:"name"`
	Address          string    `json:"address,omitempty"`
	Types            []string  `json:"types,omitempty"`
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	Rating           float64   `json:"rating,omitempty"`
	UserRatingCount  int32     `json:"userRatingCount,omitempty"`
	EditorialSummary string    `json:"editorialSummary,omitempty"`
}

type ZoneSeedCharacterDraft struct {
	DraftID     uuid.UUID `json:"draftId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PlaceID     string    `json:"placeId"`
}

type ZoneSeedQuestDraft struct {
	DraftID             uuid.UUID                     `json:"draftId"`
	Name                string                        `json:"name"`
	Description         string                        `json:"description"`
	AcceptanceDialogue  []string                      `json:"acceptanceDialogue,omitempty"`
	PlaceID             string                        `json:"placeId"`
	QuestGiverDraftID   uuid.UUID                     `json:"questGiverDraftId"`
	ChallengeQuestion   string                        `json:"challengeQuestion,omitempty"`
	ChallengeDifficulty int                           `json:"challengeDifficulty,omitempty"`
	Gold                int                           `json:"gold"`
	RewardItem          *ZoneSeedQuestRewardItemDraft `json:"rewardItem,omitempty"`
}

type ZoneSeedQuestRewardItemDraft struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	RarityTier  string `json:"rarityTier,omitempty"`
}

func (d ZoneSeedDraft) Value() (driver.Value, error) {
	if d.FantasyName == "" &&
		d.ZoneDescription == "" &&
		len(d.PointsOfInterest) == 0 &&
		len(d.Characters) == 0 &&
		len(d.Quests) == 0 {
		return nil, nil
	}
	return json.Marshal(d)
}

func (d *ZoneSeedDraft) Scan(value interface{}) error {
	if value == nil {
		*d = ZoneSeedDraft{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to scan ZoneSeedDraft: unsupported type")
	}

	if len(bytes) == 0 {
		*d = ZoneSeedDraft{}
		return nil
	}

	return json.Unmarshal(bytes, d)
}
