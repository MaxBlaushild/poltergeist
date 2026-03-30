package dungeonmaster

import "errors"

type nonRetriableQuestGenerationError struct {
	err error
}

func (e *nonRetriableQuestGenerationError) Error() string {
	if e == nil || e.err == nil {
		return "non-retriable quest generation error"
	}
	return e.err.Error()
}

func (e *nonRetriableQuestGenerationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func markNonRetriableQuestGenerationError(err error) error {
	if err == nil {
		return nil
	}
	var target *nonRetriableQuestGenerationError
	if errors.As(err, &target) {
		return err
	}
	return &nonRetriableQuestGenerationError{err: err}
}

func IsNonRetriableQuestGenerationError(err error) bool {
	var target *nonRetriableQuestGenerationError
	return errors.As(err, &target)
}
