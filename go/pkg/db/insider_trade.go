package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type insiderTradeHandle struct {
	db *gorm.DB
}

func (h *insiderTradeHandle) Upsert(ctx context.Context, trade *models.InsiderTrade) (bool, error) {
	if trade.ID == uuid.Nil {
		trade.ID = uuid.New()
	}
	if trade.CreatedAt.IsZero() {
		trade.CreatedAt = time.Now()
	}
	trade.UpdatedAt = time.Now()
	if trade.DetectedAt.IsZero() {
		trade.DetectedAt = time.Now()
	}

	result := h.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "external_id"}},
		DoNothing: true,
	}).Create(trade)
	return result.RowsAffected == 1, result.Error
}

func (h *insiderTradeHandle) List(ctx context.Context, limit, offset int) ([]models.InsiderTrade, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var trades []models.InsiderTrade
	if err := h.db.WithContext(ctx).
		Order("detected_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&trades).Error; err != nil {
		return nil, err
	}
	return trades, nil
}

func (h *insiderTradeHandle) LatestTradeTime(ctx context.Context) (*time.Time, error) {
	var latest sql.NullTime
	row := h.db.WithContext(ctx).Model(&models.InsiderTrade{}).Select("MAX(trade_time)").Row()
	if err := row.Scan(&latest); err != nil {
		return nil, err
	}
	if !latest.Valid {
		return nil, nil
	}
	return &latest.Time, nil
}
