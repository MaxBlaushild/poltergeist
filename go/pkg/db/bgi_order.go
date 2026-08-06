package db

import (
	"context"
	"errors"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type bgiOrderHandle struct {
	db *gorm.DB
}

// Create persists an order and its line items in one transaction — mirrors
// reefOrderHandle.Create exactly.
func (h *bgiOrderHandle) Create(ctx context.Context, order *models.BgiOrder) (*models.BgiOrder, error) {
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

func (h *bgiOrderHandle) FindByToken(ctx context.Context, token string) (*models.BgiOrder, error) {
	var order models.BgiOrder
	if err := h.db.WithContext(ctx).Preload("Items").Where("order_token = ?", token).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (h *bgiOrderHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.BgiOrder, error) {
	var order models.BgiOrder
	if err := h.db.WithContext(ctx).Preload("Items").Where("id = ?", id).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (h *bgiOrderHandle) FindByStripeSessionID(ctx context.Context, sessionID string) (*models.BgiOrder, error) {
	var order models.BgiOrder
	err := h.db.WithContext(ctx).Preload("Items").Where("stripe_session_id = ?", sessionID).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (h *bgiOrderHandle) Update(ctx context.Context, order *models.BgiOrder) error {
	return h.db.WithContext(ctx).Omit("Items").Save(order).Error
}
