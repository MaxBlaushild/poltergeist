package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type tradesARGlassesLeadHandle struct {
	db *gorm.DB
}

func (h *tradesARGlassesLeadHandle) CreateOrGetByEmail(ctx context.Context, lead *models.TradesARGlassesLead) (bool, error) {
	if lead.ID == uuid.Nil {
		lead.ID = uuid.New()
	}
	if lead.CreatedAt.IsZero() {
		lead.CreatedAt = time.Now()
	}
	lead.UpdatedAt = time.Now()

	result := h.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "email"}},
		DoNothing: true,
	}).Create(lead)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 1 {
		return true, nil
	}

	var existing models.TradesARGlassesLead
	if err := h.db.WithContext(ctx).First(&existing, "email = ?", lead.Email).Error; err != nil {
		return false, err
	}
	*lead = existing
	return false, nil
}

func (h *tradesARGlassesLeadHandle) ListRecent(ctx context.Context, limit int) ([]models.TradesARGlassesLead, error) {
	if limit <= 0 {
		limit = 100
	}
	var leads []models.TradesARGlassesLead
	if err := h.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Find(&leads).Error; err != nil {
		return nil, err
	}
	return leads, nil
}
