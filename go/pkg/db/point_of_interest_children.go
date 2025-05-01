package db

import (
	"context"

	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pointOfInterestChildrenHandle struct {
	db *gorm.DB
}

func (c *pointOfInterestChildrenHandle) Create(ctx context.Context, pointOfInterestGroupMemberID uuid.UUID, nextPointOfInterestGroupMemberID uuid.UUID, pointOfInterestChallengeID uuid.UUID) error {
	return c.db.Create(&models.PointOfInterestChildren{
		ID:                               uuid.New(),
		PointOfInterestGroupMemberID:     pointOfInterestGroupMemberID,
		NextPointOfInterestGroupMemberID: nextPointOfInterestGroupMemberID,
		PointOfInterestChallengeID:       pointOfInterestChallengeID,
		CreatedAt:                        time.Now(),
		UpdatedAt:                        time.Now(),
	}).Error
}

func (c *pointOfInterestChildrenHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return c.db.Delete(&models.PointOfInterestChildren{}, "id = ?", id).Error
}
