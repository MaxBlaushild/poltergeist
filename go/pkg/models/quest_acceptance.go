package models

import (
	"time"

	"github.com/google/uuid"
)

type QuestAcceptance struct {
	ID                     uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt              time.Time            `json:"createdAt"`
	UpdatedAt              time.Time            `json:"updatedAt"`
	UserID                 uuid.UUID            `json:"userId" gorm:"type:uuid"`
	User                   User                 `json:"user,omitempty" gorm:"foreignKey:UserID"`
	PointOfInterestGroupID uuid.UUID            `json:"pointOfInterestGroupId" gorm:"type:uuid"`
	PointOfInterestGroup   PointOfInterestGroup `json:"pointOfInterestGroup,omitempty" gorm:"foreignKey:PointOfInterestGroupID"`
	CharacterID            uuid.UUID            `json:"characterId" gorm:"type:uuid"`
	Character              Character            `json:"character,omitempty" gorm:"foreignKey:CharacterID"`
	TurnedInAt             *time.Time           `json:"turnedInAt,omitempty"`
}

func (q *QuestAcceptance) TableName() string {
	return "quest_acceptances"
}
