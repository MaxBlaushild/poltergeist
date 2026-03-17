package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type scenarioTemplateHandle struct {
	db *gorm.DB
}

func (h *scenarioTemplateHandle) Create(ctx context.Context, template *models.ScenarioTemplate) error {
	if template == nil {
		return nil
	}
	template.ID = uuid.New()
	template.CreatedAt = time.Now()
	template.UpdatedAt = template.CreatedAt
	return h.db.WithContext(ctx).Create(template).Error
}

func (h *scenarioTemplateHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ScenarioTemplate, error) {
	var template models.ScenarioTemplate
	if err := h.db.WithContext(ctx).First(&template, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

func (h *scenarioTemplateHandle) FindAll(ctx context.Context) ([]models.ScenarioTemplate, error) {
	var templates []models.ScenarioTemplate
	if err := h.db.WithContext(ctx).Order("created_at DESC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (h *scenarioTemplateHandle) FindRecent(ctx context.Context, limit int) ([]models.ScenarioTemplate, error) {
	var templates []models.ScenarioTemplate
	q := h.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (h *scenarioTemplateHandle) Update(ctx context.Context, id uuid.UUID, updates *models.ScenarioTemplate) error {
	if updates == nil {
		return nil
	}
	updates.UpdatedAt = time.Now()
	payload := map[string]interface{}{
		"prompt":                       updates.Prompt,
		"image_url":                    updates.ImageURL,
		"thumbnail_url":                updates.ThumbnailURL,
		"scale_with_user_level":        updates.ScaleWithUserLevel,
		"reward_mode":                  updates.RewardMode,
		"random_reward_size":           updates.RandomRewardSize,
		"difficulty":                   updates.Difficulty,
		"reward_experience":            updates.RewardExperience,
		"reward_gold":                  updates.RewardGold,
		"open_ended":                   updates.OpenEnded,
		"failure_penalty_mode":         updates.FailurePenaltyMode,
		"failure_health_drain_type":    updates.FailureHealthDrainType,
		"failure_health_drain_value":   updates.FailureHealthDrainValue,
		"failure_mana_drain_type":      updates.FailureManaDrainType,
		"failure_mana_drain_value":     updates.FailureManaDrainValue,
		"failure_statuses":             updates.FailureStatuses,
		"success_reward_mode":          updates.SuccessRewardMode,
		"success_health_restore_type":  updates.SuccessHealthRestoreType,
		"success_health_restore_value": updates.SuccessHealthRestoreValue,
		"success_mana_restore_type":    updates.SuccessManaRestoreType,
		"success_mana_restore_value":   updates.SuccessManaRestoreValue,
		"success_statuses":             updates.SuccessStatuses,
		"options":                      updates.Options,
		"item_rewards":                 updates.ItemRewards,
		"item_choice_rewards":          updates.ItemChoiceRewards,
		"spell_rewards":                updates.SpellRewards,
		"updated_at":                   updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).Model(&models.ScenarioTemplate{}).Where("id = ?", id).Updates(payload).Error
}

func (h *scenarioTemplateHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.ScenarioTemplate{}, "id = ?", id).Error
}
