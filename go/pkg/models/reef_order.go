package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	ReefOrderStatusPendingPayment = "pending_payment"
	ReefOrderStatusPaid           = "paid"
	ReefOrderStatusFulfilled      = "fulfilled"
	ReefOrderStatusCancelled      = "cancelled"

	ReefFulfillmentProviderManual = "manual"
	ReefFulfillmentProviderSlant  = "slant"
)

// ReefOrder is looked up with no login via order_token (R-8.2: /orders/[token]).
type ReefOrder struct {
	ID                    uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt             time.Time      `json:"createdAt"`
	UpdatedAt             time.Time      `json:"updatedAt"`
	OrderToken            string         `json:"orderToken" gorm:"column:order_token;uniqueIndex"`
	StripeSessionID       string         `json:"stripeSessionId" gorm:"column:stripe_session_id;index"`
	CustomerEmail         string         `json:"customerEmail" gorm:"column:customer_email"`
	ShippingAddress       datatypes.JSON `json:"shippingAddress" gorm:"column:shipping_address"`
	Status                string         `json:"status"`
	FulfillmentProvider   string         `json:"fulfillmentProvider" gorm:"column:fulfillment_provider"`
	FulfillmentStatus     string         `json:"fulfillmentStatus" gorm:"column:fulfillment_status"`
	FulfillmentExternalID string         `json:"fulfillmentExternalId" gorm:"column:fulfillment_external_id"`
	SubtotalCents         int64          `json:"subtotalCents" gorm:"column:subtotal_cents"`
	ShippingCents         int64          `json:"shippingCents" gorm:"column:shipping_cents"`
	TotalCents            int64          `json:"totalCents" gorm:"column:total_cents"`
	CogsCents             *int64         `json:"cogsCents" gorm:"column:cogs_cents"`
	ReprintCount          int            `json:"reprintCount" gorm:"column:reprint_count"`

	Items []ReefOrderItem `json:"items" gorm:"foreignKey:OrderID"`
}

func (ReefOrder) TableName() string {
	return "reef_orders"
}

type ReefOrderItem struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt       time.Time  `json:"createdAt"`
	OrderID         uuid.UUID  `json:"orderId" gorm:"type:uuid;index"`
	ProductID       uuid.UUID  `json:"productId" gorm:"type:uuid"`
	ConfigurationID *uuid.UUID `json:"configurationId" gorm:"type:uuid"`
	VariantKey      string     `json:"variantKey" gorm:"column:variant_key"`
	Quantity        int        `json:"quantity"`
	UnitPriceCents  int64      `json:"unitPriceCents" gorm:"column:unit_price_cents"`
}

func (ReefOrderItem) TableName() string {
	return "reef_order_items"
}
