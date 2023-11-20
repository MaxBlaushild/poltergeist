package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type howManySubscriptionHandle struct {
	db *gorm.DB
}

func (h *howManySubscriptionHandle) Insert(ctx context.Context, userID uuid.UUID) (*models.HowManySubscription, error) {
	subscription := models.HowManySubscription{
		UserID: userID,
	}

	if err := h.db.WithContext(ctx).Model(&models.HowManySubscription{}).Create(&subscription).Error; err != nil {
		return nil, err
	}

	return &subscription, nil
}

func (h *howManySubscriptionHandle) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.HowManySubscription, error) {
	var guessHowManySubscription models.HowManySubscription

	if err := h.db.WithContext(ctx).Model(&models.HowManySubscription{}).Where(&models.HowManySubscription{
		UserID: userID,
	}).First(&guessHowManySubscription).Error; err != nil {
		return nil, err
	}

	return &guessHowManySubscription, nil
}

func (h *howManySubscriptionHandle) IncrementNumFreeQuestions(ctx context.Context, userID uuid.UUID) error {
	guessHowManySubscription, err := h.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	return h.db.WithContext(ctx).Where("user_id = ?", userID).Updates(&models.HowManySubscription{
		NumFreeQuestions: guessHowManySubscription.NumFreeQuestions + 1,
	}).Error
}

func (h *howManySubscriptionHandle) FindAll(ctx context.Context) ([]models.HowManySubscription, error) {
	var subscriptions []models.HowManySubscription

	if err := h.db.WithContext(ctx).Preload("User").Find(&subscriptions).Error; err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (h *howManySubscriptionHandle) SetSubscribed(ctx context.Context, userID uuid.UUID, subscribed bool) error {
	return h.db.WithContext(ctx).Where("user_id = ?", userID).Updates(&models.HowManySubscription{
		Subscribed: subscribed,
	}).Error
}
