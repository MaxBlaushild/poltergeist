package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	ReefProductKindConfigurable = "configurable"
	ReefProductKindFixed        = "fixed"
)

type ReefProduct struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	Slug           string         `json:"slug" gorm:"uniqueIndex"`
	Name           string         `json:"name"`
	Kind           string         `json:"kind"`
	Description    string         `json:"description"`
	Material       string         `json:"material"`
	BasePriceCents int64          `json:"basePriceCents" gorm:"column:base_price_cents"`
	Images         datatypes.JSON `json:"images"`
	Active         bool           `json:"active"`
}

func (ReefProduct) TableName() string {
	return "reef_products"
}

type ReefProductVariant struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	ProductID  uuid.UUID `json:"productId" gorm:"type:uuid;index"`
	VariantKey string    `json:"variantKey" gorm:"column:variant_key"`
	Label      string    `json:"label"`
	PriceCents int64     `json:"priceCents" gorm:"column:price_cents"`
	Active     bool      `json:"active"`
}

func (ReefProductVariant) TableName() string {
	return "reef_product_variants"
}
