package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/billing"
	"github.com/MaxBlaushild/poltergeist/pkg/email"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/fulfillment"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/pricing"
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

// POST /api/bgi/checkout (R-8.1). Prices every line server-side (same code
// path as POST /cart), persists a bgi_order + line items, and hands off to
// the repo's existing Stripe integration (go/pkg/billing). Platform is left
// empty so go/billing's existing routing falls through to the shared/legacy
// key rather than reef's dedicated account — standing up a dedicated bgi
// Stripe account is deferred until this vertical is proven out.
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

	orderItems := make([]models.BgiOrderItem, 0, len(priced))
	for i, item := range priced {
		orderItem := models.BgiOrderItem{
			ProductID:      mustProductID(ctx, s, req.Items[i].ProductSlug),
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

	order, err := s.deps.DbClient.BgiOrder().Create(ctx, &models.BgiOrder{
		OrderToken:          orderToken,
		CustomerEmail:       req.CustomerEmail,
		Status:              models.BgiOrderStatusPendingPayment,
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
		lineItems = append(lineItems, billing.PaymentLineItem{
			Name:          item.ProductName,
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

	// See reef-site's checkout.go for why the client sends a ":orderToken:"
	// placeholder rather than knowing the token up front.
	successURL := strings.ReplaceAll(req.SuccessURL, ":orderToken:", order.OrderToken)

	session, err := s.deps.BillingClient.NewPaymentCheckoutSession(ctx, &billing.PaymentCheckoutSessionParams{
		SessionSuccessRedirectUrl:  successURL,
		SessionCancelRedirectUrl:   req.CancelURL,
		LineItems:                  lineItems,
		AutomaticTax:               false,
		CollectShippingAddress:     true,
		PaymentCompleteCallbackUrl: s.deps.Config.Public.BaseURL + "/api/bgi/webhooks/stripe",
		Metadata: map[string]string{
			"bgi_order_id":    order.ID.String(),
			"bgi_order_token": order.OrderToken,
			"bgi_session_id":  req.SessionID,
		},
	})
	if err != nil {
		internalError(c, "create stripe checkout session", err)
		return
	}

	c.JSON(http.StatusOK, checkoutResponse{CheckoutURL: session.URL, OrderToken: order.OrderToken})
}

// POST /api/bgi/webhooks/stripe (R-8.1) — receives the already-verified
// completion callback go/billing forwards, same mechanism every other
// domain's payment flow in this repo uses (see reef-site's own handler for
// the full reasoning).
func (s *server) postStripeWebhook(c *gin.Context) {
	var payload billing.OnPaymentComplete
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orderID, err := uuid.Parse(payload.Metadata["bgi_order_id"])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid bgi_order_id in metadata"})
		return
	}

	ctx := c.Request.Context()
	order, err := s.deps.DbClient.BgiOrder().FindByID(ctx, orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	// Idempotent: webhook forwards can be retried by Stripe/billing.
	if order.Status == models.BgiOrderStatusPaid || order.Status == models.BgiOrderStatusFulfilled {
		c.JSON(http.StatusOK, gin.H{"status": "already processed"})
		return
	}

	order.Status = models.BgiOrderStatusPaid
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
		log.Printf("[bgi] fulfillment submission failed for order %s: %v", order.OrderToken, err)
		order.FulfillmentStatus = "submission_failed: " + err.Error()
	} else {
		order.FulfillmentExternalID = externalID
		order.FulfillmentStatus = string(fulfillment.StatusSubmitted)
	}

	if err := s.deps.DbClient.BgiOrder().Update(ctx, order); err != nil {
		internalError(c, "update order", err)
		return
	}

	if err := s.sendOrderConfirmationEmail(ctx, order); err != nil {
		log.Printf("[bgi] failed to send order confirmation email for order %s: %v", order.OrderToken, err)
	}

	if err := s.deps.DbClient.BgiEvent().Create(ctx, &models.BgiEvent{
		EventType: models.BgiEventPurchaseCompleted,
		SessionID: payload.Metadata["bgi_session_id"],
		Metadata:  datatypes.JSON([]byte(fmt.Sprintf(`{"orderToken":%q,"totalCents":%d}`, order.OrderToken, order.TotalCents))),
	}); err != nil {
		log.Printf("[bgi] failed to record purchase_completed event: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// sendOrderConfirmationEmail mirrors reef-site's — states a lead-time
// expectation (R-7.4) since a full tray-set print genuinely takes days, and
// silent long lead times are a support-ticket/chargeback generator.
func (s *server) sendOrderConfirmationEmail(ctx context.Context, order *models.BgiOrder) error {
	if order.CustomerEmail == "" {
		return fmt.Errorf("order %s has no customer email", order.OrderToken)
	}

	orderWithItems, err := s.deps.DbClient.BgiOrder().FindByID(ctx, order.ID)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	fmt.Fprintf(&body, "Thanks for your order!\n\n")
	fmt.Fprintf(&body, "Order %s\n\n", order.OrderToken)
	for _, item := range orderWithItems.Items {
		name := "Tray set"
		if product, err := s.deps.DbClient.BgiProduct().FindByID(ctx, item.ProductID); err == nil {
			name = product.Name
		}
		fmt.Fprintf(&body, "- %s x%d — $%.2f\n", name, item.Quantity, float64(item.UnitPriceCents*int64(item.Quantity))/100)
	}
	fmt.Fprintf(&body, "\nShipping: $%.2f\n", float64(order.ShippingCents)/100)
	fmt.Fprintf(&body, "Total: $%.2f\n\n", float64(order.TotalCents)/100)
	fmt.Fprintf(&body, "This is a made-to-order print — expect several days of fulfillment lead time before it ships.\n")

	orderURL := s.deps.Config.Public.SiteURL + "/orders/" + order.OrderToken
	fmt.Fprintf(&body, "Track your order: %s\n", orderURL)

	return s.deps.EmailClient.SendMail(email.Email{
		Subject:          "Your tray set order is confirmed — " + order.OrderToken,
		Name:             order.CustomerEmail,
		Email:            order.CustomerEmail,
		PlainTextContent: body.String(),
		HtmlContent:      "<pre>" + html.EscapeString(body.String()) + "</pre>",
	})
}

func (s *server) buildFulfillmentOrder(ctx context.Context, order *models.BgiOrder, addr *billing.ShippingAddress) (fulfillment.Order, error) {
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

	orderWithItems, err := s.deps.DbClient.BgiOrder().FindByID(ctx, order.ID)
	if err != nil {
		return fo, err
	}

	for _, item := range orderWithItems.Items {
		fulfillmentItem := fulfillment.OrderItem{Quantity: item.Quantity}
		if product, err := s.deps.DbClient.BgiProduct().FindByID(ctx, item.ProductID); err == nil {
			fulfillmentItem.ProductSlug = product.Slug
		}
		// A tray-set order's "one item" is really N generated trays — the
		// manifest CSV records the set's config_hash as a stand-in for
		// STLKey (a single STL key doesn't make sense for a multi-part
		// order); the operator looks up the full tray list via
		// bgi_set_resolutions keyed by that hash.
		if item.ConfigurationID != nil {
			if cfg, err := s.deps.DbClient.BgiConfiguration().FindByID(ctx, *item.ConfigurationID); err == nil && cfg.ConfigHash != nil {
				fulfillmentItem.STLKey = "config_hash:" + *cfg.ConfigHash
			}
		}
		fo.Items = append(fo.Items, fulfillmentItem)
	}

	return fo, nil
}

func (s *server) fulfillmentAdapter() (fulfillment.Adapter, error) {
	switch s.deps.Config.Public.FulfillmentProvider {
	case models.BgiFulfillmentProviderManual, "":
		return fulfillment.NewManualAdapter(
			s.deps.AwsClient,
			s.deps.EmailClient,
			s.deps.Config.Public.S3Bucket,
			s.deps.Config.Public.OperatorEmail,
			s.deps.Config.Public.EmailFromAddress,
			"bgi",
		), nil
	default:
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
	product, err := s.deps.DbClient.BgiProduct().FindBySlug(ctx, slug)
	if err != nil {
		return uuid.Nil
	}
	return product.ID
}
