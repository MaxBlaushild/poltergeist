package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type userHandler struct {
	db *gorm.DB
}

func (c *userHandler) Insert(ctx context.Context, userID string, phoneNumber string) (*models.User, error) {
	user := models.User{
		UserID:             userID,
		PhoneNumber:        phoneNumber,
		HowManyQuestionSub: true,
	}

	if err := c.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *userHandler) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error) {
	user := models.User{}

	if err := c.db.WithContext(ctx).Where(&models.User{
		PhoneNumber: phoneNumber,
	}).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *userHandler) FindByID(ctx context.Context, id uint) (*models.User, error) {
	user := models.User{}

	if err := c.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *userHandler) FindByUserID(ctx context.Context, userID string) (*models.User, error) {
	user := models.User{}

	if err := c.db.WithContext(ctx).Where(&models.User{
		UserID: userID,
	}).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *userHandler) Verify(ctx context.Context, id uint) error {
	return c.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("verified", true).Error
}

func (c *userHandler) FindAll(ctx context.Context) ([]models.User, error) {
	users := []models.User{}
	if err := c.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
