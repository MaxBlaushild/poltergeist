package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/billing"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/reef-site/internal/fulfillment"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type operatorOrderItemResponse struct {
	models.ReefOrderItem
	ProductName string `json:"productName"`
	ProductSlug string `json:"productSlug"`
	STLURL      string `json:"stlUrl,omitempty"`
}

type operatorOrderResponse struct {
	models.ReefOrder
	Items           []operatorOrderItemResponse `json:"items"`
	ShippingAddress *billing.ShippingAddress    `json:"shippingAddress"`
}

// GET /api/reef/operator/orders (print queue). Every paid/fulfilled order,
// newest first, with enough on each line item (product name, STL download
// link) for an operator to actually work from without touching the
// database — the print queue this fills in for was previously "read the
// manifest CSV out of the notification email."
func (s *server) listOperatorOrders(c *gin.Context) {
	ctx := c.Request.Context()
	orders, err := s.deps.DbClient.ReefOrder().FindPaid(ctx)
	if err != nil {
		internalError(c, "list orders", err)
		return
	}

	resp := make([]operatorOrderResponse, 0, len(orders))
	for _, order := range orders {
		resp = append(resp, s.toOperatorOrderResponse(ctx, order))
	}
	c.JSON(http.StatusOK, resp)
}

func (s *server) toOperatorOrderResponse(ctx context.Context, order models.ReefOrder) operatorOrderResponse {
	items := make([]operatorOrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		itemResp := operatorOrderItemResponse{ReefOrderItem: item}
		if product, err := s.deps.DbClient.ReefProduct().FindByID(ctx, item.ProductID); err == nil {
			itemResp.ProductName = product.Name
			itemResp.ProductSlug = product.Slug
		}
		if item.ConfigurationID != nil {
			if cfg, err := s.deps.DbClient.ReefConfiguration().FindByID(ctx, *item.ConfigurationID); err == nil && cfg.GeometryHash != nil {
				if sliceResult, err := s.deps.DbClient.ReefSliceResult().FindByGeometryHash(ctx, *cfg.GeometryHash); err == nil && sliceResult != nil && sliceResult.STLKey != "" {
					itemResp.STLURL = s.previewURL(sliceResult.STLKey)
				}
			}
		}
		items = append(items, itemResp)
	}

	return operatorOrderResponse{
		ReefOrder:       order,
		Items:           items,
		ShippingAddress: decodeShippingAddress(json.RawMessage(order.ShippingAddress)),
	}
}

type fulfillmentStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// PATCH /api/reef/operator/orders/:id/fulfillment. The only write the print
// queue needs: move an order through printed -> shipped. Shipping also
// closes out the order's own top-level Status (R-8's ReefOrderStatusFulfilled
// existed but nothing ever set it before this).
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
	order, err := s.deps.DbClient.ReefOrder().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	order.FulfillmentStatus = req.Status
	if fulfillment.Status(req.Status) == fulfillment.StatusShipped {
		order.Status = models.ReefOrderStatusFulfilled
	}
	if err := s.deps.DbClient.ReefOrder().Update(ctx, order); err != nil {
		internalError(c, "update order fulfillment status", err)
		return
	}

	c.JSON(http.StatusOK, s.toOperatorOrderResponse(ctx, *order))
}

// POST /api/reef/operator/orders/:id/fulfill-slant. Explicit, one-order-
// at-a-time operator action (see fulfillment.SlantAdapter's own doc
// comment for why this is deliberately not the checkout-time default).
func (s *server) fulfillOrderWithSlant(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	ctx := c.Request.Context()
	order, err := s.deps.DbClient.ReefOrder().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if order.FulfillmentProvider == models.ReefFulfillmentProviderSlant && order.FulfillmentExternalID != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order was already sent to Slant"})
		return
	}

	addr := decodeShippingAddress(json.RawMessage(order.ShippingAddress))
	fulfillmentOrder, err := s.buildFulfillmentOrder(ctx, order, addr)
	if err != nil {
		internalError(c, "build fulfillment order for slant", err)
		return
	}

	externalID, err := s.slantAdapter().SubmitOrder(ctx, fulfillmentOrder)
	if err != nil {
		// Surfaced verbatim (not internalError's generic message) — an
		// operator acting on this needs to see exactly what Slant rejected
		// (bad API key, missing platform ID, no printable items, etc).
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	order.FulfillmentProvider = models.ReefFulfillmentProviderSlant
	order.FulfillmentExternalID = externalID
	order.FulfillmentStatus = string(fulfillment.StatusSubmitted)
	if err := s.deps.DbClient.ReefOrder().Update(ctx, order); err != nil {
		internalError(c, "update order after slant submission", err)
		return
	}

	c.JSON(http.StatusOK, s.toOperatorOrderResponse(ctx, *order))
}

// POST /api/reef/operator/orders/:id/refresh-slant-status. Polling, not a
// webhook listener — Slant's own docs describe webhook support as not yet
// fully implemented ("order.shipped" is the only event type so far), so an
// operator-triggered refresh is the reliable path for now.
func (s *server) refreshSlantStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	ctx := c.Request.Context()
	order, err := s.deps.DbClient.ReefOrder().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if order.FulfillmentProvider != models.ReefFulfillmentProviderSlant || order.FulfillmentExternalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order was not sent to Slant"})
		return
	}

	status, err := s.slantAdapter().GetStatus(ctx, order.FulfillmentExternalID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	order.FulfillmentStatus = string(status)
	if status == fulfillment.StatusShipped {
		order.Status = models.ReefOrderStatusFulfilled
	}
	if err := s.deps.DbClient.ReefOrder().Update(ctx, order); err != nil {
		internalError(c, "update order after slant status refresh", err)
		return
	}

	c.JSON(http.StatusOK, s.toOperatorOrderResponse(ctx, *order))
}

func decodeShippingAddress(raw json.RawMessage) *billing.ShippingAddress {
	if len(raw) == 0 {
		return nil
	}
	var addr billing.ShippingAddress
	if err := json.Unmarshal(raw, &addr); err != nil {
		return nil
	}
	if addr == (billing.ShippingAddress{}) {
		return nil
	}
	return &addr
}
