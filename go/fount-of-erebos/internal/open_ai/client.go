package open_ai

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type client struct {
	ai *openai.Client
}

type ClientConfig struct {
	ApiKey string
}

func NewClient(config ClientConfig) OpenAiClient {
	ai := openai.NewClient(config.ApiKey)

	return &client{
		ai: ai,
	}
}

func (c *client) GetAnswer(ctx context.Context, q string) (string, error) {
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
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *client) GetAnswerWithImage(ctx context.Context, q string, imageUrl string) (string, error) {
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
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
