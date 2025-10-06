package open_ai

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	openai "github.com/sashabaranov/go-openai"
)

type client struct {
	ai *openai.Client
}

type ClientConfig struct {
	ApiKey string
}

func NewClient(config ClientConfig) OpenAiClient {
	log.Println("Initializing OpenAI client")
	ai := openai.NewClient(config.ApiKey)

	return &client{
		ai: ai,
	}
}

func (c *client) GetAnswer(ctx context.Context, q string) (string, error) {
	log.Printf("Getting answer for question: %s", q)
	resp, err := c.ai.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
			Temperature: 0.1,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: q,
				},
			},
		},
	)

	if err != nil {
		log.Printf("Error getting answer: %v", err)
		return "", err
	}

	log.Printf("Successfully got answer: %s", resp.Choices[0].Message.Content)
	return resp.Choices[0].Message.Content, nil
}

func (c *client) GenerateImage(ctx context.Context, request deep_priest.GenerateImageRequest) (string, error) {
	log.Printf("Generating image with prompt: %s", request.Prompt)
	resp, err := c.ai.CreateImage(
		ctx,
		openai.ImageRequest{
			Prompt:         request.Prompt,
			N:              request.N,
			Size:           request.Size,
			Style:          request.Style,
			User:           request.User,
			Quality:        request.Quality,
			ResponseFormat: request.ResponseFormat,
			Model:          request.Model,
		},
	)

	if err != nil {
		log.Printf("Error generating image: %v", err)
		return "", err
	}
	log.Printf("Image generation response: %+v", resp)

	log.Printf("Successfully generated image with URL: %s", resp.Data[0].B64JSON)
	return resp.Data[0].B64JSON, nil
}

func (c *client) EditImage(ctx context.Context, request deep_priest.EditImageRequest) (string, error) {
	log.Printf("Editing image with prompt: %s", request.Prompt)

	resp, err := http.Get(request.ImageUrl)
	if err != nil {
		log.Printf("Error downloading image: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error downloading image: status code %d", resp.StatusCode)
		return "", fmt.Errorf("failed to download image: status code %d", resp.StatusCode)
	}

	newImage, err := c.ai.CreateEditImage(
		ctx,
		openai.ImageEditRequest{
			Prompt:         request.Prompt,
			Image:          resp.Body,
			Model:          request.Model,
			N:              request.N,
			Quality:        request.Quality,
			Size:           request.Size,
			ResponseFormat: request.ResponseFormat,
			User:           request.User,
		},
	)

	if err != nil {
		log.Printf("Error editing image: %v", err)
		return "", err
	}
	log.Printf("Image edit response: %+v", newImage)

	log.Printf("Successfully edited image with URL: %s", newImage.Data[0].B64JSON)
	return newImage.Data[0].B64JSON, nil
}

func (c *client) GetAnswerWithImage(ctx context.Context, q string, imageUrl string) (string, error) {
	log.Printf("Getting answer for question with image. Question: %s, Image URL: %s", q, imageUrl)
	resp, err := c.ai.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
			Temperature: 0.1,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleUser,
					MultiContent: []openai.ChatMessagePart{
						{
							Type: openai.ChatMessagePartTypeText,
							Text: q,
						},
						{
							Type: openai.ChatMessagePartTypeImageURL,
							ImageURL: &openai.ChatMessageImageURL{
								URL:    imageUrl,
								Detail: openai.ImageURLDetailAuto,
							},
						},
					},
				},
			},
		},
	)

	if err != nil {
		log.Printf("Error getting answer with image: %v", err)
		return "", err
	}

	log.Printf("Successfully got answer with image: %s", resp.Choices[0].Message.Content)
	return resp.Choices[0].Message.Content, nil
}
