package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type baseResourceBalanceHandle struct {
	db *gorm.DB
}

func (h *baseResourceBalanceHandle) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.BaseResourceBalance, error) {
	var balances []models.BaseResourceBalance
	if err := h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("resource_key ASC").
		Find(&balances).Error; err != nil {
		return nil, err
	}
	return balances, nil
}

func (h *baseResourceBalanceHandle) GrantToUser(ctx context.Context, userID uuid.UUID, deltas []models.BaseResourceDelta, sourceType string, sourceID *uuid.UUID, notes *string) error {
	normalized := normalizeBaseResourceDeltas(deltas)
	if len(normalized) == 0 {
		return nil
	}
	now := time.Now()
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, delta := range normalized {
			balance := &models.BaseResourceBalance{
				UserID:      userID,
				ResourceKey: delta.ResourceKey,
				Amount:      delta.Amount,
				UpdatedAt:   now,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "user_id"}, {Name: "resource_key"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"amount":     gorm.Expr("base_resource_balances.amount + EXCLUDED.amount"),
					"updated_at": now,
				}),
			}).Create(balance).Error; err != nil {
				return err
			}
			entry := &models.BaseResourceLedger{
				ID:          uuid.New(),
				UserID:      userID,
				ResourceKey: delta.ResourceKey,
				Delta:       delta.Amount,
				SourceType:  sourceType,
				SourceID:    sourceID,
				Notes:       notes,
				CreatedAt:   now,
			}
			if err := tx.Create(entry).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type baseResourceLedgerHandle struct {
	db *gorm.DB
}

func (h *baseResourceLedgerHandle) ListRecentByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]models.BaseResourceLedger, error) {
	if limit <= 0 {
		limit = 25
	}
	var entries []models.BaseResourceLedger
	if err := h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

type baseStructureDefinitionHandle struct {
	db *gorm.DB
}

func (h *baseStructureDefinitionHandle) FindAll(ctx context.Context) ([]models.BaseStructureDefinition, error) {
	var definitions []models.BaseStructureDefinition
	if err := h.db.WithContext(ctx).
		Preload("LevelCosts", func(db *gorm.DB) *gorm.DB {
			return db.Order("level ASC").Order("resource_key ASC")
		}).
		Preload("LevelVisuals", func(db *gorm.DB) *gorm.DB {
			return db.Order("level ASC")
		}).
		Order("sort_order ASC").
		Order("name ASC").
		Find(&definitions).Error; err != nil {
		return nil, err
	}
	return definitions, nil
}

func (h *baseStructureDefinitionHandle) FindAllActive(ctx context.Context) ([]models.BaseStructureDefinition, error) {
	var definitions []models.BaseStructureDefinition
	if err := h.db.WithContext(ctx).
		Preload("LevelCosts", func(db *gorm.DB) *gorm.DB {
			return db.Order("level ASC").Order("resource_key ASC")
		}).
		Preload("LevelVisuals", func(db *gorm.DB) *gorm.DB {
			return db.Order("level ASC")
		}).
		Where("active = ?", true).
		Order("sort_order ASC").
		Order("name ASC").
		Find(&definitions).Error; err != nil {
		return nil, err
	}
	return definitions, nil
}

func (h *baseStructureDefinitionHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.BaseStructureDefinition, error) {
	var definition models.BaseStructureDefinition
	if err := h.db.WithContext(ctx).
		Preload("LevelCosts", func(db *gorm.DB) *gorm.DB {
			return db.Order("level ASC").Order("resource_key ASC")
		}).
		Preload("LevelVisuals", func(db *gorm.DB) *gorm.DB {
			return db.Order("level ASC")
		}).
		First(&definition, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &definition, nil
}

func (h *baseStructureDefinitionHandle) FindActiveByKey(ctx context.Context, key string) (*models.BaseStructureDefinition, error) {
	var definition models.BaseStructureDefinition
	if err := h.db.WithContext(ctx).
		Preload("LevelCosts", func(db *gorm.DB) *gorm.DB {
			return db.Order("level ASC").Order("resource_key ASC")
		}).
		Preload("LevelVisuals", func(db *gorm.DB) *gorm.DB {
			return db.Order("level ASC")
		}).
		Where("active = ? AND key = ?", true, key).
		First(&definition).Error; err != nil {
		return nil, err
	}
	return &definition, nil
}

func (h *baseStructureDefinitionHandle) UpdatePrompts(ctx context.Context, id uuid.UUID, imagePrompt string, topDownImagePrompt string) error {
	return h.db.WithContext(ctx).
		Model(&models.BaseStructureDefinition{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"image_prompt":          imagePrompt,
			"top_down_image_prompt": topDownImagePrompt,
			"updated_at":            time.Now(),
		}).Error
}

type baseStructureLevelVisualHandle struct {
	db *gorm.DB
}

func (h *baseStructureLevelVisualHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.BaseStructureLevelVisual, error) {
	var visual models.BaseStructureLevelVisual
	if err := h.db.WithContext(ctx).First(&visual, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &visual, nil
}

func (h *baseStructureLevelVisualHandle) FindByDefinitionIDAndLevel(ctx context.Context, definitionID uuid.UUID, level int) (*models.BaseStructureLevelVisual, error) {
	var visual models.BaseStructureLevelVisual
	if err := h.db.WithContext(ctx).
		Where("structure_definition_id = ? AND level = ?", definitionID, level).
		First(&visual).Error; err != nil {
		return nil, err
	}
	return &visual, nil
}

func (h *baseStructureLevelVisualHandle) Upsert(ctx context.Context, visual *models.BaseStructureLevelVisual) error {
	if visual == nil {
		return fmt.Errorf("base structure level visual is required")
	}
	now := time.Now()
	if visual.ID == uuid.Nil {
		visual.ID = uuid.New()
	}
	if visual.CreatedAt.IsZero() {
		visual.CreatedAt = now
	}
	visual.UpdatedAt = now
	return h.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "structure_definition_id"}, {Name: "level"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"image_url":                        visual.ImageURL,
			"thumbnail_url":                    visual.ThumbnailURL,
			"image_generation_status":          visual.ImageGenerationStatus,
			"image_generation_error":           visual.ImageGenerationError,
			"top_down_image_url":               visual.TopDownImageURL,
			"top_down_thumbnail_url":           visual.TopDownThumbnailURL,
			"top_down_image_generation_status": visual.TopDownImageGenerationStatus,
			"top_down_image_generation_error":  visual.TopDownImageGenerationError,
			"updated_at":                       now,
		}),
	}).Create(visual).Error
}

