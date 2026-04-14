package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type expositionTemplateHandle struct {
	db *gorm.DB
}

func (h *expositionTemplateHandle) Create(ctx context.Context, template *models.ExpositionTemplate) error {
	if template == nil {
		return nil
	}
	template.ID = uuid.New()
	template.CreatedAt = time.Now()
	template.UpdatedAt = template.CreatedAt
	return h.db.WithContext(ctx).Create(template).Error
}

func (h *expositionTemplateHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ExpositionTemplate, error) {
	var template models.ExpositionTemplate
	if err := h.db.WithContext(ctx).First(&template, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

func (h *expositionTemplateHandle) FindAll(ctx context.Context) ([]models.ExpositionTemplate, error) {
	var templates []models.ExpositionTemplate
	if err := h.db.WithContext(ctx).Order("created_at DESC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (h *expositionTemplateHandle) Update(ctx context.Context, id uuid.UUID, updates *models.ExpositionTemplate) error {
	if updates == nil {
		return nil
	}
	updates.UpdatedAt = time.Now()
	payload := map[string]interface{}{
		"title":                 updates.Title,
		"description":           updates.Description,
		"dialogue":              updates.Dialogue,
		"required_story_flags":  updates.RequiredStoryFlags,
		"image_url":             updates.ImageURL,
		"thumbnail_url":         updates.ThumbnailURL,
		"reward_mode":           updates.RewardMode,
		"random_reward_size":    updates.RandomRewardSize,
		"reward_experience":     updates.RewardExperience,
		"reward_gold":           updates.RewardGold,
		"material_rewards_json": updates.MaterialRewards,
		"item_rewards_json":     updates.ItemRewards,
		"spell_rewards_json":    updates.SpellRewards,
		"updated_at":            updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).Model(&models.ExpositionTemplate{}).Where("id = ?", id).Updates(payload).Error
}

func (h *expositionTemplateHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.ExpositionTemplate{}, "id = ?", id).Error
}
