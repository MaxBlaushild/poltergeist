package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Event types required by R-9.1.
const (
	ReefEventConfiguratorOpened = "configurator_opened"
	ReefEventParameterChanged   = "parameter_changed"
	ReefEventPreviewRendered    = "preview_rendered"
	ReefEventValidationRejected = "validation_rejected"
	ReefEventAddToCart          = "add_to_cart"
	ReefEventCheckoutStarted    = "checkout_started"
	ReefEventPurchaseCompleted  = "purchase_completed"
	ReefEventShareLinkCreated   = "share_link_created"
	ReefEventShareLinkOpened    = "share_link_opened"
)

// ReefEvent is the analytics event log (R-9.1). Written directly to Postgres:
// there is no existing repo-wide telemetry path to CONFORM to (see
// go/reef-site/INVENTORY.md), and R-9.1 names this as the explicit fallback.
type ReefEvent struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt       time.Time      `json:"createdAt"`
	EventType       string         `json:"eventType" gorm:"column:event_type"`
	SessionID       string         `json:"sessionId" gorm:"column:session_id"`
	ProductSlug     string         `json:"productSlug" gorm:"column:product_slug"`
	ConfigurationID *uuid.UUID     `json:"configurationId" gorm:"type:uuid;column:configuration_id"`
	Rule            string         `json:"rule"`
	Metadata        datatypes.JSON `json:"metadata"`
}

func (ReefEvent) TableName() string {
	return "reef_events"
}
