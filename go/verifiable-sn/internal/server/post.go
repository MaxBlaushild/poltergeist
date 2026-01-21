package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PostWithUser struct {
	models.Post
	User *models.User `json:"user"`
}

func (s *server) CreatePost(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var requestBody struct {
		ImageURL string  `json:"imageUrl" binding:"required"`
		Caption  *string `json:"caption"`
	}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	post, err := s.dbClient.Post().Create(ctx, user.ID, requestBody.ImageURL, requestBody.Caption)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, post)
}

func (s *server) GetFeed(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	posts, err := s.dbClient.Post().FindAllFriendsPosts(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get user IDs from posts
	userIDs := make(map[uuid.UUID]bool)
	for _, post := range posts {
		userIDs[post.UserID] = true
	}

	// Convert map to slice
	userIDSlice := make([]uuid.UUID, 0, len(userIDs))
	for id := range userIDs {
		userIDSlice = append(userIDSlice, id)
	}

	// Get users
	users, err := s.dbClient.User().FindUsersByIDs(ctx, userIDSlice)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create user map
	userMap := make(map[uuid.UUID]*models.User)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	// Create posts with user info
	postsWithUsers := make([]PostWithUser, len(posts))
	for i, post := range posts {
		postsWithUsers[i] = PostWithUser{
			Post: post,
			User: userMap[post.UserID],
		}
	}

	ctx.JSON(http.StatusOK, postsWithUsers)
}

func (s *server) GetUserPosts(ctx *gin.Context) {
	_, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	userIDStr := ctx.Param("userId")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "user id is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id format",
		})
		return
	}

	posts, err := s.dbClient.Post().FindByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, posts)
}

func (s *server) DeletePost(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	postIDStr := ctx.Param("id")
	if postIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "post id is required",
		})
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid post id format",
		})
		return
	}

	// Verify the post belongs to the user
	post, err := s.dbClient.Post().FindByID(ctx, postID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "post not found",
		})
		return
	}

	if post.UserID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "cannot delete posts from another user",
		})
		return
	}

	if err := s.dbClient.Post().Delete(ctx, postID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}

