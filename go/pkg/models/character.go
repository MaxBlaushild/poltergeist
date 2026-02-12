package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Character struct {
	ID                    uuid.UUID           `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt             time.Time           `json:"createdAt"`
	UpdatedAt             time.Time           `json:"updatedAt"`
	Name                  string              `json:"name"`
	Description           string              `json:"description"`
	MapIconURL            string              `json:"mapIconUrl"`
	DialogueImageURL      string              `json:"dialogueImageUrl"`
	ThumbnailURL          string              `json:"thumbnailUrl"`
	ImageGenerationStatus string              `json:"imageGenerationStatus" gorm:"column:image_generation_status"`
	ImageGenerationError  *string             `json:"imageGenerationError,omitempty" gorm:"column:image_generation_error"`
	Locations             []CharacterLocation `json:"locations" gorm:"foreignKey:CharacterID"`
	PointOfInterestID     *uuid.UUID          `json:"pointOfInterestId,omitempty" gorm:"type:uuid"`
	PointOfInterest       *PointOfInterest    `json:"pointOfInterest,omitempty" gorm:"foreignKey:PointOfInterestID"`
	Geometry              string              `json:"geometry" gorm:"type:geometry(Point,4326)"`
	MovementPattern       MovementPattern     `json:"movementPattern"`
	MovementPatternID     uuid.UUID           `json:"movementPatternId" gorm:"type:uuid"`
}

func (n *Character) TableName() string {
	return "characters"
}

// SetGeometry creates a PostGIS geometry point from lat/lng coordinates
func (c *Character) SetGeometry(lat float64, lng float64) {
	// Create WKT (Well-Known Text) format: 'SRID=4326;POINT(lng lat)'
	c.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", lng, lat)
}

// BeforeSave hook to auto-populate geometry from movement pattern's starting position if geometry is empty
func (c *Character) BeforeSave(tx *gorm.DB) error {
	// If geometry is empty and movement pattern is loaded, use starting position
	if c.Geometry == "" && c.MovementPattern.StartingLatitude != 0 && c.MovementPattern.StartingLongitude != 0 {
		c.SetGeometry(c.MovementPattern.StartingLatitude, c.MovementPattern.StartingLongitude)
	}
	return nil
}

type MovementPatternType string

const (
	MovementPatternStatic MovementPatternType = "static"
	MovementPatternRandom MovementPatternType = "random"
	MovementPatternPath   MovementPatternType = "path"
)

const (
	CharacterImageGenerationStatusNone       = "none"
	CharacterImageGenerationStatusQueued     = "queued"
	CharacterImageGenerationStatusInProgress = "in_progress"
	CharacterImageGenerationStatusComplete   = "complete"
	CharacterImageGenerationStatusFailed     = "failed"
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
