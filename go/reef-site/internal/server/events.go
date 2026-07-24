package server

import (
	"encoding/json"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// validReefEventTypes whitelists R-9.1's nine required events so this
// endpoint can't be used to write arbitrary event_type strings into the
// analytics table.
var validReefEventTypes = map[string]bool{
	models.ReefEventConfiguratorOpened: true,
	models.ReefEventParameterChanged:   true,
	models.ReefEventPreviewRendered:    true,
	models.ReefEventValidationRejected: true,
	models.ReefEventAddToCart:          true,
	models.ReefEventCheckoutStarted:    true,
	models.ReefEventPurchaseCompleted:  true,
	models.ReefEventShareLinkCreated:   true,
	models.ReefEventShareLinkOpened:    true,
}

type eventRequest struct {
	EventType       string                 `json:"eventType" binding:"required"`
	SessionID       string                 `json:"sessionId"`
	ProductSlug     string                 `json:"productSlug"`
	ConfigurationID string                 `json:"configurationId"`
	Rule            string                 `json:"rule"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// POST /api/reef/events (R-9.1). validation_rejected is normally emitted
// server-side by the job-runner processor itself (it has the authoritative
// rule/reason); the other eight are client-observed UI events with no other
// place to originate from.
func (s *server) postEvent(c *gin.Context) {
	var req eventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validReefEventTypes[req.EventType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown eventType"})
		return
	}

	event := &models.ReefEvent{
		EventType:   req.EventType,
		SessionID:   req.SessionID,
		ProductSlug: req.ProductSlug,
		Rule:        req.Rule,
		Metadata:    datatypes.JSON([]byte(`{}`)), // metadata is NOT NULL; overwritten below if the caller sent any
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

	if err := s.deps.DbClient.ReefEvent().Create(c.Request.Context(), event); err != nil {
		internalError(c, "record event", err)
		return
	}
	c.Status(http.StatusNoContent)
}
