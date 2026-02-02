package server

import (
	"context"
	"net/http"
	"strconv"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// createNotification inserts a notification and optionally sends push. Non-blocking on errors.
func (s *server) createNotification(ctx context.Context, n *models.Notification) error {
	if err := s.dbClient.Notification().Create(ctx, n); err != nil {
		return err
	}
	go s.sendPushForNotification(n)
	return nil
}

// sendPushForNotification sends FCM push to the recipient's devices. Runs in goroutine.
func (s *server) sendPushForNotification(n *models.Notification) {
	if s.pushClient == nil {
		return
	}
	ctx := context.Background()
	tokens, err := s.dbClient.UserDeviceToken().FindByUserID(ctx, n.UserID)
	if err != nil || len(tokens) == 0 {
		return
	}
	actor, _ := s.dbClient.User().FindByID(ctx, n.ActorID)
	album, _ := s.dbClient.Album().FindByID(ctx, n.AlbumID)
	actorName := "Someone"
	if actor != nil {
		if actor.Username != nil && *actor.Username != "" {
			actorName = *actor.Username
		} else if actor.PhoneNumber != "" {
			actorName = actor.PhoneNumber
		}
	}
	albumName := "an album"
	if album != nil {
		albumName = album.Name
	}

	var title, body string
	switch n.Type {
	case "album_invite":
		title = "Album Invite"
		body = actorName + " invited you to album " + albumName
	case "album_invite_accepted":
		title = "Invite Accepted"
		body = actorName + " accepted your invite to " + albumName
	case "album_photo_added":
		title = "New Photo"
		body = actorName + " added a photo to " + albumName
	default:
		return
	}

	data := map[string]string{
		"type":    n.Type,
		"albumId": n.AlbumID.String(),
	}
	if n.PostID != nil {
		data["postId"] = n.PostID.String()
	}
	if n.InviteID != nil {
		data["inviteId"] = n.InviteID.String()
	}

	for _, t := range tokens {
		_ = s.pushClient.Send(ctx, t.Token, title, body, data)
	}
}

func (s *server) GetNotifications(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	limit := 50
	offset := 0
	if l := ctx.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if o := ctx.Query("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}
	notifications, err := s.dbClient.Notification().FindByUserID(ctx.Request.Context(), user.ID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	unreadCount, _ := s.dbClient.Notification().CountUnreadByUserID(ctx.Request.Context(), user.ID)

	actorIDs := make(map[uuid.UUID]bool)
	for _, n := range notifications {
		actorIDs[n.ActorID] = true
	}
	actorIDSlice := make([]uuid.UUID, 0, len(actorIDs))
	for id := range actorIDs {
		actorIDSlice = append(actorIDSlice, id)
	}
	actors, _ := s.dbClient.User().FindUsersByIDs(ctx.Request.Context(), actorIDSlice)
	actorMap := make(map[uuid.UUID]*models.User)
	for i := range actors {
		actorMap[actors[i].ID] = &actors[i]
	}

	albumIDs := make(map[uuid.UUID]bool)
	for _, n := range notifications {
		albumIDs[n.AlbumID] = true
	}
	albumIDSlice := make([]uuid.UUID, 0, len(albumIDs))
	for id := range albumIDs {
		albumIDSlice = append(albumIDSlice, id)
	}
	albums, _ := s.dbClient.Album().FindByIDs(ctx.Request.Context(), albumIDSlice)
	albumMap := make(map[uuid.UUID]*models.Album)
	for i := range albums {
		albumMap[albums[i].ID] = &albums[i]
	}

	type NotificationWithDetails struct {
		models.Notification
		Actor   *models.User  `json:"actor"`
		Album   *models.Album `json:"album"`
	}
	result := make([]NotificationWithDetails, len(notifications))
	for i, n := range notifications {
		result[i] = NotificationWithDetails{
			Notification: n,
			Actor:        actorMap[n.ActorID],
			Album:        albumMap[n.AlbumID],
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"notifications": result,
		"unreadCount":   unreadCount,
	})
}

func (s *server) MarkNotificationRead(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	idStr := ctx.Param("id")
	if idStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "notification id required"})
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}
	if err := s.dbClient.Notification().MarkAsRead(ctx.Request.Context(), id, user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (s *server) MarkAllNotificationsRead(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Notification().MarkAllAsRead(ctx.Request.Context(), user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (s *server) RegisterDeviceToken(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	var req struct {
		Token    string `json:"token" binding:"required"`
		Platform string `json:"platform" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "token and platform required"})
		return
	}
	if req.Platform != "ios" && req.Platform != "android" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "platform must be ios or android"})
		return
	}
	if err := s.dbClient.UserDeviceToken().Upsert(ctx.Request.Context(), user.ID, req.Token, req.Platform); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

// getAlbumFollowerIDs returns user IDs who "follow" the album (owner + members + accepted invitees), excluding excludeID.
func (s *server) getAlbumFollowerIDs(ctx context.Context, album *models.Album, excludeID uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]bool)
	seen[excludeID] = true

	if !seen[album.UserID] {
		seen[album.UserID] = true
	}

	members, _ := s.dbClient.AlbumMember().FindByAlbumID(ctx, album.ID)
	for _, m := range members {
		if !seen[m.UserID] {
			seen[m.UserID] = true
		}
	}

	acceptedInvs, _ := s.dbClient.AlbumInvite().FindByAlbumIDAndStatus(ctx, album.ID, "accepted")
	for _, inv := range acceptedInvs {
		if !seen[inv.InvitedUserID] {
			seen[inv.InvitedUserID] = true
		}
	}

	ids := make([]uuid.UUID, 0, len(seen)-1)
	for id := range seen {
		if id != excludeID {
			ids = append(ids, id)
		}
	}
	return ids
}
