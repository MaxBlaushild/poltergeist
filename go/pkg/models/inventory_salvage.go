package models

import (
	"database/sql/driver"
	"encoding/json"
)

type InventorySalvageOutput struct {
	ItemID   int `json:"itemId"`
	Quantity int `json:"quantity"`
}

type InventorySalvageRecipe struct {
	Tier    int                      `json:"tier"`
	Outputs []InventorySalvageOutput `json:"outputs"`
}

type InventorySalvageRecipes []InventorySalvageRecipe

func (r InventorySalvageRecipes) Value() (driver.Value, error) {
	if r == nil {
		return json.Marshal([]InventorySalvageRecipe{})
	}
	return json.Marshal(r)
}

func (r *InventorySalvageRecipes) Scan(value interface{}) error {
	if value == nil {
		*r = InventorySalvageRecipes{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*r = InventorySalvageRecipes{}
		return nil
	}

	return json.Unmarshal(bytes, r)
}
