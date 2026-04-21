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
			Stage:     models.TutorialStageWelcome,
			CreatedAt: now,
			UpdatedAt: now,
		}).Error
	})
}

func (h *tutorialHandle) FindStateByUserID(ctx context.Context, userID uuid.UUID) (*models.UserTutorialState, error) {
	var state models.UserTutorialState
	if err := h.db.WithContext(ctx).
		Preload("TutorialScenario").
		Preload("TutorialMonsterEncounter").
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
		state, err := h.getOrCreateStateLocked(ctx, tx, userID)
		if err != nil {
			return err
		}

		if force {
			if err := h.clearTutorialContent(ctx, tx, state); err != nil {
				return err
			}
		} else {
			if state.CompletedAt != nil {
				activatedState = state
				return nil
			}
			if state.Stage == models.TutorialStageScenario && state.TutorialScenarioID != nil {
				existing, err := (&scenarioHandle{db: tx}).FindByID(ctx, *state.TutorialScenarioID)
				if err == nil && existing != nil {
					activatedState = state
					activatedScenario = existing
					return nil
				}
			}
		}

		now := time.Now()
		scenario.ID = uuid.New()
		scenario.CreatedAt = now
		scenario.UpdatedAt = now
		resolvedScenarioGenreID, err := resolveScenarioGenreID(ctx, tx, scenario)
		if err != nil {
			return err
		}
		scenario.GenreID = resolvedScenarioGenreID
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

		state.Stage = models.TutorialStageScenario
		state.TutorialScenarioID = &scenario.ID
		state.SelectedScenarioOptionID = nil
		state.RequiredEquipItemIDs = []int{}
		state.CompletedEquipItemIDs = []int{}
		state.RequiredUseItemIDs = []int{}
		state.CompletedUseItemIDs = []int{}
		state.TutorialMonsterEncounterID = nil
		state.ActivatedAt = &now
		state.CompletedAt = nil
		state.UpdatedAt = now
		if err := tx.Save(state).Error; err != nil {
			return err
		}

		activatedState = state
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

func (h *tutorialHandle) MarkScenarioResolved(
	ctx context.Context,
	userID uuid.UUID,
	scenarioID uuid.UUID,
	selectedOptionID *uuid.UUID,
	requiredEquipItemIDs []int,
	requiredUseItemIDs []int,
) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		state, err := h.getOrCreateStateLocked(ctx, tx, userID)
		if err != nil {
			return err
		}
		if state.TutorialScenarioID == nil || *state.TutorialScenarioID != scenarioID {
			return nil
		}
		state.Stage = models.TutorialStageLoadout
		state.SelectedScenarioOptionID = selectedOptionID
		state.RequiredEquipItemIDs = requiredEquipItemIDs
		state.CompletedEquipItemIDs = []int{}
		state.RequiredUseItemIDs = requiredUseItemIDs
		state.CompletedUseItemIDs = []int{}
		state.UpdatedAt = time.Now()
		return tx.Save(state).Error
	})
}

func (h *tutorialHandle) AdvanceToPostWelcomeDialogue(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		state, err := h.getOrCreateStateLocked(ctx, tx, userID)
		if err != nil {
			return err
		}
		if state.CompletedAt != nil || state.Stage != models.TutorialStageWelcome {
			return nil
		}
		state.Stage = models.TutorialStagePostWelcomeDialogue
		state.UpdatedAt = time.Now()
		return tx.Save(state).Error
	})
}

func (h *tutorialHandle) AdvanceToPostScenarioDialogue(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		state, err := h.getOrCreateStateLocked(ctx, tx, userID)
		if err != nil {
			return err
		}
		if state.CompletedAt != nil || state.Stage != models.TutorialStageLoadout {
			return nil
		}
		state.Stage = models.TutorialStagePostScenarioDialogue
		state.UpdatedAt = time.Now()
		return tx.Save(state).Error
	})
}

func (h *tutorialHandle) RecordEquippedItem(ctx context.Context, userID uuid.UUID, inventoryItemID int) error {
	return h.recordProgressItem(ctx, userID, inventoryItemID, true)
}

func (h *tutorialHandle) RecordUsedItem(ctx context.Context, userID uuid.UUID, inventoryItemID int) error {
	return h.recordProgressItem(ctx, userID, inventoryItemID, false)
}

