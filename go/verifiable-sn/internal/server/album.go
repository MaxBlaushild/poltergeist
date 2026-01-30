package server

import (
	"fmt"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AlbumWithTags struct {
	models.Album
	Tags []string `json:"tags"`
}

func (s *server) CreateAlbum(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Name string   `json:"name" binding:"required"`
		Tags []string `json:"tags" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request: %v", err)})
		return
	}
	if len(req.Tags) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "at least one tag is required"})
		return
	}

	album, err := s.dbClient.Album().Create(ctx, user.ID, req.Name, req.Tags)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, AlbumWithTags{Album: *album, Tags: req.Tags})
}

func (s *server) GetAlbums(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	albums, err := s.dbClient.Album().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]AlbumWithTags, len(albums))
	for i, a := range albums {
		tags, _ := s.dbClient.Album().GetTags(ctx, a.ID)
		result[i] = AlbumWithTags{Album: a, Tags: tags}
	}

	ctx.JSON(http.StatusOK, result)
}

func (s *server) GetAlbum(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	idStr := ctx.Param("id")
	if idStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "album id is required"})
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid album id"})
		return
	}

	album, err := s.dbClient.Album().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}
	if album.UserID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "cannot access another user's album"})
		return
	}

	tags, err := s.dbClient.Album().GetTags(ctx, album.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	posts, err := s.dbClient.Album().FindPostsForAlbum(ctx, user.ID, tags)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.attachTagsToPosts(ctx, posts)

	userIDs := make(map[uuid.UUID]bool)
	for _, p := range posts {
		userIDs[p.UserID] = true
	}
	userIDSlice := make([]uuid.UUID, 0, len(userIDs))
	for uid := range userIDs {
		userIDSlice = append(userIDSlice, uid)
	}

	users, err := s.dbClient.User().FindUsersByIDs(ctx, userIDSlice)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	userMap := make(map[uuid.UUID]*models.User)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	postIDs := make([]uuid.UUID, len(posts))
	for i, p := range posts {
		postIDs[i] = p.ID
	}
	reactionMap, err := s.aggregateReactions(ctx, postIDs, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	postsWithUsers := make([]PostWithUser, len(posts))
	for i, p := range posts {
		postsWithUsers[i] = PostWithUser{
			Post:      p,
			User:      userMap[p.UserID],
			Reactions: reactionMap[p.ID],
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"album": AlbumWithTags{Album: *album, Tags: tags},
		"posts": postsWithUsers,
	})
}

func (s *server) DeleteAlbum(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	idStr := ctx.Param("id")
	if idStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "album id is required"})
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid album id"})
		return
	}

	album, err := s.dbClient.Album().FindByID(ctx, id)
	if err != nil || album == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}
	if album.UserID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "cannot delete another user's album"})
		return
	}

	if err := s.dbClient.Album().Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}
