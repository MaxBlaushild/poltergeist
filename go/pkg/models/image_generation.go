package models

import (
	"time"

	"github.com/google/uuid"
)

type GenerationStatus int

const (
	GenerationStatusUnidentified GenerationStatus = iota
	GenerationStatusPending
	GenerationStatusInProgress
	GenerationStatusComplete
	GenerationStatusFailed
)

type GenerationBackend int

const (
	GenerationBackendUnidentified GenerationBackend = iota
	GenerationBackendImagine
)

type ImageGeneration struct {
	ID                  uuid.UUID         `json:"id"`
	CreatedAt           time.Time         `json:"createdAt"`
	UpdatedAt           time.Time         `json:"updatedAt"`
	UserID              uuid.UUID         `json:"userID"`
	GenerationID        string            `json:"generationID"`
	GenerationBackendID GenerationBackend `json:"generationBackendID"`
	Status              GenerationStatus  `json:"status"`
	OptionOne           *string           `json:"optionOne"`
	OptionTwo           *string           `json:"optionTwo"`
	OptionThree         *string           `json:"optionThree"`
	OptionFour          *string           `json:"optionFour"`
}
