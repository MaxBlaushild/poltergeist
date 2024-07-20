package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sonarActivityHandle struct {
	db *gorm.DB
}

func (h *sonarActivityHandle) GetAllActivities(ctx context.Context) ([]models.SonarActivity, error) {
	var activities []models.SonarActivity
	result := h.db.WithContext(ctx).Find(&activities)
	return activities, result.Error
}

func (h *sonarActivityHandle) CreateActivity(ctx context.Context, activity models.SonarActivity) (models.SonarActivity, error) {
	if err := h.db.WithContext(ctx).Create(&activity).Error; err != nil {
		return models.SonarActivity{}, err
	}
	return activity, nil
}

func (h *sonarActivityHandle) GetActivityByID(ctx context.Context, id uuid.UUID) (models.SonarActivity, error) {
	var activity models.SonarActivity
	result := h.db.WithContext(ctx).Where("id = ?", id).First(&activity)
	return activity, result.Error
}

func (h *sonarActivityHandle) UpdateActivity(ctx context.Context, activity models.SonarActivity) (models.SonarActivity, error) {
	if result := h.db.WithContext(ctx).Model(&activity).Updates(activity); result.Error != nil {
		return models.SonarActivity{}, result.Error
	}
	return activity, nil
}

func (h *sonarActivityHandle) DeleteActivity(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&models.SonarActivity{}).Error
}
