package db

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type scenarioHandle struct {
	db *gorm.DB
}

type scenarioAdminListRow struct {
	ID        uuid.UUID `gorm:"column:id"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func normalizeScenarioFailurePenaltyDefaults(scenario *models.Scenario) {
	if scenario == nil {
		return
	}
	if scenario.FailurePenaltyMode == "" {
		scenario.FailurePenaltyMode = models.ScenarioFailurePenaltyModeShared
	}
	if scenario.FailureHealthDrainType == "" {
		scenario.FailureHealthDrainType = models.ScenarioFailureDrainTypeNone
	}
	if scenario.FailureManaDrainType == "" {
		scenario.FailureManaDrainType = models.ScenarioFailureDrainTypeNone
	}
	if scenario.FailureHealthDrainType == models.ScenarioFailureDrainTypeNone {
		scenario.FailureHealthDrainValue = 0
	}
	if scenario.FailureManaDrainType == models.ScenarioFailureDrainTypeNone {
		scenario.FailureManaDrainValue = 0
	}
	if scenario.FailureHealthDrainValue < 0 {
		scenario.FailureHealthDrainValue = 0
	}
	if scenario.FailureManaDrainValue < 0 {
		scenario.FailureManaDrainValue = 0
	}
	if scenario.FailureStatuses == nil {
		scenario.FailureStatuses = models.ScenarioFailureStatusTemplates{}
	}

	if scenario.SuccessRewardMode == "" {
		scenario.SuccessRewardMode = models.ScenarioSuccessRewardModeShared
	}
	if scenario.SuccessHealthRestoreType == "" {
		scenario.SuccessHealthRestoreType = models.ScenarioFailureDrainTypeNone
	}
	if scenario.SuccessManaRestoreType == "" {
		scenario.SuccessManaRestoreType = models.ScenarioFailureDrainTypeNone
	}
	if scenario.SuccessHealthRestoreType == models.ScenarioFailureDrainTypeNone {
		scenario.SuccessHealthRestoreValue = 0
	}
	if scenario.SuccessManaRestoreType == models.ScenarioFailureDrainTypeNone {
		scenario.SuccessManaRestoreValue = 0
	}
	if scenario.SuccessHealthRestoreValue < 0 {
		scenario.SuccessHealthRestoreValue = 0
	}
	if scenario.SuccessManaRestoreValue < 0 {
		scenario.SuccessManaRestoreValue = 0
	}
	if scenario.SuccessStatuses == nil {
		scenario.SuccessStatuses = models.ScenarioFailureStatusTemplates{}
	}
	if strings.TrimSpace(string(scenario.RewardMode)) == "" {
		if scenario.RewardExperience > 0 || scenario.RewardGold > 0 {
			scenario.RewardMode = models.RewardModeExplicit
		} else {
			scenario.RewardMode = models.RewardModeRandom
		}
	}
	scenario.RewardMode = models.NormalizeRewardMode(string(scenario.RewardMode))
	scenario.RandomRewardSize = models.NormalizeRandomRewardSize(string(scenario.RandomRewardSize))
}

func normalizeScenarioOptionFailurePenaltyDefaults(option *models.ScenarioOption) {
	if option == nil {
		return
	}
	if option.FailureHealthDrainType == "" {
		option.FailureHealthDrainType = models.ScenarioFailureDrainTypeNone
	}
	if option.FailureManaDrainType == "" {
		option.FailureManaDrainType = models.ScenarioFailureDrainTypeNone
	}
	if option.FailureHealthDrainType == models.ScenarioFailureDrainTypeNone {
		option.FailureHealthDrainValue = 0
	}
	if option.FailureManaDrainType == models.ScenarioFailureDrainTypeNone {
		option.FailureManaDrainValue = 0
	}
	if option.FailureHealthDrainValue < 0 {
		option.FailureHealthDrainValue = 0
	}
	if option.FailureManaDrainValue < 0 {
		option.FailureManaDrainValue = 0
	}
	if option.FailureStatuses == nil {
		option.FailureStatuses = models.ScenarioFailureStatusTemplates{}
	}

	if option.SuccessHealthRestoreType == "" {
		option.SuccessHealthRestoreType = models.ScenarioFailureDrainTypeNone
	}
	if option.SuccessManaRestoreType == "" {
		option.SuccessManaRestoreType = models.ScenarioFailureDrainTypeNone
	}
	if option.SuccessHealthRestoreType == models.ScenarioFailureDrainTypeNone {
		option.SuccessHealthRestoreValue = 0
	}
	if option.SuccessManaRestoreType == models.ScenarioFailureDrainTypeNone {
		option.SuccessManaRestoreValue = 0
	}
	if option.SuccessHealthRestoreValue < 0 {
		option.SuccessHealthRestoreValue = 0
	}
	if option.SuccessManaRestoreValue < 0 {
		option.SuccessManaRestoreValue = 0
	}
	if option.SuccessStatuses == nil {
		option.SuccessStatuses = models.ScenarioFailureStatusTemplates{}
	}
}

func (h *scenarioHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("Zone").
		Preload("PointOfInterest").
		Preload("Options").
		Preload("Options.ItemRewards").
		Preload("Options.ItemRewards.InventoryItem").
		Preload("Options.ItemChoiceRewards").
		Preload("Options.ItemChoiceRewards.InventoryItem").
		Preload("Options.SpellRewards").
		Preload("Options.SpellRewards.Spell").
		Preload("ItemRewards").
		Preload("ItemRewards.InventoryItem").
		Preload("ItemChoiceRewards").
		Preload("ItemChoiceRewards.InventoryItem").
		Preload("SpellRewards").
		Preload("SpellRewards.Spell")
}

func (h *scenarioHandle) visibleQuery(ctx context.Context) *gorm.DB {
	return h.preloadBase(ctx).Where("retired_at IS NULL")
}

func (h *scenarioHandle) Create(ctx context.Context, scenario *models.Scenario) error {
	scenario.ID = uuid.New()
	scenario.CreatedAt = time.Now()
	scenario.UpdatedAt = time.Now()
	normalizeScenarioFailurePenaltyDefaults(scenario)
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
	if err := h.visibleQuery(ctx).Find(&scenarios).Error; err != nil {
		return nil, err
	}
	return scenarios, nil
}

func (h *scenarioHandle) adminListBaseQuery(
	ctx context.Context,
	params ScenarioAdminListParams,
	userID *uuid.UUID,
) *gorm.DB {
	query := h.db.WithContext(ctx).
		Model(&models.Scenario{}).
		Where("scenarios.retired_at IS NULL").
		Joins("LEFT JOIN zones ON zones.id = scenarios.zone_id")

	if userID == nil {
		query = query.Where("scenarios.owner_user_id IS NULL")
	} else {
		query = query.Where(
			"(scenarios.owner_user_id IS NULL OR scenarios.owner_user_id = ?)",
			*userID,
		)
	}

	if normalizedQuery := strings.TrimSpace(strings.ToLower(params.Query)); normalizedQuery != "" {
		searchTerm := "%" + normalizedQuery + "%"
		query = query.Where(
			`(
				LOWER(scenarios.prompt) LIKE ?
				OR LOWER(CAST(scenarios.id AS text)) LIKE ?
				OR LOWER(COALESCE(zones.name, '')) LIKE ?
			)`,
			searchTerm,
			searchTerm,
			searchTerm,
		)
	}

	if normalizedZoneQuery := strings.TrimSpace(strings.ToLower(params.ZoneQuery)); normalizedZoneQuery != "" {
		searchTerm := "%" + normalizedZoneQuery + "%"
		query = query.Where("LOWER(COALESCE(zones.name, '')) LIKE ?", searchTerm)
	}

	return query
}

func (h *scenarioHandle) ListAdmin(
	ctx context.Context,
	params ScenarioAdminListParams,
	userID *uuid.UUID,
) (*ScenarioAdminListResult, error) {
	var total int64
	if err := h.adminListBaseQuery(ctx, params, userID).
		Distinct("scenarios.id").
		Count(&total).Error; err != nil {
		return nil, err
	}

	rows := []scenarioAdminListRow{}
	if err := h.adminListBaseQuery(ctx, params, userID).
		Select("scenarios.id, scenarios.updated_at").
		Distinct().
		Order("scenarios.updated_at DESC").
		Limit(params.PageSize).
		Offset((params.Page - 1) * params.PageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}

	scenarios := make([]models.Scenario, 0, len(ids))
	if len(ids) > 0 {
		loaded := []models.Scenario{}
		if err := h.preloadBase(ctx).
			Where("scenarios.id IN ?", ids).
			Where("scenarios.retired_at IS NULL").
			Find(&loaded).Error; err != nil {
			return nil, err
		}
		byID := make(map[uuid.UUID]models.Scenario, len(loaded))
		for _, scenario := range loaded {
			byID[scenario.ID] = scenario
		}
		for _, id := range ids {
			scenario, ok := byID[id]
			if ok {
				scenarios = append(scenarios, scenario)
			}
		}
	}

	attempted := map[uuid.UUID]bool{}
	if userID != nil && len(ids) > 0 {
		var attempts []models.UserScenarioAttempt
		if err := h.db.WithContext(ctx).
			Select("scenario_id").
			Where("user_id = ? AND scenario_id IN ?", *userID, ids).
			Find(&attempts).Error; err != nil {
			return nil, err
		}
		for _, attempt := range attempts {
			attempted[attempt.ScenarioID] = true
		}
	}

	return &ScenarioAdminListResult{
		Scenarios:         scenarios,
		AttemptedByUserID: attempted,
		Total:             total,
	}, nil
}

func (h *scenarioHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]models.Scenario, error) {
	var scenarios []models.Scenario
	if err := h.visibleQuery(ctx).
		Where("zone_id = ?", zoneID).
		Find(&scenarios).Error; err != nil {
		return nil, err
	}
	return scenarios, nil
}

func (h *scenarioHandle) FindByZoneIDExcludingQuestNodes(ctx context.Context, zoneID uuid.UUID) ([]models.Scenario, error) {
	var scenarios []models.Scenario
	if err := h.visibleQuery(ctx).
		Where("zone_id = ?", zoneID).
		Where("NOT EXISTS (SELECT 1 FROM quest_nodes qn WHERE qn.scenario_id = scenarios.id)").
		Find(&scenarios).Error; err != nil {
		return nil, err
	}
	return scenarios, nil
}

func (h *scenarioHandle) Update(ctx context.Context, id uuid.UUID, updates *models.Scenario) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	normalizeScenarioFailurePenaltyDefaults(updates)
	if err := updates.SetGeometry(updates.Latitude, updates.Longitude); err != nil {
		return err
	}

	payload := map[string]interface{}{
		"zone_id":                      updates.ZoneID,
		"point_of_interest_id":         updates.PointOfInterestID,
		"latitude":                     updates.Latitude,
		"longitude":                    updates.Longitude,
		"geometry":                     updates.Geometry,
		"prompt":                       updates.Prompt,
		"image_url":                    updates.ImageURL,
		"scale_with_user_level":        updates.ScaleWithUserLevel,
		"recurring_scenario_id":        updates.RecurringScenarioID,
		"recurrence_frequency":         updates.RecurrenceFrequency,
		"next_recurrence_at":           updates.NextRecurrenceAt,
		"retired_at":                   updates.RetiredAt,
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
		"updated_at":                   updates.UpdatedAt,
		"thumbnail_url":                updates.ThumbnailURL,
	}

	return h.db.WithContext(ctx).Model(&models.Scenario{}).Where("id = ?", id).Updates(payload).Error
}

func (h *scenarioHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Scenario{}, "id = ?", id).Error
}

func (h *scenarioHandle) FindDueRecurring(ctx context.Context, asOf time.Time, limit int) ([]models.Scenario, error) {
	var scenarios []models.Scenario
	query := h.db.WithContext(ctx).
		Where("retired_at IS NULL").
		Where("recurrence_frequency IS NOT NULL AND recurrence_frequency <> ''").
		Where("next_recurrence_at IS NOT NULL AND next_recurrence_at <= ?", asOf).
		Order("next_recurrence_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&scenarios).Error; err != nil {
		return nil, err
	}
	return scenarios, nil
}

func (h *scenarioHandle) ReplaceOptions(ctx context.Context, scenarioID uuid.UUID, options []models.ScenarioOption) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		optionIDs := tx.Model(&models.ScenarioOption{}).Select("id").Where("scenario_id = ?", scenarioID)
		if err := tx.Where("scenario_option_id IN (?)", optionIDs).Delete(&models.ScenarioOptionItemReward{}).Error; err != nil {
			return err
		}
		if err := tx.Where("scenario_option_id IN (?)", optionIDs).Delete(&models.ScenarioOptionItemChoiceReward{}).Error; err != nil {
			return err
		}
		if err := tx.Where("scenario_option_id IN (?)", optionIDs).Delete(&models.ScenarioOptionSpellReward{}).Error; err != nil {
			return err
		}
		if err := tx.Where("scenario_id = ?", scenarioID).Delete(&models.ScenarioOption{}).Error; err != nil {
			return err
		}

		now := time.Now()
		for _, option := range options {
			normalizeScenarioOptionFailurePenaltyDefaults(&option)
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
			for _, reward := range option.ItemChoiceRewards {
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

func (h *scenarioHandle) ReplaceItemChoiceRewards(ctx context.Context, scenarioID uuid.UUID, rewards []models.ScenarioItemChoiceReward) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("scenario_id = ?", scenarioID).Delete(&models.ScenarioItemChoiceReward{}).Error; err != nil {
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

func (h *scenarioHandle) ReplaceSpellRewards(ctx context.Context, scenarioID uuid.UUID, rewards []models.ScenarioSpellReward) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("scenario_id = ?", scenarioID).Delete(&models.ScenarioSpellReward{}).Error; err != nil {
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

func (h *scenarioHandle) UpsertItemChoicePending(ctx context.Context, userID uuid.UUID, scenarioID uuid.UUID, scenarioOptionID *uuid.UUID) error {
	now := time.Now()
	record := models.UserScenarioItemChoicePending{
		ID:               uuid.New(),
		CreatedAt:        now,
		UpdatedAt:        now,
		UserID:           userID,
		ScenarioID:       scenarioID,
		ScenarioOptionID: scenarioOptionID,
	}
	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "scenario_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"updated_at":         now,
				"scenario_option_id": scenarioOptionID,
			}),
		}).
		Create(&record).Error
}

func (h *scenarioHandle) FindItemChoicePendingByUserAndScenario(ctx context.Context, userID uuid.UUID, scenarioID uuid.UUID) (*models.UserScenarioItemChoicePending, error) {
	var pending models.UserScenarioItemChoicePending
	if err := h.db.WithContext(ctx).
		Where("user_id = ? AND scenario_id = ?", userID, scenarioID).
		First(&pending).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &pending, nil
}

func (h *scenarioHandle) DeleteItemChoicePending(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.UserScenarioItemChoicePending{}, "id = ?", id).Error
}

func (h *scenarioHandle) FindAllWithUserStatus(ctx context.Context, userID *uuid.UUID) ([]models.Scenario, map[uuid.UUID]bool, error) {
	query := h.visibleQuery(ctx)
	if userID == nil {
		query = query.Where("owner_user_id IS NULL")
	} else {
		query = query.Where("(owner_user_id IS NULL OR owner_user_id = ?)", *userID)
	}
	var scenarios []models.Scenario
	if err := query.Find(&scenarios).Error; err != nil {
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
	query := h.visibleQuery(ctx).Where("zone_id = ?", zoneID)
	if userID == nil {
		query = query.Where("owner_user_id IS NULL")
	} else {
		query = query.Where("(owner_user_id IS NULL OR owner_user_id = ?)", *userID)
	}
	var scenarios []models.Scenario
	if err := query.Find(&scenarios).Error; err != nil {
		return nil, nil, err
	}
	attempted := map[uuid.UUID]bool{}
	if userID == nil {
		return scenarios, attempted, nil
	}

	var attempts []models.UserScenarioAttempt
	if err := h.db.WithContext(ctx).
		Select("scenario_id").
		Where("user_id = ? AND scenario_id IN (?)", *userID, h.db.Model(&models.Scenario{}).Select("id").Where("zone_id = ? AND retired_at IS NULL", zoneID)).
		Find(&attempts).Error; err != nil {
		return nil, nil, err
	}
	for _, attempt := range attempts {
		attempted[attempt.ScenarioID] = true
	}
	return scenarios, attempted, nil
}

func (h *scenarioHandle) FindByZoneIDWithUserStatusExcludingQuestNodes(ctx context.Context, zoneID uuid.UUID, userID *uuid.UUID) ([]models.Scenario, map[uuid.UUID]bool, error) {
	query := h.visibleQuery(ctx).
		Where("zone_id = ?", zoneID).
		Where("NOT EXISTS (SELECT 1 FROM quest_nodes qn WHERE qn.scenario_id = scenarios.id)")
	if userID == nil {
		query = query.Where("owner_user_id IS NULL")
	} else {
		query = query.Where("(owner_user_id IS NULL OR owner_user_id = ?)", *userID)
	}
	var scenarios []models.Scenario
	if err := query.Find(&scenarios).Error; err != nil {
		return nil, nil, err
	}
	attempted := map[uuid.UUID]bool{}
	if userID == nil {
		return scenarios, attempted, nil
	}

	scenarioIDs := make([]uuid.UUID, 0, len(scenarios))
	for _, scenario := range scenarios {
		scenarioIDs = append(scenarioIDs, scenario.ID)
	}
	if len(scenarioIDs) == 0 {
		return scenarios, attempted, nil
	}

	var attempts []models.UserScenarioAttempt
	if err := h.db.WithContext(ctx).
		Select("scenario_id").
		Where("user_id = ? AND scenario_id IN ?", *userID, scenarioIDs).
		Find(&attempts).Error; err != nil {
		return nil, nil, err
	}
	for _, attempt := range attempts {
		attempted[attempt.ScenarioID] = true
	}
	return scenarios, attempted, nil
}
