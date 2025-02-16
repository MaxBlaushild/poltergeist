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

func (p *pointOfInterestChallengeHandle) Delete(ctx context.Context, id uuid.UUID) error {
	// First delete all related submissions
	if err := p.db.Delete(&models.PointOfInterestChallengeSubmission{}, "point_of_interest_challenge_id = ?", id).Error; err != nil {
		return err
	}
	// Then delete the challenge itself
	return p.db.Delete(&models.PointOfInterestChallenge{}, "id = ?", id).Error
}

func (p *pointOfInterestChallengeHandle) Edit(ctx context.Context, id uuid.UUID, question string, inventoryItemID int, tier int) (*models.PointOfInterestChallenge, error) {
	challenge := models.PointOfInterestChallenge{
		Question:        question,
		InventoryItemID: inventoryItemID,
		Tier:            tier,
		UpdatedAt:       time.Now(),
	}

	if err := p.db.Model(&models.PointOfInterestChallenge{}).Where("id = ?", id).Updates(challenge).Error; err != nil {
		return nil, err
	}

	return &challenge, nil
}

func (p *pointOfInterestChallengeHandle) Create(ctx context.Context, pointOfInterestID uuid.UUID, tier int, question string, inventoryItemID int) (*models.PointOfInterestChallenge, error) {
	challenge := models.PointOfInterestChallenge{
		PointOfInterestID: pointOfInterestID,
		Tier:              tier,
		Question:          question,
		InventoryItemID:   inventoryItemID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		ID:                uuid.New(),
	}

	if err := p.db.Create(&challenge).Error; err != nil {
		return nil, err
	}

	return &challenge, nil
}

func (p *pointOfInterestChallengeHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.PointOfInterestChallenge, error) {
	var challenge models.PointOfInterestChallenge
	if err := p.db.WithContext(ctx).First(&challenge, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (p *pointOfInterestChallengeHandle) SubmitAnswerForChallenge(ctx context.Context, challengeID uuid.UUID, teamID *uuid.UUID, userID *uuid.UUID, text string, imageURL string, isCorrect bool) (*models.PointOfInterestChallengeSubmission, error) {
	challenge := models.PointOfInterestChallengeSubmission{
		PointOfInterestChallengeID: challengeID,
		TeamID:                     teamID,
		UserID:                     userID,
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

func (p *pointOfInterestChallengeHandle) GetChallengeForPointOfInterest(ctx context.Context, pointOfInterestID uuid.UUID, tier int) (*models.PointOfInterestChallenge, error) {
	var challenge models.PointOfInterestChallenge
	if err := p.db.WithContext(ctx).First(&challenge, "point_of_interest_id = ? AND tier = ?", pointOfInterestID, tier).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}
