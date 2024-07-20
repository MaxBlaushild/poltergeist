package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sonarCategoryHandle struct {
	db *gorm.DB
}

func (h *sonarCategoryHandle) CreateCategory(ctx context.Context, category models.SonarCategory) (models.SonarCategory, error) {
	if err := h.db.WithContext(ctx).Create(&category).Error; err != nil {
		return models.SonarCategory{}, err
	}
	return category, nil
}

func (h *sonarCategoryHandle) GetAllCategoriesWithActivities(ctx context.Context) ([]models.SonarCategory, error) {
	var categories []models.SonarCategory
	if err := h.db.WithContext(ctx).Preload("SonarActivities").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (h *sonarCategoryHandle) GetCategoriesByUserID(ctx context.Context, userID uuid.UUID) ([]models.SonarCategory, error) {
	var categories []models.SonarCategory
	if err := h.db.WithContext(ctx).Where("user_id = ?", userID).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (h *sonarCategoryHandle) GetCategoryByID(ctx context.Context, id uuid.UUID) (models.SonarCategory, error) {
	var category models.SonarCategory
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&category).Error; err != nil {
		return models.SonarCategory{}, err
	}
	return category, nil
}

func (h *sonarCategoryHandle) UpdateCategory(ctx context.Context, category models.SonarCategory) (models.SonarCategory, error) {
	if result := h.db.WithContext(ctx).Model(&category).Updates(category); result.Error != nil {
		return models.SonarCategory{}, result.Error
	}
	return category, nil
}

func (h *sonarCategoryHandle) DeleteCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&models.SonarCategory{}).Error
}
