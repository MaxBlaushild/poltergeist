package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type sonarCategoryHandle struct {
	db *gorm.DB
}

func (h *sonarCategoryHandle) GetAllCategoriesWithActivities(ctx context.Context) ([]models.SonarCategory, error) {
	var categories []models.SonarCategory
	if err := h.db.WithContext(ctx).Preload("SonarActivities").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}
