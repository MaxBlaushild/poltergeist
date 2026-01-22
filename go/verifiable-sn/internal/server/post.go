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
		ImageURL        string  `json:"imageUrl" binding:"required"`
		Caption         *string `json:"caption"`
		ManifestURL     *string `json:"manifestUrl"`
		ManifestHash    *string `json:"manifestHash"`
		CertFingerprint *string `json:"certFingerprint"`
		AssetID         *string `json:"assetId"`
	}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var manifestHashBytes []byte
	var manifestURI *string
	var certFingerprintBytes []byte
	var assetID *string

	// If manifest data is provided, validate it
	if requestBody.ManifestURL != nil && requestBody.ManifestHash != nil && requestBody.CertFingerprint != nil {
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

