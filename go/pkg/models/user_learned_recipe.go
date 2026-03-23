package models

import (
	"time"

	"github.com/google/uuid"
)

type UserLearnedRecipe struct {
	ID                         uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt                  time.Time  `json:"createdAt"`
	UpdatedAt                  time.Time  `json:"updatedAt"`
	UserID                     uuid.UUID  `json:"userId" gorm:"column:user_id;type:uuid"`
	RecipeID                   string     `json:"recipeId" gorm:"column:recipe_id"`
	LearnedFromInventoryItemID *int       `json:"learnedFromInventoryItemId,omitempty" gorm:"column:learned_from_inventory_item_id"`
	LearnedFromOwnedItemID     *uuid.UUID `json:"learnedFromOwnedItemId,omitempty" gorm:"column:learned_from_owned_item_id;type:uuid"`
}

func (UserLearnedRecipe) TableName() string {
	return "user_learned_recipes"
}
