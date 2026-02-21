package db

import (
	"context"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userProficiencyHandle struct {
	db *gorm.DB
}

func (h *userProficiencyHandle) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserProficiency, error) {
	var rows []models.UserProficiency
	if err := h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("level DESC, proficiency ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (h *userProficiencyHandle) Increment(ctx context.Context, userID uuid.UUID, proficiency string, delta int) error {
	if delta == 0 {
		return nil
	}
	value := strings.TrimSpace(proficiency)
	if value == "" {
		return nil
	}
	now := time.Now()
	record := models.UserProficiency{
		ID:          uuid.New(),
		CreatedAt:   now,
		UpdatedAt:   now,
		UserID:      userID,
		Proficiency: value,
		Level:       delta,
	}
	return h.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "proficiency"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"level":      gorm.Expr("user_proficiencies.level + ?", delta),
			"updated_at": now,
		}),
	}).Create(&record).Error
}

func (h *userProficiencyHandle) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	return h.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserProficiency{}).Error
}
