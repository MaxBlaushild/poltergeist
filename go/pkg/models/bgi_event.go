package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Event types: R-9.1's platform set plus bgi's own. FitIndicatorShown is new
// — every time the configurator computes assembled height vs. box depth
// (client preview or server validate), this is logged so R-6.3's rejection
// frequency can reveal whether the seeded box/sleeve numbers need
// correcting before customers find out physically.
const (
	BgiEventGameSelected       = "game_selected"
	BgiEventExpansionToggled   = "expansion_toggled"
	BgiEventSleeveSelected     = "sleeve_selected"
	BgiEventBoxSelected        = "box_selected"
	BgiEventConfiguratorOpened = "configurator_opened"
	BgiEventParameterChanged   = "parameter_changed"
	BgiEventPreviewRendered    = "preview_rendered"
	BgiEventFitIndicatorShown  = "fit_indicator_shown"
	BgiEventValidationRejected = "validation_rejected"
	BgiEventFitCheckFailed     = "fit_check_failed"
	BgiEventAddToCart          = "add_to_cart"
	BgiEventCheckoutStarted    = "checkout_started"
	BgiEventPurchaseCompleted  = "purchase_completed"
	BgiEventWaitlistSubmitted  = "waitlist_submitted"
	BgiEventShareLinkCreated   = "share_link_created"
	BgiEventShareLinkOpened    = "share_link_opened"
)

// BgiEvent is the analytics event log — structural clone of ReefEvent
// (R-9.1's explicit fallback: no repo-wide telemetry path exists to CONFORM
// to), keyed by game slug rather than product slug since a fit check can
// fail before a specific product/configuration exists.
type BgiEvent struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt       time.Time      `json:"createdAt"`
	EventType       string         `json:"eventType" gorm:"column:event_type"`
	SessionID       string         `json:"sessionId" gorm:"column:session_id"`
	GameSlug        string         `json:"gameSlug" gorm:"column:game_slug"`
	ConfigurationID *uuid.UUID     `json:"configurationId" gorm:"type:uuid;column:configuration_id"`
	Rule            string         `json:"rule"`
	Metadata        datatypes.JSON `json:"metadata"`
}

func (BgiEvent) TableName() string {
	return "bgi_events"
}
