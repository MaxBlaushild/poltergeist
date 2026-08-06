package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/pricing"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	errInvalidQuantity       = errors.New("quantity must be at least 1")
	errProductNotFound       = errors.New("product not found")
	errConfigurationRequired = errors.New("configurationId is required")
	errConfigurationNotFound = errors.New("configuration not found")
	errConfigurationNotValid = errors.New("configuration has not passed server-side validation yet")
)

type cartItemRequest struct {
	ProductSlug     string `json:"productSlug" binding:"required"`
	ConfigurationID string `json:"configurationId" binding:"required"`
	Quantity        int    `json:"quantity" binding:"required"`
}

type cartRequest struct {
	Items []cartItemRequest `json:"items" binding:"required"`
}

type cartItemResponse struct {
	ProductSlug     string `json:"productSlug"`
	ProductName     string `json:"productName"`
	ConfigurationID string `json:"configurationId"`
	Quantity        int    `json:"quantity"`
	UnitPriceCents  int64  `json:"unitPriceCents"`
	LineTotalCents  int64  `json:"lineTotalCents"`
}

type cartResponse struct {
	Items         []cartItemResponse `json:"items"`
	SubtotalCents int64              `json:"subtotalCents"`
	ShippingCents int64              `json:"shippingCents"`
	TotalCents    int64              `json:"totalCents"`
}

// POST /api/bgi/cart (R-8.1). Prices every line server-side from the stored
// configuration price — the client never computes or supplies a price
// (R-7.2). R-7.3: shipping is baked into price via BGI_SET_ASSEMBLY_FEE_CENTS
// (see the job processor), so BGI_FREE_SHIPPING_THRESHOLD_CENTS defaults to
// 0 — always free — rather than reusing reef's threshold mechanic, which
// this vertical's AOV would clear trivially anyway.
func (s *server) postCart(c *gin.Context) {
	var req cartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()

	items := make([]cartItemResponse, 0, len(req.Items))
	var subtotal int64

	for _, reqItem := range req.Items {
		item, err := s.priceCartItem(ctx, reqItem)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		items = append(items, *item)
		subtotal += item.LineTotalCents
	}

	shippingRates := pricing.ShippingRates{
		FreeShippingThresholdCents: s.deps.Config.Public.FreeShippingThresholdCents,
		FlatShippingCents:          s.deps.Config.Public.FlatShippingCents,
	}
	shippingCents, _ := pricing.Shipping(subtotal, shippingRates)

	c.JSON(http.StatusOK, cartResponse{
		Items:         items,
		SubtotalCents: subtotal,
		ShippingCents: shippingCents,
		TotalCents:    subtotal + shippingCents,
	})
}

func (s *server) priceCartItem(ctx context.Context, req cartItemRequest) (*cartItemResponse, error) {
	if req.Quantity < 1 {
		return nil, errInvalidQuantity
	}

	product, err := s.deps.DbClient.BgiProduct().FindBySlug(ctx, req.ProductSlug)
	if err != nil {
		return nil, errProductNotFound
	}

	id, err := uuid.Parse(req.ConfigurationID)
	if err != nil {
		return nil, errConfigurationRequired
	}
	cfg, err := s.deps.DbClient.BgiConfiguration().FindByID(ctx, id)
	if err != nil {
		return nil, errConfigurationNotFound
	}
	// R-5.1: nothing enters a cart without a passing server-side assembly+slice.
	if cfg.Status != models.BgiConfigurationStatusValid || cfg.PriceCents == nil {
		return nil, errConfigurationNotValid
	}

	item := &cartItemResponse{
		ProductSlug:     product.Slug,
		ProductName:     product.Name,
		ConfigurationID: cfg.ID.String(),
		Quantity:        req.Quantity,
		UnitPriceCents:  *cfg.PriceCents,
	}
	item.LineTotalCents = item.UnitPriceCents * int64(req.Quantity)
	return item, nil
}
