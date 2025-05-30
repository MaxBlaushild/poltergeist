// Mock DeepPriest client for testing purposes.
package mocks

import (
	// "context" // Removed unused import

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
)

// MockDeepPriestClient is a mock implementation of the deep_priest.DeepPriest interface.
type MockDeepPriestClient struct {
	PetitionTheFountFunc           func(question *deep_priest.Question) (*deep_priest.Answer, error)
	PetitionTheFountWithImageFunc  func(question *deep_priest.QuestionWithImage) (*deep_priest.Answer, error)
	GenerateImageFunc              func(request deep_priest.GenerateImageRequest) (string, error)
}

// PetitionTheFount mocks the PetitionTheFount operation.
func (m *MockDeepPriestClient) PetitionTheFount(question *deep_priest.Question) (*deep_priest.Answer, error) {
	if m.PetitionTheFountFunc != nil {
		return m.PetitionTheFountFunc(question)
	}
	return &deep_priest.Answer{}, nil
}

// PetitionTheFountWithImage mocks the PetitionTheFountWithImage operation.
func (m *MockDeepPriestClient) PetitionTheFountWithImage(question *deep_priest.QuestionWithImage) (*deep_priest.Answer, error) {
	if m.PetitionTheFountWithImageFunc != nil {
		return m.PetitionTheFountWithImageFunc(question)
	}
	return &deep_priest.Answer{}, nil
}

// GenerateImage mocks the GenerateImage operation.
func (m *MockDeepPriestClient) GenerateImage(request deep_priest.GenerateImageRequest) (string, error) {
	if m.GenerateImageFunc != nil {
		return m.GenerateImageFunc(request)
	}
	return "http://example.com/mocked_image.png", nil
}