func (h *tutorialHandle) ActivateMonsterForUser(
	ctx context.Context,
	userID uuid.UUID,
	encounter *models.MonsterEncounter,
	members []models.MonsterEncounterMember,
	monsters []models.Monster,
) (*models.UserTutorialState, *models.MonsterEncounter, error) {
	var activatedState *models.UserTutorialState
	var activatedEncounter *models.MonsterEncounter

	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		state, err := h.getOrCreateStateLocked(ctx, tx, userID)
		if err != nil {
			return err
		}
		if state.CompletedAt != nil {
			activatedState = state
			return nil
		}
		if state.Stage == models.TutorialStageMonster && state.TutorialMonsterEncounterID != nil {
			existing, err := (&monsterEncounterHandle{db: tx}).FindByID(ctx, *state.TutorialMonsterEncounterID)
			if err == nil && existing != nil {
				activatedState = state
				activatedEncounter = existing
				return nil
			}
		}

		now := time.Now()
		createdMonsters := make([]models.Monster, 0, len(monsters))
		for _, monster := range monsters {
			monster.ID = uuid.New()
			monster.CreatedAt = now
			monster.UpdatedAt = now
			resolvedMonsterGenreID, err := resolveMonsterGenreID(ctx, tx, &monster)
			if err != nil {
				return err
			}
			monster.GenreID = resolvedMonsterGenreID
			if err := monster.SetGeometry(monster.Latitude, monster.Longitude); err != nil {
				return err
			}
			rewards := monster.ItemRewards
			monster.ItemRewards = nil
			if err := tx.Omit(clause.Associations).Create(&monster).Error; err != nil {
				return err
			}
			for _, reward := range rewards {
				reward.ID = uuid.New()
				reward.MonsterID = monster.ID
				reward.CreatedAt = now
				reward.UpdatedAt = now
				if err := tx.Create(&reward).Error; err != nil {
					return err
				}
			}
			monster.ItemRewards = rewards
			createdMonsters = append(createdMonsters, monster)
		}

		encounter.ID = uuid.New()
		encounter.CreatedAt = now
		encounter.UpdatedAt = now
		if err := encounter.SetGeometry(encounter.Latitude, encounter.Longitude); err != nil {
			return err
		}
		if err := tx.Omit(clause.Associations).Create(encounter).Error; err != nil {
			return err
		}

		for index, member := range members {
			member.ID = uuid.New()
			member.CreatedAt = now
			member.UpdatedAt = now
			member.MonsterEncounterID = encounter.ID
			if index < len(createdMonsters) {
				member.MonsterID = createdMonsters[index].ID
			}
			if err := tx.Create(&member).Error; err != nil {
				return err
			}
		}

		state.Stage = models.TutorialStageMonster
		state.TutorialMonsterEncounterID = &encounter.ID
		state.UpdatedAt = now
		if err := tx.Save(state).Error; err != nil {
			return err
		}

		activatedState = state
		loadedEncounter, err := (&monsterEncounterHandle{db: tx}).FindByID(ctx, encounter.ID)
		if err != nil {
			return err
		}
		activatedEncounter = loadedEncounter
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return activatedState, activatedEncounter, nil
}

func (h *tutorialHandle) ResetForReplay(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		state, err := h.getOrCreateStateLocked(ctx, tx, userID)
		if err != nil {
			return err
		}
		return h.clearTutorialContent(ctx, tx, state)
	})
}

func (h *tutorialHandle) MarkCompleted(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return h.db.WithContext(ctx).Model(&models.UserTutorialState{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"stage":                         models.TutorialStageCompleted,
			"tutorial_scenario_id":          nil,
			"tutorial_monster_encounter_id": nil,
			"completed_at":                  now,
			"updated_at":                    now,
		}).Error
}

func (h *tutorialHandle) AdvanceToBaseKit(
	ctx context.Context,
	userID uuid.UUID,
	requiredUseItemIDs []int,
) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		state, err := h.getOrCreateStateLocked(ctx, tx, userID)
		if err != nil {
			return err
		}
		if state.CompletedAt != nil || state.Stage != models.TutorialStagePostMonsterDialogue {
			return nil
		}
		state.Stage = models.TutorialStageBaseKit
		state.RequiredEquipItemIDs = []int{}
		state.CompletedEquipItemIDs = []int{}
		state.RequiredUseItemIDs = requiredUseItemIDs
		state.CompletedUseItemIDs = []int{}
		state.UpdatedAt = time.Now()
		return tx.Save(state).Error
	})
}

