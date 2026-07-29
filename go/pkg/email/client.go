package email

// Sends through Twilio's Comms API (https://comms.twilio.com/v1/Emails),
// which authenticates with the same Account SID / Auth Token as Twilio's
// SMS API rather than a separate SendGrid API key.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const sendEmailURL = "https://comms.twilio.com/v1/Emails"

type client struct {
	httpClient  *http.Client
	accountSid  string
	authToken   string
	fromAddress string
	fromName    string
	webHost     string
}

type ClientConfig struct {
	AccountSid  string
	AuthToken   string
	FromAddress string
	FromName    string
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
	fromName := cfg.FromName
	if fromName == "" {
		fromName = "Max Blaushild"
	}
	return &client{
		httpClient:  http.DefaultClient,
		accountSid:  cfg.AccountSid,
		authToken:   cfg.AuthToken,
		fromAddress: cfg.FromAddress,
		fromName:    fromName,
		webHost:     cfg.WebHost,
	}
}

type emailAddress struct {
	Address string `json:"address"`
	Name    string `json:"name,omitempty"`
}

type emailContent struct {
	Subject string `json:"subject"`
	Html    string `json:"html,omitempty"`
	Text    string `json:"text,omitempty"`
}

type sendEmailRequest struct {
	From    emailAddress   `json:"from"`
	To      []emailAddress `json:"to"`
	Content emailContent   `json:"content"`
}

func (c *client) SendMail(email Email) error {
	reqBody, err := json.Marshal(sendEmailRequest{
		From: emailAddress{Address: c.fromAddress, Name: c.fromName},
		To:   []emailAddress{{Address: email.Email, Name: email.Name}},
		Content: emailContent{
			Subject: email.Subject,
			Html:    email.HtmlContent,
			Text:    email.PlainTextContent,
		},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, sendEmailURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.accountSid, c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("twilio comms email send failed with status %d: %s", resp.StatusCode, respBody)
	}

	return nil
}
