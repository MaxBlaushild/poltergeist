package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *server) CreateAlbumShare(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	albumID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid album id"})
		return
	}

	album, err := s.dbClient.Album().FindByID(ctx, albumID)
	if err != nil || album == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}

	if !s.canAdminAlbum(ctx, album, user.ID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "cannot share this album"})
		return
	}

	share, err := s.dbClient.AlbumShare().FindByAlbumID(ctx, albumID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if share == nil {
		token := uuid.NewString()
		share, err = s.dbClient.AlbumShare().Create(ctx, albumID, user.ID, token)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token":    share.Token,
		"shareUrl": s.albumShareOpenURL(ctx, share.Token),
	})
}

func (s *server) GetAlbumShare(ctx *gin.Context) {
	token := strings.TrimSpace(ctx.Param("token"))
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "share token is required"})
		return
	}

	share, err := s.dbClient.AlbumShare().FindByToken(ctx, token)
	if err != nil || share == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
		return
	}

	album, err := s.dbClient.Album().FindByID(ctx, share.AlbumID)
	if err != nil || album == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}

	tags, err := s.dbClient.Album().GetTags(ctx, album.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var posts []models.Post
	if len(tags) > 0 {
		posts, err = s.dbClient.Album().FindPostsForAlbum(ctx, album.UserID, tags)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
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
	reactionMap, err := s.aggregateReactions(ctx, postIDs, uuid.Nil)
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

func (s *server) OpenAlbumShare(ctx *gin.Context) {
	token := strings.TrimSpace(ctx.Param("token"))
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "share token is required"})
		return
	}

	share, err := s.dbClient.AlbumShare().FindByToken(ctx, token)
	if err != nil || share == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
		return
	}

	ctx.Redirect(http.StatusFound, fmt.Sprintf("vera://album/%s", share.Token))
}

func (s *server) albumShareOpenURL(ctx *gin.Context, token string) string {
	scheme := "http"
	if proto := ctx.Request.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if ctx.Request.TLS != nil {
		scheme = "https"
	}
	host := ctx.Request.Host
	if forwardedHost := ctx.Request.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}
	return fmt.Sprintf("%s://%s/verifiable-sn/album-shares/%s/open", scheme, host, token)
}
