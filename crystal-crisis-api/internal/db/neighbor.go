package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/crystal-crisis-api/internal/models"
	"gorm.io/gorm"
)

type neighborHandle struct {
	db *gorm.DB
}

func (h *neighborHandle) Create(ctx context.Context, crystalOneID uint, crystalTwoID uint) error {
	return h.db.WithContext(ctx).Create(&models.Neighbor{
		CrystalOneID: crystalOneID,
		CrystalTwoID: crystalTwoID,
	}).Error
}

func (h *neighborHandle) FindAll(ctx context.Context) ([]models.Neighbor, error) {
	var neighbors []models.Neighbor

	if err := h.db.WithContext(ctx).Find(&neighbors).Error; err != nil {
		return nil, err
	}

	return neighbors, nil
}
