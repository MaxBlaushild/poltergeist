package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type challengeHandle struct {
	db *gorm.DB
}

func (h *challengeHandle) preloadBase(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).Preload("Zone")
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

func (h *challengeHandle) Update(ctx context.Context, id uuid.UUID, updates *models.Challenge) error {
	if updates == nil {
		return nil
	}
	if err := updates.SetGeometry(updates.Latitude, updates.Longitude); err != nil {
		return err
	}
	payload := map[string]interface{}{
		"zone_id":           updates.ZoneID,
		"latitude":          updates.Latitude,
		"longitude":         updates.Longitude,
		"geometry":          updates.Geometry,
		"question":          updates.Question,
		"image_url":         updates.ImageURL,
		"thumbnail_url":     updates.ThumbnailURL,
		"reward":            updates.Reward,
		"inventory_item_id": updates.InventoryItemID,
		"submission_type":   updates.SubmissionType,
		"difficulty":        updates.Difficulty,
		"stat_tags":         updates.StatTags,
		"proficiency":       updates.Proficiency,
		"updated_at":        updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).Model(&models.Challenge{}).Where("id = ?", id).Updates(payload).Error
}

func (h *challengeHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Challenge{}, "id = ?", id).Error
}
