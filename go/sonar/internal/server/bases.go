package server

import (
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func serializeBase(base *models.Base) gin.H {
	if base == nil {
		return gin.H{}
	}

	owner := gin.H{}
	if base.User.ID != uuid.Nil {
		username := ""
		if base.User.Username != nil {
			username = *base.User.Username
		}
		owner = gin.H{
			"id":                base.User.ID,
			"name":              base.User.Name,
			"username":          username,
			"profilePictureUrl": base.User.ProfilePictureUrl,
		}
	}

	return gin.H{
		"id":          base.ID,
		"userId":      base.UserID,
		"owner":       owner,
		"latitude":    base.Latitude,
		"longitude":   base.Longitude,
		"description": base.Description,
		"imageUrl":    base.ImageURL,
		"thumbnailUrl": func() string {
			if strings.TrimSpace(base.ThumbnailURL) != "" {
				return base.ThumbnailURL
			}
			if strings.TrimSpace(base.ImageURL) != "" {
				return base.ImageURL
			}
			return staticThumbnailURL(baseDiscoveredIconKey)
		}(),
		"createdAt": base.CreatedAt,
		"updatedAt": base.UpdatedAt,
	}
}

func (s *server) getVisibleBases(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	friends, err := s.dbClient.Friend().FindAllFriends(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userIDs := make([]uuid.UUID, 0, len(friends)+1)
	userIDs = append(userIDs, user.ID)
	for _, friend := range friends {
		userIDs = append(userIDs, friend.ID)
	}

	bases, err := s.dbClient.Base().FindByUserIDs(ctx, userIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]gin.H, 0, len(bases))
	for i := range bases {
		response = append(response, serializeBase(&bases[i]))
	}

	ctx.JSON(http.StatusOK, response)
}

func (s *server) getAllBases(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	bases, err := s.dbClient.Base().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]gin.H, 0, len(bases))
	for i := range bases {
		response = append(response, serializeBase(&bases[i]))
	}

	ctx.JSON(http.StatusOK, response)
}

func (s *server) deleteBase(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	baseID, err := uuid.Parse(ctx.Param("id"))
	if err != nil || baseID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid base ID"})
		return
	}

	if err := s.dbClient.Base().Delete(ctx, baseID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "base deleted"})
}
