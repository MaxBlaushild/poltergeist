package open_ai

import (
	"context"
	"log"

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

	log.Printf("Successfully generated image with URL: %s", resp.Data[0].URL)
	return resp.Data[0].URL, nil
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
