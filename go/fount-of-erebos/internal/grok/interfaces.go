package grok

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
)

type GrokClient interface {
	GenerateImage(ctx context.Context, request deep_priest.GenerateImageRequest) (string, error)
}
