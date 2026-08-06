package server

import (
	"encoding/json"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// validBgiEventTypes whitelists R-9.1's platform set plus bgi's own
// (game_selected, expansion_toggled, sleeve_selected, box_selected,
// fit_indicator_shown, fit_check_failed, waitlist_submitted) — mirrors
// reef's own whitelist so this endpoint can't be used to write arbitrary
// event_type strings into the analytics table.
var validBgiEventTypes = map[string]bool{
	models.BgiEventGameSelected:       true,
	models.BgiEventExpansionToggled:   true,
	models.BgiEventSleeveSelected:     true,
	models.BgiEventBoxSelected:        true,
	models.BgiEventConfiguratorOpened: true,
	models.BgiEventParameterChanged:   true,
	models.BgiEventPreviewRendered:    true,
	models.BgiEventFitIndicatorShown:  true,
	models.BgiEventValidationRejected: true,
	models.BgiEventFitCheckFailed:     true,
	models.BgiEventAddToCart:          true,
	models.BgiEventCheckoutStarted:    true,
	models.BgiEventPurchaseCompleted:  true,
	models.BgiEventWaitlistSubmitted:  true,
	models.BgiEventShareLinkCreated:   true,
	models.BgiEventShareLinkOpened:    true,
}

type eventRequest struct {
	EventType       string                 `json:"eventType" binding:"required"`
	SessionID       string                 `json:"sessionId"`
	GameSlug        string                 `json:"gameSlug"`
	ConfigurationID string                 `json:"configurationId"`
	Rule            string                 `json:"rule"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// POST /api/bgi/events (R-9.1).
func (s *server) postEvent(c *gin.Context) {
	var req eventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validBgiEventTypes[req.EventType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown eventType"})
		return
	}

	event := &models.BgiEvent{
		EventType: req.EventType,
		SessionID: req.SessionID,
		GameSlug:  req.GameSlug,
		Rule:      req.Rule,
		Metadata:  datatypes.JSON([]byte(`{}`)),
	}
	if req.ConfigurationID != "" {
		if id, err := uuid.Parse(req.ConfigurationID); err == nil {
			event.ConfigurationID = &id
		}
	}
	if req.Metadata != nil {
		if b, err := json.Marshal(req.Metadata); err == nil {
			event.Metadata = datatypes.JSON(b)
		}
	}

	if err := s.deps.DbClient.BgiEvent().Create(c.Request.Context(), event); err != nil {
		internalError(c, "record event", err)
		return
	}
	c.Status(http.StatusNoContent)
}
