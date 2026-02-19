package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type newUserStarterConfigHandle struct {
	db *gorm.DB
}

func (h *newUserStarterConfigHandle) Get(ctx context.Context) (*models.NewUserStarterConfig, error) {
	return h.getOrCreate(ctx, h.db.WithContext(ctx))
}

func (h *newUserStarterConfigHandle) Upsert(ctx context.Context, config *models.NewUserStarterConfig) (*models.NewUserStarterConfig, error) {
	if config == nil {
		return nil, gorm.ErrInvalidData
	}
	config.ID = 1
	if config.Items == nil {
		config.Items = []models.NewUserStarterItem{}
	}
	if err := h.db.WithContext(ctx).Save(config).Error; err != nil {
		return nil, err
	}
	return config, nil
}

func (h *newUserStarterConfigHandle) ApplyToUser(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var grant models.NewUserStarterGrant
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&grant).Error; err == nil {
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		config, err := h.getOrCreate(ctx, tx)
		if err != nil {
			return err
		}

		if config.Gold > 0 {
			if err := tx.Model(&models.User{}).
				Where("id = ?", userID).
				UpdateColumn("gold", gorm.Expr("gold + ?", config.Gold)).Error; err != nil {
				return err
			}
		}

		for _, item := range config.Items {
			if item.InventoryItemID <= 0 || item.Quantity <= 0 {
				continue
			}
			var owned models.OwnedInventoryItem
			err := tx.Where("user_id = ? AND inventory_item_id = ?", userID, item.InventoryItemID).
				First(&owned).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					newItem := models.OwnedInventoryItem{
						ID:              uuid.New(),
						UserID:          &userID,
						InventoryItemID: item.InventoryItemID,
						Quantity:        item.Quantity,
					}
					if err := tx.Create(&newItem).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				owned.Quantity += item.Quantity
				if err := tx.Save(&owned).Error; err != nil {
					return err
				}
			}
		}

		return tx.Create(&models.NewUserStarterGrant{
			UserID:    userID,
			AppliedAt: time.Now(),
		}).Error
	})
}

func (h *newUserStarterConfigHandle) getOrCreate(ctx context.Context, db *gorm.DB) (*models.NewUserStarterConfig, error) {
	var config models.NewUserStarterConfig
	err := db.WithContext(ctx).First(&config, 1).Error
	if err == nil {
		return &config, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	config = models.NewUserStarterConfig{
		ID:    1,
		Gold:  0,
		Items: []models.NewUserStarterItem{},
	}
	if err := db.WithContext(ctx).Create(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}
