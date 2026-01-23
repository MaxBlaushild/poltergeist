package server

import (
	"encoding/hex"
	"fmt"
	"net/http"

	ethereum_transactor "github.com/MaxBlaushild/poltergeist/pkg/ethereum_transactor"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReactionSummary struct {
	Emoji      string `json:"emoji"`
	Count      int    `json:"count"`
	UserReacted bool  `json:"userReacted"`
}

type PostWithUser struct {
	models.Post
	User     *models.User        `json:"user"`
	Reactions []ReactionSummary  `json:"reactions,omitempty"`
	CommentCount *int64          `json:"commentCount,omitempty"`
}

type CommentWithUser struct {
	models.PostComment
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
		ImageURL        string  `json:"imageUrl" binding:"required"`
		Caption         *string `json:"caption"`
		ManifestURL     *string `json:"manifestUrl"`
		ManifestHash    *string `json:"manifestHash"`
		CertFingerprint *string `json:"certFingerprint"`
		AssetID         *string `json:"assetId"`
	}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid request body: %v", err),
		})
		return
	}

	var manifestHashBytes []byte
	var manifestURI *string
	var certFingerprintBytes []byte
	var assetID *string

	// If manifest data is provided, validate it
	if requestBody.ManifestURL != nil && *requestBody.ManifestURL != "" &&
		requestBody.ManifestHash != nil && *requestBody.ManifestHash != "" &&
		requestBody.CertFingerprint != nil && *requestBody.CertFingerprint != "" {
		// Download manifest from S3
		manifestBytes, err := DownloadManifestFromS3(*requestBody.ManifestURL)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("failed to download manifest: %v", err),
			})
			return
		}

		// Validate manifest
		computedHash, computedFingerprint, err := ValidateManifest(manifestBytes)
		if err != nil {
			fmt.Printf("Manifest validation error: %v\n", err)
			fmt.Printf("Manifest size: %d bytes\n", len(manifestBytes))
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("manifest validation failed: %v", err),
			})
			return
		}

		// Verify provided hash matches computed hash
		providedHashBytes, err := HexToBytes(*requestBody.ManifestHash)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("invalid manifest hash format: %v", err),
			})
			return
		}

		if len(providedHashBytes) != len(computedHash) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "manifest hash mismatch",
			})
			return
		}

		// Compare hashes
		hashMatch := true
		for i := range providedHashBytes {
			if providedHashBytes[i] != computedHash[i] {
				hashMatch = false
				break
			}
		}

		if !hashMatch {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "manifest hash mismatch",
			})
			return
		}

		// Verify certificate fingerprint matches
		providedFingerprintBytes, err := HexToBytes(*requestBody.CertFingerprint)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("invalid certificate fingerprint format: %v", err),
			})
			return
		}

		if len(providedFingerprintBytes) != len(computedFingerprint) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "certificate fingerprint mismatch",
			})
			return
		}

		// Compare fingerprints
		fingerprintMatch := true
		for i := range providedFingerprintBytes {
			if providedFingerprintBytes[i] != computedFingerprint[i] {
				fingerprintMatch = false
				break
			}
		}

		if !fingerprintMatch {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "certificate fingerprint mismatch",
			})
			return
		}

		// Verify certificate is active
		cert, err := s.dbClient.UserCertificate().FindByFingerprint(ctx, computedFingerprint)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to find certificate: %v", err),
			})
			return
		}

		if cert == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "certificate not found",
			})
			return
		}

		if !cert.Active {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "certificate is not active",
			})
			return
		}

		// Verify certificate belongs to the user
		if cert.UserID != user.ID {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": "certificate does not belong to user",
			})
			return
		}

		// Set manifest data
		manifestHashBytes = computedHash
		manifestURI = requestBody.ManifestURL
		certFingerprintBytes = computedFingerprint
		if requestBody.AssetID != nil {
			assetID = requestBody.AssetID
		}
	}

	// Create post
	post, err := s.dbClient.Post().Create(
		ctx,
		user.ID,
		requestBody.ImageURL,
		requestBody.Caption,
		manifestHashBytes,
		manifestURI,
		certFingerprintBytes,
		assetID,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// If manifest was provided, create blockchain transaction for anchoring
	if manifestHashBytes != nil && manifestURI != nil && certFingerprintBytes != nil {
		// Encode anchorManifest function call
		assetIDStr := ""
		if assetID != nil {
			assetIDStr = *assetID
		}

		encodedData, err := encodeAnchorManifestCall(manifestHashBytes, *manifestURI, assetIDStr, certFingerprintBytes)
		if err != nil {
			// Log error but don't fail the post creation
			fmt.Printf("Warning: failed to encode anchorManifest call: %v\n", err)
		} else {
			anchorManifestType := string(models.AnchorManifestType)

			// Create blockchain transaction via ethereum-transactor service
			dataHex := "0x" + hex.EncodeToString(encodedData)
			_, err = s.ethereumTransactorClient.CreateTransaction(ctx, ethereum_transactor.CreateTransactionRequest{
				To:    &s.c2PAContractAddress,
				Value: "0",
				Data:  &dataHex,
				Type:  &anchorManifestType,
			})
			if err != nil {
				// Log error but don't fail the post creation
				// In production, you might want to queue this for retry
				fmt.Printf("Warning: failed to create blockchain transaction for manifest anchoring: %v\n", err)
			}
		}
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

	// Get post IDs for reaction aggregation
	postIDs := make([]uuid.UUID, len(posts))
	for i, post := range posts {
		postIDs[i] = post.ID
	}

	// Aggregate reactions
	reactionMap, err := s.aggregateReactions(ctx, postIDs, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create posts with user info and reactions
	postsWithUsers := make([]PostWithUser, len(posts))
	for i, post := range posts {
		postsWithUsers[i] = PostWithUser{
			Post:      post,
			User:      userMap[post.UserID],
			Reactions: reactionMap[post.ID],
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

	// Get current user for reaction aggregation
	currentUser, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get post IDs for reaction aggregation
	postIDs := make([]uuid.UUID, len(posts))
	for i, post := range posts {
		postIDs[i] = post.ID
	}

	// Aggregate reactions
	reactionMap, err := s.aggregateReactions(ctx, postIDs, currentUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create posts with reactions
	postsWithReactions := make([]PostWithUser, len(posts))
	for i, post := range posts {
		postsWithReactions[i] = PostWithUser{
			Post:      post,
			Reactions: reactionMap[post.ID],
		}
	}

	ctx.JSON(http.StatusOK, postsWithReactions)
}

func (s *server) GetPost(ctx *gin.Context) {
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

	// Get post by ID
	post, err := s.dbClient.Post().FindByID(ctx, postID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "post not found",
		})
		return
	}

	// Get user information
	postUser, err := s.dbClient.User().FindByID(ctx, post.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Aggregate reactions
	reactionMap, err := s.aggregateReactions(ctx, []uuid.UUID{postID}, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create post with user info and reactions
	postWithUser := PostWithUser{
		Post:      *post,
		User:      postUser,
		Reactions: reactionMap[postID],
	}

	ctx.JSON(http.StatusOK, postWithUser)
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

func (s *server) CreateReaction(ctx *gin.Context) {
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

	// Verify post exists
	_, err = s.dbClient.Post().FindByID(ctx, postID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "post not found",
		})
		return
	}

	var requestBody struct {
		Emoji string `json:"emoji" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid request body: %v", err),
		})
		return
	}

	// Validate emoji is not empty
	if requestBody.Emoji == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "emoji is required",
		})
		return
	}

	reaction, err := s.dbClient.PostReaction().CreateOrUpdate(ctx, postID, user.ID, requestBody.Emoji)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, reaction)
}

func (s *server) DeleteReaction(ctx *gin.Context) {
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

	if err := s.dbClient.PostReaction().Delete(ctx, postID, user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// aggregateReactions groups reactions by emoji and creates summaries
func (s *server) aggregateReactions(ctx *gin.Context, postIDs []uuid.UUID, currentUserID uuid.UUID) (map[uuid.UUID][]ReactionSummary, error) {
	if len(postIDs) == 0 {
		return make(map[uuid.UUID][]ReactionSummary), nil
	}

	reactions, err := s.dbClient.PostReaction().FindByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, err
	}

	// Group reactions by post ID and emoji
	postReactionMap := make(map[uuid.UUID]map[string]int)
	userReactionMap := make(map[uuid.UUID]string) // postID -> emoji that user reacted with

	for _, reaction := range reactions {
		if postReactionMap[reaction.PostID] == nil {
			postReactionMap[reaction.PostID] = make(map[string]int)
		}
		postReactionMap[reaction.PostID][reaction.Emoji]++

		if reaction.UserID == currentUserID {
			userReactionMap[reaction.PostID] = reaction.Emoji
		}
	}

	// Create summaries
	result := make(map[uuid.UUID][]ReactionSummary)
	for postID, emojiCounts := range postReactionMap {
		summaries := make([]ReactionSummary, 0, len(emojiCounts))
		for emoji, count := range emojiCounts {
			summaries = append(summaries, ReactionSummary{
				Emoji:       emoji,
				Count:       count,
				UserReacted: userReactionMap[postID] == emoji,
			})
		}
		result[postID] = summaries
	}

	// Ensure all posts have an entry (even if empty)
	for _, postID := range postIDs {
		if _, exists := result[postID]; !exists {
			result[postID] = []ReactionSummary{}
		}
	}

	return result, nil
}

func (s *server) CreateComment(ctx *gin.Context) {
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

	// Verify post exists
	_, err = s.dbClient.Post().FindByID(ctx, postID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "post not found",
		})
		return
	}

	var requestBody struct {
		Text string `json:"text" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid request body: %v", err),
		})
		return
	}

	// Validate text is not empty
	if requestBody.Text == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "comment text is required",
		})
		return
	}

	comment, err := s.dbClient.PostComment().Create(ctx, postID, user.ID, requestBody.Text)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Return comment with user info
	commentWithUser := CommentWithUser{
		PostComment: *comment,
		User:        user,
	}

	ctx.JSON(http.StatusOK, commentWithUser)
}

