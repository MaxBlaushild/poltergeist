package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrackedPointOfInterestGroup struct {
	ID                     uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt              time.Time            `json:"createdAt"`
	UpdatedAt              time.Time            `json:"updatedAt"`
	DeletedAt              gorm.DeletedAt       `json:"deletedAt" gorm:"index"`
	PointOfInterestGroupID uuid.UUID            `json:"pointOfInterestGroupID"`
	PointOfInterestGroup   PointOfInterestGroup `json:"pointOfInterestGroup" gorm:"foreignKey:PointOfInterestGroupID"`
	UserID                 uuid.UUID            `json:"userId"`
	User                   User                 `json:"user" gorm:"foreignKey:UserID"`
}
