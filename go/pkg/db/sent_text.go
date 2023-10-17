package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type sentTextHandle struct {
	db *gorm.DB
}

func (h *sentTextHandle) Insert(ctx context.Context, textType string, phoneNumber string, text string) (*models.SentText, error) {
	t := models.SentText{
		TextType:    textType,
		PhoneNumber: phoneNumber,
		Text:        text,
	}

	if err := h.db.WithContext(ctx).Model(&models.SentText{}).Create(&t).Error; err != nil {
		return nil, err
	}

	return &t, nil
}

func (h *sentTextHandle) GetCount(ctx context.Context, phoneNumber string, textType string) (int64, error) {
	var count int64

	if err := h.db.WithContext(ctx).Model(&models.SentText{}).Where(&models.SentText{PhoneNumber: phoneNumber, TextType: textType}).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
