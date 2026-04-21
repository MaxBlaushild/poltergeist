package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HealingFountain struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ThumbnailURL string    `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	ZoneID       uuid.UUID `json:"zoneId"`
	ZoneKind     string    `json:"zoneKind,omitempty" gorm:"column:zone_kind"`
	Zone         Zone      `json:"zone"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	Geometry     string    `json:"geometry" gorm:"type:geometry(Point,4326)"`
	Invalidated  bool      `json:"invalidated"`
}

func (h *HealingFountain) TableName() string {
	return "healing_fountains"
}

func (h *HealingFountain) BeforeSave(tx *gorm.DB) error {
	if h.Latitude != 0 || h.Longitude != 0 {
		if err := h.SetGeometry(h.Latitude, h.Longitude); err != nil {
			return err
		}
	}
	return nil
}

func (h *HealingFountain) SetGeometry(latitude, longitude float64) error {
	h.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", longitude, latitude)
	return nil
}
