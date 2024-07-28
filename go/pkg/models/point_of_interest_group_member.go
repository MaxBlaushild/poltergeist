package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterestGroupMember struct {
	ID                     uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	PointOfInterestGroupID uuid.UUID
	PointOfInterestGroup   PointOfInterestGroup
	PointOfInterestID      uuid.UUID
	PointOfInterest        PointOfInterest
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
