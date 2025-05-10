package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Point struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Geometry  string    `json:"geometry" gorm:"type:geometry(Point,4326)"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
}

func (p *Point) BeforeSave(tx *gorm.DB) error {
	if p.Latitude != 0 && p.Longitude != 0 {
		if err := p.SetGeometry(p.Latitude, p.Longitude); err != nil {
			return err
		}
	}
	return nil
}

func (p *Point) SetGeometry(latitude float64, longitude float64) error {
	// Create WKT (Well-Known Text) format: 'SRID=4326;POINT(lng lat)'
	p.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", longitude, latitude)
	return nil
}

func (p *Point) TableName() string {
	return "points"
}
