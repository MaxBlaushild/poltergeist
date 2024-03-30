package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sonarSurveyHandle struct {
	db *gorm.DB
}

func (h *sonarSurveyHandle) GetSurveys(ctx context.Context, userID uuid.UUID) ([]models.SonarSurvey, error) {
	var surveys []models.SonarSurvey

	if err := h.db.WithContext(ctx).Preload("SonarActivities").Where("user_id = ?", userID.String()).Find(&surveys).Error; err != nil {
		return nil, err
	}

	return surveys, nil
}

func (h *sonarSurveyHandle) GetSurveyByID(ctx context.Context, surveyID uuid.UUID) (*models.SonarSurvey, error) {
	var survey models.SonarSurvey

	if err := h.db.WithContext(ctx).Preload("SonarActivities").Where("id = ?", surveyID).First(&survey).Error; err != nil {
		return nil, err
	}

	return &survey, nil
}

func (h *sonarSurveyHandle) CreateSurvey(ctx context.Context, userID uuid.UUID, title string, activityIDs []uuid.UUID) (*models.SonarSurvey, error) {
	// transaction
	id := uuid.New()
	now := time.Now()

	survey := models.SonarSurvey{
		ID:        id,
		UserID:    userID,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}

	var surveyActivities []models.SonarSurveyActivity
	for _, activityID := range activityIDs {
		surveyActivities = append(surveyActivities, models.SonarSurveyActivity{
			ID:              uuid.New(),
			CreatedAt:       now,
			UpdatedAt:       now,
			SonarSurveyID:   id,
			SonarActivityID: activityID,
		})
	}

	if err := h.db.WithContext(ctx).Create(&survey).Error; err != nil {
		return nil, err
	}

	if err := h.db.WithContext(ctx).Create(&surveyActivities).Error; err != nil {
		return nil, err
	}

	return &survey, nil
}
