package models

import (
	"time"

	"github.com/google/uuid"
)

type Character struct {
	ID               uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	MapIconURL       string          `json:"mapIconUrl"`
	DialogueImageURL string          `json:"dialogueUrl"`
	LocationID       uuid.UUID       `json:"locationId" gorm:"type:uuid"`
	Location         Location        `json:"location" gorm:"foreignKey:LocationID"`
	MovementPattern  MovementPattern `json:"movementPattern"`
}

func (n *Character) TableName() string {
	return "characters"
}

type MovementPatternType string

const (
	MovementPatternStatic MovementPatternType = "static"
	MovementPatternRandom MovementPatternType = "random"
	MovementPatternPath   MovementPatternType = "path"
)

type MovementPattern struct {
	ID                  uuid.UUID           `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt           time.Time           `json:"createdAt"`
	UpdatedAt           time.Time           `json:"updatedAt"`
	MovementPatternType MovementPatternType `json:"movementPatternType"`
	ZoneID              *uuid.UUID          `json:"zoneId" gorm:"type:uuid"`
	StartingLatitude    float64             `json:"startingLatitude"`
	StartingLongitude   float64             `json:"startingLongitude"`
	Path                []Location          `json:"path"`
}

func (m *MovementPattern) TableName() string {
	return "movement_patterns"
}
