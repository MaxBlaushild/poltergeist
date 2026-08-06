package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/fulfillment"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fulfillmentStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// PATCH /api/bgi/operator/orders/:id/fulfillment — mirrors reef-site's own
// print-queue status update exactly. No /operator/metrics or
// /fulfill-slant routes: the operator dashboard and Slant integration are
// both out of scope for this vertical slice (see PLATFORM_FINDINGS.md).
func (s *server) updateOrderFulfillment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	var req fulfillmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !fulfillment.OperatorSettableStatuses[fulfillment.Status(req.Status)] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be one of: printed, shipped"})
		return
	}

	ctx := c.Request.Context()
	order, err := s.deps.DbClient.BgiOrder().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	order.FulfillmentStatus = req.Status
	if fulfillment.Status(req.Status) == fulfillment.StatusShipped {
		order.Status = models.BgiOrderStatusFulfilled
	}
	if err := s.deps.DbClient.BgiOrder().Update(ctx, order); err != nil {
		internalError(c, "update order fulfillment status", err)
		return
	}

	c.JSON(http.StatusOK, order)
}
