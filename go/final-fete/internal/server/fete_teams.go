package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *server) getAllFeteTeams(ctx *gin.Context) {
	teams, err := s.dbClient.FeteTeam().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, teams)
}

func (s *server) getFeteTeam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	team, err := s.dbClient.FeteTeam().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if team == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	ctx.JSON(http.StatusOK, team)
}

func (s *server) createFeteTeam(ctx *gin.Context) {
	var requestBody struct {
		Name string `json:"name" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team := &models.FeteTeam{
		Name: requestBody.Name,
	}

	if err := s.dbClient.FeteTeam().Create(ctx, team); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create team: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, team)
}

func (s *server) updateFeteTeam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	var requestBody struct {
		Name *string `json:"name"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch existing team
	existingTeam, err := s.dbClient.FeteTeam().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingTeam == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	// Update fields if provided
	updates := &models.FeteTeam{}
	if requestBody.Name != nil {
		updates.Name = *requestBody.Name
	} else {
		updates.Name = existingTeam.Name
	}

	if err := s.dbClient.FeteTeam().Update(ctx, id, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update team: " + err.Error()})
		return
	}

	// Fetch the updated team
	updatedTeam, err := s.dbClient.FeteTeam().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedTeam)
}

func (s *server) deleteFeteTeam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	if err := s.dbClient.FeteTeam().Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete team: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "team deleted successfully"})
}

func (s *server) getTeamUsers(ctx *gin.Context) {
	idStr := ctx.Param("id")
	teamID, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	// Verify team exists
	team, err := s.dbClient.FeteTeam().FindByID(ctx, teamID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if team == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	users, err := s.dbClient.FeteTeam().GetUsersByTeamID(ctx, teamID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func (s *server) getCurrentUserTeam(ctx *gin.Context) {
	// Get user from context
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}
	user, ok := userInterface.(*models.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type in context"})
		return
	}

	// Get user's team
	team, err := s.dbClient.FeteTeam().FindTeamByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find user's team: " + err.Error()})
		return
	}
	if team == nil {
		ctx.JSON(http.StatusOK, nil)
		return
	}

	ctx.JSON(http.StatusOK, team)
}

func (s *server) addUserToTeam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	teamID, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	var requestBody struct {
		UserID string `json:"userId" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(requestBody.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := s.dbClient.FeteTeam().AddUserToTeam(ctx, teamID, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add user to team: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user added to team successfully"})
}

func (s *server) removeUserFromTeam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	teamID, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	userIDStr := ctx.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := s.dbClient.FeteTeam().RemoveUserFromTeam(ctx, teamID, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove user from team: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user removed from team successfully"})
}

func (s *server) searchUsers(ctx *gin.Context) {
	query := ctx.Query("query")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}

	var results []models.User

	// Try to find user by phone number first
	user, err := s.dbClient.User().FindByPhoneNumber(ctx, query)
	if err == nil && user != nil {
		results = append(results, *user)
	}

	// Also search by username (partial match)
	usernameUsers, err := s.dbClient.User().FindLikeByUsername(ctx, query)
	if err == nil {
		for _, u := range usernameUsers {
			// Avoid duplicates if phone search already found this user
			found := false
			for _, r := range results {
				if r.ID == u.ID {
					found = true
					break
				}
			}
			if !found {
				results = append(results, *u)
			}
		}
	}

	ctx.JSON(http.StatusOK, results)
}

