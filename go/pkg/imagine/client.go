package imagine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ImagineClient interface {
	GenerateImage(ctx context.Context, prompt string) ([]byte, error)
}

type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
		baseURL:    "http://cl.imagineapi.dev",
	}
}

// GenerateImage sends a request to generate an image based on the given prompt
func (c *Client) GenerateImage(ctx context.Context, prompt string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/items/images/", c.baseURL)

	requestBody, err := json.Marshal(map[string]string{
		"prompt": prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
