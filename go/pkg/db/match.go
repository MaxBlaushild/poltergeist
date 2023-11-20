package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type matchHandle struct {
	db *gorm.DB
}

func (m *matchHandle) Insert(ctx context.Context, match *models.Match) error {
	return m.db.WithContext(ctx).Create(match).Error
}

func (m *matchHandle) GetCurrentMatchForUser(ctx context.Context, userID uuid.UUID) (*models.Match, error) {
	match := models.Match{}

	if err := m.db.
		Preload("QuestionSet.Questions.Category").
		Preload(clause.Associations).
		Order("created_at DESC").
		First(&match, "home_id = ? OR away_id = ?", userID, userID).
		Error; err != nil {
		return nil, err
	}

	return &match, nil
}
