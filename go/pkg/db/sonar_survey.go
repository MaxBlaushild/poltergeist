package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sonarSurveyHandle struct {
	db *gorm.DB
}

func (h *sonarSurveyHandle) GetSurveys(ctx context.Context, userID uuid.UUID) ([]models.SonarSurvey, error) {
	var surveys []models.SonarSurvey

	if err := h.db.WithContext(ctx).Where("user_id = ?", userID.String()).Find(&surveys).Error; err != nil {
		return nil, err
	}

	return surveys, nil
}
