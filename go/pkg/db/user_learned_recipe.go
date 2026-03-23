package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userLearnedRecipeHandle struct {
	db *gorm.DB
}

func (h *userLearnedRecipeHandle) FindByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.UserLearnedRecipe, error) {
	var recipes []models.UserLearnedRecipe
	if err := h.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&recipes).Error; err != nil {
		return nil, err
	}
	return recipes, nil
}

func (h *userLearnedRecipeHandle) Upsert(
	ctx context.Context,
	recipe *models.UserLearnedRecipe,
) error {
	if recipe == nil {
		return nil
	}
	return h.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "recipe_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at", "learned_from_inventory_item_id", "learned_from_owned_item_id"}),
	}).Create(recipe).Error
}
