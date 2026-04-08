package judge

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
)

type stubDeepPriest struct {
	imageCall func(*deep_priest.QuestionWithImage) (*deep_priest.Answer, error)
	textCall  func(*deep_priest.Question) (*deep_priest.Answer, error)
}

func (s *stubDeepPriest) PetitionTheFount(question *deep_priest.Question) (*deep_priest.Answer, error) {
	if s.textCall != nil {
		return s.textCall(question)
	}
	return &deep_priest.Answer{Answer: `{"score": 10, "reason": "ok"}`}, nil
}

func (s *stubDeepPriest) PetitionTheFountWithImage(question *deep_priest.QuestionWithImage) (*deep_priest.Answer, error) {
	if s.imageCall != nil {
		return s.imageCall(question)
	}
	return &deep_priest.Answer{Answer: `{"score": 10, "reason": "ok"}`}, nil
}

func (s *stubDeepPriest) GenerateImage(request deep_priest.GenerateImageRequest) (string, error) {
	return "", nil
}

func (s *stubDeepPriest) EditImage(request deep_priest.EditImageRequest) (string, error) {
	return "", nil
}

func TestJudgeFreeformHonorsContextTimeoutForImageSubmissions(t *testing.T) {
	t.Parallel()

	client := &client{
		deepPriest: &stubDeepPriest{
			imageCall: func(*deep_priest.QuestionWithImage) (*deep_priest.Answer, error) {
				time.Sleep(200 * time.Millisecond)
				return &deep_priest.Answer{Answer: `{"score": 12, "reason": "ok"}`}, nil
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	startedAt := time.Now()
	_, err := client.JudgeFreeform(ctx, FreeformJudgeSubmissionRequest{
		Question:           "Take a picture of a bakery.",
		ImageSubmissionUrl: "https://example.com/bakery.png",
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}
	if elapsed := time.Since(startedAt); elapsed > 150*time.Millisecond {
		t.Fatalf("expected timeout to return quickly, took %s", elapsed)
	}
}
