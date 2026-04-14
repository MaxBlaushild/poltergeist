package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type characterTemplateUpsertRequest struct {
	Name             string                         `json:"name"`
	Description      string                         `json:"description"`
	InternalTags     []string                       `json:"internalTags"`
	StoryVariants    []models.CharacterStoryVariant `json:"storyVariants"`
	MapIconURL       string                         `json:"mapIconUrl"`
	DialogueImageURL string                         `json:"dialogueImageUrl"`
	ThumbnailURL     string                         `json:"thumbnailUrl"`
}

func parseCharacterTemplateUpsertRequest(
	body characterTemplateUpsertRequest,
) (*models.CharacterTemplate, error) {
	name := strings.TrimSpace(body.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	dialogueImageURL := strings.TrimSpace(body.DialogueImageURL)
	thumbnailURL := strings.TrimSpace(body.ThumbnailURL)
	if thumbnailURL == "" && dialogueImageURL != "" {
		thumbnailURL = dialogueImageURL
	}
	return &models.CharacterTemplate{
		Name:             name,
		Description:      strings.TrimSpace(body.Description),
		InternalTags:     parseCharacterInternalTags(body.InternalTags),
		StoryVariants:    normalizeCharacterStoryVariants(body.StoryVariants),
		MapIconURL:       strings.TrimSpace(body.MapIconURL),
		DialogueImageURL: dialogueImageURL,
		ThumbnailURL:     thumbnailURL,
		ImageGenerationStatus: func() string {
			if dialogueImageURL != "" {
				return models.CharacterImageGenerationStatusComplete
			}
			return models.CharacterImageGenerationStatusNone
		}(),
	}, nil
}

func (s *server) getCharacterTemplates(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	templates, err := s.dbClient.CharacterTemplate().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, templates)
}

func (s *server) getCharacterTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character template ID"})
		return
	}
	template, err := s.dbClient.CharacterTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if template == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character template not found"})
		return
	}
	ctx.JSON(http.StatusOK, template)
}

func (s *server) createCharacterTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	var body characterTemplateUpsertRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, err := parseCharacterTemplateUpsertRequest(body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.CharacterTemplate().Create(ctx, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	created, err := s.dbClient.CharacterTemplate().FindByID(ctx, template.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

func (s *server) updateCharacterTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character template ID"})
		return
	}
	existing, err := s.dbClient.CharacterTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "character template not found"})
		return
	}
	var body characterTemplateUpsertRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, err := parseCharacterTemplateUpsertRequest(body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.CharacterTemplate().Update(ctx, id, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updated, err := s.dbClient.CharacterTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteCharacterTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid character template ID"})
		return
	}
	if err := s.dbClient.CharacterTemplate().Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "character template deleted successfully"})
}
