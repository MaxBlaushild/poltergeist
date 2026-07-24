package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// POST /api/reef/checkout and POST /api/reef/webhooks/stripe (R-8.1, R-2.8).
// Implemented in full alongside the ManualAdapter fulfillment integration —
// see internal/fulfillment and the Stripe checkout-session wiring.
func (s *server) postCheckout(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "checkout not yet implemented"})
}

func (s *server) postStripeWebhook(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "stripe webhook not yet implemented"})
}
