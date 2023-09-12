package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type userHandle struct {
	db *gorm.DB
}

func (h *userHandle) Insert(ctx context.Context, name string, phoneNumber string) (*models.User, error) {
	user := models.User{
		Name:        name,
		PhoneNumber: phoneNumber,
	}

	if err := h.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *userHandle) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User

	if err := h.db.WithContext(ctx).Preload("Credentials").First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *userHandle) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error) {
	var user models.User

	if err := h.db.WithContext(ctx).Preload("Credentials").Where(&models.User{PhoneNumber: phoneNumber}).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *userHandle) FindUsersByIDs(ctx context.Context, userIDs []uint) ([]models.User, error) {
	var users []models.User

	if err := h.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (h *userHandle) FindAll(ctx context.Context) ([]models.User, error) {
	var users []models.User

	if err := h.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (h *userHandle) Delete(ctx context.Context, userID uint) error {
	return h.db.WithContext(ctx).Delete(&models.User{}, userID).Error
}

func (h *userHandle) DeleteAll(ctx context.Context) error {
	return h.db.WithContext(ctx).Where("1 = 1").Delete(&models.User{}).Error
}
