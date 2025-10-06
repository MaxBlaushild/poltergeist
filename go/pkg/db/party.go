package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type partyHandle struct {
	db *gorm.DB
}

func (h *partyHandle) Create(ctx context.Context) (*models.Party, error) {
	party := &models.Party{}

	if err := h.db.WithContext(ctx).Create(party).Error; err != nil {
		return nil, err
	}

	return party, nil
}
