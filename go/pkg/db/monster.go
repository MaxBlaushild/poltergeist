package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type monsterHandler struct {
	db *gorm.DB
}

func (h *monsterHandler) GetAll(ctx context.Context) ([]models.Monster, error) {
	var monsters []models.Monster
	err := h.db.WithContext(ctx).Where("active = ?", true).Find(&monsters).Error
	return monsters, err
}

func (h *monsterHandler) GetByID(ctx context.Context, id uuid.UUID) (*models.Monster, error) {
	var monster models.Monster
	err := h.db.WithContext(ctx).Where("id = ? AND active = ?", id, true).First(&monster).Error
	if err != nil {
		return nil, err
	}
	return &monster, nil
}

func (h *monsterHandler) GetByName(ctx context.Context, name string) (*models.Monster, error) {
	var monster models.Monster
	err := h.db.WithContext(ctx).Where("name = ? AND active = ?", name, true).First(&monster).Error
	if err != nil {
		return nil, err
	}
	return &monster, nil
}

func (h *monsterHandler) GetByChallengeRating(ctx context.Context, cr float64) ([]models.Monster, error) {
	var monsters []models.Monster
	err := h.db.WithContext(ctx).Where("challenge_rating = ? AND active = ?", cr, true).Find(&monsters).Error
	return monsters, err
}

func (h *monsterHandler) GetByType(ctx context.Context, monsterType string) ([]models.Monster, error) {
	var monsters []models.Monster
	err := h.db.WithContext(ctx).Where("type = ? AND active = ?", monsterType, true).Find(&monsters).Error
	return monsters, err
}

func (h *monsterHandler) GetBySize(ctx context.Context, size string) ([]models.Monster, error) {
	var monsters []models.Monster
	err := h.db.WithContext(ctx).Where("size = ? AND active = ?", size, true).Find(&monsters).Error
	return monsters, err
}

func (h *monsterHandler) Create(ctx context.Context, monster *models.Monster) error {
	return h.db.WithContext(ctx).Create(monster).Error
}

func (h *monsterHandler) Update(ctx context.Context, monster *models.Monster) error {
	return h.db.WithContext(ctx).Save(monster).Error
}

func (h *monsterHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.Monster{}).Where("id = ?", id).Update("active", false).Error
}

func (h *monsterHandler) Search(ctx context.Context, query string) ([]models.Monster, error) {
	var monsters []models.Monster
	searchTerm := "%" + query + "%"
	err := h.db.WithContext(ctx).Where(
		"active = ? AND (name ILIKE ? OR type ILIKE ? OR description ILIKE ?)",
		true, searchTerm, searchTerm, searchTerm,
	).Find(&monsters).Error
	return monsters, err
}