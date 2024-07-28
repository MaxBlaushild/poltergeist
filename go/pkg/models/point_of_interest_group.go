package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterestGroup struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Name             string
	GroupMembers     []PointOfInterestGroupMember `gorm:"foreignKey:PointOfInterestGroupID"`
	PointsOfInterest []PointOfInterest            `gorm:"many2many:point_of_interest_group_members;associationForeignKey:PointOfInterestID;foreignKey:ID;joinForeignKey:PointOfInterestGroupID;joinReferences:PointOfInterestID"`
}
