package db

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type verificationCodeHandler struct {
	db *gorm.DB
}

func generateRandomCode(length int) string {
	charset := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

func (h *verificationCodeHandler) Create(ctx context.Context) (*models.VerificationCode, error) {
	var verificationCode models.VerificationCode
	for {
		// Generate a 6-digit alphanumeric code
		code := generateRandomCode(6)

		// Create a new verification code instance
		verificationCode = models.VerificationCode{
			ID:        uuid.New(),
			Code:      code,
			CreatedAt: time.Now(),
		}

		// Check if a similar code exists and is less than an hour old
		var existingCode models.VerificationCode
		if err := h.db.WithContext(ctx).Where("code = ? AND created_at > ?", code, time.Now().Add(-time.Hour)).First(&existingCode).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
			// Save the new code if no recent duplicate exists
			if err := h.db.WithContext(ctx).Create(&verificationCode).Error; err != nil {
				return nil, err
			}
			break
		}
	}

	return &verificationCode, nil
}
