package models

import (
	"time"

	"github.com/google/uuid"
)

type Quest struct {
	ID                    uuid.UUID         `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt             time.Time         `json:"createdAt"`
	UpdatedAt             time.Time         `json:"updatedAt"`
	Name                  string            `json:"name"`
	Description           string            `json:"description"`
	ImageURL              string            `json:"imageUrl"`
	ZoneID                *uuid.UUID        `json:"zoneId" gorm:"type:uuid"`
	QuestArchetypeID      *uuid.UUID        `json:"questArchetypeId" gorm:"type:uuid"`
	QuestGiverCharacterID *uuid.UUID        `json:"questGiverCharacterId" gorm:"type:uuid"`
	Gold                  int               `json:"gold"`
	ItemRewards           []QuestItemReward `json:"itemRewards" gorm:"foreignKey:QuestID"`
	Nodes                 []QuestNode       `json:"nodes" gorm:"foreignKey:QuestID"`
}

func (q *Quest) TableName() string {
	return "quests"
}
