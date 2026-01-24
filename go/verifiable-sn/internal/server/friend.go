package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *server) GetFriends(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	friends, err := s.dbClient.Friend().FindAllFriends(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, friends)
}

func (s *server) GetFriendInvites(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	invites, err := s.dbClient.FriendInvite().FindAllInvites(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, invites)
}

func (s *server) CreateFriendInvite(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var requestBody struct {
		InviteeID string `json:"inviteeID" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	inviteeID, err := uuid.Parse(requestBody.InviteeID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid inviteeID format",
		})
		return
	}

	// Prevent self-invites
	if user.ID == inviteeID {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "cannot invite yourself",
		})
		return
	}

	invite, err := s.dbClient.FriendInvite().Create(ctx, user.ID, inviteeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, invite)
}

func (s *server) AcceptFriendInvite(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var requestBody struct {
		InviteID string `json:"inviteID" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	inviteID, err := uuid.Parse(requestBody.InviteID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid inviteID format",
		})
		return
	}

	invite, err := s.dbClient.FriendInvite().FindByID(ctx, inviteID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "invite not found",
		})
		return
	}

	// Verify the invite is for the current user
	if invite.InviteeID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "cannot accept invites for another user",
		})
		return
	}

	// Check if friendship already exists
	exists, err := s.dbClient.Friend().Exists(ctx, invite.InviterID, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create the friend relationship only if it doesn't already exist
	if !exists {
		_, err = s.dbClient.Friend().Create(ctx, invite.InviterID, user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	// Delete the invite
	if err = s.dbClient.FriendInvite().Delete(ctx, inviteID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete invite: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "friend invite accepted successfully",
	})
}

func (s *server) DeleteFriendInvite(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	inviteIDStr := ctx.Param("id")
	if inviteIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invite id is required",
		})
		return
	}

	inviteID, err := uuid.Parse(inviteIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid invite id format",
		})
		return
	}

	// Verify the invite belongs to the user (either as inviter or invitee)
	invite, err := s.dbClient.FriendInvite().FindByID(ctx, inviteID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "invite not found",
		})
		return
	}

	if invite.InviterID != user.ID && invite.InviteeID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "cannot delete invites for another user",
		})
		return
	}

	if err = s.dbClient.FriendInvite().Delete(ctx, inviteID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (s *server) SearchUsers(ctx *gin.Context) {
	_, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	query := ctx.Query("query")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "query parameter is required",
		})
		return
	}

	users, err := s.dbClient.User().FindLikeByUsername(ctx, query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Filter out invalid users (similar to travel-angels)
	filteredUsers := make([]models.User, 0)
	for _, user := range users {
		if user != nil && user.ID != uuid.Nil {
			filteredUsers = append(filteredUsers, *user)
		}
	}

	ctx.JSON(http.StatusOK, filteredUsers)
}

