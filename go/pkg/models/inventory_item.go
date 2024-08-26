package models

import "time"

type InventoryItem struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Name       string    `json:"name"`
	ImageURL   string    `json:"imageUrl"`
	FlavorText string    `json:"flavorText"`
	EffectText string    `json:"effectText"`
}
