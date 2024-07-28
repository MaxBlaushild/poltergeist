package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type matchHandle struct {
	db *gorm.DB
}

func (h *matchHandle) Create(ctx context.Context, match models.Match) error {
	return h.db.WithContext(ctx).Create(&match).Error
}
