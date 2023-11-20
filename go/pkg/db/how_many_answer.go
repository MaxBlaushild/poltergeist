package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type howManyAnswerHandle struct {
	db *gorm.DB
}

func (h *howManyAnswerHandle) FindByQuestionIDAndUserID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.HowManyAnswer, error) {
	howManyAnswer := models.HowManyAnswer{}

	if err := h.db.WithContext(ctx).Where(&models.HowManyAnswer{
		HowManyQuestionID: id,
		UserID:            userID,
	}).First(&howManyAnswer).Error; err != nil {
		return nil, err
	}

	return &howManyAnswer, nil
}

func (h *howManyAnswerHandle) Insert(ctx context.Context, a *models.HowManyAnswer) (*models.HowManyAnswer, error) {
	if err := h.db.WithContext(ctx).Create(&a).Error; err != nil {
		return nil, err
	}

	return a, nil
}
