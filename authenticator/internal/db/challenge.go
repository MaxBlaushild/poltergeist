package db

import (
	"context"

	"github.com/MaxBlaushild/authenticator/internal/models"
	"gorm.io/gorm"
)

type challengeHandle struct {
	db *gorm.DB
}

func (h *challengeHandle) Insert(ctx context.Context, challenge string, userID uint) error {
	return h.db.WithContext(ctx).Create(&models.Challenge{
		AuthUserID: userID,
		Challenge:  challenge,
	}).Error
}

func (h *challengeHandle) Find(ctx context.Context, challenge string) (*models.Challenge, error) {
	var cha models.Challenge

	if err := h.db.WithContext(ctx).Where(&models.Challenge{
		Challenge: challenge,
	}).First(&cha).Error; err != nil {
		return nil, err
	}

	return &cha, nil
}
