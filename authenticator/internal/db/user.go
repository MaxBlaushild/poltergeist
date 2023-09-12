package db

import (
	"context"

	"github.com/MaxBlaushild/authenticator/internal/models"
	"gorm.io/gorm"
)

type userHandle struct {
	db *gorm.DB
}

func (h *userHandle) Insert(ctx context.Context, name string, phoneNumber string) (*models.AuthUser, error) {
	user := models.AuthUser{
		Name:        name,
		PhoneNumber: phoneNumber,
	}

	if err := h.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *userHandle) FindByID(ctx context.Context, id uint) (*models.AuthUser, error) {
	var user models.AuthUser

	if err := h.db.WithContext(ctx).Preload("Credentials").First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *userHandle) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*models.AuthUser, error) {
	var user models.AuthUser

	if err := h.db.WithContext(ctx).Preload("Credentials").Where(&models.AuthUser{PhoneNumber: phoneNumber}).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *userHandle) FindUsersByIDs(ctx context.Context, userIDs []uint) ([]models.AuthUser, error) {
	var users []models.AuthUser

	if err := h.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (h *userHandle) FindAll(ctx context.Context) ([]models.AuthUser, error) {
	var users []models.AuthUser

	if err := h.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (h *userHandle) Delete(ctx context.Context, userID uint) error {
	return h.db.WithContext(ctx).Delete(&models.AuthUser{}, userID).Error
}

func (h *userHandle) DeleteAll(ctx context.Context) error {
	return h.db.WithContext(ctx).Where("1 = 1").Delete(&models.AuthUser{}).Error
}
