package deep_priest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type deepPriest struct {
	baseURL       string
	consultClient *http.Client
	imageClient   *http.Client
}

type DeepPriest interface {
	PetitionTheFount(*Question) (*Answer, error)
	PetitionTheFountWithImage(*QuestionWithImage) (*Answer, error)
	GenerateImage(request GenerateImageRequest) (string, error)
	EditImage(request EditImageRequest) (string, error)
}

type GenerateImageRequest struct {
	Prompt         string `json:"prompt,omitempty"`
	Model          string `json:"model,omitempty"`
	N              int    `json:"n,omitempty"`
	Quality        string `json:"quality,omitempty"`
	Size           string `json:"size,omitempty"`
	Style          string `json:"style,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	User           string `json:"user,omitempty"`
}

type EditImageRequest struct {
	Prompt         string `json:"prompt,omitempty"`
	Model          string `json:"model,omitempty"`
	N              int    `json:"n,omitempty"`
	Quality        string `json:"quality,omitempty"`
	Size           string `json:"size,omitempty"`
	Style          string `json:"style,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	User           string `json:"user,omitempty"`
	ImageUrl       string `json:"image_url,omitempty"`
}

type Question struct {
	Question string `json:"question" binding:"required"`
}

type Answer struct {
	Answer string `json:"answer"`
}

type ImageGenerationResponse struct {
	ImageUrl string `json:"imageUrl"`
}

type QuestionWithImage struct {
	Question string `json:"question" binding:"required"`
	Image    string `json:"image" binding:"required"`
}

const (
	defaultBaseURL         = "http://localhost:8081"
	defaultConsultTimeout  = 30 * time.Second
	defaultImageGenTimeout = 2 * time.Minute
)

func SummonDeepPriest() DeepPriest {
	return &deepPriest{
		baseURL:       defaultBaseURL,
		consultClient: &http.Client{Timeout: defaultConsultTimeout},
		imageClient:   &http.Client{Timeout: defaultImageGenTimeout},
	}
}

func (d *deepPriest) PetitionTheFount(question *Question) (*Answer, error) {
	var answer Answer
	if err := d.postJSON(
		d.consultHTTPClient(),
		"consultation",
		"/consult",
		question,
		&answer,
	); err != nil {
		return nil, err
	}
	return &answer, nil
}

func (d *deepPriest) PetitionTheFountWithImage(question *QuestionWithImage) (*Answer, error) {
	var answer Answer
	if err := d.postJSON(
		d.consultHTTPClient(),
		"image consultation",
		"/consultWithImage",
		question,
		&answer,
	); err != nil {
		return nil, err
	}
	return &answer, nil
}

func (d *deepPriest) GenerateImage(request GenerateImageRequest) (string, error) {
	ApplyGenerateImageDefaults(&request)
	var response ImageGenerationResponse
	if err := d.postJSON(
		d.imageHTTPClient(),
		"image generation",
		"/generateImage",
		request,
		&response,
	); err != nil {
		return "", err
	}
	if strings.TrimSpace(response.ImageUrl) == "" {
		return "", fmt.Errorf("image generation returned empty payload")
	}

	return response.ImageUrl, nil
}

func (d *deepPriest) EditImage(request EditImageRequest) (string, error) {
	ApplyEditImageDefaults(&request)
	var response ImageGenerationResponse
	if err := d.postJSON(
		d.imageHTTPClient(),
		"image edit",
		"/editImage",
		request,
		&response,
	); err != nil {
		return "", err
	}
	if strings.TrimSpace(response.ImageUrl) == "" {
		return "", fmt.Errorf("image edit returned empty payload")
	}

	return response.ImageUrl, nil
}

func (d *deepPriest) normalizedBaseURL() string {
	if strings.TrimSpace(d.baseURL) == "" {
		return defaultBaseURL
	}
	return strings.TrimRight(strings.TrimSpace(d.baseURL), "/")
}

func (d *deepPriest) consultHTTPClient() *http.Client {
	if d.consultClient == nil {
		d.consultClient = &http.Client{Timeout: defaultConsultTimeout}
	}
	return d.consultClient
}

func (d *deepPriest) imageHTTPClient() *http.Client {
	if d.imageClient == nil {
		d.imageClient = &http.Client{Timeout: defaultImageGenTimeout}
	}
	return d.imageClient
}

func (d *deepPriest) postJSON(
	client *http.Client,
	operationName string,
	path string,
	payload interface{},
	out interface{},
) error {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal %s request: %w", operationName, err)
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		d.normalizedBaseURL()+path,
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return fmt.Errorf("failed to build %s request: %w", operationName, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return d.normalizeRequestError(operationName, path, client.Timeout, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read %s response: %w", operationName, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf(
			"%s failed: status %d: %s",
			operationName,
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf(
			"failed to decode %s response (%s): %w",
			operationName,
			strings.TrimSpace(string(body)),
			err,
		)
	}
	return nil
}

func (d *deepPriest) normalizeRequestError(
	operationName string,
	path string,
	timeout time.Duration,
	err error,
) error {
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		if timeout > 0 {
			return fmt.Errorf(
				"%s timed out after %s while calling %s: %w",
				operationName,
				timeout,
				path,
				err,
			)
		}
		return fmt.Errorf("%s timed out while calling %s: %w", operationName, path, err)
	}
	return fmt.Errorf("%s request failed for %s: %w", operationName, path, err)
}
