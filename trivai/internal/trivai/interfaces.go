package trivai

import (
	"context"
)

type TrivaiClient interface {
	GenerateQuestions(ctx context.Context) ([]Question, error)
	GradeUserSubmission(ctx context.Context, questionSet []Question, answers []string) ([]bool, error)
	GradeAnswer(ctx context.Context, question Question, answer string) (bool, error)
	GenerateNewHowManyQuestion(ctx context.Context, promptSeed string) (*HowManyQuestion, error)
	GradeHowManyQuestion(ctx context.Context, guess int, answer int) (float64, int)
}
