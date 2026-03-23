package models

import (
	"database/sql/driver"
	"encoding/json"
)

type CraftingStation string

const (
	CraftingStationAlchemy  CraftingStation = "alchemy"
	CraftingStationWorkshop CraftingStation = "workshop"
)

func NormalizeCraftingStation(raw string) CraftingStation {
	switch CraftingStation(raw) {
	case CraftingStationAlchemy:
		return CraftingStationAlchemy
	case CraftingStationWorkshop:
		return CraftingStationWorkshop
	default:
		return ""
	}
}

type InventoryRecipeIngredient struct {
	ItemID   int `json:"itemId"`
	Quantity int `json:"quantity"`
}

type InventoryRecipe struct {
	ID          string                      `json:"id"`
	Tier        int                         `json:"tier"`
	IsPublic    bool                        `json:"isPublic"`
	Ingredients []InventoryRecipeIngredient `json:"ingredients"`
}

type InventoryRecipes []InventoryRecipe

func (r InventoryRecipes) Value() (driver.Value, error) {
	if r == nil {
		return json.Marshal([]InventoryRecipe{})
	}
	return json.Marshal(r)
}

func (r *InventoryRecipes) Scan(value interface{}) error {
	if value == nil {
		*r = InventoryRecipes{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*r = InventoryRecipes{}
		return nil
	}

	return json.Unmarshal(bytes, r)
}
