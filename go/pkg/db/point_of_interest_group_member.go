package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pointOfInterestGroupMemberHandle struct {
	db *gorm.DB
}

func (h *pointOfInterestGroupMemberHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestGroupMember, error) {
	var member models.PointOfInterestGroupMember
	if err := h.db.WithContext(ctx).First(&member, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (h *pointOfInterestGroupMemberHandle) FindByPointOfInterestAndGroup(ctx context.Context, pointOfInterestID uuid.UUID, groupID uuid.UUID) (*models.PointOfInterestGroupMember, error) {
	var member models.PointOfInterestGroupMember
	if err := h.db.WithContext(ctx).Where("point_of_interest_id = ? AND point_of_interest_group_id = ?", pointOfInterestID, groupID).First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}
