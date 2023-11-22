package texter

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/http"
)

type Text struct {
	Body     string `json:"body" binding:"required"`
	To       string `json:"to" binding:"required"`
	From     string `json:"from" binding:"required"`
	TextType string `json:"textType" binding:"required"`
}

type client struct {
	httpClient http.Client
}

type Client interface {
	Text(context.Context, *Text) error
}

const (
	baseUrl = "http://localhost:8084"
)

func NewClient() Client {
	httpClient := http.NewClient(baseUrl, http.ApplicationJson)
	return &client{httpClient: httpClient}
}

func (c *client) Text(ctx context.Context, text *Text) error {
	_, err := c.httpClient.Post(ctx, "/texter/send-sms", text)
	if err != nil {
		return err
	}

	return nil
}
