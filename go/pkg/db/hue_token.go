package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type hueTokenHandler struct {
	db *gorm.DB
}

func (h *hueTokenHandler) Create(ctx context.Context, token *models.HueToken) error {
	// Set timestamps if not set
	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}
	token.UpdatedAt = time.Now()

	// Set ID if not set
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}

	// If UserID is provided, check if token exists for user
	if token.UserID != nil {
		existingToken, err := h.FindByUserID(ctx, *token.UserID)
		if err == nil && existingToken != nil {
			// Update existing token
			existingToken.AccessToken = token.AccessToken
			existingToken.RefreshToken = token.RefreshToken
			existingToken.ExpiresAt = token.ExpiresAt
			existingToken.TokenType = token.TokenType
			existingToken.UpdatedAt = time.Now()

			return h.db.WithContext(ctx).Save(existingToken).Error
		}
	}

	return h.db.WithContext(ctx).Create(token).Error
}

func (h *hueTokenHandler) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.HueToken, error) {
	var token models.HueToken
	if err := h.db.WithContext(ctx).Where("user_id = ?", userID).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (h *hueTokenHandler) FindLatest(ctx context.Context) (*models.HueToken, error) {
	var token models.HueToken
	if err := h.db.WithContext(ctx).Order("created_at DESC").First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (h *hueTokenHandler) Update(ctx context.Context, token *models.HueToken) error {
	token.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Save(token).Error
}

func (h *hueTokenHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.HueToken{}, id).Error
}

func (h *hueTokenHandler) RefreshToken(ctx context.Context, id uuid.UUID, newAccessToken string, expiresAt time.Time) error {
	var token models.HueToken
	if err := h.db.WithContext(ctx).First(&token, "id = ?", id).Error; err != nil {
		return err
	}

	token.AccessToken = newAccessToken
	token.ExpiresAt = expiresAt
	token.UpdatedAt = time.Now()

	return h.db.WithContext(ctx).Save(&token).Error
}

