package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type expositionTemplateUpsertRequest struct {
	Title              string                       `json:"title"`
	Description        string                       `json:"description"`
	Dialogue           []models.DialogueMessage     `json:"dialogue"`
	RequiredStoryFlags []string                     `json:"requiredStoryFlags"`
	ImageURL           string                       `json:"imageUrl"`
	ThumbnailURL       string                       `json:"thumbnailUrl"`
	RewardMode         string                       `json:"rewardMode"`
	RandomRewardSize   string                       `json:"randomRewardSize"`
	RewardExperience   int                          `json:"rewardExperience"`
	RewardGold         int                          `json:"rewardGold"`
	MaterialRewards    []baseMaterialRewardPayload  `json:"materialRewards"`
	ItemRewards        []scenarioRewardItemPayload  `json:"itemRewards"`
	SpellRewards       []scenarioRewardSpellPayload `json:"spellRewards"`
}

func (s *server) parseExpositionTemplateUpsertRequest(
	ctx *gin.Context,
	body expositionTemplateUpsertRequest,
) (*models.ExpositionTemplate, error) {
	title := strings.TrimSpace(body.Title)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	dialogue, err := s.parseExpositionDialogue(ctx, body.Dialogue)
	if err != nil {
		return nil, err
	}
	if body.RewardExperience < 0 || body.RewardGold < 0 {
		return nil, fmt.Errorf("reward values must be zero or greater")
	}
	materialRewards, err := parseBaseMaterialRewards(body.MaterialRewards, "materialRewards")
	if err != nil {
		return nil, err
	}
	itemRewards, err := s.parseExpositionItemRewards(body.ItemRewards)
	if err != nil {
		return nil, err
	}
	spellRewards, err := s.parseExpositionSpellRewards(ctx, body.SpellRewards)
	if err != nil {
		return nil, err
	}
	rewardMode := models.NormalizeRewardMode(body.RewardMode)
	if strings.TrimSpace(body.RewardMode) == "" {
		if body.RewardExperience > 0 ||
			body.RewardGold > 0 ||
			len(materialRewards) > 0 ||
			len(itemRewards) > 0 ||
			len(spellRewards) > 0 {
			rewardMode = models.RewardModeExplicit
		}
	}
	imageURL := strings.TrimSpace(body.ImageURL)
	thumbnailURL := strings.TrimSpace(body.ThumbnailURL)
	if thumbnailURL == "" && imageURL != "" {
		thumbnailURL = imageURL
	}
	templateItemRewards := make(models.ExpositionTemplateItemRewards, 0, len(itemRewards))
	for _, reward := range itemRewards {
		templateItemRewards = append(templateItemRewards, models.ExpositionTemplateItemReward{
			InventoryItemID: reward.InventoryItemID,
			Quantity:        reward.Quantity,
		})
	}
	templateSpellRewards := make(models.ExpositionTemplateSpellRewards, 0, len(spellRewards))
	for _, reward := range spellRewards {
		templateSpellRewards = append(templateSpellRewards, models.ExpositionTemplateSpellReward{
			SpellID: reward.SpellID,
		})
	}
	return &models.ExpositionTemplate{
		Title:              title,
		Description:        strings.TrimSpace(body.Description),
		Dialogue:           dialogue,
		RequiredStoryFlags: normalizeStoryFlagKeys(body.RequiredStoryFlags),
		ImageURL:           imageURL,
		ThumbnailURL:       thumbnailURL,
		RewardMode:         rewardMode,
		RandomRewardSize:   models.NormalizeRandomRewardSize(body.RandomRewardSize),
		RewardExperience:   body.RewardExperience,
		RewardGold:         body.RewardGold,
		MaterialRewards:    materialRewards,
		ItemRewards:        templateItemRewards,
		SpellRewards:       templateSpellRewards,
	}, nil
}

func (s *server) getExpositionTemplates(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	templates, err := s.dbClient.ExpositionTemplate().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, templates)
}

func (s *server) getExpositionTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid exposition template ID"})
		return
	}
	template, err := s.dbClient.ExpositionTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if template == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "exposition template not found"})
		return
	}
	ctx.JSON(http.StatusOK, template)
}

func (s *server) createExpositionTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	var body expositionTemplateUpsertRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, err := s.parseExpositionTemplateUpsertRequest(ctx, body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.ExpositionTemplate().Create(ctx, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	created, err := s.dbClient.ExpositionTemplate().FindByID(ctx, template.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

func (s *server) updateExpositionTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid exposition template ID"})
		return
	}
	existing, err := s.dbClient.ExpositionTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "exposition template not found"})
		return
	}
	var body expositionTemplateUpsertRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	template, err := s.parseExpositionTemplateUpsertRequest(ctx, body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.ExpositionTemplate().Update(ctx, id, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updated, err := s.dbClient.ExpositionTemplate().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteExpositionTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid exposition template ID"})
		return
	}
	if err := s.dbClient.ExpositionTemplate().Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "exposition template deleted successfully"})
}
