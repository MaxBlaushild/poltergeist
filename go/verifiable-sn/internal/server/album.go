package server

import (
	"fmt"
	"net/http"
	"strings"

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
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request: %v", err)})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "album name is required"})
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

	ids, err := s.dbClient.Album().FindAccessibleAlbumIDs(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	albums, err := s.dbClient.Album().FindByIDs(ctx, ids)
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
	if !s.canAccessAlbum(ctx, album, user.ID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "cannot access this album"})
		return
	}

	tags, err := s.dbClient.Album().GetTags(ctx, album.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var posts []models.Post
	hasExplicit, _ := s.dbClient.AlbumPost().HasAny(ctx, album.ID)
	if hasExplicit {
		postIDs, _ := s.dbClient.AlbumPost().FindPostIDsByAlbumID(ctx, album.ID)
		if len(postIDs) > 0 {
			posts, _ = s.dbClient.Post().FindByIDs(ctx, postIDs)
		}
	}
	if len(posts) == 0 && len(tags) > 0 {
		posts, err = s.dbClient.Album().FindPostsForAlbum(ctx, user.ID, tags)
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

	role := s.getAlbumRole(ctx, album, user.ID)
	if role == "" {
		role = "viewer" // has access via accepted invite but not yet member
	}

	resp := gin.H{
		"album": AlbumWithTags{Album: *album, Tags: tags},
		"posts": postsWithUsers,
		"role":  role,
	}
	if s.canAdminAlbum(ctx, album, user.ID) {
		members, _ := s.dbClient.AlbumMember().FindByAlbumID(ctx, album.ID)
		userIDs := make([]uuid.UUID, 0, len(members)+1)
		userIDs = append(userIDs, album.UserID)
		for _, m := range members {
			userIDs = append(userIDs, m.UserID)
		}
		users, _ := s.dbClient.User().FindUsersByIDs(ctx, userIDs)
		userMap := make(map[uuid.UUID]*models.User)
		for i := range users {
			userMap[users[i].ID] = &users[i]
		}
		type MemberInfo struct {
			UserID uuid.UUID   `json:"userId"`
			Role   string      `json:"role"`
			User   *models.User `json:"user"`
		}
		memberList := []MemberInfo{{UserID: album.UserID, Role: "owner", User: userMap[album.UserID]}}
		for _, m := range members {
			memberList = append(memberList, MemberInfo{UserID: m.UserID, Role: m.Role, User: userMap[m.UserID]})
		}
		invs, _ := s.dbClient.AlbumInvite().FindPendingByAlbumID(ctx, album.ID)
		invList := make([]gin.H, len(invs))
		for i, inv := range invs {
			u, _ := s.dbClient.User().FindByID(ctx, inv.InvitedUserID)
			invList[i] = gin.H{"id": inv.ID, "invitedUser": u, "role": inv.Role}
		}
		resp["members"] = memberList
		resp["pendingInvites"] = invList
	}
	ctx.JSON(http.StatusOK, resp)
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
