package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type tutorialHandle struct {
	db *gorm.DB
}

func (h *tutorialHandle) GetConfig(ctx context.Context) (*models.TutorialConfig, error) {
	return h.getOrCreateConfig(ctx, h.db.WithContext(ctx))
}

func (h *tutorialHandle) UpsertConfig(ctx context.Context, config *models.TutorialConfig) (*models.TutorialConfig, error) {
	if config == nil {
		return nil, gorm.ErrInvalidData
	}
	config.ID = 1
	if err := h.db.WithContext(ctx).Save(config).Error; err != nil {
		return nil, err
	}
	return h.GetConfig(ctx)
}

func (h *tutorialHandle) InitializeForNewUser(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		config, err := h.getOrCreateConfig(ctx, tx)
		if err != nil {
			return err
		}
		if !config.IsConfigured() {
			return nil
		}

		now := time.Now()
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.UserTutorialState{
			UserID:    userID,
			CreatedAt: now,
			UpdatedAt: now,
		}).Error
	})
}

func (h *tutorialHandle) FindStateByUserID(ctx context.Context, userID uuid.UUID) (*models.UserTutorialState, error) {
	var state models.UserTutorialState
	if err := h.db.WithContext(ctx).
		Preload("TutorialScenario").
		Where("user_id = ?", userID).
		First(&state).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &state, nil
}

func (h *tutorialHandle) ActivateForUser(
	ctx context.Context,
	userID uuid.UUID,
	scenario *models.Scenario,
	options []models.ScenarioOption,
	itemRewards []models.ScenarioItemReward,
	spellRewards []models.ScenarioSpellReward,
	force bool,
) (*models.UserTutorialState, *models.Scenario, error) {
	var activatedState *models.UserTutorialState
	var activatedScenario *models.Scenario

	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var state models.UserTutorialState
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&state).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				now := time.Now()
				state = models.UserTutorialState{
					UserID:    userID,
					CreatedAt: now,
					UpdatedAt: now,
				}
				if err := tx.Create(&state).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		if state.CompletedAt != nil && !force {
			activatedState = &state
			if state.TutorialScenarioID != nil {
				existing, err := (&scenarioHandle{db: tx}).FindByID(ctx, *state.TutorialScenarioID)
				if err == nil {
					activatedScenario = existing
				}
			}
			return nil
		}

		if state.TutorialScenarioID != nil && !force {
			existing, err := (&scenarioHandle{db: tx}).FindByID(ctx, *state.TutorialScenarioID)
			if err == nil && existing != nil {
				activatedState = &state
				activatedScenario = existing
				return nil
			}
		}

		now := time.Now()
		scenario.ID = uuid.New()
		scenario.CreatedAt = now
		scenario.UpdatedAt = now
		normalizeScenarioFailurePenaltyDefaults(scenario)
		if err := scenario.SetGeometry(scenario.Latitude, scenario.Longitude); err != nil {
			return err
		}
		if err := tx.Omit(clause.Associations).Create(scenario).Error; err != nil {
			return err
		}

		for _, option := range options {
			normalizeScenarioOptionFailurePenaltyDefaults(&option)
			option.ID = uuid.New()
			option.ScenarioID = scenario.ID
			option.CreatedAt = now
			option.UpdatedAt = now
			if option.Proficiencies == nil {
				option.Proficiencies = models.StringArray{}
			}
			if err := tx.Omit(clause.Associations).Create(&option).Error; err != nil {
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
			for _, reward := range option.SpellRewards {
				reward.ID = uuid.New()
				reward.ScenarioOptionID = option.ID
				reward.CreatedAt = now
				reward.UpdatedAt = now
				if err := tx.Create(&reward).Error; err != nil {
					return err
				}
			}
		}

		for _, reward := range itemRewards {
			reward.ID = uuid.New()
			reward.ScenarioID = scenario.ID
			reward.CreatedAt = now
			reward.UpdatedAt = now
			if err := tx.Create(&reward).Error; err != nil {
				return err
			}
		}
		for _, reward := range spellRewards {
			reward.ID = uuid.New()
			reward.ScenarioID = scenario.ID
			reward.CreatedAt = now
			reward.UpdatedAt = now
			if err := tx.Create(&reward).Error; err != nil {
				return err
			}
		}

		state.TutorialScenarioID = &scenario.ID
		state.ActivatedAt = &now
		state.CompletedAt = nil
		state.UpdatedAt = now
		if err := tx.Model(&models.UserTutorialState{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"tutorial_scenario_id": scenario.ID,
				"activated_at":         now,
				"completed_at":         nil,
				"updated_at":           now,
			}).Error; err != nil {
			return err
		}

		stateCopy := state
		activatedState = &stateCopy
		loadedScenario, err := (&scenarioHandle{db: tx}).FindByID(ctx, scenario.ID)
		if err != nil {
			return err
		}
		activatedScenario = loadedScenario
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return activatedState, activatedScenario, nil
}

func (h *tutorialHandle) MarkCompleted(ctx context.Context, userID uuid.UUID, scenarioID uuid.UUID) error {
	now := time.Now()
	return h.db.WithContext(ctx).Model(&models.UserTutorialState{}).
		Where("user_id = ? AND tutorial_scenario_id = ?", userID, scenarioID).
		Updates(map[string]interface{}{
			"completed_at": now,
			"updated_at":   now,
		}).Error
}

func (h *tutorialHandle) getOrCreateConfig(ctx context.Context, db *gorm.DB) (*models.TutorialConfig, error) {
	var config models.TutorialConfig
	err := db.WithContext(ctx).Preload("Character").First(&config, 1).Error
	if err == nil {
		return &config, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	created := models.TutorialConfig{
		ID:                    1,
		Dialogue:              []string{},
		ScenarioPrompt:        "You hear a commotion outside of your door.",
		ScenarioImageURL:      "",
		ImageGenerationStatus: models.TutorialImageGenerationStatusNone,
		Options:               []models.TutorialScenarioOption{},
		ItemRewards:           []models.TutorialItemReward{},
		SpellRewards:          []models.TutorialSpellReward{},
	}
	if err := db.WithContext(ctx).Create(&created).Error; err != nil {
		return nil, err
	}
	return h.getOrCreateConfig(ctx, db)
}
