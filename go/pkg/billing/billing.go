package billing

import (
	"context"
	"encoding/json"

	"github.com/MaxBlaushild/poltergeist/pkg/http"
)

type client struct {
	httpClient http.Client
}

type Client interface {
	NewCheckoutSession(ctx context.Context, params *CheckoutSessionParams) (*CheckoutSessionResponse, error)
	NewPaymentCheckoutSession(ctx context.Context, params *PaymentCheckoutSessionParams) (*CheckoutSessionResponse, error)
	CancelSubscription(ctx context.Context, params *CancelSubscriptionParams) (*CancelSubscriptionResponse, error)
}

const (
	baseUrl = "http://localhost:8022"
)

type CheckoutSessionParams struct {
	SessionSuccessRedirectUrl     string            `json:"successUrl" binding:"required"`
	SessionCancelRedirectUrl      string            `json:"cancelUrl" binding:"required"`
	PlanID                        string            `json:"planId" binding:"required"`
	SubscriptionCreateCallbackUrl string            `json:"subscriptionCreateCallbackUrl" binding:"required"`
	SubscriptionCancelCallbackUrl string            `json:"subscriptionCancelCallbackUrl" binding:"required"`
	Metadata                      map[string]string `json:"metadata"`
}

type CancelSubscriptionParams struct {
	StripeID string `json:"stripeId" binding:"required"`
}

type OnSubscribe struct {
	Metadata       map[string]string `json:"metadata"`
	SubscriptionID string            `json:"subscriptionId"`
}

type OnSubscriptionDelete struct {
	Metadata       map[string]string `json:"metadata"`
	SubscriptionID string            `json:"subscriptionId"`
}

// PaymentLineItem is one itemized Stripe Checkout line (R-2.8/R-6.2: the
// customer sees what they're paying for, not one lump sum).
type PaymentLineItem struct {
	Name          string `json:"name" binding:"required"`
	AmountInCents int64  `json:"amountInCents" binding:"required"`
	Quantity      int64  `json:"quantity" binding:"required"`
}

type PaymentCheckoutSessionParams struct {
	SessionSuccessRedirectUrl string `json:"successUrl" binding:"required"`
	SessionCancelRedirectUrl  string `json:"cancelUrl" binding:"required"`
	// AmountInCents is the original single-line-item shape, kept for
	// existing callers. Set LineItems instead for an itemized session; if
	// both are set, LineItems wins. Exactly one must be set.
	AmountInCents int64             `json:"amountInCents"`
	LineItems     []PaymentLineItem `json:"lineItems"`
	// AutomaticTax enables Stripe Tax on the session (R-2.8). Off by
	// default so existing callers see no behavior change.
	AutomaticTax bool `json:"automaticTax"`
	// CollectShippingAddress adds Stripe's shipping address collection step
	// (US-only, matching R-1.2's no-international-shipping scope for v1).
	CollectShippingAddress     bool              `json:"collectShippingAddress"`
	PaymentCompleteCallbackUrl string            `json:"paymentCompleteCallbackUrl" binding:"required"`
	Metadata                   map[string]string `json:"metadata"`
}

type ShippingAddress struct {
	Name       string `json:"name"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
}

type OnPaymentComplete struct {
	Metadata        map[string]string `json:"metadata"`
	SessionID       string            `json:"sessionId"`
	AmountInCents   int64             `json:"amountInCents"`
	CustomerEmail   string            `json:"customerEmail"`
	ShippingAddress *ShippingAddress  `json:"shippingAddress,omitempty"`
}

type CheckoutSessionResponse struct {
	URL string `json:"url" binding:"required"`
}

type CancelSubscriptionResponse struct {
}

func NewClient() Client {
	httpClient := http.NewClient(baseUrl, http.ApplicationJson)
	return &client{httpClient: httpClient}
}

func (c *client) NewCheckoutSession(ctx context.Context, params *CheckoutSessionParams) (*CheckoutSessionResponse, error) {
	respBytes, err := c.httpClient.Post(ctx, "/billing/checkout-session", params)
	if err != nil {
		return nil, err
	}

	var res CheckoutSessionResponse
	err = json.Unmarshal(respBytes, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *client) NewPaymentCheckoutSession(ctx context.Context, params *PaymentCheckoutSessionParams) (*CheckoutSessionResponse, error) {
	respBytes, err := c.httpClient.Post(ctx, "/billing/payment-checkout-session", params)
	if err != nil {
		return nil, err
	}

	var res CheckoutSessionResponse
	err = json.Unmarshal(respBytes, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *client) CancelSubscription(ctx context.Context, params *CancelSubscriptionParams) (*CancelSubscriptionResponse, error) {
	respBytes, err := c.httpClient.Post(ctx, "/billing/subscriptions/cancel", params)
	if err != nil {
		return nil, err
	}

	var res CancelSubscriptionResponse
	err = json.Unmarshal(respBytes, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
