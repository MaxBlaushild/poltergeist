package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Resource struct {
	ID             uuid.UUID    `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
	ZoneID         uuid.UUID    `json:"zoneId"`
	Zone           Zone         `json:"zone"`
	ResourceTypeID uuid.UUID    `json:"resourceTypeId" gorm:"column:resource_type_id"`
	ResourceType   ResourceType `json:"resourceType"`
	Quantity       int          `json:"quantity"`
	Latitude       float64      `json:"latitude"`
	Longitude      float64      `json:"longitude"`
	Geometry       string       `json:"geometry" gorm:"type:geometry(Point,4326)"`
	Invalidated    bool         `json:"invalidated"`
}

func (Resource) TableName() string {
	return "resources"
}

func (r *Resource) BeforeSave(tx *gorm.DB) error {
	if r.Latitude != 0 || r.Longitude != 0 {
		if err := r.SetGeometry(r.Latitude, r.Longitude); err != nil {
			return err
		}
	}
	return nil
}

func (r *Resource) SetGeometry(latitude, longitude float64) error {
	r.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", longitude, latitude)
	return nil
}
