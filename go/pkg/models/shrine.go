package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Shrine struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	ShrineTemplateID uuid.UUID      `json:"shrineTemplateId" gorm:"column:shrine_template_id;type:uuid"`
	Template         ShrineTemplate `json:"template" gorm:"foreignKey:ShrineTemplateID"`
	ZoneID           uuid.UUID      `json:"zoneId"`
	ZoneKind         string         `json:"zoneKind,omitempty" gorm:"column:zone_kind"`
	Zone             Zone           `json:"zone"`
	Latitude         float64        `json:"latitude"`
	Longitude        float64        `json:"longitude"`
	Geometry         string         `json:"geometry" gorm:"type:geometry(Point,4326)"`
	CooldownSeconds  int            `json:"cooldownSeconds" gorm:"column:cooldown_seconds"`
	Invalidated      bool           `json:"invalidated"`
	MapMarkerURL     string         `json:"mapMarkerUrl,omitempty" gorm:"-"`
}

func (Shrine) TableName() string {
	return "shrines"
}

func (s *Shrine) BeforeSave(tx *gorm.DB) error {
	if s.Latitude != 0 || s.Longitude != 0 {
		if err := s.SetGeometry(s.Latitude, s.Longitude); err != nil {
			return err
		}
	}
	s.ZoneKind = NormalizeZoneKind(s.ZoneKind)
	if s.CooldownSeconds < 0 {
		s.CooldownSeconds = 0
	}
	return nil
}

func (s *Shrine) SetGeometry(latitude, longitude float64) error {
	s.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", longitude, latitude)
	return nil
}
