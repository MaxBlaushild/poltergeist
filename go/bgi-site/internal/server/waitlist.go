package server

import (
	"encoding/json"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type waitlistRequest struct {
	Email         string `json:"email" binding:"required"`
	RequestedGame string `json:"requestedGame" binding:"required"`
	SessionID     string `json:"sessionId"`
}

// POST /api/bgi/waitlist (R-1.3/R-8.1). "A 'request a game' waitlist form is
// permitted (it is demand data); building the game it requests is not."
// Recorded as a bgi_events row rather than a dedicated table — the
// waitlist_submitted event type already models this shape (R-9.2's
// "most-requested waitlist games" reads directly off event metadata), and a
// v1 vertical slice with one launch game doesn't yet need a table of its
// own to manage.
func (s *server) postWaitlist(c *gin.Context) {
	var req waitlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metadata, err := json.Marshal(map[string]string{"email": req.Email, "requestedGame": req.RequestedGame})
	if err != nil {
		internalError(c, "encode waitlist metadata", err)
		return
	}

	if err := s.deps.DbClient.BgiEvent().Create(c.Request.Context(), &models.BgiEvent{
		EventType: models.BgiEventWaitlistSubmitted,
		SessionID: req.SessionID,
		Metadata:  datatypes.JSON(metadata),
	}); err != nil {
		internalError(c, "record waitlist submission", err)
		return
	}

	c.Status(http.StatusNoContent)
}
