package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type guessHowManySubscriptionHandle struct {
	db *gorm.DB
}

func (h *guessHowManySubscriptionHandle) Insert(ctx context.Context, userID uint) (*models.GuessHowManySubscription, error) {
	subscription := models.GuessHowManySubscription{
		UserID: userID,
	}

	if err := h.db.WithContext(ctx).Model(&models.GuessHowManySubscription{}).Create(&subscription).Error; err != nil {
		return nil, err
	}

	return &subscription, nil
}

func (h *guessHowManySubscriptionHandle) FindByUserID(ctx context.Context, userID uint) (*models.GuessHowManySubscription, error) {
	var guessHowManySubscription models.GuessHowManySubscription

	if err := h.db.WithContext(ctx).Model(&models.GuessHowManySubscription{}).Where(&models.GuessHowManySubscription{
		UserID: userID,
	}).First(&guessHowManySubscription).Error; err != nil {
		return nil, err
	}

	return &guessHowManySubscription, nil
}

func (h *guessHowManySubscriptionHandle) IncrementNumFreeQuestions(ctx context.Context, userID uint) error {
	guessHowManySubscription, err := h.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	return h.db.WithContext(ctx).Where("user_id = ?", userID).Updates(&models.GuessHowManySubscription{
		NumFreeQuestions: guessHowManySubscription.NumFreeQuestions + 1,
	}).Error
}

func (h *guessHowManySubscriptionHandle) FindAll(ctx context.Context) ([]models.GuessHowManySubscription, error) {
	var subscriptions []models.GuessHowManySubscription

	if err := h.db.WithContext(ctx).Preload("User").Find(&subscriptions).Error; err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (h *guessHowManySubscriptionHandle) SetSubscribed(ctx context.Context, userID uint, subscribed bool) error {
	return h.db.WithContext(ctx).Where("user_id = ?", userID).Updates(&models.GuessHowManySubscription{
		Subscribed: subscribed,
	}).Error
}
