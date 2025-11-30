package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *server) getAllFeteRoomLinkedListTeams(ctx *gin.Context) {
	items, err := s.dbClient.FeteRoomLinkedListTeam().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, items)
}

func (s *server) getFeteRoomLinkedListTeam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid linked list item ID"})
		return
	}

	item, err := s.dbClient.FeteRoomLinkedListTeam().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if item == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "linked list item not found"})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

func (s *server) createFeteRoomLinkedListTeam(ctx *gin.Context) {
	var requestBody struct {
		FeteRoomID   uuid.UUID `json:"feteRoomId" binding:"required"`
		FirstTeamID  uuid.UUID `json:"firstTeamId" binding:"required"`
		SecondTeamID uuid.UUID `json:"secondTeamId" binding:"required"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := &models.FeteRoomLinkedListTeam{
		FeteRoomID:   requestBody.FeteRoomID,
		FirstTeamID:  requestBody.FirstTeamID,
		SecondTeamID: requestBody.SecondTeamID,
	}

	if err := s.dbClient.FeteRoomLinkedListTeam().Create(ctx, item); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create linked list item: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, item)
}

func (s *server) updateFeteRoomLinkedListTeam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid linked list item ID"})
		return
	}

	var requestBody struct {
		FeteRoomID   *uuid.UUID `json:"feteRoomId"`
		FirstTeamID  *uuid.UUID `json:"firstTeamId"`
		SecondTeamID *uuid.UUID `json:"secondTeamId"`
	}

	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch existing item
	existingItem, err := s.dbClient.FeteRoomLinkedListTeam().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingItem == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "linked list item not found"})
		return
	}

	// Update fields if provided
	updates := &models.FeteRoomLinkedListTeam{}
	if requestBody.FeteRoomID != nil {
		updates.FeteRoomID = *requestBody.FeteRoomID
	} else {
		updates.FeteRoomID = existingItem.FeteRoomID
	}
	if requestBody.FirstTeamID != nil {
		updates.FirstTeamID = *requestBody.FirstTeamID
	} else {
		updates.FirstTeamID = existingItem.FirstTeamID
	}
	if requestBody.SecondTeamID != nil {
		updates.SecondTeamID = *requestBody.SecondTeamID
	} else {
		updates.SecondTeamID = existingItem.SecondTeamID
	}

	if err := s.dbClient.FeteRoomLinkedListTeam().Update(ctx, id, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update linked list item: " + err.Error()})
		return
	}

	// Fetch the updated item
	updatedItem, err := s.dbClient.FeteRoomLinkedListTeam().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedItem)
}

func (s *server) deleteFeteRoomLinkedListTeam(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid linked list item ID"})
		return
	}

	if err := s.dbClient.FeteRoomLinkedListTeam().Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete linked list item: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "linked list item deleted successfully"})
}