func (h *tutorialHandle) AdvanceToPostBaseDialogue(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		state, err := h.getOrCreateStateLocked(ctx, tx, userID)
		if err != nil {
			return err
		}
		if state.CompletedAt != nil ||
			(state.Stage != models.TutorialStageBaseKit &&
				state.Stage != models.TutorialStagePostBasePlacement &&
				state.Stage != models.TutorialStageHearth) {
			return nil
		}
		state.Stage = models.TutorialStagePostBaseDialogue
		state.RequiredEquipItemIDs = []int{}
		state.CompletedEquipItemIDs = []int{}
		state.RequiredUseItemIDs = []int{}
		state.CompletedUseItemIDs = []int{}
		state.UpdatedAt = time.Now()
		return tx.Save(state).Error
	})
}

func (h *tutorialHandle) MarkMonsterCompleted(ctx context.Context, userID uuid.UUID, monsterEncounterID uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		state, err := h.getOrCreateStateLocked(ctx, tx, userID)
		if err != nil {
			return err
		}
		if state.TutorialMonsterEncounterID == nil || *state.TutorialMonsterEncounterID != monsterEncounterID {
			return nil
		}
		now := time.Now()
		state.Stage = models.TutorialStagePostMonsterDialogue
		state.CompletedAt = nil
		state.TutorialScenarioID = nil
		state.TutorialMonsterEncounterID = nil
		state.RequiredEquipItemIDs = []int{}
		state.CompletedEquipItemIDs = []int{}
		state.RequiredUseItemIDs = []int{}
		state.CompletedUseItemIDs = []int{}
		state.UpdatedAt = now
		return tx.Save(state).Error
	})
}

func (h *tutorialHandle) recordProgressItem(
	ctx context.Context,
	userID uuid.UUID,
	inventoryItemID int,
	equip bool,
) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		state, err := h.getOrCreateStateLocked(ctx, tx, userID)
		if err != nil {
			return err
		}
		if state.Stage != models.TutorialStageLoadout && state.Stage != models.TutorialStageBaseKit {
			return nil
		}

		updated := false
		if equip {
			if containsTutorialInt(state.RequiredEquipItemIDs, inventoryItemID) &&
				!containsTutorialInt(state.CompletedEquipItemIDs, inventoryItemID) {
				state.CompletedEquipItemIDs = append(state.CompletedEquipItemIDs, inventoryItemID)
				updated = true
			}
		} else {
			if containsTutorialInt(state.RequiredUseItemIDs, inventoryItemID) &&
				!containsTutorialInt(state.CompletedUseItemIDs, inventoryItemID) {
				state.CompletedUseItemIDs = append(state.CompletedUseItemIDs, inventoryItemID)
				updated = true
			}
		}
		if !updated {
			return nil
		}

		state.UpdatedAt = time.Now()
		return tx.Save(state).Error
	})
}

