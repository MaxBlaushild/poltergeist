package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pointOfInterestGroupHandle struct {
	db *gorm.DB
}

func (c *pointOfInterestGroupHandle) Create(ctx context.Context, pointOfInterestIDs []uuid.UUID, name string) (*models.PointOfInterestGroup, error) {

	pointOfInterestGroup := models.PointOfInterestGroup{
		Name: name,
	}

	if err := c.db.Create(&pointOfInterestGroup).Error; err != nil {
		return nil, err
	}

	var pointOfInterestGroupMembers []models.PointOfInterestGroupMember
	for _, pointOfInterestID := range pointOfInterestIDs {
		pointOfInterestGroupMembers = append(pointOfInterestGroupMembers, models.PointOfInterestGroupMember{
			PointOfInterestID:      pointOfInterestID,
			PointOfInterestGroupID: pointOfInterestGroup.ID,
		})
	}

	if err := c.db.Create(&pointOfInterestGroupMembers).Error; err != nil {
		return nil, err
	}

	pointOfInterestGroup.GroupMembers = pointOfInterestGroupMembers

	return &pointOfInterestGroup, nil
}

func (c *pointOfInterestGroupHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestGroup, error) {
	var pointOfInterestGroup models.PointOfInterestGroup
	if err := c.db.Preload("PointsOfInterest").First(&pointOfInterestGroup, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &pointOfInterestGroup, nil
}

func (c *pointOfInterestGroupHandle) FindAll(ctx context.Context) ([]*models.PointOfInterestGroup, error) {
	var pointOfInterestGroups []*models.PointOfInterestGroup
	if err := c.db.Preload("PointsOfInterest").Find(&pointOfInterestGroups).Error; err != nil {
		return nil, err
	}
	return pointOfInterestGroups, nil
}
