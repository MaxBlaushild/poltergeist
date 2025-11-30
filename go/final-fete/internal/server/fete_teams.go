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

