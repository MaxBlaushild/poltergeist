package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
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
