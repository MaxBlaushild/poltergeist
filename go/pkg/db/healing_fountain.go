package db

import (
	"context"
	stdErrors "errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type healingFountainHandle struct {
	db *gorm.DB
}

func (h *healingFountainHandle) Create(ctx context.Context, fountain *models.HealingFountain) error {
	fountain.ID = uuid.New()
	fountain.CreatedAt = time.Now()
	fountain.UpdatedAt = time.Now()
	if err := fountain.SetGeometry(fountain.Latitude, fountain.Longitude); err != nil {
		return err
	}
	return h.db.WithContext(ctx).Create(fountain).Error
}

func (h *healingFountainHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.HealingFountain, error) {
	var fountain models.HealingFountain
	if err := h.db.WithContext(ctx).
		Preload("Zone").
		First(&fountain, id).Error; err != nil {
		return nil, err
	}
	return &fountain, nil
}

func (h *healingFountainHandle) FindAll(ctx context.Context) ([]models.HealingFountain, error) {
	var fountains []models.HealingFountain
	if err := h.db.WithContext(ctx).
		Preload("Zone").
		Find(&fountains).Error; err != nil {
		return nil, err
	}
	return fountains, nil
}

func (h *healingFountainHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]models.HealingFountain, error) {
	var fountains []models.HealingFountain
	if err := h.db.WithContext(ctx).
		Where("zone_id = ? AND invalidated = false", zoneID).
		Preload("Zone").
		Find(&fountains).Error; err != nil {
		return nil, err
	}
	return fountains, nil
}

func (h *healingFountainHandle) Update(ctx context.Context, id uuid.UUID, updates *models.HealingFountain) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	if updates.Latitude != 0 || updates.Longitude != 0 {
		if err := updates.SetGeometry(updates.Latitude, updates.Longitude); err != nil {
			return err
		}
	}

	payload := map[string]interface{}{
		"name":          updates.Name,
		"description":   updates.Description,
		"thumbnail_url": updates.ThumbnailURL,
		"zone_id":       updates.ZoneID,
		"latitude":      updates.Latitude,
		"longitude":     updates.Longitude,
		"geometry":      updates.Geometry,
		"invalidated":   updates.Invalidated,
		"updated_at":    updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).Model(&models.HealingFountain{}).Where("id = ?", id).Updates(payload).Error
}

func (h *healingFountainHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("healing_fountain_id = ?", id).Delete(&models.UserHealingFountainVisit{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.HealingFountain{}, "id = ?", id).Error
	})
}

func (h *healingFountainHandle) DeleteByIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("healing_fountain_id IN ?", ids).Delete(&models.UserHealingFountainVisit{}).Error; err != nil {
			return err
		}
		return tx.Where("id IN ?", ids).Delete(&models.HealingFountain{}).Error
	})
}

func (h *healingFountainHandle) FindLatestVisitByUserAndFountain(ctx context.Context, userID uuid.UUID, healingFountainID uuid.UUID) (*models.UserHealingFountainVisit, error) {
	var visit models.UserHealingFountainVisit
	err := h.db.WithContext(ctx).
		Where("user_id = ? AND healing_fountain_id = ?", userID, healingFountainID).
		Order("visited_at DESC").
		First(&visit).Error
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &visit, nil
}

func (h *healingFountainHandle) GetDiscoveriesForUser(ctx context.Context, userID uuid.UUID) ([]models.UserHealingFountainDiscovery, error) {
	var discoveries []models.UserHealingFountainDiscovery
	if err := h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&discoveries).Error; err != nil {
		return nil, err
	}
	return discoveries, nil
}

func (h *healingFountainHandle) FindDiscoveryByUserAndFountain(ctx context.Context, userID uuid.UUID, healingFountainID uuid.UUID) (*models.UserHealingFountainDiscovery, error) {
	var discovery models.UserHealingFountainDiscovery
	err := h.db.WithContext(ctx).
		Where("user_id = ? AND healing_fountain_id = ?", userID, healingFountainID).
		First(&discovery).Error
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &discovery, nil
}

func (h *healingFountainHandle) CreateUserHealingFountainDiscovery(ctx context.Context, discovery *models.UserHealingFountainDiscovery) error {
	discovery.ID = uuid.New()
	discovery.CreatedAt = time.Now()
	discovery.UpdatedAt = discovery.CreatedAt
	return h.db.WithContext(ctx).Create(discovery).Error
}

func (h *healingFountainHandle) FindLatestVisitsByUser(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]*models.UserHealingFountainVisit, error) {
	var visits []models.UserHealingFountainVisit
	if err := h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("visited_at DESC").
		Find(&visits).Error; err != nil {
		return nil, err
	}

	out := make(map[uuid.UUID]*models.UserHealingFountainVisit, len(visits))
	for i := range visits {
		visit := visits[i]
		if _, exists := out[visit.HealingFountainID]; exists {
			continue
		}
		v := visit
		out[visit.HealingFountainID] = &v
	}
	return out, nil
}

func (h *healingFountainHandle) CreateUserHealingFountainVisit(ctx context.Context, visit *models.UserHealingFountainVisit) error {
	visit.ID = uuid.New()
	visit.CreatedAt = time.Now()
	visit.UpdatedAt = visit.CreatedAt
	if visit.VisitedAt.IsZero() {
		visit.VisitedAt = visit.CreatedAt
	}
	return h.db.WithContext(ctx).Create(visit).Error
}
