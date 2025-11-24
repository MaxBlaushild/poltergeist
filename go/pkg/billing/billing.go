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

type PaymentCheckoutSessionParams struct {
	SessionSuccessRedirectUrl  string            `json:"successUrl" binding:"required"`
	SessionCancelRedirectUrl   string            `json:"cancelUrl" binding:"required"`
	AmountInCents              int64             `json:"amountInCents" binding:"required"`
	PaymentCompleteCallbackUrl string            `json:"paymentCompleteCallbackUrl" binding:"required"`
	Metadata                   map[string]string `json:"metadata"`
}

type OnPaymentComplete struct {
	Metadata      map[string]string `json:"metadata"`
	SessionID     string            `json:"sessionId"`
	AmountInCents int64             `json:"amountInCents"`
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
