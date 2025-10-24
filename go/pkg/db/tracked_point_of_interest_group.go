package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type trackedPointOfInterestGroupHandle struct {
	db *gorm.DB
}

func (h *trackedPointOfInterestGroupHandle) Create(ctx context.Context, pointOfInterestGroupID uuid.UUID, userID uuid.UUID) error {
	// Check if already exists to make this idempotent
	var existing models.TrackedPointOfInterestGroup
	err := h.db.WithContext(ctx).Where("user_id = ? AND point_of_interest_group_id = ?", userID, pointOfInterestGroupID).First(&existing).Error
	if err == nil {
		// Already exists, return nil (no error)
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		// Some other error occurred
		return err
	}

	// Create new record
	return h.db.Create(&models.TrackedPointOfInterestGroup{
		ID:                     uuid.New(),
		UserID:                 userID,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		PointOfInterestGroupID: pointOfInterestGroupID,
	}).Error
}

func (h *trackedPointOfInterestGroupHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.Unscoped().Delete(&models.TrackedPointOfInterestGroup{}, "point_of_interest_group_id = ?", id).Error
}

func (h *trackedPointOfInterestGroupHandle) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.TrackedPointOfInterestGroup, error) {
	var trackedPointOfInterestGroups []models.TrackedPointOfInterestGroup
	if err := h.db.Where("user_id = ?", userID).Find(&trackedPointOfInterestGroups).Error; err != nil {
		return nil, err
	}
	return trackedPointOfInterestGroups, nil
}
func (h *trackedPointOfInterestGroupHandle) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	return h.db.Unscoped().Where("user_id = ?", userID).Delete(&models.TrackedPointOfInterestGroup{}).Error
}
