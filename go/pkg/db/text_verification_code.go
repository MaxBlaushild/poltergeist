package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type textVerificationCodeHandle struct {
	db *gorm.DB
}

func (c *textVerificationCodeHandle) Insert(ctx context.Context, phoneNumber string) (*models.TextVerificationCode, error) {
	if err := c.db.WithContext(ctx).Where(&models.TextVerificationCode{
		PhoneNumber: phoneNumber,
	}).Updates(&models.TextVerificationCode{
		Used: true,
	}).Error; err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(1000000)
	code := fmt.Sprintf("%06d", num)

	textVerificationCode := models.TextVerificationCode{
		Code:        code,
		PhoneNumber: phoneNumber,
	}

	if err := c.db.WithContext(ctx).Create(&textVerificationCode).Error; err != nil {
		return nil, err
	}

	return &textVerificationCode, nil
}

func (c *textVerificationCodeHandle) Find(ctx context.Context, phoneNumber string, code string) (*models.TextVerificationCode, error) {
	var textVerificationCode models.TextVerificationCode

	if err := c.db.WithContext(ctx).Model(&models.TextVerificationCode{}).Where(&models.TextVerificationCode{
		PhoneNumber: phoneNumber,
		Code:        code,
		Used:        false,
	}).First(&textVerificationCode).Error; err != nil {
		return nil, err
	}

	return &textVerificationCode, nil
}

func (c *textVerificationCodeHandle) MarkUsed(ctx context.Context, id uuid.UUID) error {
	return c.db.WithContext(ctx).Model(&models.TextVerificationCode{}).Where("id = ?", id).Update("used", true).Error
}
