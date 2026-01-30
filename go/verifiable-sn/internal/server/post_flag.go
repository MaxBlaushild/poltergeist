package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *server) FlagPost(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	postIDStr := ctx.Param("id")
	if postIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "post id is required"})
		return
	}
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	// Verify post exists
	_, err = s.dbClient.Post().FindByID(ctx, postID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	// Idempotent: if already flagged by user, no-op
	flagged, err := s.dbClient.PostFlag().IsFlaggedByUser(ctx, postID, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if flagged {
		ctx.Status(http.StatusNoContent)
		return
	}

	if err := s.dbClient.PostFlag().Create(ctx, postID, user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (s *server) GetFlaggedPosts(ctx *gin.Context) {
	_, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	postIDs, err := s.dbClient.PostFlag().FindFlaggedPostIDs(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(postIDs) == 0 {
		ctx.JSON(http.StatusOK, []interface{}{})
		return
	}

	type FlaggedPostItem struct {
		Post      *models.Post   `json:"post"`
		User      *models.User   `json:"user"`
		Reactions []ReactionSummary `json:"reactions,omitempty"`
		FlagCount int64          `json:"flagCount"`
	}

	items := make([]FlaggedPostItem, 0, len(postIDs))
	for _, postID := range postIDs {
		post, err := s.dbClient.Post().FindByID(ctx, postID)
		if err != nil || post == nil {
			continue
		}

		flagCount, _ := s.dbClient.PostFlag().GetFlagCount(ctx, postID)

		users, err := s.dbClient.User().FindUsersByIDs(ctx, []uuid.UUID{post.UserID})
		if err != nil || len(users) == 0 {
			continue
		}
		u := &users[0]

		reactionMap, _ := s.aggregateReactions(ctx, []uuid.UUID{postID}, uuid.Nil)
		reactions := reactionMap[postID]
		if reactions == nil {
			reactions = []ReactionSummary{}
		}

		postSlice := []models.Post{*post}
		s.attachTagsToPosts(ctx, postSlice)
		post.Tags = postSlice[0].Tags

		items = append(items, struct {
			Post      *models.Post        `json:"post"`
			User      *models.User        `json:"user"`
			Reactions []ReactionSummary   `json:"reactions,omitempty"`
			FlagCount int64               `json:"flagCount"`
		}{
			Post:      post,
			User:      u,
			Reactions: reactions,
			FlagCount: flagCount,
		})
	}

	ctx.JSON(http.StatusOK, items)
}

func (s *server) AdminDeletePost(ctx *gin.Context) {
	_, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	postIDStr := ctx.Param("id")
	if postIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "post id is required"})
		return
	}
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	_, err = s.dbClient.Post().FindByID(ctx, postID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	if err := s.dbClient.Post().Delete(ctx, postID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (s *server) DismissFlaggedPost(ctx *gin.Context) {
	_, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	postIDStr := ctx.Param("id")
	if postIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "post id is required"})
		return
	}
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	if err := s.dbClient.PostFlag().DeleteByPostID(ctx, postID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}
