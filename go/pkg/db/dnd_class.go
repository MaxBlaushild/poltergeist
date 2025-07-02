package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type dndClassHandler struct {
	db *gorm.DB
}

func (h *dndClassHandler) GetAll(ctx context.Context) ([]models.DndClass, error) {
	var classes []models.DndClass
	err := h.db.WithContext(ctx).Where("active = ?", true).Find(&classes).Error
	return classes, err
}

func (h *dndClassHandler) GetByID(ctx context.Context, id uuid.UUID) (*models.DndClass, error) {
	var class models.DndClass
	err := h.db.WithContext(ctx).Where("id = ? AND active = ?", id, true).First(&class).Error
	if err != nil {
		return nil, err
	}
	return &class, nil
}

func (h *dndClassHandler) GetByName(ctx context.Context, name string) (*models.DndClass, error) {
	var class models.DndClass
	err := h.db.WithContext(ctx).Where("name = ? AND active = ?", name, true).First(&class).Error
	if err != nil {
		return nil, err
	}
	return &class, nil
}

func (h *dndClassHandler) Create(ctx context.Context, class *models.DndClass) error {
	return h.db.WithContext(ctx).Create(class).Error
}

func (h *dndClassHandler) Update(ctx context.Context, class *models.DndClass) error {
	return h.db.WithContext(ctx).Save(class).Error
}

func (h *dndClassHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.DndClass{}).Where("id = ?", id).Update("active", false).Error
}