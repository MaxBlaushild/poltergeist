package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type dropboxTokenHandler struct {
	db *gorm.DB
}

func (h *dropboxTokenHandler) Create(ctx context.Context, token *models.DropboxToken) error {
	// Set timestamps if not set
	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}
	token.UpdatedAt = time.Now()

	// Check if token exists for user
	existingToken, err := h.FindByUserID(ctx, token.UserID)
	if err == nil && existingToken != nil {
		// Update existing token
		existingToken.AccessToken = token.AccessToken
		existingToken.RefreshToken = token.RefreshToken
		existingToken.ExpiresAt = token.ExpiresAt
		existingToken.TokenType = token.TokenType
		existingToken.UpdatedAt = time.Now()

		return h.db.WithContext(ctx).Save(existingToken).Error
	}

	// Set ID if not set
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}

	return h.db.WithContext(ctx).Create(token).Error
}

func (h *dropboxTokenHandler) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.DropboxToken, error) {
	var token models.DropboxToken
	if err := h.db.WithContext(ctx).Where("user_id = ?", userID).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (h *dropboxTokenHandler) Update(ctx context.Context, token *models.DropboxToken) error {
	token.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Save(token).Error
}

func (h *dropboxTokenHandler) Delete(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.DropboxToken{}).Error
}

func (h *dropboxTokenHandler) RefreshToken(ctx context.Context, userID uuid.UUID, newAccessToken string, expiresAt time.Time) error {
	token, err := h.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	token.AccessToken = newAccessToken
	token.ExpiresAt = expiresAt
	token.UpdatedAt = time.Now()

	return h.db.WithContext(ctx).Save(token).Error
}

