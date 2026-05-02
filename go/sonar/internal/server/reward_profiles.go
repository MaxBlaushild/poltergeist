package server

import (
	stdErrors "errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var errRewardProfileSlugExists = fmt.Errorf("reward profile slug already exists")

type rewardProfilePayload struct {
	Name                      string   `json:"name"`
	Slug                      string   `json:"slug"`
	Description               string   `json:"description"`
	Active                    bool     `json:"active"`
	PreferredItemTags         []string `json:"preferredItemTags"`
	PreferredMaterialKeys     []string `json:"preferredMaterialKeys"`
	PreferredDamageAffinities []string `json:"preferredDamageAffinities"`
	PreferredResourceTypeIDs  []string `json:"preferredResourceTypeIds"`
	PreferEquipment           bool     `json:"preferEquipment"`
	PreferUtility             bool     `json:"preferUtility"`
	PreferKnowledge           bool     `json:"preferKnowledge"`
	PreferNonEquipment        bool     `json:"preferNonEquipment"`
}

func normalizeRewardProfileMaterialKeys(values []string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	seen := map[models.BaseResourceKey]struct{}{}
	for idx, raw := range values {
		resourceKey := models.NormalizeBaseResourceKey(raw)
		if resourceKey == "" {
			if strings.TrimSpace(raw) == "" {
				continue
			}
			return nil, fmt.Errorf("preferredMaterialKeys[%d] is invalid", idx)
		}
		if _, exists := seen[resourceKey]; exists {
			continue
		}
		seen[resourceKey] = struct{}{}
		normalized = append(normalized, string(resourceKey))
	}
	return normalized, nil
}

func normalizeRewardProfileResourceTypeIDs(values []string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	seen := map[uuid.UUID]struct{}{}
	for idx, raw := range values {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		id, err := uuid.Parse(trimmed)
		if err != nil || id == uuid.Nil {
			return nil, fmt.Errorf("preferredResourceTypeIds[%d] is invalid", idx)
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id.String())
	}
	return normalized, nil
}

func normalizeRewardProfilePayload(payload rewardProfilePayload) (*models.RewardProfile, error) {
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	preferredMaterialKeys, err := normalizeRewardProfileMaterialKeys(payload.PreferredMaterialKeys)
	if err != nil {
		return nil, err
	}
	preferredResourceTypeIDs, err := normalizeRewardProfileResourceTypeIDs(payload.PreferredResourceTypeIDs)
	if err != nil {
		return nil, err
	}

	rewardProfile := &models.RewardProfile{
		Slug:                      payload.Slug,
		Name:                      name,
		Description:               strings.TrimSpace(payload.Description),
		Active:                    payload.Active,
		PreferredItemTags:         models.StringArray(models.NormalizeTagList(payload.PreferredItemTags)),
		PreferredMaterialKeys:     models.StringArray(preferredMaterialKeys),
		PreferredDamageAffinities: models.StringArray(models.NormalizeTagList(payload.PreferredDamageAffinities)),
		PreferredResourceTypeIDs:  models.StringArray(preferredResourceTypeIDs),
		PreferEquipment:           payload.PreferEquipment,
		PreferUtility:             payload.PreferUtility,
		PreferKnowledge:           payload.PreferKnowledge,
		PreferNonEquipment:        payload.PreferNonEquipment,
	}
	if rewardProfile.Slug == "" {
		rewardProfile.Slug = models.NormalizeRewardProfileSlug(rewardProfile.Name)
	}
	return rewardProfile, nil
}

func (s *server) ensureRewardProfileSlugAvailable(
	ctx *gin.Context,
	slug string,
	currentID *uuid.UUID,
) error {
	existing, err := s.dbClient.RewardProfile().FindBySlug(ctx, slug)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if existing == nil {
		return nil
	}
	if currentID != nil && existing.ID == *currentID {
		return nil
	}
	return errRewardProfileSlugExists
}

func (s *server) getRewardProfiles(ctx *gin.Context) {
	rewardProfiles, err := s.dbClient.RewardProfile().FindAll(ctx, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, rewardProfiles)
}

func (s *server) createRewardProfile(ctx *gin.Context) {
	var payload rewardProfilePayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rewardProfile, err := normalizeRewardProfilePayload(payload)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.ensureRewardProfileSlugAvailable(ctx, rewardProfile.Slug, nil); err != nil {
		if stdErrors.Is(err, errRewardProfileSlugExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.RewardProfile().Create(ctx, rewardProfile); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, rewardProfile)
}

func (s *server) updateRewardProfile(ctx *gin.Context) {
	id, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid reward profile ID"})
		return
	}

	existing, err := s.dbClient.RewardProfile().FindByID(ctx, id)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "reward profile not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var payload rewardProfilePayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := normalizeRewardProfilePayload(payload)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated.ID = existing.ID
	updated.CreatedAt = existing.CreatedAt
	if err := s.ensureRewardProfileSlugAvailable(ctx, updated.Slug, &existing.ID); err != nil {
		if stdErrors.Is(err, errRewardProfileSlugExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.RewardProfile().Update(ctx, updated); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteRewardProfile(ctx *gin.Context) {
	id, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid reward profile ID"})
		return
	}

	existing, err := s.dbClient.RewardProfile().FindByID(ctx, id)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "reward profile not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.RewardProfile().Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, existing)
}
