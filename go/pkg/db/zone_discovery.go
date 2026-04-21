package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type zoneDiscoveryHandle struct {
	db *gorm.DB
}

func (h *zoneDiscoveryHandle) GetDiscoveriesForUser(
	userID uuid.UUID,
) ([]models.ZoneDiscovery, error) {
	var discoveries []models.ZoneDiscovery
	if err := h.db.Where("user_id = ?", userID).Find(&discoveries).Error; err != nil {
		return nil, err
	}
	return discoveries, nil
}

func (h *zoneDiscoveryHandle) ExistsForUserAndZone(
	ctx context.Context,
	userID uuid.UUID,
	zoneID uuid.UUID,
) (bool, error) {
	var count int64
	if err := h.db.WithContext(ctx).
		Model(&models.ZoneDiscovery{}).
		Where("user_id = ? AND zone_id = ?", userID, zoneID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (h *zoneDiscoveryHandle) DeleteByUserID(
	ctx context.Context,
	userID uuid.UUID,
) error {
	return h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.ZoneDiscovery{}).Error
}

func (h *zoneDiscoveryHandle) Create(
	ctx context.Context,
	discovery *models.ZoneDiscovery,
) error {
	return h.db.WithContext(ctx).Create(discovery).Error
}
