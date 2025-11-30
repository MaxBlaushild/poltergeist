package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *server) getAllFeteRooms(ctx *gin.Context) {
	rooms, err := s.dbClient.FeteRoom().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, rooms)
}

func (s *server) getFeteRoom(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}

	room, err := s.dbClient.FeteRoom().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if room == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	ctx.JSON(http.StatusOK, room)
}

func (s *server) createFeteRoom(ctx *gin.Context) {
	var requestBody struct {
		Name          string    `json:"name" binding:"required"`
		Open          bool      `json:"open"`
		CurrentTeamID uuid.UUID `json:"currentTeamId" binding:"required"`
		HueLightID    *int      `json:"hueLightId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room := &models.FeteRoom{
		Name:          requestBody.Name,
		Open:          requestBody.Open,
		CurrentTeamID: requestBody.CurrentTeamID,
		HueLightID:    requestBody.HueLightID,
	}

	if err := s.dbClient.FeteRoom().Create(ctx, room); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create room: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, room)
}

func (s *server) updateFeteRoom(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}

	var requestBody struct {
		Name          *string    `json:"name"`
		Open          *bool      `json:"open"`
		CurrentTeamID *uuid.UUID `json:"currentTeamId"`
		HueLightID    *int       `json:"hueLightId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch existing room
	existingRoom, err := s.dbClient.FeteRoom().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingRoom == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	// Update fields if provided
	updates := &models.FeteRoom{}
	if requestBody.Name != nil {
		updates.Name = *requestBody.Name
	} else {
		updates.Name = existingRoom.Name
	}
	if requestBody.Open != nil {
		updates.Open = *requestBody.Open
	} else {
		updates.Open = existingRoom.Open
	}
	if requestBody.CurrentTeamID != nil {
		updates.CurrentTeamID = *requestBody.CurrentTeamID
	} else {
		updates.CurrentTeamID = existingRoom.CurrentTeamID
	}
	if requestBody.HueLightID != nil {
		updates.HueLightID = requestBody.HueLightID
	} else {
		updates.HueLightID = existingRoom.HueLightID
	}

	if err := s.dbClient.FeteRoom().Update(ctx, id, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update room: " + err.Error()})
		return
	}

	// Fetch the updated room
	updatedRoom, err := s.dbClient.FeteRoom().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedRoom)
}

func (s *server) deleteFeteRoom(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}

	if err := s.dbClient.FeteRoom().Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete room: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "room deleted successfully"})
}
