package trivai

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
)

type client struct {
	deepPriest deep_priest.DeepPriest
}

type Question struct {
	Prompt   string
	Category string
	Answer   string
}

type HowManyQuestion struct {
	Text        string
	Explanation string
	HowMany     int
}

func NewClient(deepPriest deep_priest.DeepPriest) TrivaiClient {
	return &client{
		deepPriest: deepPriest,
	}
}

func (c *client) GradeHowManyQuestion(ctx context.Context, guess int, answer int) (float64, int) {
	var correctness float64
	var offBy int

	if guess < answer {
		correctness = float64(guess) / float64(answer)
		offBy = answer - guess
	} else {
		correctness = float64(answer) / float64(guess)
		offBy = guess - answer
	}

	return correctness, offBy
}

func (c *client) GenerateNewHowManyQuestion(ctx context.Context, promptSeed string) (*HowManyQuestion, error) {
	answer, err := c.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: howManyQuestionPrompt(promptSeed),
	})
	if err != nil {
		return nil, err
	}

	parts := strings.Split(answer.Answer, promptDelimiter)

	if len(parts) < 3 {
		return nil, errors.New("botched gen job: not enough parts")
	}

	howMany, err := util.ParseNumber(parts[1])
	if err != nil {
		return nil, errors.New("botched gen job: invalid how many")
	}

	return &HowManyQuestion{
		Text:        strings.TrimSpace(parts[0]),
		HowMany:     howMany,
		Explanation: strings.TrimSpace(parts[2]),
	}, nil
}

func (c *client) GenerateQuestions(ctx context.Context) ([]Question, error) {
	answer, err := c.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: questionsPrompt(),
	})
	if err != nil {
		return nil, err
	}

	parts := strings.Split(answer.Answer, promptDelimiter)

	var questions []Question
	for _, part := range parts {
		if strings.Contains(part, categoryDelimiter) {
			questionParts := strings.Split(part, categoryDelimiter)
			questions = append(questions, Question{
				Prompt:   questionParts[0],
				Category: questionParts[1],
				Answer:   questionParts[2],
			})
		}
	}

	return questions, nil
}

func (c *client) GradeUserSubmission(ctx context.Context, questions []Question, answers []string) ([]bool, error) {
	if len(questions) != len(answers) {
		return nil, errors.New("mismatch betweeen number of questions of answers")
	}

	var grades []bool
	for i, question := range questions {
		grade, err := c.GradeAnswer(ctx, question, answers[i])
		if err != nil {
			return nil, err
		}

		grades = append(grades, grade)
	}

	return grades, nil
}

func (c *client) GradeAnswer(ctx context.Context, question Question, answer string) (bool, error) {
	prompt := gradePrompt(question.Prompt, answer)
	fmt.Println(prompt)
	isCorrect, err := c.deepPriest.PetitionTheFount(&deep_priest.Question{
		Question: prompt,
	})

	if err != nil {
		return false, err
	}

	return strings.Contains(strings.ToLower(isCorrect.Answer), "true"), nil
}