func (h *tutorialHandle) getOrCreateConfig(ctx context.Context, db *gorm.DB) (*models.TutorialConfig, error) {
	var config models.TutorialConfig
	err := db.WithContext(ctx).
		Preload("Character").
		Preload("BaseQuestArchetype").
		Preload("BaseQuestGiverCharacter").
		Preload("BaseQuestGiverCharacterTemplate").
		Preload("MonsterEncounter").
		First(&config, 1).Error
	if err == nil {
		return &config, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	created := models.TutorialConfig{
		ID:                      1,
		Dialogue:                models.DialogueSequence{},
		PostWelcomeDialogue:     models.DialogueSequence{},
		GuideSupportGreeting:    "",
		GuideSupportPersonality: "",
		GuideSupportBehavior:    "",
		ScenarioObjectiveCopy:   "Complete the tutorial scenario to continue.",
		PostScenarioDialogue:    models.DialogueSequence{},
		LoadoutDialogue: models.DialogueSequence{
			{
				Speaker: "character",
				Text:    "Equip your new gear and use the spellbook before you head back out.",
				Order:   0,
			},
		},
		LoadoutObjectiveCopy: "Equip your new gear and use the spellbook to continue.",
		PostMonsterDialogue:  models.DialogueSequence{},
		BaseKitDialogue: models.DialogueSequence{
			{
				Speaker: "character",
				Text:    "Use the home base kit you just earned and claim a safe place for yourself.",
				Order:   0,
			},
		},
		BaseKitObjectiveCopy:      "Use your home base kit to establish your base.",
		PostBasePlacementDialogue: models.DialogueSequence{},
		HearthObjectiveCopy:       "Use your hearth to heal yourself before the tutorial continues.",
		PostBaseDialogue:          models.DialogueSequence{},
		ScenarioPrompt:            "You hear a commotion outside of your door.",
		ScenarioImageURL:          "",
		ImageGenerationStatus:     models.TutorialImageGenerationStatusNone,
		Options:                   []models.TutorialScenarioOption{},
		MonsterObjectiveCopy:      "Defeat the tutorial monster encounter to continue.",
		MonsterItemRewards:        []models.TutorialItemReward{},
		ItemRewards:               []models.TutorialItemReward{},
		SpellRewards:              []models.TutorialSpellReward{},
	}
	if err := db.WithContext(ctx).Create(&created).Error; err != nil {
		return nil, err
	}
	return h.getOrCreateConfig(ctx, db)
}

func (h *tutorialHandle) getOrCreateStateLocked(
	ctx context.Context,
	tx *gorm.DB,
	userID uuid.UUID,
) (*models.UserTutorialState, error) {
	var state models.UserTutorialState
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&state).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		now := time.Now()
		state = models.UserTutorialState{
			UserID:    userID,
			Stage:     models.TutorialStageWelcome,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := tx.Create(&state).Error; err != nil {
			return nil, err
		}
	}
	return &state, nil
}

func (h *tutorialHandle) clearTutorialContent(
	ctx context.Context,
	tx *gorm.DB,
	state *models.UserTutorialState,
) error {
	if state == nil {
		return nil
	}
	if state.TutorialScenarioID != nil {
		if err := tx.Delete(&models.Scenario{}, "id = ?", *state.TutorialScenarioID).Error; err != nil {
			return err
		}
	}
	if state.TutorialMonsterEncounterID != nil {
		if err := tx.WithContext(ctx).
			Where(
				"user_id = ? AND monster_encounter_id = ?",
				state.UserID,
				*state.TutorialMonsterEncounterID,
			).
			Delete(&models.UserMonsterEncounterVictory{}).Error; err != nil {
			return err
		}
		if err := h.deleteTutorialMonsterEncounter(ctx, tx, *state.TutorialMonsterEncounterID); err != nil {
			return err
		}
	}
	state.Stage = models.TutorialStageWelcome
	state.TutorialScenarioID = nil
	state.SelectedScenarioOptionID = nil
	state.RequiredEquipItemIDs = []int{}
	state.CompletedEquipItemIDs = []int{}
	state.RequiredUseItemIDs = []int{}
	state.CompletedUseItemIDs = []int{}
	state.TutorialMonsterEncounterID = nil
	state.CompletedAt = nil
	state.ActivatedAt = nil
	state.UpdatedAt = time.Now()
	return tx.Save(state).Error
}

func (h *tutorialHandle) deleteTutorialMonsterEncounter(
	ctx context.Context,
	tx *gorm.DB,
	encounterID uuid.UUID,
) error {
	var members []models.MonsterEncounterMember
	if err := tx.WithContext(ctx).
		Where("monster_encounter_id = ?", encounterID).
		Find(&members).Error; err != nil {
		return err
	}
	monsterIDs := make([]uuid.UUID, 0, len(members))
	for _, member := range members {
		if member.MonsterID != uuid.Nil {
			monsterIDs = append(monsterIDs, member.MonsterID)
		}
	}
	if err := tx.Delete(&models.MonsterEncounter{}, "id = ?", encounterID).Error; err != nil {
		return err
	}
	if len(monsterIDs) > 0 {
		if err := tx.Delete(&models.Monster{}, "id IN ?", monsterIDs).Error; err != nil {
			return err
		}
	}
	return nil
}

func containsTutorialInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
