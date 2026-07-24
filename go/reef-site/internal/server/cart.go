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
	errVariantNotFound       = errors.New("variant not found for this product")
	errConfigurationRequired = errors.New("configurationId is required for a configurable product")
	errConfigurationNotFound = errors.New("configuration not found")
	errConfigurationNotValid = errors.New("configuration has not passed server-side validation yet")
)

type cartItemRequest struct {
	ProductSlug     string `json:"productSlug" binding:"required"`
	VariantKey      string `json:"variantKey"`
	ConfigurationID string `json:"configurationId"`
	Quantity        int    `json:"quantity" binding:"required"`
}

type cartRequest struct {
	Items []cartItemRequest `json:"items" binding:"required"`
}

type cartItemResponse struct {
	ProductSlug     string `json:"productSlug"`
	ProductName     string `json:"productName"`
	VariantKey      string `json:"variantKey,omitempty"`
	VariantLabel    string `json:"variantLabel,omitempty"`
	ConfigurationID string `json:"configurationId,omitempty"`
	Quantity        int    `json:"quantity"`
	UnitPriceCents  int64  `json:"unitPriceCents"`
	LineTotalCents  int64  `json:"lineTotalCents"`
}

type cartResponse struct {
	Items                        []cartItemResponse `json:"items"`
	SubtotalCents                int64              `json:"subtotalCents"`
	ShippingCents                int64              `json:"shippingCents"`
	TotalCents                   int64              `json:"totalCents"`
	RemainingToFreeShippingCents int64              `json:"remainingToFreeShippingCents"`
	CrossSell                    []productResponse  `json:"crossSell,omitempty"`
}

// POST /api/reef/cart (R-8.1). Prices every line server-side from stored
// configuration/variant prices — the client never computes or supplies a
// price (R-6.2). Below the free-shipping threshold, includes up to two
// fixed SKUs as cross-sell (R-6.4).
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
	shippingCents, remaining := pricing.Shipping(subtotal, shippingRates)

	resp := cartResponse{
		Items:                        items,
		SubtotalCents:                subtotal,
		ShippingCents:                shippingCents,
		TotalCents:                   subtotal + shippingCents,
		RemainingToFreeShippingCents: remaining,
	}

	if remaining > 0 {
		crossSell, err := s.crossSellProducts(ctx, req.Items)
		if err != nil {
			internalError(c, "load cross-sell products", err)
			return
		}
		resp.CrossSell = crossSell
	}

	c.JSON(http.StatusOK, resp)
}

func (s *server) priceCartItem(ctx context.Context, req cartItemRequest) (*cartItemResponse, error) {
	if req.Quantity < 1 {
		return nil, errInvalidQuantity
	}

	product, err := s.deps.DbClient.ReefProduct().FindBySlug(ctx, req.ProductSlug)
	if err != nil {
		return nil, errProductNotFound
	}

	item := &cartItemResponse{
		ProductSlug: product.Slug,
		ProductName: product.Name,
		Quantity:    req.Quantity,
	}

	switch product.Kind {
	case models.ReefProductKindConfigurable:
		id, err := uuid.Parse(req.ConfigurationID)
		if err != nil {
			return nil, errConfigurationRequired
		}
		cfg, err := s.deps.DbClient.ReefConfiguration().FindByID(ctx, id)
		if err != nil {
			return nil, errConfigurationNotFound
		}
		// R-5.1: nothing enters a cart without a passing server-side slice.
		if cfg.Status != models.ReefConfigurationStatusValid || cfg.PriceCents == nil {
			return nil, errConfigurationNotValid
		}
		item.ConfigurationID = cfg.ID.String()
		item.UnitPriceCents = *cfg.PriceCents

	case models.ReefProductKindFixed:
		variant, err := s.deps.DbClient.ReefProductVariant().FindByProductAndKey(ctx, product.ID, req.VariantKey)
		if err != nil {
			return nil, errVariantNotFound
		}
		item.VariantKey = variant.VariantKey
		item.VariantLabel = variant.Label
		item.UnitPriceCents = variant.PriceCents
	}

	item.LineTotalCents = item.UnitPriceCents * int64(req.Quantity)
	return item, nil
}

// crossSellProducts returns up to two active fixed SKUs not already in the
// cart (R-6.4).
func (s *server) crossSellProducts(ctx context.Context, requested []cartItemRequest) ([]productResponse, error) {
	inCart := make(map[string]bool, len(requested))
	for _, r := range requested {
		inCart[r.ProductSlug] = true
	}

	products, err := s.deps.DbClient.ReefProduct().FindActive(ctx)
	if err != nil {
		return nil, err
	}

	var crossSell []productResponse
	for _, p := range products {
		if p.Kind != models.ReefProductKindFixed || inCart[p.Slug] {
			continue
		}
		resp := productResponse{ReefProduct: p}
		if variants, err := s.deps.DbClient.ReefProductVariant().FindByProductID(ctx, p.ID); err == nil {
			resp.Variants = variants
		}
		crossSell = append(crossSell, resp)
		if len(crossSell) == 2 {
			break
		}
	}
	return crossSell, nil
}
