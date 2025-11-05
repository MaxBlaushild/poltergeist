package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TreasureChest struct {
	ID          uuid.UUID           `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
	Latitude    float64             `json:"latitude"`
	Longitude   float64             `json:"longitude"`
	ZoneID      uuid.UUID           `json:"zoneId"`
	Zone        Zone                `json:"zone"`
	Gold        *int                `json:"gold"`
	Geometry    string              `json:"geometry" gorm:"type:geometry(Point,4326)"`
	Invalidated bool                `json:"invalidated"`
	Items       []TreasureChestItem `json:"items" gorm:"foreignKey:TreasureChestID"`
}

func (t *TreasureChest) TableName() string {
	return "treasure_chests"
}

func (t *TreasureChest) BeforeSave(tx *gorm.DB) error {
	if t.Latitude != 0 && t.Longitude != 0 {
		if err := t.SetGeometry(t.Latitude, t.Longitude); err != nil {
			return err
		}
	}
	return nil
}

func (t *TreasureChest) SetGeometry(latitude float64, longitude float64) error {
	// Create WKT (Well-Known Text) format: 'SRID=4326;POINT(lng lat)'
	t.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", longitude, latitude)
	return nil
}
