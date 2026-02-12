package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type InsiderTrade struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt  time.Time      `gorm:"not null" json:"createdAt"`
	UpdatedAt  time.Time      `gorm:"not null" json:"updatedAt"`
	ExternalID string         `gorm:"type:text;uniqueIndex;not null" json:"externalId"`
	MarketID   string         `gorm:"type:text" json:"marketId"`
	MarketName string         `gorm:"type:text" json:"marketName"`
	Outcome    string         `gorm:"type:text" json:"outcome"`
	Side       string         `gorm:"type:text" json:"side"`
	Price      float64        `gorm:"type:numeric" json:"price"`
	Size       float64        `gorm:"type:numeric" json:"size"`
	Notional   float64        `gorm:"type:numeric" json:"notional"`
	Trader     string         `gorm:"type:text" json:"trader"`
	TradeTime  time.Time      `gorm:"type:timestamp" json:"tradeTime"`
	DetectedAt time.Time      `gorm:"type:timestamp" json:"detectedAt"`
	Reason     string         `gorm:"type:text" json:"reason"`
	Raw        datatypes.JSON `gorm:"type:jsonb" json:"raw"`
}
