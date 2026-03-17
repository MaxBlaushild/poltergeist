package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserID    uuid.UUID `json:"userId"`
	User      User      `json:"user"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Geometry  string    `json:"geometry" gorm:"type:geometry(Point,4326)"`
}

func (Base) TableName() string {
	return "bases"
}

func (b *Base) BeforeSave(tx *gorm.DB) error {
	if b.Latitude != 0 || b.Longitude != 0 {
		if err := b.SetGeometry(b.Latitude, b.Longitude); err != nil {
			return err
		}
	}
	return nil
}

func (b *Base) SetGeometry(latitude, longitude float64) error {
	b.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", longitude, latitude)
	return nil
}
