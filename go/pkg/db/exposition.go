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

type expositionHandle struct {
	db *gorm.DB
}

func (h *expositionHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("Zone").
		Preload("PointOfInterest").
		Preload("ItemRewards").
		Preload("ItemRewards.InventoryItem").
		Preload("SpellRewards").
		Preload("SpellRewards.Spell")
}

func (h *expositionHandle) visibleQuery(ctx context.Context) *gorm.DB {
	return h.preloadBase(ctx)
}

func normalizeExpositionRewards(exposition *models.Exposition) {
	if exposition == nil {
		return
	}
	exposition.RequiredStoryFlags = normalizeJSONStringArray(exposition.RequiredStoryFlags)
	if exposition.Dialogue == nil {
		exposition.Dialogue = models.DialogueSequence{}
	}
	if strings.TrimSpace(string(exposition.RewardMode)) == "" {
		if exposition.RewardExperience > 0 ||
			exposition.RewardGold > 0 ||
			len(exposition.ItemRewards) > 0 ||
			len(exposition.SpellRewards) > 0 {
			exposition.RewardMode = models.RewardModeExplicit
		} else {
			exposition.RewardMode = models.RewardModeRandom
		}
	}
	exposition.RewardMode = models.NormalizeRewardMode(string(exposition.RewardMode))
	exposition.RandomRewardSize = models.NormalizeRandomRewardSize(string(exposition.RandomRewardSize))
	if exposition.RewardExperience < 0 {
		exposition.RewardExperience = 0
	}
	if exposition.RewardGold < 0 {
		exposition.RewardGold = 0
	}
}

func (h *expositionHandle) Create(ctx context.Context, exposition *models.Exposition) error {
	if exposition == nil {
		return nil
	}
	normalizeExpositionRewards(exposition)
	return h.db.WithContext(ctx).Create(exposition).Error
}

func (h *expositionHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Exposition, error) {
	var exposition models.Exposition
	if err := h.preloadBase(ctx).First(&exposition, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &exposition, nil
}

func (h *expositionHandle) FindAll(ctx context.Context) ([]models.Exposition, error) {
	var expositions []models.Exposition
	if err := h.visibleQuery(ctx).Find(&expositions).Error; err != nil {
		return nil, err
	}
	return expositions, nil
}

func (h *expositionHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]models.Exposition, error) {
	var expositions []models.Exposition
	if err := h.visibleQuery(ctx).
		Where("zone_id = ?", zoneID).
		Find(&expositions).Error; err != nil {
		return nil, err
	}
	return expositions, nil
}

func (h *expositionHandle) FindByZoneIDExcludingQuestNodes(ctx context.Context, zoneID uuid.UUID) ([]models.Exposition, error) {
	var expositions []models.Exposition
	if err := h.visibleQuery(ctx).
		Where("zone_id = ?", zoneID).
		Where("NOT EXISTS (SELECT 1 FROM quest_nodes qn WHERE qn.exposition_id = expositions.id)").
		Find(&expositions).Error; err != nil {
		return nil, err
	}
	return expositions, nil
}

func (h *expositionHandle) IsLinkedToQuestNode(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := h.db.WithContext(ctx).
		Table("quest_nodes").
		Where("exposition_id = ?", id).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (h *expositionHandle) Update(ctx context.Context, id uuid.UUID, updates *models.Exposition) error {
	if updates == nil {
		return nil
	}
	normalizeExpositionRewards(updates)
	payload := map[string]interface{}{
		"zone_id":               updates.ZoneID,
		"point_of_interest_id":  updates.PointOfInterestID,
		"latitude":              updates.Latitude,
		"longitude":             updates.Longitude,
		"geometry":              updates.Geometry,
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
		"updated_at":            updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).Model(&models.Exposition{}).Where("id = ?", id).Updates(payload).Error
}

func (h *expositionHandle) ReplaceItemRewards(ctx context.Context, expositionID uuid.UUID, rewards []models.ExpositionItemReward) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("exposition_id = ?", expositionID).Delete(&models.ExpositionItemReward{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, reward := range rewards {
			reward.ID = uuid.New()
			reward.ExpositionID = expositionID
			reward.CreatedAt = now
			reward.UpdatedAt = now
			if err := tx.Create(&reward).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *expositionHandle) ReplaceSpellRewards(ctx context.Context, expositionID uuid.UUID, rewards []models.ExpositionSpellReward) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("exposition_id = ?", expositionID).Delete(&models.ExpositionSpellReward{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, reward := range rewards {
			reward.ID = uuid.New()
			reward.ExpositionID = expositionID
			reward.CreatedAt = now
			reward.UpdatedAt = now
			if err := tx.Create(&reward).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *expositionHandle) UpsertCompletion(ctx context.Context, userID uuid.UUID, expositionID uuid.UUID) error {
	now := time.Now()
	record := models.UserExpositionCompletion{
		ID:           uuid.New(),
		CreatedAt:    now,
		UpdatedAt:    now,
		UserID:       userID,
		ExpositionID: expositionID,
	}
	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "exposition_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"updated_at": now}),
		}).
		Create(&record).Error
}

func (h *expositionHandle) FindCompletionByUserAndExposition(ctx context.Context, userID uuid.UUID, expositionID uuid.UUID) (*models.UserExpositionCompletion, error) {
	var completion models.UserExpositionCompletion
	if err := h.db.WithContext(ctx).
		Where("user_id = ? AND exposition_id = ?", userID, expositionID).
		First(&completion).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &completion, nil
}

func (h *expositionHandle) FindCompletedExpositionIDsByUser(ctx context.Context, userID uuid.UUID, expositionIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(expositionIDs) == 0 {
		return nil, nil
	}
	var completedIDs []uuid.UUID
	if err := h.db.WithContext(ctx).
		Model(&models.UserExpositionCompletion{}).
		Where("user_id = ?", userID).
		Where("exposition_id IN ?", expositionIDs).
		Pluck("exposition_id", &completedIDs).Error; err != nil {
		return nil, err
	}
	return completedIDs, nil
}

func (h *expositionHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Exposition{}, "id = ?", id).Error
}
