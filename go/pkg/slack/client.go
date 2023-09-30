package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type client struct {
	webhookUrl string
}

type SlackMessage struct {
	Text string `json:"text"`
}

type SlackClient interface {
	Post(ctx context.Context, text *SlackMessage) error
}

func NewSlackClient(webhookUrl string) SlackClient {
	return &client{
		webhookUrl: webhookUrl,
	}
}

func (s *client) Post(ctx context.Context, msg *SlackMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Failed to encode message payload:", err)
		return err
	}

	resp, err := http.Post(s.webhookUrl, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Failed to send message:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: received non-200 status code, got %d\n", resp.StatusCode)
		return fmt.Errorf("Error: received non-200 status code, got %d\n", resp.StatusCode)
	}

	return nil
}
