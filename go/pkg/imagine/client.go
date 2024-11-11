package imagine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	maxImageGenerationRetries = 24
)

type ImagineClient interface {
	InitiateImageGeneration(ctx context.Context, prompt string) (*ImageGenerationResponse, error)
	GetImageGenerationStatus(ctx context.Context, id string) (*ImageGenerationResponse, error)
	GenerateImageSync(ctx context.Context, prompt string) (*ImageGenerationResponse, error)
}

type client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

func NewClient(apiKey string) ImagineClient {
	return &client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
		baseURL:    "http://cl.imagineapi.dev",
	}
}

type ImageGenerationResponse struct {
	Data struct {
		ID           string   `json:"id"`
		Prompt       string   `json:"prompt"`
		Results      string   `json:"results"`
		UserCreated  string   `json:"user_created"`
		DateCreated  string   `json:"date_created"`
		Status       string   `json:"status"`
		Progress     *float64 `json:"progress"`
		URL          string   `json:"url"`
		Error        *string  `json:"error"`
		UpscaledURLs []string `json:"upscaled_urls"`
		Upscaled     []string `json:"upscaled"`
		Ref          *string  `json:"ref"`
	} `json:"data"`
}

func (c *client) InitiateImageGeneration(ctx context.Context, prompt string) (*ImageGenerationResponse, error) {
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

	var imageGenerationResponse ImageGenerationResponse
	if err := json.Unmarshal(body, &imageGenerationResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return &imageGenerationResponse, nil
}

func (c *client) GetImageGenerationStatus(ctx context.Context, id string) (*ImageGenerationResponse, error) {
	endpoint := fmt.Sprintf("%s/items/images/%s", c.baseURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
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

	var imageGenerationResponse ImageGenerationResponse
	if err := json.Unmarshal(body, &imageGenerationResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return &imageGenerationResponse, nil
}

func (c *client) GenerateImageSync(ctx context.Context, prompt string) (*ImageGenerationResponse, error) {
	initiateImageGenerationResponse, err := c.InitiateImageGeneration(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("error initiating image generation: %w", err)
	}

	for i := 0; i < maxImageGenerationRetries; i++ {
		response, err := c.GetImageGenerationStatus(ctx, initiateImageGenerationResponse.Data.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting image generation status: %w", err)
		}

		if response.Data.Status == "completed" || response.Data.Status == "failed" {
			fmt.Println("Completed image details:")
			fmt.Printf("%+v\n", response.Data)
			return response, nil
		} else {
			fmt.Printf("Image is not finished generation. Status: %s (Attempt %d/%d)\n", response.Data.Status, i+1, maxImageGenerationRetries)
			time.Sleep(5 * time.Second)
		}
	}
	return nil, fmt.Errorf("image generation timed out after %d attempts", maxImageGenerationRetries)
}
