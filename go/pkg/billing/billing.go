package billing

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type client struct{}

type Client interface {
	NewCheckoutSession(params *CheckoutSessionParams) (*CheckoutSessionResponse, error)
}

const (
	baseUrl = "http://localhost:8022/billing"
)

func NewClient() Client {
	return &client{}
}

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

func (c *client) NewCheckoutSession(params *CheckoutSessionParams) (*CheckoutSessionResponse, error) {
	jsonBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(baseUrl+"/checkout-session", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res CheckoutSessionResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
