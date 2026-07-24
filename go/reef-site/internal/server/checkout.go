package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/billing"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/pricing"
	"github.com/MaxBlaushild/poltergeist/reef-site/internal/fulfillment"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type checkoutRequest struct {
	Items         []cartItemRequest `json:"items" binding:"required"`
	CustomerEmail string            `json:"customerEmail" binding:"required"`
	SuccessURL    string            `json:"successUrl" binding:"required"`
	CancelURL     string            `json:"cancelUrl" binding:"required"`
	SessionID     string            `json:"sessionId"`
}

type checkoutResponse struct {
	CheckoutURL string `json:"checkoutUrl"`
	OrderToken  string `json:"orderToken"`
}

// POST /api/reef/checkout (R-8.1, R-2.8). Prices every line server-side
// (same code path as POST /cart), persists a reef_order + line items, and
// hands off to the repo's existing Stripe integration (go/pkg/billing) for
// an itemized, tax-enabled Checkout Session — see
// go/reef-site/INVENTORY.md for why that integration needed additive
// extension rather than a second Stripe client.
func (s *server) postCheckout(c *gin.Context) {
	var req checkoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cart is empty"})
		return
	}
	ctx := c.Request.Context()

	priced := make([]*cartItemResponse, 0, len(req.Items))
	var subtotal int64
	for _, reqItem := range req.Items {
		item, err := s.priceCartItem(ctx, reqItem)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		priced = append(priced, item)
		subtotal += item.LineTotalCents
	}

	shippingCents, _ := pricing.Shipping(subtotal, pricing.ShippingRates{
		FreeShippingThresholdCents: s.deps.Config.Public.FreeShippingThresholdCents,
		FlatShippingCents:          s.deps.Config.Public.FlatShippingCents,
	})

	orderToken, err := randomOrderToken()
	if err != nil {
		internalError(c, "generate order token", err)
		return
	}

	orderItems := make([]models.ReefOrderItem, 0, len(priced))
	for i, item := range priced {
		orderItem := models.ReefOrderItem{
			ProductID:      mustProductID(ctx, s, req.Items[i].ProductSlug),
			VariantKey:     item.VariantKey,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
		}
		if item.ConfigurationID != "" {
			if id, err := uuid.Parse(item.ConfigurationID); err == nil {
				orderItem.ConfigurationID = &id
			}
		}
		orderItems = append(orderItems, orderItem)
	}

	order, err := s.deps.DbClient.ReefOrder().Create(ctx, &models.ReefOrder{
		OrderToken:          orderToken,
		CustomerEmail:       req.CustomerEmail,
		Status:              models.ReefOrderStatusPendingPayment,
		FulfillmentProvider: s.deps.Config.Public.FulfillmentProvider,
		SubtotalCents:       subtotal,
		ShippingCents:       shippingCents,
		TotalCents:          subtotal + shippingCents,
		ShippingAddress:     datatypes.JSON([]byte(`{}`)),
		Items:               orderItems,
	})
	if err != nil {
		internalError(c, "create order", err)
		return
	}

	lineItems := make([]billing.PaymentLineItem, 0, len(priced)+1)
	for _, item := range priced {
		name := item.ProductName
		if item.VariantLabel != "" {
			name = fmt.Sprintf("%s (%s)", name, item.VariantLabel)
		}
		lineItems = append(lineItems, billing.PaymentLineItem{
			Name:          name,
			AmountInCents: item.UnitPriceCents,
			Quantity:      int64(item.Quantity),
		})
	}
	if shippingCents > 0 {
		lineItems = append(lineItems, billing.PaymentLineItem{
			Name:          "Shipping",
			AmountInCents: shippingCents,
			Quantity:      1,
		})
	}

	session, err := s.deps.BillingClient.NewPaymentCheckoutSession(ctx, &billing.PaymentCheckoutSessionParams{
		SessionSuccessRedirectUrl:  req.SuccessURL,
		SessionCancelRedirectUrl:   req.CancelURL,
		LineItems:                  lineItems,
		AutomaticTax:               true,
		CollectShippingAddress:     true,
		PaymentCompleteCallbackUrl: s.deps.Config.Public.BaseURL + "/api/reef/webhooks/stripe",
		Metadata: map[string]string{
			"reef_order_id":    order.ID.String(),
			"reef_order_token": order.OrderToken,
			"reef_session_id":  req.SessionID,
		},
	})
	if err != nil {
		internalError(c, "create stripe checkout session", err)
		return
	}

	c.JSON(http.StatusOK, checkoutResponse{CheckoutURL: session.URL, OrderToken: order.OrderToken})
}

