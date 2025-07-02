package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userHandle struct {
	db *gorm.DB
}

func (h *userHandle) Insert(ctx context.Context, name string, phoneNumber string, id *uuid.UUID) (*models.User, error) {
	user := models.User{
		Name:        name,
		PhoneNumber: phoneNumber,
	}

	if id != nil {
		user.ID = *id
	}

	if err := h.db.WithContext(ctx).Model(&models.User{}).Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *userHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := h.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *userHandle) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error) {
	var user models.User
	if err := h.db.WithContext(ctx).Where(&models.User{PhoneNumber: phoneNumber}).First(&user).Error; err != nil {
		return nil, err
	}

	if uuid.Nil == user.ID {
		return nil, gorm.ErrRecordNotFound
	}

	return &user, nil
}

func (h *userHandle) FindUsersByIDs(ctx context.Context, userIDs []uuid.UUID) ([]models.User, error) {
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

func (h *userHandle) Delete(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.User{}, userID).Error
}

func (h *userHandle) DeleteAll(ctx context.Context) error {
	return h.db.WithContext(ctx).Where("1 = 1").Delete(&models.User{}).Error
}

func (h *userHandle) UpdateProfilePictureUrl(ctx context.Context, userID uuid.UUID, url string) error {
	return h.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("profile_picture_url", url).Error
}

func (h *userHandle) UpdateHasSeenTutorial(ctx context.Context, userID uuid.UUID, hasSeenTutorial bool) error {
	return h.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("has_seen_tutorial", hasSeenTutorial).Error
}

func (h *userHandle) UpdateDndClass(ctx context.Context, userID uuid.UUID, dndClassID uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("dnd_class_id", dndClassID).Error
}

func (h *userHandle) FindByIDWithDndClass(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := h.db.WithContext(ctx).Preload("DndClass").First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
