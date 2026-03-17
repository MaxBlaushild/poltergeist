package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type challengeTemplateHandle struct {
	db *gorm.DB
}

func (h *challengeTemplateHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).Preload("LocationArchetype")
}

func (h *challengeTemplateHandle) Create(ctx context.Context, template *models.ChallengeTemplate) error {
	if template == nil {
		return nil
	}
	template.ID = uuid.New()
	template.CreatedAt = time.Now()
	template.UpdatedAt = template.CreatedAt
	return h.db.WithContext(ctx).Create(template).Error
}

func (h *challengeTemplateHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ChallengeTemplate, error) {
	var template models.ChallengeTemplate
	if err := h.preloadBase(ctx).First(&template, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

func (h *challengeTemplateHandle) FindAll(ctx context.Context) ([]models.ChallengeTemplate, error) {
	var templates []models.ChallengeTemplate
	if err := h.preloadBase(ctx).Order("created_at DESC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (h *challengeTemplateHandle) FindRecentByLocationArchetypeID(ctx context.Context, locationArchetypeID uuid.UUID, limit int) ([]models.ChallengeTemplate, error) {
	var templates []models.ChallengeTemplate
	q := h.preloadBase(ctx).Where("location_archetype_id = ?", locationArchetypeID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (h *challengeTemplateHandle) Update(ctx context.Context, id uuid.UUID, updates *models.ChallengeTemplate) error {
	if updates == nil {
		return nil
	}
	updates.UpdatedAt = time.Now()
	payload := map[string]interface{}{
		"location_archetype_id": updates.LocationArchetypeID,
		"question":              updates.Question,
		"description":           updates.Description,
		"image_url":             updates.ImageURL,
		"thumbnail_url":         updates.ThumbnailURL,
		"scale_with_user_level": updates.ScaleWithUserLevel,
		"reward_mode":           updates.RewardMode,
		"random_reward_size":    updates.RandomRewardSize,
		"reward_experience":     updates.RewardExperience,
		"reward":                updates.Reward,
		"inventory_item_id":     updates.InventoryItemID,
		"item_choice_rewards":   updates.ItemChoiceRewards,
		"submission_type":       updates.SubmissionType,
		"difficulty":            updates.Difficulty,
		"stat_tags":             updates.StatTags,
		"proficiency":           updates.Proficiency,
		"updated_at":            updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).Model(&models.ChallengeTemplate{}).Where("id = ?", id).Updates(payload).Error
}

func (h *challengeTemplateHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.ChallengeTemplate{}, "id = ?", id).Error
}
