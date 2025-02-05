package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const discordApiBaseUrl = "https://discord.com/api/v10"

type DiscordClient interface {
	SendMessage(ctx context.Context, message string) error
}

type discordClient struct {
	authToken  string
	channelId  string
	httpClient *http.Client
}

type messagePayload struct {
	Content string `json:"content"`
}

func NewClient(authToken, channelId string) DiscordClient {
	return &discordClient{
		authToken:  authToken,
		channelId:  channelId,
		httpClient: &http.Client{},
	}
}

func (c *discordClient) SendMessage(ctx context.Context, message string) error {
	payload := messagePayload{
		Content: message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message payload: %w", err)
	}

	url := fmt.Sprintf("%s/channels/%s/messages", discordApiBaseUrl, c.channelId)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", c.authToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord API returned non-2xx status code: %d", resp.StatusCode)
	}

	return nil
}

