package server

import (
	"fmt"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getAlbumRole returns "owner", "admin", "poster", or ""
func (s *server) getAlbumRole(ctx *gin.Context, album *models.Album, userID uuid.UUID) string {
	if album.UserID == userID {
		return "owner"
	}
	role, err := s.dbClient.AlbumMember().GetRole(ctx, album.ID, userID)
	if err != nil {
		return ""
	}
	return role
}

// canAccess returns true if user can view the album (owner, member, or accepted invite)
func (s *server) canAccessAlbum(ctx *gin.Context, album *models.Album, userID uuid.UUID) bool {
	if s.getAlbumRole(ctx, album, userID) != "" {
		return true
	}
	inv, err := s.dbClient.AlbumInvite().FindByAlbumAndUser(ctx, album.ID, userID)
	if err != nil || inv == nil {
		return false
	}
	return inv.Status == "accepted"
}

// canAdmin returns true if user can manage album (owner or admin)
func (s *server) canAdminAlbum(ctx *gin.Context, album *models.Album, userID uuid.UUID) bool {
	role := s.getAlbumRole(ctx, album, userID)
	return role == "owner" || role == "admin"
}

// canAddRemoveAnyPost returns true for owner/admin
func (s *server) canAddRemoveAnyPost(ctx *gin.Context, album *models.Album, userID uuid.UUID) bool {
	return s.canAdminAlbum(ctx, album, userID)
}

// canAddRemoveOwnPost returns true for owner/admin/poster
func (s *server) canAddRemoveOwnPost(ctx *gin.Context, album *models.Album, userID uuid.UUID) bool {
	role := s.getAlbumRole(ctx, album, userID)
	return role == "owner" || role == "admin" || role == "poster"
}

func (s *server) AddAlbumTag(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	albumID, tag, ok := s.parseAlbumIDAndBody(ctx, "tag")
	if !ok {
		return
	}
	album, err := s.dbClient.Album().FindByID(ctx, albumID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}
	if !s.canAdminAlbum(ctx, album, user.ID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}
	if err := s.dbClient.Album().AddTag(ctx, albumID, tag); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (s *server) RemoveAlbumTag(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	albumID, tag, ok := s.parseAlbumIDAndBody(ctx, "tag")
	if !ok {
		return
	}
	album, err := s.dbClient.Album().FindByID(ctx, albumID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}
	if !s.canAdminAlbum(ctx, album, user.ID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}
	if err := s.dbClient.Album().RemoveTag(ctx, albumID, tag); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (s *server) AddAlbumPost(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	albumID, postID, ok := s.parseAlbumAndPostID(ctx)
	if !ok {
		return
	}
	album, err := s.dbClient.Album().FindByID(ctx, albumID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}
	if !s.canAddRemoveOwnPost(ctx, album, user.ID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	post, err := s.dbClient.Post().FindByID(ctx, postID)
	if err != nil || post == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	if !s.canAddRemoveAnyPost(ctx, album, user.ID) && post.UserID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "posters can only add their own posts"})
		return
	}
	if err := s.dbClient.AlbumPost().Add(ctx, albumID, postID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	followerIDs := s.getAlbumFollowerIDs(ctx.Request.Context(), album, user.ID)
	postIDPtr := &postID
	for _, uid := range followerIDs {
		_ = s.createNotification(ctx, &models.Notification{
			UserID:  uid,
			Type:    "album_photo_added",
			ActorID: user.ID,
			AlbumID: albumID,
			PostID:  postIDPtr,
		})
	}
	ctx.Status(http.StatusNoContent)
}

func (s *server) RemoveAlbumPost(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	albumID, postID, ok := s.parseAlbumAndPostID(ctx)
	if !ok {
		return
	}
	album, err := s.dbClient.Album().FindByID(ctx, albumID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}
	if !s.canAddRemoveOwnPost(ctx, album, user.ID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	post, err := s.dbClient.Post().FindByID(ctx, postID)
	if err != nil || post == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	if !s.canAddRemoveAnyPost(ctx, album, user.ID) && post.UserID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "posters can only remove their own posts"})
		return
	}
	if err := s.dbClient.AlbumPost().Remove(ctx, albumID, postID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (s *server) InviteToAlbum(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	albumID, _ := uuid.Parse(ctx.Param("id"))
	if albumID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "album id required"})
		return
	}
	var req struct {
		UserID uuid.UUID `json:"userId" binding:"required"`
		Role   string    `json:"role" binding:"required"` // "admin" or "poster"
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request: %v", err)})
		return
	}
	if req.Role != "admin" && req.Role != "poster" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "role must be admin or poster"})
		return
	}
	album, err := s.dbClient.Album().FindByID(ctx, albumID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}
	if !s.canAdminAlbum(ctx, album, user.ID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}
	if req.UserID == user.ID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot invite yourself"})
		return
	}
	if _, err := s.dbClient.AlbumMember().GetRole(ctx, albumID, req.UserID); err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user is already a member"})
		return
	}
	existing, _ := s.dbClient.AlbumInvite().FindByAlbumAndUser(ctx, albumID, req.UserID)
	if existing != nil {
		if existing.Status == "pending" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invite already pending"})
			return
		}
		if existing.Status == "rejected" {
			inv, reinvErr := s.dbClient.AlbumInvite().Reinvite(ctx, albumID, req.UserID, req.Role)
			if reinvErr != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": reinvErr.Error()})
				return
			}
			_ = s.createNotification(ctx, &models.Notification{
				UserID:   req.UserID,
				Type:     "album_invite",
				ActorID:  user.ID,
				AlbumID:  albumID,
				InviteID: &inv.ID,
			})
			ctx.JSON(http.StatusOK, inv)
			return
		}
	}
	inv, err := s.dbClient.AlbumInvite().Create(ctx, albumID, req.UserID, user.ID, req.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = s.createNotification(ctx, &models.Notification{
		UserID:   req.UserID,
		Type:     "album_invite",
		ActorID:  user.ID,
		AlbumID:  albumID,
		InviteID: &inv.ID,
	})
	ctx.JSON(http.StatusOK, inv)
}

