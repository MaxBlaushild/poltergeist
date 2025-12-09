package db

import (
	"context"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type feteTeamHandler struct {
	db *gorm.DB
}

func (h *feteTeamHandler) Create(ctx context.Context, team *models.FeteTeam) error {
	team.ID = uuid.New()
	team.CreatedAt = time.Now()
	team.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Create(team).Error
}

func (h *feteTeamHandler) FindByID(ctx context.Context, id uuid.UUID) (*models.FeteTeam, error) {
	var team models.FeteTeam
	if err := h.db.WithContext(ctx).First(&team, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &team, nil
}

func (h *feteTeamHandler) FindAll(ctx context.Context) ([]models.FeteTeam, error) {
	var teams []models.FeteTeam
	if err := h.db.WithContext(ctx).Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}

func (h *feteTeamHandler) Update(ctx context.Context, id uuid.UUID, updates *models.FeteTeam) error {
	updates.ID = id
	updates.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Model(&models.FeteTeam{}).Where("id = ?", id).Updates(updates).Error
}

func (h *feteTeamHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.FeteTeam{}, id).Error
}

func (h *feteTeamHandler) FindTeamByUserID(ctx context.Context, userID uuid.UUID) (*models.FeteTeam, error) {
	// Query fete_team_users join table to find user's team
	type FeteTeamUser struct {
		FeteTeamID uuid.UUID `gorm:"column:fete_team_id"`
		UserID     uuid.UUID `gorm:"column:user_id"`
	}
	
	var teamUser FeteTeamUser
	if err := h.db.WithContext(ctx).
		Table("fete_team_users").
		Where("user_id = ?", userID).
		First(&teamUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	// Get the team
	return h.FindByID(ctx, teamUser.FeteTeamID)
}

func (h *feteTeamHandler) AddUserToTeam(ctx context.Context, teamID, userID uuid.UUID) error {
	// Check if team exists
	team, err := h.FindByID(ctx, teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return gorm.ErrRecordNotFound
	}

	// Check if user-team relationship already exists
	type FeteTeamUser struct {
		ID         uuid.UUID `gorm:"column:id"`
		FeteTeamID uuid.UUID `gorm:"column:fete_team_id"`
		UserID     uuid.UUID `gorm:"column:user_id"`
	}
	
	var existing FeteTeamUser
	err = h.db.WithContext(ctx).
		Table("fete_team_users").
		Where("fete_team_id = ? AND user_id = ?", teamID, userID).
		First(&existing).Error
	
	if err == nil {
		// Relationship already exists, return nil (no error, just skip)
		return nil
	}
	
	if err != gorm.ErrRecordNotFound {
		// Some other error occurred
		return err
	}

	// Insert new relationship
	// Using raw SQL to avoid model type conflicts
	result := h.db.WithContext(ctx).Exec(
		"INSERT INTO fete_team_users (id, created_at, updated_at, fete_team_id, user_id) VALUES (gen_random_uuid(), NOW(), NOW(), ?, ?)",
		teamID, userID,
	)
	
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

func (h *feteTeamHandler) GetUsersByTeamID(ctx context.Context, teamID uuid.UUID) ([]models.User, error) {
	var users []models.User
	
	// Join fete_team_users with users table
	err := h.db.WithContext(ctx).
		Table("fete_team_users").
		Select("users.*").
		Joins("JOIN users ON fete_team_users.user_id = users.id").
		Where("fete_team_users.fete_team_id = ? AND fete_team_users.deleted_at IS NULL", teamID).
		Find(&users).Error
	
	if err != nil {
		return nil, err
	}
	
	return users, nil
}

func (h *feteTeamHandler) RemoveUserFromTeam(ctx context.Context, teamID, userID uuid.UUID) error {
	// Soft delete the relationship from fete_team_users table
	result := h.db.WithContext(ctx).
		Table("fete_team_users").
		Where("fete_team_id = ? AND user_id = ?", teamID, userID).
		Update("deleted_at", time.Now())
	
	if result.Error != nil {
		return result.Error
	}
	
	// If no rows were affected, that's okay - relationship might not exist
	return nil
}

