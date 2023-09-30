package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"gorm.io/gorm"
)

type scoreHandler struct {
	db *gorm.DB
}

func (h *scoreHandler) Upsert(ctx context.Context, username string) (*models.Score, error) {
	score := models.Score{}
	err := h.db.WithContext(ctx).Where(&models.Score{Username: username}).First(&score).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if err != nil && err == gorm.ErrRecordNotFound {
		if err := h.db.WithContext(ctx).Create(&models.Score{
			Username: username,
			Score:    1,
		}).Error; err != nil {
			return nil, err
		}

		score.Score++

		return &score, nil
	}

	if err := h.db.WithContext(ctx).Where("id = ?", score.ID).Updates(models.Score{
		Score: score.Score + 1,
	}).Error; err != nil {
		return nil, err
	}

	score.Score++

	return &score, nil
}

func (h *scoreHandler) FindAll(ctx context.Context) ([]models.Score, error) {
	scores := []models.Score{}
	if err := h.db.WithContext(ctx).Find(&scores).Error; err != nil {
		return nil, err
	}

	return scores, nil
}
