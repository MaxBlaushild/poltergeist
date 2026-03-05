package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type challengeHandle struct {
	db *gorm.DB
}

func (h *challengeHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).Preload("Zone").Preload("PointOfInterest")
}

func (h *challengeHandle) Create(ctx context.Context, challenge *models.Challenge) error {
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
	if err := h.preloadBase(ctx).Find(&challenges).Error; err != nil {
		return nil, err
	}
	return challenges, nil
}

func (h *challengeHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]models.Challenge, error) {
	var challenges []models.Challenge
	if err := h.preloadBase(ctx).
		Where("zone_id = ?", zoneID).
		Find(&challenges).Error; err != nil {
		return nil, err
	}
	return challenges, nil
}

func (h *challengeHandle) FindByZoneIDExcludingQuestNodes(ctx context.Context, zoneID uuid.UUID) ([]models.Challenge, error) {
	var challenges []models.Challenge
	if err := h.preloadBase(ctx).
		Where("zone_id = ?", zoneID).
		Where("NOT EXISTS (SELECT 1 FROM quest_nodes qn WHERE qn.challenge_id = challenges.id)").
		Find(&challenges).Error; err != nil {
		return nil, err
	}
	return challenges, nil
}

func (h *challengeHandle) Update(ctx context.Context, id uuid.UUID, updates *models.Challenge) error {
	if updates == nil {
		return nil
	}
	if err := updates.SetGeometry(updates.Latitude, updates.Longitude); err != nil {
		return err
	}
	payload := map[string]interface{}{
		"zone_id":                updates.ZoneID,
		"point_of_interest_id":   updates.PointOfInterestID,
		"latitude":               updates.Latitude,
		"longitude":              updates.Longitude,
		"geometry":               updates.Geometry,
		"question":               updates.Question,
		"description":            updates.Description,
		"image_url":              updates.ImageURL,
		"thumbnail_url":          updates.ThumbnailURL,
		"scale_with_user_level":  updates.ScaleWithUserLevel,
		"recurring_challenge_id": updates.RecurringChallengeID,
		"recurrence_frequency":   updates.RecurrenceFrequency,
		"next_recurrence_at":     updates.NextRecurrenceAt,
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