// POST /api/reef/webhooks/stripe (R-8.1). Despite the name (kept for R-8.1's
// literal route), this doesn't verify raw Stripe webhook signatures itself
// — it receives the already-verified completion callback that go/billing
// forwards after processing the real Stripe webhook (the same mechanism
// every other domain's payment flow in this repo uses). "No custom card
// handling" (R-2.8) is satisfied by never touching Stripe directly here.
func (s *server) postStripeWebhook(c *gin.Context) {
	var payload billing.OnPaymentComplete
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orderID, err := uuid.Parse(payload.Metadata["reef_order_id"])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid reef_order_id in metadata"})
		return
	}

	ctx := c.Request.Context()
	order, err := s.deps.DbClient.ReefOrder().FindByID(ctx, orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	// Idempotent: webhook forwards can be retried by Stripe/billing.
	if order.Status == models.ReefOrderStatusPaid || order.Status == models.ReefOrderStatusFulfilled {
		c.JSON(http.StatusOK, gin.H{"status": "already processed"})
		return
	}

	order.Status = models.ReefOrderStatusPaid
	order.StripeSessionID = payload.SessionID
	if payload.CustomerEmail != "" {
		order.CustomerEmail = payload.CustomerEmail
	}
	if payload.ShippingAddress != nil {
		if b, err := json.Marshal(payload.ShippingAddress); err == nil {
			order.ShippingAddress = datatypes.JSON(b)
		}
	}

	fulfillmentOrder, err := s.buildFulfillmentOrder(ctx, order, payload.ShippingAddress)
	if err != nil {
		internalError(c, "build fulfillment order", err)
		return
	}

	adapter, err := s.fulfillmentAdapter()
	if err != nil {
		internalError(c, "resolve fulfillment adapter", err)
		return
	}
	externalID, err := adapter.SubmitOrder(ctx, fulfillmentOrder)
	if err != nil {
		log.Printf("[reef] fulfillment submission failed for order %s: %v", order.OrderToken, err)
		order.FulfillmentStatus = "submission_failed: " + err.Error()
	} else {
		order.FulfillmentExternalID = externalID
		order.FulfillmentStatus = string(fulfillment.StatusSubmitted)
	}

	if err := s.deps.DbClient.ReefOrder().Update(ctx, order); err != nil {
		internalError(c, "update order", err)
		return
	}

	if err := s.deps.DbClient.ReefEvent().Create(ctx, &models.ReefEvent{
		EventType: models.ReefEventPurchaseCompleted,
		SessionID: payload.Metadata["reef_session_id"],
		Metadata:  datatypes.JSON([]byte(fmt.Sprintf(`{"orderToken":%q,"totalCents":%d}`, order.OrderToken, order.TotalCents))),
	}); err != nil {
		log.Printf("[reef] failed to record purchase_completed event: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *server) buildFulfillmentOrder(ctx context.Context, order *models.ReefOrder, addr *billing.ShippingAddress) (fulfillment.Order, error) {
	fo := fulfillment.Order{
		OrderToken:    order.OrderToken,
		CustomerEmail: order.CustomerEmail,
	}
	if addr != nil {
		fo.ShippingName = addr.Name
		fo.ShippingLine1 = addr.Line1
		fo.ShippingLine2 = addr.Line2
		fo.ShippingCity = addr.City
		fo.ShippingState = addr.State
		fo.ShippingZip = addr.PostalCode
		fo.ShippingCountry = addr.Country
	}

	orderWithItems, err := s.deps.DbClient.ReefOrder().FindByID(ctx, order.ID)
	if err != nil {
		return fo, err
	}

	for _, item := range orderWithItems.Items {
		fulfillmentItem := fulfillment.OrderItem{
			VariantKey: item.VariantKey,
			Quantity:   item.Quantity,
		}
		if product, err := s.deps.DbClient.ReefProduct().FindByID(ctx, item.ProductID); err == nil {
			fulfillmentItem.ProductSlug = product.Slug
		}
		if item.ConfigurationID != nil {
			if cfg, err := s.deps.DbClient.ReefConfiguration().FindByID(ctx, *item.ConfigurationID); err == nil && cfg.GeometryHash != nil {
				if sliceResult, err := s.deps.DbClient.ReefSliceResult().FindByGeometryHash(ctx, *cfg.GeometryHash); err == nil && sliceResult != nil {
					fulfillmentItem.STLKey = sliceResult.STLKey
				}
			}
		}
		fo.Items = append(fo.Items, fulfillmentItem)
	}

	return fo, nil
}

func (s *server) fulfillmentAdapter() (fulfillment.Adapter, error) {
	switch s.deps.Config.Public.FulfillmentProvider {
	case models.ReefFulfillmentProviderManual, "":
		return fulfillment.NewManualAdapter(
			s.deps.AwsClient,
			s.deps.EmailClient,
			s.deps.Config.Public.S3Bucket,
			s.deps.Config.Public.OperatorEmail,
			s.deps.Config.Public.EmailFromAddress,
		), nil
	default:
		// R-7.3: SlantAdapter is v1.1, deliberately not implemented until
		// sample parts have been ordered and inspected.
		return nil, fmt.Errorf("fulfillment provider %q not implemented", s.deps.Config.Public.FulfillmentProvider)
	}
}

func randomOrderToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func mustProductID(ctx context.Context, s *server, slug string) uuid.UUID {
	product, err := s.deps.DbClient.ReefProduct().FindBySlug(ctx, slug)
	if err != nil {
		return uuid.Nil
	}
	return product.ID
}
