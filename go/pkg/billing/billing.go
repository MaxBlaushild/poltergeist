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
}

const (
	baseUrl = "http://localhost:8022"
)

type CheckoutSessionParams struct {
	SuccessUrl  string            `json:"successUrl" binding:"required"`
	CancelUrl   string            `json:"cancelUrl" binding:"required"`
	PlanID      string            `json:"planId" binding:"required"`
	CallbackUrl string            `json:"callbackUrl" binding:"required"`
	Metadata    map[string]string `json:"metadata"`
}

type OnSubscribe struct {
	Metadata map[string]string `json:"metadata"`
}

type CheckoutSessionResponse struct {
	URL string `json:"url" binding:"required"`
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
