package email

// using SendGrid's Go Library
// https://github.com/sendgrid/sendgrid-go

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type client struct {
	sendgridClient *sendgrid.Client
	fromAddress    *mail.Email
	webHost        string
}

type ClientConfig struct {
	ApiKey      string
	FromAddress string
	WebHost     string
}

type Email struct {
	Subject          string
	Name             string
	Email            string
	PlainTextContent string
	HtmlContent      string
}

func NewClient(cfg ClientConfig) EmailClient {
	sendgridClient := sendgrid.NewSendClient(cfg.ApiKey)

	return &client{
		sendgridClient: sendgridClient,
		fromAddress:    mail.NewEmail("Max Blaushild", cfg.FromAddress),
		webHost:        cfg.WebHost,
	}
}

func (c *client) SendMail(email Email) error {
	to := mail.NewEmail(email.Name, email.Email)
	message := mail.NewSingleEmail(c.fromAddress, email.Subject, to, email.PlainTextContent, email.HtmlContent)
	if _, err := c.sendgridClient.Send(message); err != nil {
		return err
	}

	return nil
}
