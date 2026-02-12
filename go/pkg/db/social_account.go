package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type socialAccountHandler struct {
	db *gorm.DB
}

func (h *socialAccountHandler) Upsert(ctx context.Context, account *models.SocialAccount) error {
	if account.CreatedAt.IsZero() {
		account.CreatedAt = time.Now()
	}
	account.UpdatedAt = time.Now()

	var existing models.SocialAccount
	err := h.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", account.UserID, account.Provider).
		First(&existing).Error
	if err == nil {
		existing.AccountID = account.AccountID
		existing.Username = account.Username
		existing.AccessToken = account.AccessToken
		existing.RefreshToken = account.RefreshToken
		existing.ExpiresAt = account.ExpiresAt
		existing.Scopes = account.Scopes
		existing.UpdatedAt = time.Now()
		return h.db.WithContext(ctx).Save(&existing).Error
	}

	if account.ID == uuid.Nil {
		account.ID = uuid.New()
	}
	return h.db.WithContext(ctx).Create(account).Error
}

func (h *socialAccountHandler) FindByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) (*models.SocialAccount, error) {
	var account models.SocialAccount
	if err := h.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (h *socialAccountHandler) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.SocialAccount, error) {
	var accounts []models.SocialAccount
	if err := h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("provider ASC").
		Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (h *socialAccountHandler) DeleteByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) error {
	return h.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		Delete(&models.SocialAccount{}).Error
}
