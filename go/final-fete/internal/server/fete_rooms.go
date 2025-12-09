package server

import (
	"log"
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

func (s *server) toggleFeteRoom(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}

	// Fetch room by ID
	room, err := s.dbClient.FeteRoom().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if room == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	// Prepare updates
	updates := &models.FeteRoom{
		Name:          room.Name,
		Open:          !room.Open,
		CurrentTeamID: room.CurrentTeamID,
		HueLightID:    room.HueLightID,
	}

	// If room is currently open, close it and set light to red
	if room.Open {
		// Update room to closed
		if err := s.dbClient.FeteRoom().Update(ctx, id, updates); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update room: " + err.Error()})
			return
		}

		// Update light to red if room has a light ID
		if room.HueLightID != nil {
			if s.hueClient != nil {
				if err := s.hueClient.SetColorRGB(ctx, *room.HueLightID, 255, 0, 0); err != nil {
					log.Printf("Warning: Failed to set light to red for room %s: %v", id, err)
					// Don't fail the request if light update fails
				}
			}
		}
	} else {
		// Room is closed, need to open it
		// Find linked list item where current team is the first team
		linkedListItem, err := s.dbClient.FeteRoomLinkedListTeam().FindByRoomIDAndFirstTeamID(ctx, room.ID, room.CurrentTeamID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find linked list item: " + err.Error()})
			return
		}
		if linkedListItem == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot open room: no linked list item found for current team"})
			return
		}

		// Set room to open and update current team to second team
		updates.CurrentTeamID = linkedListItem.SecondTeamID

		// Update room
		if err := s.dbClient.FeteRoom().Update(ctx, id, updates); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update room: " + err.Error()})
			return
		}

		// Update light to green if room has a light ID
		if room.HueLightID != nil {
			if s.hueClient != nil {
				if err := s.hueClient.SetColorRGB(ctx, *room.HueLightID, 0, 255, 0); err != nil {
					log.Printf("Warning: Failed to set light to green for room %s: %v", id, err)
					// Don't fail the request if light update fails
				}
			}
		}
	}

	// Fetch the updated room
	updatedRoom, err := s.dbClient.FeteRoom().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedRoom)
}

func (s *server) unlockFeteRoom(ctx *gin.Context) {
	idStr := ctx.Param("id")
	roomID, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}

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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user is not part of any team"})
		return
	}

	// Check if room exists
	room, err := s.dbClient.FeteRoom().FindByID(ctx, roomID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if room == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	// Check if room-team relationship already exists
	existing, err := s.dbClient.FeteRoomTeam().FindByRoomIDAndTeamID(ctx, roomID, team.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing relationship: " + err.Error()})
		return
	}
	if existing != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "room is already unlocked for this team"})
		return
	}

	// Create new room-team relationship
	roomTeam := &models.FeteRoomTeam{
		FeteRoomID: roomID,
		TeamID:     team.ID,
	}

	if err := s.dbClient.FeteRoomTeam().Create(ctx, roomTeam); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlock room: " + err.Error()})
		return
	}

	// Fetch the created record with preloaded relationships
	created, err := s.dbClient.FeteRoomTeam().FindByID(ctx, roomTeam.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, created)
}
