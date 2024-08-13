package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pointOfInterestHandle struct {
	db *gorm.DB
}

func (c *pointOfInterestHandle) FindAll(ctx context.Context) ([]models.PointOfInterest, error) {
	var pointsOfInterest []models.PointOfInterest

	if err := c.db.WithContext(ctx).Find(&pointsOfInterest).Error; err != nil {
		return nil, err
	}

	return pointsOfInterest, nil
}

func (c *pointOfInterestHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterest, error) {
	var pointOfInterest models.PointOfInterest

	if err := c.db.WithContext(ctx).First(&pointOfInterest, id).Error; err != nil {
		return nil, err
	}

	return &pointOfInterest, nil
}

func (c *pointOfInterestHandle) FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]models.PointOfInterest, error) {
	var pointsOfInterestGroupMembers []models.PointOfInterestGroupMember

	if err := c.db.WithContext(ctx).Where("point_of_interest_group_id = ?", groupID).Find(&pointsOfInterestGroupMembers).Error; err != nil {
		return nil, err
	}

	var ids []uuid.UUID
	for _, member := range pointsOfInterestGroupMembers {
		ids = append(ids, member.PointOfInterestID)
	}

	pointsOfInterest, err := c.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	return pointsOfInterest, nil
}

func (c *pointOfInterestHandle) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.PointOfInterest, error) {
	var pointsOfInterest []models.PointOfInterest

	if err := c.db.WithContext(ctx).Where("id IN (?)", ids).Find(&pointsOfInterest).Error; err != nil {
		return nil, err
	}

	return pointsOfInterest, nil
}

func (c *pointOfInterestHandle) FindByMatchID(ctx context.Context, matchID uuid.UUID) ([]models.PointOfInterest, error) {
	var pointsOfInterest []models.PointOfInterest

	if err := c.db.WithContext(ctx).Where("match_id = ?", matchID).Find(&pointsOfInterest).Error; err != nil {
		return nil, err
	}

	return pointsOfInterest, nil
}

func (c *pointOfInterestHandle) Capture(ctx context.Context, pointOfInterestID uuid.UUID, teamID uuid.UUID, tier int) error {
	updates := models.PointOfInterestTeam{
		CaptureTier: tier,
	}

	return c.db.WithContext(ctx).Model(&models.PointOfInterestTeam{}).Where("team_id = ? AND point_of_interest_id = ?", teamID, pointOfInterestID).Updates(&updates).Error
}

func (c *pointOfInterestHandle) Unlock(ctx context.Context, pointOfInterestID uuid.UUID, teamID uuid.UUID) error {
	unlock := models.PointOfInterestTeam{
		TeamID:            teamID,
		PointOfInterestID: pointOfInterestID,
		Unlocked:          true,
	}

	return c.db.WithContext(ctx).Create(&unlock).Error
}

func (c *pointOfInterestHandle) Create(ctx context.Context, pointOfInterest models.PointOfInterest) error {
	return c.db.WithContext(ctx).Create(&pointOfInterest).Error
}
