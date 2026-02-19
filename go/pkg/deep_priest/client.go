package deep_priest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type deepPriest struct{}

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
	baseUrl = "http://localhost:8081"
)

func SummonDeepPriest() DeepPriest {
	return &deepPriest{}
}

func (d *deepPriest) PetitionTheFount(question *Question) (*Answer, error) {
	jsonBody, err := json.Marshal(question)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(baseUrl+"/consult", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var answer Answer
	err = json.Unmarshal(body, &answer)
	if err != nil {
		return nil, err
	}

	return &answer, nil
}

func (d *deepPriest) PetitionTheFountWithImage(question *QuestionWithImage) (*Answer, error) {
	jsonBody, err := json.Marshal(question)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(baseUrl+"/consultWithImage", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var answer Answer
	err = json.Unmarshal(body, &answer)
	if err != nil {
		return nil, err
	}

	return &answer, nil
}

func (d *deepPriest) GenerateImage(request GenerateImageRequest) (string, error) {
	ApplyGenerateImageDefaults(&request)
	jsonBody, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(baseUrl+"/generateImage", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("image generation failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var response ImageGenerationResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal image generation response: %w", err)
	}
	if strings.TrimSpace(response.ImageUrl) == "" {
		return "", fmt.Errorf("image generation returned empty payload: %s", strings.TrimSpace(string(body)))
	}

	return response.ImageUrl, nil
}

func (d *deepPriest) EditImage(request EditImageRequest) (string, error) {
	ApplyEditImageDefaults(&request)
	jsonBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(baseUrl+"/editImage", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("image edit failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var response ImageGenerationResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if strings.TrimSpace(response.ImageUrl) == "" {
		return "", fmt.Errorf("image edit returned empty payload: %s", strings.TrimSpace(string(body)))
	}

	return response.ImageUrl, nil
}
