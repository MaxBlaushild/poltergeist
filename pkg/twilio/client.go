package twilio

import (
	"context"

	"github.com/twilio/twilio-go"

	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type client struct {
	twilioClient *twilio.RestClient
}

type Text struct {
	To   string
	Text string
	From string
}

type Client interface {
	SendText(ctx context.Context, text *Text) error
}

type ClientConfig struct {
	AccountSid string
	AuthToken  string
}

func NewClient(cfg *ClientConfig) Client {
	twilioClient := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: cfg.AccountSid,
		Password: cfg.AuthToken,
	})

	return &client{
		twilioClient: twilioClient,
	}
}

func (c *client) SendText(ctx context.Context, text *Text) error {
	params := &twilioApi.CreateMessageParams{}
	params.SetTo(text.To)
	params.SetFrom(text.From)
	params.SetBody(text.Text)

	_, err := c.twilioClient.Api.CreateMessage(params)
	return err
}
