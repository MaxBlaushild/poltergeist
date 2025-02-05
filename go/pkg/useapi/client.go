package useapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type ImageOptionsGenerationRequest struct {
	Prompt             string `json:"prompt"`
	WebhookURL         string `json:"webhook_url"`
	WebhookType        string `json:"webhook_type"`
	AccountHash        string `json:"account_hash"`
	IsDisablePrefilter bool   `json:"is_disable_prefilter"`
}

type ImageOptionsGenerationStatusResponse struct {
	AccountHash string `json:"account_hash"`
	Hash        string `json:"hash"`
	WebhookURL  string `json:"webhook_url"`
	WebhookType string `json:"webhook_type"`
	Prompt      string `json:"prompt"`
	Type        string `json:"type"`
	Progress    int    `json:"progress"`
	Status      string `json:"status"`
	Result      struct {
		URL         string `json:"url"`
		ProxyURL    string `json:"proxy_url"`
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
		Size        int    `json:"size"`
	} `json:"result"`
	StatusReason    *string  `json:"status_reason"`
	PrefilterResult []string `json:"prefilter_result"`
	CreatedAt       string   `json:"created_at"`
}

type ImageUpscaleResponse struct {
	AccountHash string  `json:"account_hash"`
	Hash        string  `json:"hash"`
	WebhookURL  *string `json:"webhook_url"`
	WebhookType *string `json:"webhook_type"`
	Choice      string  `json:"choice"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	Result      struct {
		URL         string `json:"url"`
		ProxyURL    string `json:"proxy_url"`
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
		Size        int    `json:"size"`
	} `json:"result"`
	StatusReason *string `json:"status_reason"`
	CreatedAt    string  `json:"created_at"`
}

type ImageGenerationResponse struct {
	Hash string `json:"hash"`
}

const (
	useApiBaseUrl = "https://api.userapi.ai"
)

type Client interface {
	GenerateImageOptions(ctx context.Context, prompt string) (*ImageGenerationResponse, error)
	CheckImageGenerationOptionsStatus(ctx context.Context, id string) (*ImageOptionsGenerationStatusResponse, error)
	UpscaleImage(ctx context.Context, id string, choice int) (*ImageGenerationResponse, error)
	CheckUpscaleImageStatus(ctx context.Context, id string) (*ImageUpscaleResponse, error)
}

type useApiClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) Client {
	return &useApiClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (c *useApiClient) doRequest(ctx context.Context, method, endpoint string, requestBody interface{}, response interface{}) error {
	var bodyReader io.Reader
	if requestBody != nil {
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("error marshaling request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	fmt.Println("body")
	fmt.Println(string(body))
	if err := json.Unmarshal(body, response); err != nil {
		return fmt.Errorf("error unmarshalling response: %w", err)
	}

	return nil
}

func (c *useApiClient) GenerateImageOptions(ctx context.Context, prompt string) (*ImageGenerationResponse, error) {
	endpoint := fmt.Sprintf("%s/midjourney/v2/imagine", useApiBaseUrl)
	requestBody := map[string]string{
		"prompt": prompt,
	}

	var response ImageGenerationResponse
	if err := c.doRequest(ctx, http.MethodPost, endpoint, requestBody, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *useApiClient) CheckImageGenerationOptionsStatus(ctx context.Context, id string) (*ImageOptionsGenerationStatusResponse, error) {
	endpoint := fmt.Sprintf("%s/midjourney/v2/status?hash=%s", useApiBaseUrl, id)
	var response ImageOptionsGenerationStatusResponse
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *useApiClient) UpscaleImage(ctx context.Context, id string, choice int) (*ImageGenerationResponse, error) {
	endpoint := fmt.Sprintf("%s/midjourney/v2/upscale", useApiBaseUrl)
	requestBody := map[string]interface{}{
		"hash":   id,
		"choice": choice,
	}

	var response ImageGenerationResponse
	if err := c.doRequest(ctx, http.MethodPost, endpoint, requestBody, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *useApiClient) CheckUpscaleImageStatus(ctx context.Context, id string) (*ImageUpscaleResponse, error) {
	endpoint := fmt.Sprintf("%s/midjourney/v2/status?hash=%s", useApiBaseUrl, id)
	var response ImageUpscaleResponse
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}
