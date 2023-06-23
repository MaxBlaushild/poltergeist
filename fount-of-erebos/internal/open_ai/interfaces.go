package open_ai

import "context"

type OpenAiClient interface {
	GetAnswer(ctx context.Context, q string) (string, error)
}
