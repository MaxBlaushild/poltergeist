package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Character struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	MapIconURL        string          `json:"mapIconUrl"`
	DialogueImageURL  string          `json:"dialogueUrl"`
	LocationID        uuid.UUID       `json:"locationId" gorm:"type:uuid"`
	Location          PointOfInterest `json:"location" gorm:"foreignKey:LocationID"`
	MovementPattern   MovementPattern `json:"movementPattern"`
	MovementPatternID uuid.UUID       `json:"movementPatternId" gorm:"type:uuid"`
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

// LocationPath is a custom type for []Location that implements sql.Scanner and driver.Valuer
type LocationPath []Location

// Scan implements the sql.Scanner interface for reading from database
func (lp *LocationPath) Scan(value interface{}) error {
	if value == nil {
		*lp = []Location{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan LocationPath: value is not []byte")
	}

	var locations []Location
	if err := json.Unmarshal(bytes, &locations); err != nil {
		return err
	}

	*lp = locations
	return nil
}

// Value implements the driver.Valuer interface for writing to database
func (lp LocationPath) Value() (driver.Value, error) {
	if lp == nil {
		return json.Marshal([]Location{})
	}
	return json.Marshal(lp)
}

type MovementPattern struct {
	ID                  uuid.UUID           `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt           time.Time           `json:"createdAt"`
	UpdatedAt           time.Time           `json:"updatedAt"`
	MovementPatternType MovementPatternType `json:"movementPatternType"`
	ZoneID              *uuid.UUID          `json:"zoneId" gorm:"type:uuid"`
	StartingLatitude    float64             `json:"startingLatitude"`
	StartingLongitude   float64             `json:"startingLongitude"`
	Path                LocationPath        `json:"path" gorm:"type:jsonb"`
}

func (m *MovementPattern) TableName() string {
	return "movement_patterns"
}
