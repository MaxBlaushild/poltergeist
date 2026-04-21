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

var errZoneKindSlugExists = stdErrors.New("zone kind slug already exists")

type zoneKindPayload struct {
	Name                      string  `json:"name"`
	Slug                      string  `json:"slug"`
	Description               string  `json:"description"`
	PlaceCountRatio           float64 `json:"placeCountRatio"`
	MonsterCountRatio         float64 `json:"monsterCountRatio"`
	BossEncounterCountRatio   float64 `json:"bossEncounterCountRatio"`
	RaidEncounterCountRatio   float64 `json:"raidEncounterCountRatio"`
	InputEncounterCountRatio  float64 `json:"inputEncounterCountRatio"`
	OptionEncounterCountRatio float64 `json:"optionEncounterCountRatio"`
	TreasureChestCountRatio   float64 `json:"treasureChestCountRatio"`
	HealingFountainCountRatio float64 `json:"healingFountainCountRatio"`
	ResourceCountRatio        float64 `json:"resourceCountRatio"`
}

func normalizeZoneKindPayload(body zoneKindPayload) (*models.ZoneKind, error) {
	name := strings.TrimSpace(body.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	slug := strings.TrimSpace(body.Slug)
	if slug == "" {
		slug = name
	}
	slug = models.NormalizeZoneKind(slug)
	if slug == "" {
		return nil, fmt.Errorf("slug is required")
	}

	return &models.ZoneKind{
		Name:                      name,
		Slug:                      slug,
		Description:               strings.TrimSpace(body.Description),
		PlaceCountRatio:           body.PlaceCountRatio,
		MonsterCountRatio:         body.MonsterCountRatio,
		BossEncounterCountRatio:   body.BossEncounterCountRatio,
		RaidEncounterCountRatio:   body.RaidEncounterCountRatio,
		InputEncounterCountRatio:  body.InputEncounterCountRatio,
		OptionEncounterCountRatio: body.OptionEncounterCountRatio,
		TreasureChestCountRatio:   body.TreasureChestCountRatio,
		HealingFountainCountRatio: body.HealingFountainCountRatio,
		ResourceCountRatio:        body.ResourceCountRatio,
	}, nil
}

func (s *server) ensureZoneKindSlugAvailable(ctx *gin.Context, slug string, currentID *uuid.UUID) error {
	existing, err := s.dbClient.ZoneKind().FindBySlug(ctx, slug)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if currentID != nil && existing.ID == *currentID {
		return nil
	}
	return errZoneKindSlugExists
}

func normalizeZoneKindAssignmentIDs(values []string) ([]uuid.UUID, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("zoneIds array cannot be empty")
	}

	seen := make(map[uuid.UUID]struct{}, len(values))
	ids := make([]uuid.UUID, 0, len(values))
	for _, raw := range values {
		id, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("invalid zone ID: %s", raw)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("zoneIds array cannot be empty")
	}
	return ids, nil
}

func (s *server) getZoneKinds(ctx *gin.Context) {
	zoneKinds, err := s.dbClient.ZoneKind().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, zoneKinds)
}

func (s *server) createZoneKind(ctx *gin.Context) {
	var requestBody zoneKindPayload
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zoneKind, err := normalizeZoneKindPayload(requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.ensureZoneKindSlugAvailable(ctx, zoneKind.Slug, nil); err != nil {
		if stdErrors.Is(err, errZoneKindSlugExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.ZoneKind().Create(ctx, zoneKind); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, zoneKind)
}

func (s *server) updateZoneKind(ctx *gin.Context) {
	id, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone kind ID"})
		return
	}

	existing, err := s.dbClient.ZoneKind().FindByID(ctx, id)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "zone kind not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody zoneKindPayload
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := normalizeZoneKindPayload(requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated.ID = existing.ID
	updated.CreatedAt = existing.CreatedAt
	oldSlug := existing.Slug

	if err := s.ensureZoneKindSlugAvailable(ctx, updated.Slug, &existing.ID); err != nil {
		if stdErrors.Is(err, errZoneKindSlugExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.ZoneKind().Update(ctx, updated); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if oldSlug != updated.Slug {
		if _, err := s.dbClient.Zone().ReplaceKind(ctx, oldSlug, updated.Slug); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if _, err := s.dbClient.ZoneSeedJob().ReplaceZoneKind(ctx, oldSlug, updated.Slug); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteZoneKind(ctx *gin.Context) {
	id, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone kind ID"})
		return
	}

	existing, err := s.dbClient.ZoneKind().FindByID(ctx, id)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "zone kind not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.dbClient.Zone().ReplaceKind(ctx, existing.Slug, ""); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.dbClient.ZoneSeedJob().ReplaceZoneKind(ctx, existing.Slug, ""); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.ZoneKind().Delete(ctx, existing.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (s *server) assignZoneKindToZones(ctx *gin.Context) {
	var requestBody struct {
		Kind    string   `json:"kind"`
		ZoneIDs []string `json:"zoneIds"`
	}
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zoneIDs, err := normalizeZoneKindAssignmentIDs(requestBody.ZoneIDs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	normalizedKind := models.NormalizeZoneKind(requestBody.Kind)
	if normalizedKind != "" {
		if _, err := s.dbClient.ZoneKind().FindBySlug(ctx, normalizedKind); err != nil {
			if stdErrors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "zone kind not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	updatedCount, err := s.dbClient.Zone().SetKind(ctx, zoneIDs, normalizedKind)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"updatedCount": updatedCount,
		"kind":         normalizedKind,
	})
}
