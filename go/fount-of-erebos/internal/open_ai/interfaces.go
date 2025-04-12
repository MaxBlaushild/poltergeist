package open_ai

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
)

type OpenAiClient interface {
	GetAnswer(ctx context.Context, q string) (string, error)
	GetAnswerWithImage(ctx context.Context, q string, imageUrl string) (string, error)
	GenerateImage(ctx context.Context, request deep_priest.ImageGenerationRequest) (string, error)
}
