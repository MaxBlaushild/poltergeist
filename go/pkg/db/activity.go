package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type activityHandle struct {
	db *gorm.DB
}

func (h *activityHandle) GetFeed(ctx context.Context, userID uuid.UUID) ([]models.Activity, error) {
	var activities []models.Activity
	if err := h.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&activities).Error; err != nil {
		return nil, err
	}
	return activities, nil
}

func (h *activityHandle) MarkAsSeen(ctx context.Context, activityIDs []uuid.UUID) error {
	return h.db.WithContext(ctx).Where("id IN (?)", activityIDs).Updates(&models.Activity{Seen: true}).Error
}

func (h *activityHandle) CreateActivity(ctx context.Context, activity models.Activity) error {
	if activity.CreatedAt.IsZero() {
		activity.CreatedAt = time.Now()
	}
	if activity.UpdatedAt.IsZero() {
		activity.UpdatedAt = time.Now()
	}
	if activity.ID == uuid.Nil {
		activity.ID = uuid.New()
	}
	return h.db.WithContext(ctx).Create(&activity).Error
}

func (h *activityHandle) CreateActivitiesForPartyMembers(ctx context.Context, partyID *uuid.UUID, userID *uuid.UUID, activityType models.ActivityType, data []byte) error {
	var targetUserIDs []uuid.UUID

	// If user is in a party, create activities for all party members
	if partyID != nil {
		var users []models.User
		if err := h.db.WithContext(ctx).Where("party_id = ?", partyID).Find(&users).Error; err != nil {
			return err
		}
		for _, user := range users {
			targetUserIDs = append(targetUserIDs, user.ID)
		}
	} else if userID != nil {
		// If not in a party, just create for the single user
		targetUserIDs = append(targetUserIDs, *userID)
	} else {
		return nil // No user or party specified
	}

	// Create one activity per target user
	for _, targetUserID := range targetUserIDs {
		activity := models.Activity{
			UserID:       targetUserID,
			ActivityType: activityType,
			Data:         data,
			Seen:         false,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			ID:           uuid.New(),
		}
		if err := h.db.WithContext(ctx).Create(&activity).Error; err != nil {
			return err
		}
	}

	return nil
}

func (h *activityHandle) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Activity{}).Error
}

func (h *activityHandle) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.Activity{}).Error
}