func (s *server) AcceptAlbumInvite(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	inviteID, _ := uuid.Parse(ctx.Param("inviteId"))
	if inviteID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invite id required"})
		return
	}
	inv, err := s.dbClient.AlbumInvite().FindByID(ctx, inviteID)
	if err != nil || inv == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "invite not found"})
		return
	}
	if inv.InvitedUserID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "not your invite"})
		return
	}
	if inv.Status != "pending" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invite already " + inv.Status})
		return
	}
	if err := s.dbClient.AlbumInvite().UpdateStatus(ctx, inviteID, "accepted"); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	role := inv.Role
	if role != "admin" && role != "poster" {
		role = "poster"
	}
	_ = s.dbClient.AlbumMember().Add(ctx, inv.AlbumID, user.ID, role)
	ctx.Status(http.StatusNoContent)
}

func (s *server) RejectAlbumInvite(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	inviteID, _ := uuid.Parse(ctx.Param("inviteId"))
	if inviteID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invite id required"})
		return
	}
	inv, err := s.dbClient.AlbumInvite().FindByID(ctx, inviteID)
	if err != nil || inv == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "invite not found"})
		return
	}
	if inv.InvitedUserID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "not your invite"})
		return
	}
	if inv.Status != "pending" {
		ctx.Status(http.StatusNoContent)
		return
	}
	if err := s.dbClient.AlbumInvite().UpdateStatus(ctx, inviteID, "rejected"); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (s *server) GetAlbumInvites(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	invs, err := s.dbClient.AlbumInvite().FindPendingByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Enrich with album and inviter info
	type InviteWithDetails struct {
		models.AlbumInvite
		Album   *models.Album `json:"album"`
		Inviter *models.User  `json:"inviter"`
	}
	result := make([]InviteWithDetails, len(invs))
	for i, inv := range invs {
		album, _ := s.dbClient.Album().FindByID(ctx, inv.AlbumID)
		inviter, _ := s.dbClient.User().FindByID(ctx, inv.InvitedByID)
		result[i] = InviteWithDetails{
			AlbumInvite: inv,
			Album:       album,
			Inviter:     inviter,
		}
	}
	ctx.JSON(http.StatusOK, result)
}