func (h *baseStructureLevelVisualHandle) Update(ctx context.Context, id uuid.UUID, updates *models.BaseStructureLevelVisual) error {
	if updates == nil {
		return fmt.Errorf("base structure level visual update is required")
	}
	updates.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).
		Model(&models.BaseStructureLevelVisual{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (h *baseStructureLevelVisualHandle) UpdateThumbnailURL(ctx context.Context, id uuid.UUID, thumbnailURL string) error {
	return h.db.WithContext(ctx).
		Model(&models.BaseStructureLevelVisual{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"thumbnail_url": thumbnailURL,
			"updated_at":    time.Now(),
		}).Error
}

func (h *baseStructureLevelVisualHandle) UpdateTopDownThumbnailURL(ctx context.Context, id uuid.UUID, thumbnailURL string) error {
	return h.db.WithContext(ctx).
		Model(&models.BaseStructureLevelVisual{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"top_down_thumbnail_url": thumbnailURL,
			"updated_at":             time.Now(),
		}).Error
}

type userBaseStructureHandle struct {
	db *gorm.DB
}

func (h *userBaseStructureHandle) FindByBaseID(ctx context.Context, baseID uuid.UUID) ([]models.UserBaseStructure, error) {
	var structures []models.UserBaseStructure
	if err := h.db.WithContext(ctx).
		Where("base_id = ?", baseID).
		Order("grid_y ASC").
		Order("grid_x ASC").
		Order("created_at ASC").
		Order("structure_key ASC").
		Find(&structures).Error; err != nil {
		return nil, err
	}
	return structures, nil
}

func (h *userBaseStructureHandle) EnsureBuilt(ctx context.Context, baseID uuid.UUID, userID uuid.UUID, structureKey string, level int, gridX int, gridY int) error {
	if baseID == uuid.Nil || userID == uuid.Nil || structureKey == "" || level <= 0 {
		return nil
	}
	now := time.Now()
	structure := &models.UserBaseStructure{
		ID:           uuid.New(),
		CreatedAt:    now,
		UpdatedAt:    now,
		BaseID:       baseID,
		UserID:       userID,
		StructureKey: structureKey,
		Level:        level,
		GridX:        gridX,
		GridY:        gridY,
	}
	return h.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "base_id"}, {Name: "structure_key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"level":      level,
			"user_id":    userID,
			"grid_x":     gridX,
			"grid_y":     gridY,
			"updated_at": now,
		}),
	}).Create(structure).Error
}

