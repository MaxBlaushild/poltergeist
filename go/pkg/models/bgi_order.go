package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	BgiOrderStatusPendingPayment = "pending_payment"
	BgiOrderStatusPaid           = "paid"
	BgiOrderStatusFulfilled      = "fulfilled"
	BgiOrderStatusCancelled      = "cancelled"

	BgiFulfillmentProviderManual = "manual"
)

// BgiOrder is a structural clone of ReefOrder.
type BgiOrder struct {
	ID                    uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt             time.Time      `json:"createdAt"`
	UpdatedAt             time.Time      `json:"updatedAt"`
	OrderToken            string         `json:"orderToken" gorm:"column:order_token;uniqueIndex"`
	StripeSessionID       string         `json:"stripeSessionId" gorm:"column:stripe_session_id"`
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

	Items []BgiOrderItem `json:"items" gorm:"foreignKey:OrderID"`
}

func (BgiOrder) TableName() string {
	return "bgi_orders"
}

type BgiOrderItem struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt       time.Time  `json:"createdAt"`
	OrderID         uuid.UUID  `json:"orderId" gorm:"type:uuid;column:order_id;index"`
	ProductID       uuid.UUID  `json:"productId" gorm:"type:uuid;column:product_id"`
	ConfigurationID *uuid.UUID `json:"configurationId" gorm:"type:uuid;column:configuration_id"`
	Quantity        int        `json:"quantity"`
	UnitPriceCents  int64      `json:"unitPriceCents" gorm:"column:unit_price_cents"`
}

func (BgiOrderItem) TableName() string {
	return "bgi_order_items"
}
