package models

import (
	"time"

	"github.com/google/uuid"
)

type PointOfInterestGroupType int

const (
	PointOfInterestGroupTypeUnassigned PointOfInterestGroupType = iota
	PointOfInterestGroupTypeArena
	PointOfInterestGroupTypeQuest
)

type PointOfInterestGroup struct {
	ID               uuid.UUID                    `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt        time.Time                    `json:"createdAt"`
	UpdatedAt        time.Time                    `json:"updatedAt"`
	Name             string                       `json:"name"`
	Description      string                       `json:"description"`
	ImageUrl         string                       `json:"imageUrl"`
	Hidden           bool                         `json:"hidden" gorm:"default:false"`
	GroupMembers     []PointOfInterestGroupMember `json:"groupMembers" gorm:"foreignKey:PointOfInterestGroupID"`
	PointsOfInterest []PointOfInterest            `json:"pointsOfInterest" gorm:"many2many:point_of_interest_group_members;associationForeignKey:PointOfInterestID;foreignKey:ID;joinForeignKey:PointOfInterestGroupID;joinReferences:PointOfInterestID"`
	Type             PointOfInterestGroupType     `json:"type"`
}

func (p *PointOfInterestGroup) GetRootMember() *PointOfInterestGroupMember {
	for _, member := range p.GroupMembers {
		isChild := false
		for _, otherMember := range p.GroupMembers {
			for _, child := range otherMember.Children {
				if child.NextPointOfInterestGroupMemberID == member.ID {
					isChild = true
					break
				}
			}
			if isChild {
				break
			}
		}
		if !isChild {
			return &member
		}
	}
	return nil
}
