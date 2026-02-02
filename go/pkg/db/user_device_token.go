package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userDeviceTokenHandle struct {
	db *gorm.DB
}

func (h *userDeviceTokenHandle) Upsert(ctx context.Context, userID uuid.UUID, token, platform string) error {
	now := time.Now()
	var existing models.UserDeviceToken
	err := h.db.WithContext(ctx).Where("user_id = ? AND token = ?", userID, token).First(&existing).Error
	if err == nil { // found existing
		return h.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"platform":   platform,
			"updated_at": now,
		}).Error
	}
	return h.db.WithContext(ctx).Create(&models.UserDeviceToken{
		UserID:    userID,
		Token:     token,
		Platform:  platform,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error
}

func (h *userDeviceTokenHandle) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserDeviceToken, error) {
	var tokens []models.UserDeviceToken
	err := h.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tokens).Error
	return tokens, err
}
