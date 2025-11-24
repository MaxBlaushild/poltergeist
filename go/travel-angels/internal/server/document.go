package server

import (
	"io"
	"net/http"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/travel-angels/internal/parser"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateDocumentRequest struct {
	Title          string   `json:"title" binding:"required"`
	Provider       string   `json:"provider" binding:"required"`
	Link           *string  `json:"link"`
	Content        *string  `json:"content"`
	ExistingTagIds []string `json:"existingTagIds"`
	NewTagTexts    []string `json:"newTagTexts"`
}

func (s *server) CreateDocument(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	var req CreateDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate provider
	provider := models.CloudDocumentProvider(req.Provider)
	switch provider {
	case models.CloudDocumentProviderUnknown,
		models.CloudDocumentProviderGoogleDocs,
		models.CloudDocumentProviderGoogleSheets,
		models.CloudDocumentProviderInternal:
		// Valid provider
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid provider. Must be one of: unknown, google_docs, google_sheets, internal",
		})
		return
	}

	// Create document model
	document := &models.Document{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Title:     req.Title,
		Provider:  provider,
		UserID:    user.ID,
		Link:      req.Link,
		Content:   req.Content,
	}

	// Convert existing tag IDs from strings to UUIDs
	existingTagIDs := make([]uuid.UUID, 0, len(req.ExistingTagIds))
	for _, tagIDStr := range req.ExistingTagIds {
		tagID, err := uuid.Parse(tagIDStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid tag ID format: " + tagIDStr,
			})
			return
		}
		existingTagIDs = append(existingTagIDs, tagID)
	}

	// Create document with tags
	createdDocument, err := s.dbClient.Document().Create(ctx, document, existingTagIDs, req.NewTagTexts)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, createdDocument)
}

func (s *server) GetDocumentsByUserID(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Parse userId from URL param
	userIDStr := ctx.Param("userId")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "userId is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid userId format",
		})
		return
	}

	// Verify userId matches authenticated user (users can only see their own documents)
	if user.ID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "cannot access documents for another user",
		})
		return
	}

	// Get documents for user
	documents, err := s.dbClient.Document().FindByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, documents)
}

func (s *server) GetFriendsDocuments(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get user's friends (returns []User)
	friends, err := s.dbClient.Friend().FindAllFriends(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch friends: " + err.Error(),
		})
		return
	}

	// Extract friend user IDs
	if len(friends) == 0 {
		ctx.JSON(http.StatusOK, []models.Document{})
		return
	}

	friendIDs := make([]uuid.UUID, 0, len(friends))
	for _, friend := range friends {
		friendIDs = append(friendIDs, friend.ID)
	}

	// Get documents for all friends
	documents, err := s.dbClient.Document().FindByUserIDs(ctx, friendIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch friends' documents: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, documents)
}

func (s *server) ParseDocument(ctx *gin.Context) {
	_, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get the file from multipart form
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "file is required: " + err.Error(),
		})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to open file: " + err.Error(),
		})
		return
	}
	defer src.Close()

	// Read file bytes
	fileBytes := make([]byte, file.Size)
	_, err = io.ReadFull(src, fileBytes)
	if err != nil && err != io.EOF {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to read file: " + err.Error(),
		})
		return
	}

	// Initialize parser
	documentParser := parser.NewDocumentParser()

	// Parse the document
	parsedDoc, err := documentParser.ParseDocument(fileBytes)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to parse document: " + err.Error(),
		})
		return
	}

	// Return parsed document
	ctx.JSON(http.StatusOK, parsedDoc)
}

type UpdateDocumentRequest struct {
	Title          *string  `json:"title"`
	Content        *string  `json:"content"`
	ExistingTagIds []string `json:"existingTagIds"`
	NewTagTexts    []string `json:"newTagTexts"`
}

func (s *server) UpdateDocument(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Parse document ID from URL param
	documentIDStr := ctx.Param("id")
	if documentIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "document id is required",
		})
		return
	}

	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid document id format",
		})
		return
	}

	// Get the document to verify ownership
	document, err := s.dbClient.Document().FindByID(ctx, documentID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "document not found",
		})
		return
	}

	// Verify user owns the document
	if document.UserID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "cannot update documents for another user",
		})
		return
	}

	// Parse request body
	var req UpdateDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Update document fields if provided
	if req.Title != nil {
		document.Title = *req.Title
	}
	if req.Content != nil {
		document.Content = req.Content
	}

	// Update the document
	if err := s.dbClient.Document().Update(ctx, document); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Handle tag updates if provided
	if req.ExistingTagIds != nil || req.NewTagTexts != nil {
		// Convert existing tag IDs from strings to UUIDs
		existingTagIDs := make([]uuid.UUID, 0, len(req.ExistingTagIds))
		for _, tagIDStr := range req.ExistingTagIds {
			tagID, err := uuid.Parse(tagIDStr)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid tag ID format: " + tagIDStr,
				})
				return
			}
			existingTagIDs = append(existingTagIDs, tagID)
		}

		// Get document tag handler
		documentTagHandler := s.dbClient.DocumentTag()

		// Collect all tags to associate
		var tagsToAssociate []models.DocumentTag

		// Handle existing tag IDs
		if len(existingTagIDs) > 0 {
			existingTags, err := documentTagHandler.FindByIDs(ctx, existingTagIDs)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error": "failed to find existing tags: " + err.Error(),
				})
				return
			}
			tagsToAssociate = append(tagsToAssociate, existingTags...)
		}

		// Handle new tag texts
		if len(req.NewTagTexts) > 0 {
			for _, tagText := range req.NewTagTexts {
				if tagText == "" {
					continue
				}
				newTag, err := documentTagHandler.FindOrCreateByText(ctx, tagText)
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{
						"error": "failed to create tag: " + err.Error(),
					})
					return
				}
				tagsToAssociate = append(tagsToAssociate, *newTag)
			}
		}

		// Update tags using the document handler's UpdateTags method
		if err := s.dbClient.Document().UpdateTags(ctx, documentID, existingTagIDs, req.NewTagTexts); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to update document tags: " + err.Error(),
			})
			return
		}
	}

	// Reload document with tags preloaded
	updatedDocument, err := s.dbClient.Document().FindByID(ctx, documentID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve updated document",
		})
		return
	}

	ctx.JSON(http.StatusOK, updatedDocument)
}

func (s *server) DeleteDocument(ctx *gin.Context) {
	user, err := s.GetAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Parse document ID from URL param
	documentIDStr := ctx.Param("id")
	if documentIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "document id is required",
		})
		return
	}

	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid document id format",
		})
		return
	}

	// Get the document to verify ownership
	document, err := s.dbClient.Document().FindByID(ctx, documentID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "document not found",
		})
		return
	}

	// Verify user owns the document
	if document.UserID != user.ID {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "cannot delete documents for another user",
		})
		return
	}

	// Delete the document
	if err := s.dbClient.Document().Delete(ctx, documentID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}
