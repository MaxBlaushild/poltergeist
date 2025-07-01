package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userStatsHandler struct {
	db *gorm.DB
}

func (h *userStatsHandler) Create(ctx context.Context, userStats *models.UserStats) error {
	return h.db.Create(userStats).Error
}

func (h *userStatsHandler) FindOrCreateForUser(ctx context.Context, userID uuid.UUID) (*models.UserStats, error) {
	userStats, err := h.FindByUserID(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if userStats != nil {
		return userStats, nil
	}

	userStats = &models.UserStats{
		ID:                  uuid.New(),
		UserID:              userID,
		Strength:            models.DefaultStatValue,
		Dexterity:           models.DefaultStatValue,
		Constitution:        models.DefaultStatValue,
		Intelligence:        models.DefaultStatValue,
		Wisdom:              models.DefaultStatValue,
		Charisma:            models.DefaultStatValue,
		AvailableStatPoints: 0,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	return userStats, h.Create(ctx, userStats)
}

func (h *userStatsHandler) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserStats, error) {
	var userStats models.UserStats
	if err := h.db.Where("user_id = ?", userID).First(&userStats).Error; err != nil {
		return nil, err
	}
	return &userStats, nil
}

func (h *userStatsHandler) Update(ctx context.Context, userStats *models.UserStats) error {
	userStats.UpdatedAt = time.Now()
	return h.db.Save(userStats).Error
}

func (h *userStatsHandler) AllocateStatPoint(ctx context.Context, userID uuid.UUID, statName string) (*models.UserStats, error) {
	userStats, err := h.FindOrCreateForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := userStats.AllocateStatPoint(statName); err != nil {
		return nil, err
	}

	return userStats, h.Update(ctx, userStats)
}

func (h *userStatsHandler) AddStatPoints(ctx context.Context, userID uuid.UUID, points int) (*models.UserStats, error) {
	userStats, err := h.FindOrCreateForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	userStats.AddStatPoints(points)

	return userStats, h.Update(ctx, userStats)
}