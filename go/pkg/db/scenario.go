package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type scenarioHandle struct {
	db *gorm.DB
}

func (h *scenarioHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("Zone").
		Preload("Options").
		Preload("Options.ItemRewards").
		Preload("Options.ItemRewards.InventoryItem").
		Preload("ItemRewards").
		Preload("ItemRewards.InventoryItem")
}

func (h *scenarioHandle) Create(ctx context.Context, scenario *models.Scenario) error {
	scenario.ID = uuid.New()
	scenario.CreatedAt = time.Now()
	scenario.UpdatedAt = time.Now()
	if err := scenario.SetGeometry(scenario.Latitude, scenario.Longitude); err != nil {
		return err
	}
	return h.db.WithContext(ctx).Create(scenario).Error
}

func (h *scenarioHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Scenario, error) {
	var scenario models.Scenario
	if err := h.preloadBase(ctx).First(&scenario, id).Error; err != nil {
		return nil, err
	}
	return &scenario, nil
}

func (h *scenarioHandle) FindAll(ctx context.Context) ([]models.Scenario, error) {
	var scenarios []models.Scenario
	if err := h.preloadBase(ctx).Find(&scenarios).Error; err != nil {
		return nil, err
	}
	return scenarios, nil
}

func (h *scenarioHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]models.Scenario, error) {
	var scenarios []models.Scenario
	if err := h.preloadBase(ctx).
		Where("zone_id = ?", zoneID).
		Find(&scenarios).Error; err != nil {
		return nil, err
	}
	return scenarios, nil
}

func (h *scenarioHandle) Update(ctx context.Context, id uuid.UUID, updates *models.Scenario) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	if err := updates.SetGeometry(updates.Latitude, updates.Longitude); err != nil {
		return err
	}

	payload := map[string]interface{}{
		"zone_id":           updates.ZoneID,
		"latitude":          updates.Latitude,
		"longitude":         updates.Longitude,
		"geometry":          updates.Geometry,
		"prompt":            updates.Prompt,
		"image_url":         updates.ImageURL,
		"difficulty":        updates.Difficulty,
		"reward_experience": updates.RewardExperience,
		"reward_gold":       updates.RewardGold,
		"open_ended":        updates.OpenEnded,
		"updated_at":        updates.UpdatedAt,
		"thumbnail_url":     updates.ThumbnailURL,
	}

	return h.db.WithContext(ctx).Model(&models.Scenario{}).Where("id = ?", id).Updates(payload).Error
}

func (h *scenarioHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Scenario{}, "id = ?", id).Error
}

func (h *scenarioHandle) ReplaceOptions(ctx context.Context, scenarioID uuid.UUID, options []models.ScenarioOption) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		optionIDs := tx.Model(&models.ScenarioOption{}).Select("id").Where("scenario_id = ?", scenarioID)
		if err := tx.Where("scenario_option_id IN (?)", optionIDs).Delete(&models.ScenarioOptionItemReward{}).Error; err != nil {
			return err
		}
		if err := tx.Where("scenario_id = ?", scenarioID).Delete(&models.ScenarioOption{}).Error; err != nil {
			return err
		}

		now := time.Now()
		for _, option := range options {
			option.ID = uuid.New()
			option.ScenarioID = scenarioID
			option.CreatedAt = now
			option.UpdatedAt = now
			if option.Proficiencies == nil {
				option.Proficiencies = models.StringArray{}
			}
			if err := tx.Create(&option).Error; err != nil {
				return err
			}

			for _, reward := range option.ItemRewards {
				reward.ID = uuid.New()
				reward.ScenarioOptionID = option.ID
				reward.CreatedAt = now
				reward.UpdatedAt = now
				if err := tx.Create(&reward).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (h *scenarioHandle) ReplaceItemRewards(ctx context.Context, scenarioID uuid.UUID, rewards []models.ScenarioItemReward) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("scenario_id = ?", scenarioID).Delete(&models.ScenarioItemReward{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, reward := range rewards {
			reward.ID = uuid.New()
			reward.ScenarioID = scenarioID
			reward.CreatedAt = now
			reward.UpdatedAt = now
			if err := tx.Create(&reward).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *scenarioHandle) FindAttemptByUserAndScenario(ctx context.Context, userID uuid.UUID, scenarioID uuid.UUID) (*models.UserScenarioAttempt, error) {
	var attempt models.UserScenarioAttempt
	if err := h.db.WithContext(ctx).
		Where("user_id = ? AND scenario_id = ?", userID, scenarioID).
		First(&attempt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &attempt, nil
}

func (h *scenarioHandle) CreateAttempt(ctx context.Context, attempt *models.UserScenarioAttempt) error {
	attempt.ID = uuid.New()
	attempt.CreatedAt = time.Now()
	attempt.UpdatedAt = time.Now()
	if attempt.ProficienciesUsed == nil {
		attempt.ProficienciesUsed = models.StringArray{}
	}
	return h.db.WithContext(ctx).Create(attempt).Error
}

func (h *scenarioHandle) FindAllWithUserStatus(ctx context.Context, userID *uuid.UUID) ([]models.Scenario, map[uuid.UUID]bool, error) {
	scenarios, err := h.FindAll(ctx)
	if err != nil {
		return nil, nil, err
	}
	attempted := map[uuid.UUID]bool{}
	if userID == nil {
		return scenarios, attempted, nil
	}

	var attempts []models.UserScenarioAttempt
	if err := h.db.WithContext(ctx).
		Select("scenario_id").
		Where("user_id = ?", *userID).
		Find(&attempts).Error; err != nil {
		return nil, nil, err
	}
	for _, attempt := range attempts {
		attempted[attempt.ScenarioID] = true
	}
	return scenarios, attempted, nil
}

func (h *scenarioHandle) FindByZoneIDWithUserStatus(ctx context.Context, zoneID uuid.UUID, userID *uuid.UUID) ([]models.Scenario, map[uuid.UUID]bool, error) {
	scenarios, err := h.FindByZoneID(ctx, zoneID)
	if err != nil {
		return nil, nil, err
	}
	attempted := map[uuid.UUID]bool{}
	if userID == nil {
		return scenarios, attempted, nil
	}

	var attempts []models.UserScenarioAttempt
	if err := h.db.WithContext(ctx).
		Select("scenario_id").
		Where("user_id = ? AND scenario_id IN (?)", *userID, h.db.Model(&models.Scenario{}).Select("id").Where("zone_id = ?", zoneID)).
		Find(&attempts).Error; err != nil {
		return nil, nil, err
	}
	for _, attempt := range attempts {
		attempted[attempt.ScenarioID] = true
	}
	return scenarios, attempted, nil
}
