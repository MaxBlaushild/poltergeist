package trivai

import (
	"fmt"
)

var (
	promptDelimiter   = ":::"
	categoryDelimiter = "---"
	dash              = "-"
)

func template() string {
	return fmt.Sprintf("%squestion%scategory%s", promptDelimiter, categoryDelimiter, promptDelimiter)
}

func questionsPrompt() string {
	return fmt.Sprintf(
		"Can you please generate six difficult trivia questions for me. Interpolate them into this template: \":::{prompt}---{category}---{answer}:::\". Don't include any numbers, quotations or newlines in your response.",
	)
}

func gradePrompt(question string, answer string) string {
	return fmt.Sprintf(
		`Given the question "%s", would you consider the answer "%s" correct, allowing for misspellings, and only requiring the last name if the answer is a person? If correct, reply with only "true". If not correct, reply with only "false."`,
		question,
		answer,
	)
}

func howManyQuestionPrompt(promptString string) string {
	return fmt.Sprintf("Can you please generate a trivia question in which someone has to guess how many of something exists? %s Please respond with the question, followed by three colons, followed by the answer as a number, followed by three colons, follow by a bit of flavor text to explain why the number is what it is.", promptString)
}
