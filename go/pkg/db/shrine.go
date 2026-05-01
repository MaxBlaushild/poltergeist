package db

import (
	"context"
	stdErrors "errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type shrineHandle struct {
	db *gorm.DB
}

func (h *shrineHandle) Create(ctx context.Context, shrine *models.Shrine) error {
	shrine.ID = uuid.New()
	shrine.CreatedAt = time.Now()
	shrine.UpdatedAt = time.Now()
	shrine.ZoneKind = models.NormalizeZoneKind(shrine.ZoneKind)
	if err := shrine.SetGeometry(shrine.Latitude, shrine.Longitude); err != nil {
		return err
	}
	return h.db.WithContext(ctx).Create(shrine).Error
}

func (h *shrineHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Shrine, error) {
	var shrine models.Shrine
	if err := h.db.WithContext(ctx).
		Preload("Zone").
		Preload("Template").
		First(&shrine, id).Error; err != nil {
		return nil, err
	}
	return &shrine, nil
}

func (h *shrineHandle) FindAll(ctx context.Context) ([]models.Shrine, error) {
	var shrines []models.Shrine
	if err := h.db.WithContext(ctx).
		Preload("Zone").
		Preload("Template").
		Find(&shrines).Error; err != nil {
		return nil, err
	}
	return shrines, nil
}

func (h *shrineHandle) FindByZoneID(ctx context.Context, zoneID uuid.UUID) ([]models.Shrine, error) {
	var shrines []models.Shrine
	if err := h.db.WithContext(ctx).
		Where("zone_id = ? AND invalidated = false", zoneID).
		Preload("Zone").
		Preload("Template").
		Find(&shrines).Error; err != nil {
		return nil, err
	}
	return shrines, nil
}

func (h *shrineHandle) Update(ctx context.Context, id uuid.UUID, updates *models.Shrine) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	updates.ZoneKind = models.NormalizeZoneKind(updates.ZoneKind)
	if updates.Latitude != 0 || updates.Longitude != 0 {
		if err := updates.SetGeometry(updates.Latitude, updates.Longitude); err != nil {
			return err
		}
	}

	payload := map[string]interface{}{
		"shrine_template_id": updates.ShrineTemplateID,
		"zone_id":            updates.ZoneID,
		"zone_kind":          updates.ZoneKind,
		"latitude":           updates.Latitude,
		"longitude":          updates.Longitude,
		"geometry":           updates.Geometry,
		"cooldown_seconds":   updates.CooldownSeconds,
		"invalidated":        updates.Invalidated,
		"updated_at":         updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).Model(&models.Shrine{}).Where("id = ?", id).Updates(payload).Error
}

func (h *shrineHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return deleteShrines(tx, []uuid.UUID{id})
	})
}

func deleteShrines(tx *gorm.DB, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	if err := tx.Where("shrine_id IN ?", ids).Delete(&models.UserShrineUse{}).Error; err != nil {
		return err
	}
	return tx.Where("id IN ?", ids).Delete(&models.Shrine{}).Error
}

func (h *shrineHandle) DeleteByIDs(ctx context.Context, ids []uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return deleteShrines(tx, ids)
	})
}

func (h *shrineHandle) FindLatestUseByUserAndShrine(ctx context.Context, userID uuid.UUID, shrineID uuid.UUID) (*models.UserShrineUse, error) {
	var use models.UserShrineUse
	err := h.db.WithContext(ctx).
		Where("user_id = ? AND shrine_id = ?", userID, shrineID).
		Order("used_at DESC").
		First(&use).Error
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &use, nil
}

func (h *shrineHandle) FindLatestUsesByUser(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]*models.UserShrineUse, error) {
	var uses []models.UserShrineUse
	if err := h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("used_at DESC").
		Find(&uses).Error; err != nil {
		return nil, err
	}

	out := make(map[uuid.UUID]*models.UserShrineUse, len(uses))
	for i := range uses {
		use := uses[i]
		if _, exists := out[use.ShrineID]; exists {
			continue
		}
		next := use
		out[use.ShrineID] = &next
	}
	return out, nil
}

func (h *shrineHandle) CreateUserShrineUse(ctx context.Context, use *models.UserShrineUse) error {
	use.ID = uuid.New()
	use.CreatedAt = time.Now()
	use.UpdatedAt = use.CreatedAt
	if use.UsedAt.IsZero() {
		use.UsedAt = use.CreatedAt
	}
	return h.db.WithContext(ctx).Create(use).Error
}
