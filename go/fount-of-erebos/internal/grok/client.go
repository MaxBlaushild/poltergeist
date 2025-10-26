package grok

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
)

type client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

type ClientConfig struct {
	ApiKey string
}

func NewClient(config ClientConfig) GrokClient {
	log.Println("Initializing Grok client")
	return &client{
		apiKey:     config.ApiKey,
		httpClient: &http.Client{},
		baseURL:    "https://api.x.ai/v1/images/generations",
	}
}

// grokImageRequest matches the expected Grok API request format
type grokImageRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
	N      int    `json:"n,omitempty"`
}

// grokImageResponse matches the Grok API response format
type grokImageResponse struct {
	Data []struct {
		URL           string `json:"url,omitempty"`
		B64JSON       string `json:"b64_json,omitempty"`
		RevisedPrompt string `json:"revised_prompt,omitempty"`
	} `json:"data"`
	Created int `json:"created"`
}

func (c *client) GenerateImage(ctx context.Context, request deep_priest.GenerateImageRequest) (string, error) {
	log.Printf("Generating image with Grok. Prompt: %s", request.Prompt)

	// Build the Grok API request
	grokReq := grokImageRequest{
		Prompt: request.Prompt,
		Model:  "grok-2-image-1212",
		N:      request.N,
	}

	// Default to 1 image if not specified
	if grokReq.N == 0 {
		grokReq.N = 1
	}

	jsonData, err := json.Marshal(grokReq)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// Make the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("Error making request to Grok API: %v", err)
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		log.Printf("Grok API returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("grok API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var grokResp grokImageResponse
	if err := json.Unmarshal(body, &grokResp); err != nil {
		log.Printf("Error unmarshaling response: %v, body: %s", err, string(body))
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(grokResp.Data) == 0 {
		log.Printf("No images returned in response")
		return "", fmt.Errorf("no images generated")
	}

	// Return the first image (URL or base64)
	// Prefer b64_json if available (matching OpenAI behavior)
	imageData := grokResp.Data[0]
	if imageData.B64JSON != "" {
		log.Printf("Successfully generated image with Grok (base64, length: %d)", len(imageData.B64JSON))
		return imageData.B64JSON, nil
	}

	if imageData.URL != "" {
		log.Printf("Successfully generated image with Grok (URL: %s)", imageData.URL)
		return imageData.URL, nil
	}

	return "", fmt.Errorf("no image data in response")
}
