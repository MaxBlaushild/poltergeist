package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sonarSurveySubmissionHandle struct {
	db *gorm.DB
}

func (h *sonarSurveySubmissionHandle) CreateSubmission(ctx context.Context, surveyID uuid.UUID, userID uuid.UUID, activityIDs []uuid.UUID, downs []bool) (*models.SonarSurveySubmission, error) {
	now := time.Now()

	submission := models.SonarSurveySubmission{
		SonarSurveyID: surveyID,
		UserID:        userID,
		CreatedAt:     now,
		UpdatedAt:     now,
		ID:            uuid.New(),
	}

	if err := h.db.WithContext(ctx).Create(&submission).Error; err != nil {
		return nil, err
	}

	for i, activityID := range activityIDs {
		submissionActivity := models.SonarSurveySubmissionAnswer{
			SonarSurveySubmissionID: submission.ID,
			SonarActivityID:         activityID,
			Down:                    downs[i],
			SonarSurveyID:           surveyID,
			CreatedAt:               now,
			UpdatedAt:               now,
			ID:                      uuid.New(),
		}

		if err := h.db.WithContext(ctx).Create(&submissionActivity).Error; err != nil {
			return nil, err
		}
	}

	return &submission, nil
}

func (h *sonarSurveySubmissionHandle) GetAllSubmissionsForUser(ctx context.Context, userID uuid.UUID) ([]models.SonarSurveySubmission, error) {
	var submissions []models.SonarSurveySubmission
	if err := h.db.WithContext(ctx).
		Preload("User").
		Preload("SonarSurveySubmissionAnswers.SonarActivity").
		Joins("JOIN sonar_surveys ON sonar_surveys.id = sonar_survey_submissions.sonar_survey_id").
		Where("sonar_surveys.user_id = ?", userID).
		Find(&submissions).Error; err != nil {
		return nil, err
	}
	return submissions, nil
}

func (h *sonarSurveySubmissionHandle) GetUserSubmissionForSurvey(ctx context.Context, userID uuid.UUID, surveyID uuid.UUID) (*models.SonarSurveySubmission, error) {
	var submission models.SonarSurveySubmission
	if err := h.db.WithContext(ctx).Preload("SonarSurveySubmissionAnswers").Where("user_id = ? AND sonar_survey_id = ?", userID, surveyID).Order("created_at DESC").Order("created_at DESC").First(&submission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No submission found is not an error
		}
		return nil, err
	}
	return &submission, nil
}

func (h *sonarSurveySubmissionHandle) GetSubmissionByID(ctx context.Context, submissionID uuid.UUID) (*models.SonarSurveySubmission, error) {
	var submission models.SonarSurveySubmission
	if err := h.db.WithContext(ctx).Preload("User").Preload("SonarSurveySubmissionAnswers").Where("id = ?", submissionID).First(&submission).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}
