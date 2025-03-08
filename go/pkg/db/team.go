package db

import (
	"context"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type teamHandle struct {
	db *gorm.DB
}

func (h *teamHandle) GetAll(ctx context.Context) ([]models.Team, error) {
	var teams []models.Team

	if err := h.db.WithContext(ctx).Preload("Users").Find(&teams).Error; err != nil {
		return nil, err
	}

	return teams, nil
}

func (h *teamHandle) UpdateTeamName(ctx context.Context, teamID uuid.UUID, name string) (*models.Team, error) {
	var team models.Team

	if err := h.db.WithContext(ctx).Model(&models.Team{}).Where("id = ?", teamID).Update("name", name).Error; err != nil {
		return nil, err
	}

	return &team, nil
}

func (h *teamHandle) GetByMatchID(ctx context.Context, matchID uuid.UUID) ([]models.Team, error) {
	var teams []models.Team

	if err := h.db.WithContext(ctx).
		Joins("JOIN team_matches ON teams.id = team_matches.team_id").
		Preload("Users").
		Where("team_matches.match_id = ?", matchID).
		Find(&teams).Error; err != nil {
		return nil, err
	}

	return teams, nil
}

func (h *teamHandle) Create(ctx context.Context, userIDs []uuid.UUID, teamName string, matchID uuid.UUID) (*models.Team, error) {
	for _, userID := range userIDs {
		if err := h.RemoveUserFromMatch(ctx, matchID, userID); err != nil {
			return nil, err
		}
	}

	team := models.Team{Name: teamName}

	if err := h.db.WithContext(ctx).Create(&team).Error; err != nil {
		return nil, err
	}

	userTeams := []models.UserTeam{}
	for _, userID := range userIDs {
		userTeams = append(userTeams, models.UserTeam{
			UserID: userID,
			TeamID: team.ID,
		})
	}

	matchTeam := models.TeamMatch{
		TeamID:  team.ID,
		MatchID: matchID,
	}

	if err := h.db.WithContext(ctx).Create(&matchTeam).Error; err != nil {
		return nil, err
	}

	if err := h.db.WithContext(ctx).Create(&userTeams).Error; err != nil {
		return nil, err
	}

	return &team, nil
}

func (h *teamHandle) AddUserToTeam(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error {
	userTeam := models.UserTeam{
		UserID: userID,
		TeamID: teamID,
	}

	// Get the match ID for the team the user is joining
	var matchTeam models.TeamMatch
	if err := h.db.WithContext(ctx).Where("team_id = ?", teamID).First(&matchTeam).Error; err != nil {
		return err
	}

	// Remove user from any other teams in this match
	if err := h.RemoveUserFromMatch(ctx, matchTeam.MatchID, userID); err != nil {
		return err
	}

	return h.db.WithContext(ctx).Create(&userTeam).Error
}

func (h *teamHandle) RemoveUserFromMatch(ctx context.Context, matchID uuid.UUID, userID uuid.UUID) error {
	teamMatches := []models.TeamMatch{}

	if err := h.db.WithContext(ctx).Where("match_id = ?", matchID).Find(&teamMatches).Error; err != nil {
		return err
	}

	teamIDs := []uuid.UUID{}
	for _, teamMatch := range teamMatches {
		teamIDs = append(teamIDs, teamMatch.TeamID)
	}

	if err := h.db.WithContext(ctx).Where("user_id = ? AND team_id IN (?)", userID, teamIDs).Delete(&models.UserTeam{}).Error; err != nil {
		return err
	}

	return nil
}

func (h *teamHandle) GetByID(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	var team models.Team

	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&team).Error; err != nil {
		return nil, err
	}

	return &team, nil
}
