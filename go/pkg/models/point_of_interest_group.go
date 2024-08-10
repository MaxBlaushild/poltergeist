package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterestGroup struct {
	ID               uuid.UUID                    `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt        time.Time                    `json:"createdAt"`
	UpdatedAt        time.Time                    `json:"updatedAt"`
	Name             string                       `json:"name"`
	Description      string                       `json:"description"`
	GroupMembers     []PointOfInterestGroupMember `json:"groupMembers" gorm:"foreignKey:PointOfInterestGroupID"`
	PointsOfInterest []PointOfInterest            `json:"pointsOfInterest" gorm:"many2many:point_of_interest_group_members;associationForeignKey:PointOfInterestID;foreignKey:ID;joinForeignKey:PointOfInterestGroupID;joinReferences:PointOfInterestID"`
}
