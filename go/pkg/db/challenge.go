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

type challengeHandle struct {
	db *gorm.DB
}

func (h *challengeHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("Zone").
		Preload("PointOfInterest").
		Preload("ItemChoiceRewards").
		Preload("ItemChoiceRewards.InventoryItem")
}

func (h *challengeHandle) visibleQuery(ctx context.Context) *gorm.DB {
	return h.preloadBase(ctx).Where("retired_at IS NULL")
}

func (h *challengeHandle) Create(ctx context.Context, challenge *models.Challenge) error {
	if challenge != nil {
		if strings.TrimSpace(string(challenge.RewardMode)) == "" {
			if challenge.Reward > 0 || challenge.RewardExperience > 0 || challenge.InventoryItemID != nil {
				challenge.RewardMode = models.RewardModeExplicit
			} else {
				challenge.RewardMode = models.RewardModeRandom
			}
		}
		challenge.RewardMode = models.NormalizeRewardMode(string(challenge.RewardMode))
		challenge.RandomRewardSize = models.NormalizeRandomRewardSize(string(challenge.RandomRewardSize))
		if challenge.RewardExperience < 0 {
			challenge.RewardExperience = 0
		}
	}
	return h.db.WithContext(ctx).Create(challenge).Error
}

func (h *challengeHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Challenge, error) {
	var challenge models.Challenge
	if err := h.preloadBase(ctx).First(&challenge, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (h *challengeHandle) FindAll(ctx context.Context) ([]models.Challenge, error) {
	var challenges []models.Challenge
	if err := h.visibleQuery(ctx).Find(&challenges).Error; err != nil {
		return nil, err
	}
	return challenges, nil
}

func (h *challengeHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]models.Challenge, error) {
	var challenges []models.Challenge
	if err := h.visibleQuery(ctx).
		Where("zone_id = ?", zoneID).
		Find(&challenges).Error; err != nil {
		return nil, err
	}
	return challenges, nil
}

func (h *challengeHandle) FindByZoneIDExcludingQuestNodes(ctx context.Context, zoneID uuid.UUID) ([]models.Challenge, error) {
	var challenges []models.Challenge
	if err := h.visibleQuery(ctx).
		Where("zone_id = ?", zoneID).
		Where("NOT EXISTS (SELECT 1 FROM quest_nodes qn WHERE qn.challenge_id = challenges.id)").
		Find(&challenges).Error; err != nil {
		return nil, err
	}
	return challenges, nil
}

func (h *challengeHandle) IsLinkedToQuestNode(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := h.db.WithContext(ctx).
		Table("quest_nodes").
		Where("challenge_id = ?", id).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (h *challengeHandle) Update(ctx context.Context, id uuid.UUID, updates *models.Challenge) error {
	if updates == nil {
		return nil
	}
	if err := updates.SyncLocationGeometry(); err != nil {
		return err
	}
	updates.RewardMode = models.NormalizeRewardMode(string(updates.RewardMode))
	updates.RandomRewardSize = models.NormalizeRandomRewardSize(string(updates.RandomRewardSize))
	if updates.RewardExperience < 0 {
		updates.RewardExperience = 0
	}
	payload := map[string]interface{}{
		"zone_id":                updates.ZoneID,
		"point_of_interest_id":   updates.PointOfInterestID,
		"latitude":               updates.Latitude,
		"longitude":              updates.Longitude,
		"geometry":               updates.Geometry,
		"polygon":                updates.Polygon,
		"question":               updates.Question,
		"description":            updates.Description,
		"image_url":              updates.ImageURL,
		"thumbnail_url":          updates.ThumbnailURL,
		"scale_with_user_level":  updates.ScaleWithUserLevel,
		"recurring_challenge_id": updates.RecurringChallengeID,
		"recurrence_frequency":   updates.RecurrenceFrequency,
		"next_recurrence_at":     updates.NextRecurrenceAt,
		"retired_at":             updates.RetiredAt,
		"reward_mode":            updates.RewardMode,
		"random_reward_size":     updates.RandomRewardSize,
		"reward_experience":      updates.RewardExperience,
		"reward":                 updates.Reward,
		"inventory_item_id":      updates.InventoryItemID,
		"submission_type":        updates.SubmissionType,
		"difficulty":             updates.Difficulty,
		"stat_tags":              updates.StatTags,
		"proficiency":            updates.Proficiency,
		"updated_at":             updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).Model(&models.Challenge{}).Where("id = ?", id).Updates(payload).Error
}

func (h *challengeHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Challenge{}, "id = ?", id).Error
}

func (h *challengeHandle) FindDueRecurring(ctx context.Context, asOf time.Time, limit int) ([]models.Challenge, error) {
	var challenges []models.Challenge
	query := h.db.WithContext(ctx).
		Where("retired_at IS NULL").
		Where("recurrence_frequency IS NOT NULL AND recurrence_frequency <> ''").
		Where("next_recurrence_at IS NOT NULL AND next_recurrence_at <= ?", asOf).
		Order("next_recurrence_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&challenges).Error; err != nil {
		return nil, err
	}
	return challenges, nil
}

func (h *challengeHandle) ReplaceItemChoiceRewards(ctx context.Context, challengeID uuid.UUID, rewards []models.ChallengeItemChoiceReward) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("challenge_id = ?", challengeID).Delete(&models.ChallengeItemChoiceReward{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, reward := range rewards {
			reward.ID = uuid.New()
			reward.ChallengeID = challengeID
			reward.CreatedAt = now
			reward.UpdatedAt = now
			if err := tx.Create(&reward).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *challengeHandle) UpsertCompletion(ctx context.Context, userID uuid.UUID, challengeID uuid.UUID) error {
	now := time.Now()
	record := models.UserChallengeCompletion{
		ID:          uuid.New(),
		CreatedAt:   now,
		UpdatedAt:   now,
		UserID:      userID,
		ChallengeID: challengeID,
	}
	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "challenge_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"updated_at": now}),
		}).
		Create(&record).Error
}

func (h *challengeHandle) FindCompletionByUserAndChallenge(ctx context.Context, userID uuid.UUID, challengeID uuid.UUID) (*models.UserChallengeCompletion, error) {
	var completion models.UserChallengeCompletion
	if err := h.db.WithContext(ctx).
		Where("user_id = ? AND challenge_id = ?", userID, challengeID).
		First(&completion).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &completion, nil
}

func (h *challengeHandle) FindCompletedChallengeIDsByUser(ctx context.Context, userID uuid.UUID, challengeIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(challengeIDs) == 0 {
		return nil, nil
	}
	var completedIDs []uuid.UUID
	if err := h.db.WithContext(ctx).
		Model(&models.UserChallengeCompletion{}).
		Where("user_id = ?", userID).
		Where("challenge_id IN ?", challengeIDs).
		Pluck("challenge_id", &completedIDs).Error; err != nil {
		return nil, err
	}
	return completedIDs, nil
}

func (h *challengeHandle) UpsertItemChoicePending(ctx context.Context, userID uuid.UUID, challengeID uuid.UUID) error {
	now := time.Now()
	record := models.UserChallengeItemChoicePending{
		ID:          uuid.New(),
		CreatedAt:   now,
		UpdatedAt:   now,
		UserID:      userID,
		ChallengeID: challengeID,
	}
	return h.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "challenge_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"updated_at": now}),
		}).
		Create(&record).Error
}

func (h *challengeHandle) FindItemChoicePendingByUserAndChallenge(ctx context.Context, userID uuid.UUID, challengeID uuid.UUID) (*models.UserChallengeItemChoicePending, error) {
	var pending models.UserChallengeItemChoicePending
	if err := h.db.WithContext(ctx).
		Where("user_id = ? AND challenge_id = ?", userID, challengeID).
		First(&pending).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &pending, nil
}

func (h *challengeHandle) DeleteItemChoicePending(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.UserChallengeItemChoicePending{}, "id = ?", id).Error
}
