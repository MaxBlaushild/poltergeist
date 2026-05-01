package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type shrineTemplateHandle struct {
	db *gorm.DB
}

func (h *shrineTemplateHandle) Create(ctx context.Context, template *models.ShrineTemplate) error {
	if template == nil {
		return nil
	}
	template.ID = uuid.New()
	template.CreatedAt = time.Now()
	template.UpdatedAt = template.CreatedAt
	template.ZoneKind = models.NormalizeZoneKind(template.ZoneKind)
	template.EffectKind = models.NormalizeShrineEffectKind(string(template.EffectKind))
	return h.db.WithContext(ctx).Create(template).Error
}

func (h *shrineTemplateHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ShrineTemplate, error) {
	var template models.ShrineTemplate
	if err := h.db.WithContext(ctx).First(&template, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

func (h *shrineTemplateHandle) FindAll(ctx context.Context) ([]models.ShrineTemplate, error) {
	var templates []models.ShrineTemplate
	if err := h.db.WithContext(ctx).Order("created_at DESC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (h *shrineTemplateHandle) FindByZoneKind(ctx context.Context, zoneKind string) ([]models.ShrineTemplate, error) {
	var templates []models.ShrineTemplate
	query := h.db.WithContext(ctx).Order("created_at DESC")
	if normalized := models.NormalizeZoneKind(zoneKind); normalized != "" {
		query = query.Where("zone_kind = ?", normalized)
	}
	if err := query.Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (h *shrineTemplateHandle) Update(ctx context.Context, id uuid.UUID, updates *models.ShrineTemplate) error {
	if updates == nil {
		return nil
	}
	updates.UpdatedAt = time.Now()
	updates.ZoneKind = models.NormalizeZoneKind(updates.ZoneKind)
	updates.EffectKind = models.NormalizeShrineEffectKind(string(updates.EffectKind))
	payload := map[string]interface{}{
		"zone_kind":          updates.ZoneKind,
		"name":               updates.Name,
		"description":        updates.Description,
		"blessing_name":      updates.BlessingName,
		"effect_description": updates.EffectDescription,
		"effect_kind":        updates.EffectKind,
		"base_magnitude":     updates.BaseMagnitude,
		"updated_at":         updates.UpdatedAt,
	}
	return h.db.WithContext(ctx).Model(&models.ShrineTemplate{}).Where("id = ?", id).Updates(payload).Error
}

func (h *shrineTemplateHandle) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.ShrineTemplate{}, "id = ?", id).Error
}