func (s *server) GetComments(ctx *gin.Context) {
	_, err := s.GetAuthenticatedUser(ctx)
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

	comments, err := s.dbClient.PostComment().FindByPostID(ctx, postID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get user IDs from comments
	userIDs := make(map[uuid.UUID]bool)
	for _, comment := range comments {
		userIDs[comment.UserID] = true
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

	// Create comments with user info
	commentsWithUsers := make([]CommentWithUser, len(comments))
	for i, comment := range comments {
		commentsWithUsers[i] = CommentWithUser{
			PostComment: comment,
			User:        userMap[comment.UserID],
		}
	}

	ctx.JSON(http.StatusOK, commentsWithUsers)
}

func (s *server) DeleteComment(ctx *gin.Context) {
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

	commentIDStr := ctx.Param("commentId")
	if commentIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "comment id is required",
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

	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid comment id format",
		})
		return
	}

	// Get comment to check ownership
	comment, err := s.dbClient.PostComment().FindByID(ctx, commentID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if comment == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "comment not found",
		})
		return
	}

	// Verify comment belongs to the post
	if comment.PostID != postID {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "comment does not belong to this post",
		})
		return
	}

	// Get post to check if user is post owner
	post, err := s.dbClient.Post().FindByID(ctx, postID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Authorization: user must be comment author OR post owner
	if comment.UserID != user.ID && post.UserID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "cannot delete comments from another user",
		})
		return
	}

	if err := s.dbClient.PostComment().Delete(ctx, commentID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (s *server) GetBlockchainTransactionByManifestHash(ctx *gin.Context) {
	_, err := s.GetAuthenticatedUser(ctx)
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

	// Get post to retrieve manifest hash
	post, err := s.dbClient.Post().FindByID(ctx, postID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "post not found",
		})
		return
	}

	// Check if post has a manifest hash
	if post.ManifestHash == nil || len(post.ManifestHash) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "post does not have a manifest hash",
		})
		return
	}

	// Find blockchain transaction by manifest hash
	tx, err := s.dbClient.BlockchainTransaction().FindByManifestHash(ctx, post.ManifestHash)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if tx == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "blockchain transaction not found for this manifest",
		})
		return
	}

	ctx.JSON(http.StatusOK, tx)
}
