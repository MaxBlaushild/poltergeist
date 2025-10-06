package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type friendHandle struct {
	db *gorm.DB
}

func (h *friendHandle) Create(ctx context.Context, firstUserID uuid.UUID, secondUserID uuid.UUID) (*models.Friend, error) {
	friend := &models.Friend{
		FirstUserID:  firstUserID,
		SecondUserID: secondUserID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.db.WithContext(ctx).Create(friend).Error; err != nil {
		return nil, err
	}

	return friend, nil
}

func (h *friendHandle) FindAllFriends(ctx context.Context, userID uuid.UUID) ([]models.User, error) {
	var friends []models.Friend
	if err := h.db.WithContext(ctx).Find(&friends).Error; err != nil {
		return nil, err
	}

	userIDs := []uuid.UUID{}
	for _, friend := range friends {
		if friend.FirstUserID == userID {
			userIDs = append(userIDs, friend.SecondUserID)
		} else if friend.SecondUserID == userID {
			userIDs = append(userIDs, friend.FirstUserID)
		}
	}

	users := []models.User{}
	if err := h.db.WithContext(ctx).Where("id IN (?)", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}