func (h *userBaseStructureHandle) UpsertLevelWithCost(ctx context.Context, baseID uuid.UUID, userID uuid.UUID, structureKey string, level int, costs []models.BaseResourceDelta, gridX *int, gridY *int) (*models.UserBaseStructure, error) {
	normalizedCosts := normalizeBaseResourceDeltas(costs)
	if baseID == uuid.Nil || userID == uuid.Nil || structureKey == "" || level <= 0 {
		return nil, fmt.Errorf("invalid base structure update")
	}
	now := time.Now()
	var structure models.UserBaseStructure
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(normalizedCosts) > 0 {
			keys := make([]models.BaseResourceKey, 0, len(normalizedCosts))
			required := make(map[models.BaseResourceKey]int, len(normalizedCosts))
			for _, cost := range normalizedCosts {
				keys = append(keys, cost.ResourceKey)
				required[cost.ResourceKey] = cost.Amount
			}

			var balances []models.BaseResourceBalance
			if err := tx.Where("user_id = ? AND resource_key IN ?", userID, keys).Find(&balances).Error; err != nil {
				return err
			}
			available := make(map[models.BaseResourceKey]int, len(balances))
			for _, balance := range balances {
				available[balance.ResourceKey] = balance.Amount
			}
			for resourceKey, amount := range required {
				if available[resourceKey] < amount {
					return fmt.Errorf("not enough %s", resourceKey)
				}
			}
			for _, cost := range normalizedCosts {
				if err := tx.Model(&models.BaseResourceBalance{}).
					Where("user_id = ? AND resource_key = ? AND amount >= ?", userID, cost.ResourceKey, cost.Amount).
					Updates(map[string]interface{}{
						"amount":     gorm.Expr("amount - ?", cost.Amount),
						"updated_at": now,
					}).Error; err != nil {
					return err
				}
			}
		}

		upserted := &models.UserBaseStructure{
			ID:           uuid.New(),
			CreatedAt:    now,
			UpdatedAt:    now,
			BaseID:       baseID,
			UserID:       userID,
			StructureKey: structureKey,
			Level:        level,
		}
		assignments := map[string]interface{}{
			"level":      level,
			"updated_at": now,
			"user_id":    userID,
		}
		if gridX != nil {
			upserted.GridX = *gridX
			assignments["grid_x"] = *gridX
		}
		if gridY != nil {
			upserted.GridY = *gridY
			assignments["grid_y"] = *gridY
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "base_id"}, {Name: "structure_key"}},
			DoUpdates: clause.Assignments(assignments),
		}).Create(upserted).Error; err != nil {
			return err
		}
		return tx.Where("base_id = ? AND structure_key = ?", baseID, structureKey).First(&structure).Error
	})
	if err != nil {
		return nil, err
	}
	return &structure, nil
}

func (h *userBaseStructureHandle) MoveMany(ctx context.Context, baseID uuid.UUID, userID uuid.UUID, positions map[string]models.BaseGridPosition) error {
	if baseID == uuid.Nil || userID == uuid.Nil || len(positions) == 0 {
		return fmt.Errorf("invalid base structure move")
	}
	now := time.Now()
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		index := 0
		for structureKey := range positions {
			result := tx.Model(&models.UserBaseStructure{}).
				Where("base_id = ? AND user_id = ? AND structure_key = ?", baseID, userID, structureKey).
				Updates(map[string]interface{}{
					"grid_x":     -(index + 1000),
					"grid_y":     -(index + 1000),
					"updated_at": now,
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("structure %s was not found", structureKey)
			}
			index++
		}
		for structureKey, position := range positions {
			result := tx.Model(&models.UserBaseStructure{}).
				Where("base_id = ? AND user_id = ? AND structure_key = ?", baseID, userID, structureKey).
				Updates(map[string]interface{}{
					"grid_x":     position.GridX,
					"grid_y":     position.GridY,
					"updated_at": now,
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("structure %s was not found", structureKey)
			}
		}
		return nil
	})
}

func (h *userBaseStructureHandle) DeleteByStructureKey(ctx context.Context, baseID uuid.UUID, userID uuid.UUID, structureKey string) error {
	if baseID == uuid.Nil || userID == uuid.Nil || strings.TrimSpace(structureKey) == "" {
		return fmt.Errorf("invalid base structure delete")
	}
	result := h.db.WithContext(ctx).
		Where("base_id = ? AND user_id = ? AND structure_key = ?", baseID, userID, strings.TrimSpace(structureKey)).
		Delete(&models.UserBaseStructure{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("structure %s was not found", structureKey)
	}
	return nil
}

type userBaseDailyStateHandle struct {
	db *gorm.DB
}

func (h *userBaseDailyStateHandle) FindActiveByUserID(ctx context.Context, userID uuid.UUID, asOf time.Time) ([]models.UserBaseDailyState, error) {
	var states []models.UserBaseDailyState
	asOfDate := time.Date(asOf.Year(), asOf.Month(), asOf.Day(), 0, 0, 0, 0, asOf.Location())
	if err := h.db.WithContext(ctx).
		Where("user_id = ? AND resets_on >= ?", userID, asOfDate).
		Order("resets_on ASC").
		Order("state_key ASC").
		Find(&states).Error; err != nil {
		return nil, err
	}
	return states, nil
}

func normalizeBaseResourceDeltas(deltas []models.BaseResourceDelta) []models.BaseResourceDelta {
	if len(deltas) == 0 {
		return nil
	}
	merged := map[models.BaseResourceKey]int{}
	order := make([]models.BaseResourceKey, 0, len(deltas))
	for _, delta := range deltas {
		if delta.ResourceKey == "" || delta.Amount <= 0 {
			continue
		}
		if _, exists := merged[delta.ResourceKey]; !exists {
			order = append(order, delta.ResourceKey)
		}
		merged[delta.ResourceKey] += delta.Amount
	}
	result := make([]models.BaseResourceDelta, 0, len(order))
	for _, key := range order {
		amount := merged[key]
		if amount <= 0 {
			continue
		}
		result = append(result, models.BaseResourceDelta{
			ResourceKey: key,
			Amount:      amount,
		})
	}
	return result
}
