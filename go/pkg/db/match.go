package db

import (
	"context"
	"errors"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type matchHandle struct {
	db *gorm.DB
}

func (h *matchHandle) FindForTeamID(ctx context.Context, teamID uuid.UUID) (*models.TeamMatch, error) {
	match := models.TeamMatch{}
	if err := h.db.WithContext(ctx).Where("team_id = ?", teamID).First(&match).Error; err != nil {
		return nil, err
	}

	return &match, nil
}

func (h *matchHandle) FindByID(ctx context.Context, id uuid.UUID) (*models.Match, error) {
	var match models.Match
	if err := h.db.WithContext(ctx).
		Preload("PointsOfInterest.PointOfInterestChallenges.PointOfInterestChallengeSubmissions").
		Preload("Teams.Users").
		Preload("Teams.PointOfInterestDiscoveries").
		Preload("Teams.TeamInventoryItems").
		Preload("InventoryItemEffects").
		Where("id = ?", id).First(&match).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	teamIDs := make(map[uuid.UUID]bool)
	for _, team := range match.Teams {
		teamIDs[team.ID] = true
	}

	for i, poiTeam := range match.PointsOfInterest {
		for j, poiChallenge := range poiTeam.PointOfInterestChallenges {
			filteredSubmissions := []models.PointOfInterestChallengeSubmission{}
			for _, submission := range poiChallenge.PointOfInterestChallengeSubmissions {
				if teamIDs[submission.TeamID] {
					filteredSubmissions = append(filteredSubmissions, submission)
				}
			}
			match.PointsOfInterest[i].PointOfInterestChallenges[j].PointOfInterestChallengeSubmissions = filteredSubmissions
		}
	}
	return &match, nil
}

func (h *matchHandle) FindCurrentMatchForUser(ctx context.Context, userId uuid.UUID) (*models.Match, error) {
	var stringMatchId string
	sql := `
		SELECT m.id FROM matches m
		JOIN team_matches ON team_matches.match_id = m.id
		JOIN teams ON teams.id = team_matches.team_id
		JOIN user_teams ON user_teams.team_id = teams.id
		WHERE user_teams.user_id = ? AND m.ended_at IS NULL
		ORDER BY m.created_at DESC
		LIMIT 1
	`
	if err := h.db.WithContext(ctx).Raw(sql, userId).Scan(&stringMatchId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	matchID, err := uuid.Parse(stringMatchId)
	if err != nil {
		return nil, nil
	}

	match, err := h.FindByID(ctx, matchID)
	if err != nil {
		return nil, err
	}

	return match, nil
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
		Name: util.GenerateTeamName(),
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
		ID:                 uuid.New(),
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
		tx.Rollback()
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
	if err := h.db.WithContext(ctx).Preload("VerificationCodes").Joins("JOIN user_teams ON user_teams.user_id = ?", userID).Where("matches.creator_id = user_teams.creator_id").Find(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}
