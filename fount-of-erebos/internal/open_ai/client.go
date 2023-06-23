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
			Model: openai.GPT3Dot5Turbo0301,
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
