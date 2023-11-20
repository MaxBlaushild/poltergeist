package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userSubmissionHandle struct {
	db *gorm.DB
}

func (u *userSubmissionHandle) Insert(
	ctx context.Context,
	questionSetID uuid.UUID,
	userID uuid.UUID,
	userAnswers []models.UserAnswer,
) (*models.UserSubmission, error) {

	submission := models.UserSubmission{
		QuestionSetID: questionSetID,
		UserAnswers:   userAnswers,
		UserID:        userID,
	}

	if err := u.db.WithContext(ctx).Create(&submission).Error; err != nil {
		return nil, err
	}

	return &submission, nil
}

func (m *userSubmissionHandle) FindByUserAndQuestionSetID(ctx context.Context, userID uuid.UUID, questionSetID uuid.UUID) (*models.UserSubmission, error) {
	submission := models.UserSubmission{
		QuestionSetID: questionSetID,
		UserID:        userID,
	}

	if err := m.db.Where(&submission).Preload("UserAnswers").First(&submission).Error; err != nil {
		return nil, err
	}

	return &submission, nil
}
