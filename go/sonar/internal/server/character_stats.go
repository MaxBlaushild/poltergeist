package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type characterStatsResponse struct {
	Strength         int                            `json:"strength"`
	Dexterity        int                            `json:"dexterity"`
	Constitution     int                            `json:"constitution"`
	Intelligence     int                            `json:"intelligence"`
	Wisdom           int                            `json:"wisdom"`
	Charisma         int                            `json:"charisma"`
	Health           int                            `json:"health"`
	Mana             int                            `json:"mana"`
	EquipmentBonuses map[string]int                 `json:"equipmentBonuses"`
	StatusBonuses    map[string]int                 `json:"statusBonuses"`
	UnspentPoints    int                            `json:"unspentPoints"`
	Level            int                            `json:"level"`
	Proficiencies    []characterProficiencyResponse `json:"proficiencies"`
	Statuses         []characterStatusResponse      `json:"statuses"`
}

type characterProficiencyResponse struct {
	Proficiency string `json:"proficiency"`
	Level       int    `json:"level"`
}

type characterStatusResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Effect      string    `json:"effect"`
	EffectType  string    `json:"effectType"`
	StartedAt   time.Time `json:"startedAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type characterStatsAllocationRequest struct {
	Allocations map[string]int `json:"allocations"`
}

type userCharacterProfileResponse struct {
	User      models.User            `json:"user"`
	Stats     characterStatsResponse `json:"stats"`
	UserLevel *models.UserLevel      `json:"userLevel"`
}

func (s *server) getCharacterStats(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	userLevel, err := s.dbClient.UserLevel().FindOrCreateForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stats, err := s.dbClient.UserCharacterStats().EnsureLevelPoints(ctx, user.ID, userLevel.Level)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	proficiencies, err := s.dbClient.UserProficiency().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	equipmentBonuses, err := s.dbClient.UserEquipment().GetStatBonuses(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	statusBonuses, statuses, err := s.getActiveStatusBonusesAndStatuses(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, characterStatsResponseFrom(stats, userLevel.Level, proficiencies, equipmentBonuses, statusBonuses, statuses))
}

func (s *server) getUserCharacterProfile(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	idStr := ctx.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	target, err := s.dbClient.User().FindByID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	userLevel, err := s.dbClient.UserLevel().FindOrCreateForUser(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	userLevel.ExperienceToNextLevel = userLevel.XPToNextLevel()

	stats, err := s.dbClient.UserCharacterStats().EnsureLevelPoints(ctx, userID, userLevel.Level)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	proficiencies, err := s.dbClient.UserProficiency().FindByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	equipmentBonuses, err := s.dbClient.UserEquipment().GetStatBonuses(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	statusBonuses, statuses, err := s.getActiveStatusBonusesAndStatuses(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, userCharacterProfileResponse{
		User:      *target,
		Stats:     characterStatsResponseFrom(stats, userLevel.Level, proficiencies, equipmentBonuses, statusBonuses, statuses),
		UserLevel: userLevel,
	})
}

func (s *server) allocateCharacterStats(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var req characterStatsAllocationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil || len(req.Allocations) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "allocations required"})
		return
	}

	userLevel, err := s.dbClient.UserLevel().FindOrCreateForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stats, err := s.dbClient.UserCharacterStats().ApplyAllocations(ctx, user.ID, userLevel.Level, req.Allocations)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrNoStatAllocations),
			errors.Is(err, db.ErrInvalidStatAllocation),
			errors.Is(err, db.ErrInsufficientStatPoints):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	proficiencies, err := s.dbClient.UserProficiency().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	equipmentBonuses, err := s.dbClient.UserEquipment().GetStatBonuses(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	statusBonuses, statuses, err := s.getActiveStatusBonusesAndStatuses(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, characterStatsResponseFrom(stats, userLevel.Level, proficiencies, equipmentBonuses, statusBonuses, statuses))
}

func characterStatsResponseFrom(
	stats *models.UserCharacterStats,
	level int,
	proficiencies []models.UserProficiency,
	equipmentBonuses models.CharacterStatBonuses,
	statusBonuses models.CharacterStatBonuses,
	statuses []models.UserStatus,
) characterStatsResponse {
	totalBonuses := equipmentBonuses.Add(statusBonuses)
	effectiveConstitution := stats.Constitution + totalBonuses.Constitution
	effectiveIntelligence := stats.Intelligence + totalBonuses.Intelligence
	effectiveWisdom := stats.Wisdom + totalBonuses.Wisdom
	proficiencyResponse := make([]characterProficiencyResponse, 0, len(proficiencies))
	for _, proficiency := range proficiencies {
		proficiencyResponse = append(proficiencyResponse, characterProficiencyResponse{
			Proficiency: proficiency.Proficiency,
			Level:       proficiency.Level,
		})
	}
	return characterStatsResponse{
		Strength:         stats.Strength,
		Dexterity:        stats.Dexterity,
		Constitution:     stats.Constitution,
		Intelligence:     stats.Intelligence,
		Wisdom:           stats.Wisdom,
		Charisma:         stats.Charisma,
		Health:           deriveCharacterHealth(effectiveConstitution),
		Mana:             deriveCharacterMana(effectiveIntelligence, effectiveWisdom),
		EquipmentBonuses: equipmentBonuses.ToMap(),
		StatusBonuses:    statusBonuses.ToMap(),
		UnspentPoints:    stats.UnspentPoints,
		Level:            level,
		Proficiencies:    proficiencyResponse,
		Statuses:         characterStatusResponsesFrom(statuses),
	}
}

func characterStatusResponsesFrom(statuses []models.UserStatus) []characterStatusResponse {
	response := make([]characterStatusResponse, 0, len(statuses))
	for _, status := range statuses {
		effectType := string(status.EffectType)
		if effectType == "" {
			effectType = string(models.UserStatusEffectTypeStatModifier)
		}
		response = append(response, characterStatusResponse{
			ID:          status.ID,
			Name:        status.Name,
			Description: status.Description,
			Effect:      status.Effect,
			EffectType:  effectType,
			StartedAt:   status.StartedAt,
			ExpiresAt:   status.ExpiresAt,
		})
	}
	return response
}

func (s *server) getActiveStatusBonusesAndStatuses(
	ctx context.Context,
	userID uuid.UUID,
) (models.CharacterStatBonuses, []models.UserStatus, error) {
	statusBonuses, err := s.dbClient.UserStatus().GetActiveStatBonuses(ctx, userID)
	if err != nil {
		return models.CharacterStatBonuses{}, nil, err
	}
	statuses, err := s.dbClient.UserStatus().FindActiveByUserID(ctx, userID)
	if err != nil {
		return models.CharacterStatBonuses{}, nil, err
	}
	return statusBonuses, statuses, nil
}

func deriveCharacterHealth(constitution int) int {
	if constitution < 1 {
		constitution = 1
	}
	return constitution * 10
}

func deriveCharacterMana(intelligence int, wisdom int) int {
	mental := intelligence + wisdom
	if mental < 1 {
		mental = 1
	}
	return mental * 5
}
