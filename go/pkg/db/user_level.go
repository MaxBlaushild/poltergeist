package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userLevelHandler struct {
	db *gorm.DB
}

func (h *userLevelHandler) Create(ctx context.Context, userLevel *models.UserLevel) error {
	return h.db.Create(userLevel).Error
}

func (h *userLevelHandler) FindOrCreateForUser(ctx context.Context, userID uuid.UUID) (*models.UserLevel, error) {
	userLevel, err := h.FindByUserID(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if userLevel != nil {
		return userLevel, nil
	}

	userLevel = &models.UserLevel{
		ID:                      uuid.New(),
		UserID:                  userID,
		Level:                   1,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
		TotalExperiencePoints:   0,
		ExperiencePointsOnLevel: 0,
	}

	userLevel.ExperienceToNextLevel = userLevel.XPToNextLevel()

	return userLevel, h.Create(ctx, userLevel)
}

func (h *userLevelHandler) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserLevel, error) {
	var userLevel models.UserLevel
	if err := h.db.Where("user_id = ?", userID).First(&userLevel).Error; err != nil {
		return nil, err
	}
	return &userLevel, nil
}

func (h *userLevelHandler) ProcessExperiencePointAdditions(ctx context.Context, userID uuid.UUID, experiencePoints int) (*models.UserLevel, error) {
	userLevel, err := h.FindOrCreateForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	userLevel.AddExperiencePoints(experiencePoints)

	return userLevel, h.db.Save(userLevel).Error
}
