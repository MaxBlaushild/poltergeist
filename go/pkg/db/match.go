package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type matchHandle struct {
	db *gorm.DB
}

func (h *matchHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Match, error) {
	var match models.Match
	if err := h.db.WithContext(ctx).Preload("VerificationCodes").Preload("Teams.Users.UserProfile").Preload("Teams.PointOfInterestTeams").Where("id = ?", id).First(&match).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &match, nil
}

func (h *matchHandle) FindCurrentMatchForUser(ctx context.Context, userId uuid.UUID) (*models.Match, error) {
	var match models.Match
	if err := h.db.WithContext(ctx).
		Preload("VerificationCodes").
		Preload("Teams.Users.UserProfile").
		Preload("Teams.PointOfInterestTeams").
		Joins("JOIN team_users ON team_users.team_id = teams.id").
		Where("creator_id = ? OR team_users.user_id = ?", userId, userId).
		Where("ended_at IS NULL").
		Order("created_at DESC").
		First(&match).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &match, nil
}

func (h *matchHandle) Create(ctx context.Context, creatorID uuid.UUID, pointsOfInterestIDs []uuid.UUID) (*models.Match, error) {
	tx := h.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	verificationCodeHandler := &verificationCodeHandler{db: tx}
	verificationCode, err := verificationCodeHandler.Create(ctx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	match := models.Match{
		CreatorID: creatorID,
	}

	if err := tx.Create(&match).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	team := models.Team{
		Name: "Team " + verificationCode.Code,
	}

	if err := tx.Create(&team).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	teamUser := models.UserTeam{
		UserID: creatorID,
		TeamID: team.ID,
	}

	if err := tx.Create(&teamUser).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	matchTeam := models.TeamMatch{
		MatchID: match.ID,
		TeamID:  team.ID,
	}

	if err := tx.Create(&matchTeam).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	verificationCodeMatch := models.MatchVerificationCode{
		MatchID:            match.ID,
		VerificationCodeID: verificationCode.ID,
	}

	if err := tx.Create(&verificationCodeMatch).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var matchPointOfInterest []*models.MatchPointOfInterest
	for _, pointOfInterestID := range pointsOfInterestIDs {
		matchPointOfInterest = append(matchPointOfInterest, &models.MatchPointOfInterest{
			ID:                uuid.New(),
			MatchID:           match.ID,
			PointOfInterestID: pointOfInterestID,
		})
	}

	if err := tx.Create(&matchPointOfInterest).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	match.VerificationCodes = append(match.VerificationCodes, *verificationCode)

	return &match, nil
}

func (h *matchHandle) StartMatch(ctx context.Context, matchID uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.Match{}).Where("id = ?", matchID).Update("started_at", time.Now()).Error
}

func (h *matchHandle) EndMatch(ctx context.Context, matchID uuid.UUID) error {
	return h.db.WithContext(ctx).Model(&models.Match{}).Where("id = ?", matchID).Update("ended_at", time.Now()).Error
}

func (h *matchHandle) FindUsersMatches(ctx context.Context, userID uuid.UUID) ([]*models.Match, error) {
	var matches []*models.Match
	if err := h.db.WithContext(ctx).Preload("VerificationCodes").Joins("JOIN team_users ON team_users.user_id = ?", userID).Where("matches.creator_id = team_users.creator_id").Find(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}
