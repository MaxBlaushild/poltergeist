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

func (p *pointOfInterestChallengeHandle) DeleteAllForPointOfInterest(ctx context.Context, pointOfInterestID uuid.UUID) error {
	challenges := []models.PointOfInterestChallenge{}
	if err := p.db.WithContext(ctx).Where("point_of_interest_id = ?", pointOfInterestID).Find(&challenges).Error; err != nil {
		return err
	}

	for _, challenge := range challenges {
		// First delete related records in point_of_interest_children
		if err := p.db.WithContext(ctx).Delete(&models.PointOfInterestChildren{}, "point_of_interest_challenge_id = ?", challenge.ID).Error; err != nil {
			return err
		}
		// Then delete the challenge itself
		if err := p.Delete(ctx, challenge.ID); err != nil {
			return err
		}
	}
	return nil
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

func (p *pointOfInterestChallengeHandle) Create(
	ctx context.Context,
	pointOfInterestID uuid.UUID,
	tier int,
	question string,
	inventoryItemID int,
	pointOfInterestGroupID *uuid.UUID,
) (*models.PointOfInterestChallenge, error) {
	challenge := models.PointOfInterestChallenge{
		PointOfInterestID:      pointOfInterestID,
		Tier:                   tier,
		Question:               question,
		InventoryItemID:        inventoryItemID,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		ID:                     uuid.New(),
		PointOfInterestGroupID: pointOfInterestGroupID,
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

func (p *pointOfInterestChallengeHandle) GetSubmissionsForMatch(ctx context.Context, matchID uuid.UUID) ([]models.PointOfInterestChallengeSubmission, error) {
	var submissions []models.PointOfInterestChallengeSubmission
	var teams []models.TeamMatch
	if err := p.db.WithContext(ctx).Where("match_id = ?", matchID).Find(&teams).Error; err != nil {
		return nil, err
	}

	teamIDs := make([]uuid.UUID, len(teams))
	for i, team := range teams {
		teamIDs[i] = team.TeamID
	}

	if err := p.db.WithContext(ctx).Where("team_id IN ?", teamIDs).Find(&submissions).Error; err != nil {
		return nil, err
	}
	return submissions, nil
}

func (p *pointOfInterestChallengeHandle) GetSubmissionsForUser(ctx context.Context, userID uuid.UUID) ([]models.PointOfInterestChallengeSubmission, error) {
	var submissions []models.PointOfInterestChallengeSubmission
	if err := p.db.WithContext(ctx).Where("user_id = ?", userID).Find(&submissions).Error; err != nil {
		return nil, err
	}
	return submissions, nil
}
