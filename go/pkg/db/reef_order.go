package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type reefOrderHandle struct {
	db *gorm.DB
}

// Create persists an order and its line items in one transaction so a
// partially-written order (order row with no items) can never be observed.
func (h *reefOrderHandle) Create(ctx context.Context, order *models.ReefOrder) (*models.ReefOrder, error) {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		items := order.Items
		order.Items = nil
		if err := tx.Omit("Items").Create(order).Error; err != nil {
			return err
		}
		for i := range items {
			if items[i].ID == uuid.Nil {
				items[i].ID = uuid.New()
			}
			items[i].OrderID = order.ID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		order.Items = items
		return nil
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (h *reefOrderHandle) FindByToken(ctx context.Context, token string) (*models.ReefOrder, error) {
	var order models.ReefOrder
	if err := h.db.WithContext(ctx).Preload("Items").Where("order_token = ?", token).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (h *reefOrderHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.ReefOrder, error) {
	var order models.ReefOrder
	if err := h.db.WithContext(ctx).Preload("Items").Where("id = ?", id).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (h *reefOrderHandle) FindByStripeSessionID(ctx context.Context, sessionID string) (*models.ReefOrder, error) {
	var order models.ReefOrder
	err := h.db.WithContext(ctx).Preload("Items").Where("stripe_session_id = ?", sessionID).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (h *reefOrderHandle) Update(ctx context.Context, order *models.ReefOrder) error {
	return h.db.WithContext(ctx).Omit("Items").Save(order).Error
}

func (h *reefOrderHandle) FindPaid(ctx context.Context) ([]models.ReefOrder, error) {
	var orders []models.ReefOrder
	if err := h.db.WithContext(ctx).
		Where("status IN ?", []string{models.ReefOrderStatusPaid, models.ReefOrderStatusFulfilled}).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// OperatorCogsStats is the raw material for R-9.2's "mean landed COGS per
// order" and "reprint rate" — averaged only over orders with a real recorded
// COGS (R-7.4: estimated COGS is not acceptable, so unset rows are excluded
// rather than treated as zero).
type OperatorCogsStats struct {
	OrderCount      int64
	CogsRecorded    int64
	MeanCogsCents   float64
	OrdersReprinted int64
}

func (h *reefOrderHandle) CogsStatsSince(ctx context.Context, since time.Time) (*OperatorCogsStats, error) {
	stats := &OperatorCogsStats{}
	if err := h.db.WithContext(ctx).Model(&models.ReefOrder{}).
		Where("status IN ? AND created_at >= ?", []string{models.ReefOrderStatusPaid, models.ReefOrderStatusFulfilled}, since).
		Count(&stats.OrderCount).Error; err != nil {
		return nil, err
	}
	row := struct {
		CogsRecorded  int64
		MeanCogsCents float64
	}{}
	if err := h.db.WithContext(ctx).Model(&models.ReefOrder{}).
		Select("count(cogs_cents) as cogs_recorded, coalesce(avg(cogs_cents), 0) as mean_cogs_cents").
		Where("status IN ? AND created_at >= ?", []string{models.ReefOrderStatusPaid, models.ReefOrderStatusFulfilled}, since).
		Scan(&row).Error; err != nil {
		return nil, err
	}
	stats.CogsRecorded = row.CogsRecorded
	stats.MeanCogsCents = row.MeanCogsCents
	if err := h.db.WithContext(ctx).Model(&models.ReefOrder{}).
		Where("status IN ? AND created_at >= ? AND reprint_count > 0", []string{models.ReefOrderStatusPaid, models.ReefOrderStatusFulfilled}, since).
		Count(&stats.OrdersReprinted).Error; err != nil {
		return nil, err
	}
	return stats, nil
}
