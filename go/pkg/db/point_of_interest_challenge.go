package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pointOfInterestChallengeHandle struct {
	db *gorm.DB
}

func (p *pointOfInterestChallengeHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestChallenge, error) {
	var challenge models.PointOfInterestChallenge
	if err := p.db.WithContext(ctx).First(&challenge, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (p *pointOfInterestChallengeHandle) SubmitAnswerForChallenge(ctx context.Context, challengeID uuid.UUID, teamID uuid.UUID, text string, imageURL string, isCorrect bool) (*models.PointOfInterestChallengeSubmission, error) {
	challenge := models.PointOfInterestChallengeSubmission{
		PointOfInterestChallengeID: challengeID,
		TeamID:                     teamID,
		Text:                       text,
		ImageURL:                   imageURL,
		IsCorrect:                  &isCorrect,
		CreatedAt:                  time.Now(),
		UpdatedAt:                  time.Now(),
		ID:                         uuid.New(),
	}

	if err := p.db.WithContext(ctx).Create(&challenge).Error; err != nil {
		return nil, err
	}

	return &challenge, nil
}
