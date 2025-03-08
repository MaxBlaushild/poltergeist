package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type matchUserHandle struct {
	db *gorm.DB
}

func (h *matchUserHandle) Create(ctx context.Context, matchUser *models.MatchUser) error {
	matchUser.ID = uuid.New()
	matchUser.CreatedAt = time.Now()
	matchUser.UpdatedAt = time.Now()

	return h.db.WithContext(ctx).Create(matchUser).Error
}

func (h *matchUserHandle) FindByMatchID(ctx context.Context, matchID uuid.UUID) ([]models.MatchUser, error) {
	var matchUsers []models.MatchUser
	if err := h.db.WithContext(ctx).Where("match_id = ?", matchID).Find(&matchUsers).Error; err != nil {
		return nil, err
	}
	return matchUsers, nil
}

func (h *matchUserHandle) FindUsersForMatch(ctx context.Context, matchID uuid.UUID) ([]models.User, error) {
	var users []models.User
	if err := h.db.WithContext(ctx).
		Model(&models.MatchUser{}).
		Select("users.*").
		Joins("JOIN users ON users.id = match_users.user_id").
		Where("match_users.match_id = ?", matchID).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
