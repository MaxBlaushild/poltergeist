package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *server) getAllFeteRoomTeams(ctx *gin.Context) {
	items, err := s.dbClient.FeteRoomTeam().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, items)
}

func (s *server) getFeteRoomTeam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid room team ID"})
		return
	}

	item, err := s.dbClient.FeteRoomTeam().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if item == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "room team not found"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

func (s *server) createFeteRoomTeam(ctx *gin.Context) {
	var requestBody struct {
		FeteRoomID uuid.UUID `json:"feteRoomId" binding:"required"`
		TeamID     uuid.UUID `json:"teamId" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if relationship already exists
	existing, err := s.dbClient.FeteRoomTeam().FindByRoomIDAndTeamID(ctx, requestBody.FeteRoomID, requestBody.TeamID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing relationship: " + err.Error()})
		return
	}
	if existing != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "room is already unlocked for this team"})
		return
	}

	item := &models.FeteRoomTeam{
		FeteRoomID: requestBody.FeteRoomID,
		TeamID:     requestBody.TeamID,
	}

	if err := s.dbClient.FeteRoomTeam().Create(ctx, item); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create room team: " + err.Error()})
		return
	}

	// Fetch the created record with preloaded relationships
	created, err := s.dbClient.FeteRoomTeam().FindByID(ctx, item.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, created)
}

func (s *server) deleteFeteRoomTeam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid room team ID"})
		return
	}

	if err := s.dbClient.FeteRoomTeam().Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete room team: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "room team deleted successfully"})
}

