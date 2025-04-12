package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PointOfInterestZone struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt  `json:"deletedAt" gorm:"index"`
	ZoneID            uuid.UUID       `json:"zoneID"`
	Zone              Zone            `json:"zone"`
	PointOfInterestID uuid.UUID       `json:"pointOfInterestID"`
	PointOfInterest   PointOfInterest `json:"pointOfInterest"`
}
