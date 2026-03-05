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

const (
	ZoneSeedChallengeShuffleStatusQueued     = "queued"
	ZoneSeedChallengeShuffleStatusInProgress = "in_progress"
	ZoneSeedChallengeShuffleStatusCompleted  = "completed"
	ZoneSeedChallengeShuffleStatusFailed     = "failed"
)

type ZoneSeedJob struct {
	ID                   uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt            time.Time     `json:"createdAt"`
	UpdatedAt            time.Time     `json:"updatedAt"`
	ZoneID               uuid.UUID     `json:"zoneId" gorm:"type:uuid"`
	Status               string        `json:"status"`
	ErrorMessage         *string       `json:"errorMessage,omitempty"`
	PlaceCount           int           `json:"placeCount"`
	CharacterCount       int           `json:"characterCount"`
	QuestCount           int           `json:"questCount"`
	MainQuestCount       int           `json:"mainQuestCount"`
	MonsterCount         int           `json:"monsterCount"`
	InputEncounterCount  int           `json:"inputEncounterCount"`
	OptionEncounterCount int           `json:"optionEncounterCount"`
	TreasureChestCount   int           `json:"treasureChestCount"`
	RequiredPlaceTags    StringArray   `json:"requiredPlaceTags,omitempty" gorm:"type:jsonb"`
	Draft                ZoneSeedDraft `json:"draft" gorm:"type:jsonb"`
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
	MainQuests       []ZoneSeedMainQuestDraft       `json:"mainQuests,omitempty"`
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
	DraftID                uuid.UUID                     `json:"draftId"`
	Name                   string                        `json:"name"`
	Description            string                        `json:"description"`
	AcceptanceDialogue     []string                      `json:"acceptanceDialogue,omitempty"`
	PlaceID                string                        `json:"placeId"`
	QuestGiverDraftID      uuid.UUID                     `json:"questGiverDraftId"`
	ChallengeQuestion      string                        `json:"challengeQuestion,omitempty"`
	ChallengeDifficulty    int                           `json:"challengeDifficulty,omitempty"`
	ChallengeShuffleStatus string                        `json:"challengeShuffleStatus,omitempty"`
	ChallengeShuffleError  *string                       `json:"challengeShuffleError,omitempty"`
	Gold                   int                           `json:"gold"`
	RewardItem             *ZoneSeedQuestRewardItemDraft `json:"rewardItem,omitempty"`
}

type ZoneSeedQuestRewardItemDraft struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	RarityTier  string `json:"rarityTier,omitempty"`
}

type ZoneSeedMainQuestDraft struct {
	DraftID            uuid.UUID                     `json:"draftId"`
	Name               string                        `json:"name"`
	Description        string                        `json:"description"`
	AcceptanceDialogue []string                      `json:"acceptanceDialogue,omitempty"`
	QuestGiverDraftID  uuid.UUID                     `json:"questGiverDraftId"`
	Nodes              []ZoneSeedMainQuestNodeDraft  `json:"nodes,omitempty"`
	Gold               int                           `json:"gold"`
	RewardItem         *ZoneSeedQuestRewardItemDraft `json:"rewardItem,omitempty"`
}

type ZoneSeedMainQuestNodeDraft struct {
	DraftID                uuid.UUID `json:"draftId"`
	OrderIndex             int       `json:"orderIndex"`
	Title                  string    `json:"title,omitempty"`
	Story                  string    `json:"story,omitempty"`
	PlaceID                string    `json:"placeId"`
	ChallengeQuestion      string    `json:"challengeQuestion,omitempty"`
	ChallengeDifficulty    int       `json:"challengeDifficulty,omitempty"`
	ChallengeShuffleStatus string    `json:"challengeShuffleStatus,omitempty"`
	ChallengeShuffleError  *string   `json:"challengeShuffleError,omitempty"`
}

func (d ZoneSeedDraft) Value() (driver.Value, error) {
	if d.FantasyName == "" &&
		d.ZoneDescription == "" &&
		len(d.PointsOfInterest) == 0 &&
		len(d.Characters) == 0 &&
		len(d.Quests) == 0 &&
		len(d.MainQuests) == 0 {
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

func (d *ZoneSeedDraft) SetQuestChallengeShuffleStatus(questDraftID uuid.UUID, status string, errMsg *string) bool {
	for idx := range d.Quests {
		if d.Quests[idx].DraftID != questDraftID {
			continue
		}
		d.Quests[idx].ChallengeShuffleStatus = status
		d.Quests[idx].ChallengeShuffleError = errMsg
		return true
	}
	return false
}

func (d *ZoneSeedDraft) SetMainQuestNodeChallengeShuffleStatus(nodeDraftID uuid.UUID, status string, errMsg *string) bool {
	for mainIdx := range d.MainQuests {
		for nodeIdx := range d.MainQuests[mainIdx].Nodes {
			if d.MainQuests[mainIdx].Nodes[nodeIdx].DraftID != nodeDraftID {
				continue
			}
			d.MainQuests[mainIdx].Nodes[nodeIdx].ChallengeShuffleStatus = status
			d.MainQuests[mainIdx].Nodes[nodeIdx].ChallengeShuffleError = errMsg
			return true
		}
	}
	return false
}