func (s *server) GetAlbumMembers(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	albumID, _ := uuid.Parse(ctx.Param("id"))
	if albumID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "album id required"})
		return
	}
	album, err := s.dbClient.Album().FindByID(ctx, albumID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}
	if !s.canAccessAlbum(ctx, album, user.ID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	members, err := s.dbClient.AlbumMember().FindByAlbumID(ctx, albumID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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
	type MemberWithUser struct {
		UserID uuid.UUID    `json:"userId"`
		Role   string       `json:"role"`
		User   *models.User `json:"user"`
	}
	result := []MemberWithUser{
		{UserID: album.UserID, Role: "owner", User: userMap[album.UserID]},
	}
	for _, m := range members {
		result = append(result, MemberWithUser{UserID: m.UserID, Role: m.Role, User: userMap[m.UserID]})
	}
	ctx.JSON(http.StatusOK, result)
}

func (s *server) RemoveAlbumMember(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	albumID, _ := uuid.Parse(ctx.Param("id"))
	if albumID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "album id required"})
		return
	}
	var req struct {
		UserID uuid.UUID `json:"userId" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "userId required"})
		return
	}
	album, err := s.dbClient.Album().FindByID(ctx, albumID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}
	if !s.canAdminAlbum(ctx, album, user.ID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}
	if req.UserID == album.UserID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot remove album owner"})
		return
	}
	if err := s.dbClient.AlbumMember().Remove(ctx, albumID, req.UserID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (s *server) UpdateAlbumMemberRole(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	albumID, _ := uuid.Parse(ctx.Param("id"))
	if albumID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "album id required"})
		return
	}
	var req struct {
		UserID uuid.UUID `json:"userId" binding:"required"`
		Role   string    `json:"role" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "userId and role required"})
		return
	}
	if req.Role != "admin" && req.Role != "poster" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "role must be admin or poster"})
		return
	}
	album, err := s.dbClient.Album().FindByID(ctx, albumID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}
	if !s.canAdminAlbum(ctx, album, user.ID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}
	if req.UserID == album.UserID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot change owner role"})
		return
	}
	if err := s.dbClient.AlbumMember().UpdateRole(ctx, albumID, req.UserID, req.Role); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (s *server) GetAlbumPendingInvites(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	albumID, _ := uuid.Parse(ctx.Param("id"))
	if albumID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "album id required"})
		return
	}
	album, err := s.dbClient.Album().FindByID(ctx, albumID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}
	if !s.canAdminAlbum(ctx, album, user.ID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}
	invs, err := s.dbClient.AlbumInvite().FindPendingByAlbumID(ctx, albumID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type InviteWithUser struct {
		models.AlbumInvite
		InvitedUser *models.User `json:"invitedUser"`
	}
	result := make([]InviteWithUser, len(invs))
	for i, inv := range invs {
		invitedUser, _ := s.dbClient.User().FindByID(ctx, inv.InvitedUserID)
		result[i] = InviteWithUser{AlbumInvite: inv, InvitedUser: invitedUser}
	}
	ctx.JSON(http.StatusOK, result)
}

func (s *server) parseAlbumIDAndBody(ctx *gin.Context, bodyKey string) (uuid.UUID, string, bool) {
	albumID, _ := uuid.Parse(ctx.Param("id"))
	if albumID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "album id required"})
		return uuid.Nil, "", false
	}
	var req map[string]string
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return uuid.Nil, "", false
	}
	tag, ok := req[bodyKey]
	if !ok || tag == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": bodyKey + " required"})
		return uuid.Nil, "", false
	}
	return albumID, tag, true
}

func (s *server) parseAlbumAndPostID(ctx *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	albumID, _ := uuid.Parse(ctx.Param("id"))
	if albumID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "album id required"})
		return uuid.Nil, uuid.Nil, false
	}
	var req struct {
		PostID uuid.UUID `json:"postId" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "postId required"})
		return uuid.Nil, uuid.Nil, false
	}
	return albumID, req.PostID, true
}
