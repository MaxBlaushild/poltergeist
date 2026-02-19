package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type NewUserStarterItem struct {
	InventoryItemID int `json:"inventoryItemId"`
	Quantity        int `json:"quantity"`
}

type NewUserStarterConfig struct {
	ID        int            `gorm:"primaryKey" json:"id"`
	Gold      int            `json:"gold"`
	ItemsJSON datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"-"`
	Items     []NewUserStarterItem `gorm:"-" json:"items"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

func (NewUserStarterConfig) TableName() string {
	return "new_user_starter_configs"
}

func (c *NewUserStarterConfig) BeforeSave(tx *gorm.DB) (err error) {
	if c.Items == nil {
		c.Items = []NewUserStarterItem{}
	}
	raw, err := json.Marshal(c.Items)
	if err != nil {
		return err
	}
	c.ItemsJSON = datatypes.JSON(raw)
	return nil
}

func (c *NewUserStarterConfig) AfterFind(tx *gorm.DB) (err error) {
	if len(c.ItemsJSON) == 0 {
		c.Items = []NewUserStarterItem{}
		return nil
	}
	var items []NewUserStarterItem
	if err := json.Unmarshal(c.ItemsJSON, &items); err != nil {
		return err
	}
	c.Items = items
	return nil
}

type NewUserStarterGrant struct {
	UserID    uuid.UUID `gorm:"primaryKey" json:"userId"`
	AppliedAt time.Time `json:"appliedAt"`
}

func (NewUserStarterGrant) TableName() string {
	return "new_user_starter_grants"
}
