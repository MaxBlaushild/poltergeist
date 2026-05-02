package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type rewardProfileHandle struct {
	db *gorm.DB
}

func (h *rewardProfileHandle) Create(ctx context.Context, rewardProfile *models.RewardProfile) error {
	return h.db.WithContext(ctx).Create(rewardProfile).Error
}

func (h *rewardProfileHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.RewardProfile, error) {
	var rewardProfile models.RewardProfile
	if err := h.db.WithContext(ctx).First(&rewardProfile, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &rewardProfile, nil
}

func (h *rewardProfileHandle) FindBySlug(ctx context.Context, slug string) (*models.RewardProfile, error) {
	var rewardProfile models.RewardProfile
	if err := h.db.WithContext(ctx).
		First(&rewardProfile, "slug = ?", models.NormalizeRewardProfileSlug(slug)).
		Error; err != nil {
		return nil, err
	}
	return &rewardProfile, nil
}

func (h *rewardProfileHandle) FindAll(ctx context.Context, includeInactive bool) ([]models.RewardProfile, error) {
	var rewardProfiles []models.RewardProfile
	query := h.db.WithContext(ctx).Order("active DESC").Order("name ASC").Order("slug ASC")
	if !includeInactive {
		query = query.Where("active = ?", true)
	}
	if err := query.Find(&rewardProfiles).Error; err != nil {
		return nil, err
	}
	return rewardProfiles, nil
}

func (h *rewardProfileHandle) Update(ctx context.Context, rewardProfile *models.RewardProfile) error {
	return h.db.WithContext(ctx).Save(rewardProfile).Error
}

func (h *rewardProfileHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.RewardProfile{}, "id = ?", id).Error
}
