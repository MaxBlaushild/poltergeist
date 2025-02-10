package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterestGroupMember struct {
	ID                     uuid.UUID                 `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	PointOfInterestGroupID uuid.UUID                 `json:"pointOfInterestGroupId"`
	PointOfInterestGroup   PointOfInterestGroup      `gorm:"foreignKey:PointOfInterestGroupID" json:"pointOfInterestGroup"`
	PointOfInterestID      uuid.UUID                 `json:"pointOfInterestId"`
	PointOfInterest        PointOfInterest           `gorm:"foreignKey:PointOfInterestID" json:"pointOfInterest"`
	CreatedAt              time.Time                 `json:"createdAt"`
	UpdatedAt              time.Time                 `json:"updatedAt"`
	Children               []PointOfInterestChildren `gorm:"foreignKey:PointOfInterestGroupMemberID" json:"children"`
}
