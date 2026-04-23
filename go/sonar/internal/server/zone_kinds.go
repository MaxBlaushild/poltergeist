package server

import (
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

var errZoneKindSlugExists = stdErrors.New("zone kind slug already exists")

type zoneKindPayload struct {
	Name                        string  `json:"name"`
	Slug                        string  `json:"slug"`
	Description                 string  `json:"description"`
	OverlayColor                string  `json:"overlayColor"`
	PlaceCountRatio             float64 `json:"placeCountRatio"`
	MonsterCountRatio           float64 `json:"monsterCountRatio"`
	BossEncounterCountRatio     float64 `json:"bossEncounterCountRatio"`
	RaidEncounterCountRatio     float64 `json:"raidEncounterCountRatio"`
	InputEncounterCountRatio    float64 `json:"inputEncounterCountRatio"`
	OptionEncounterCountRatio   float64 `json:"optionEncounterCountRatio"`
	TreasureChestCountRatio     float64 `json:"treasureChestCountRatio"`
	HealingFountainCountRatio   float64 `json:"healingFountainCountRatio"`
	HerbalismResourceCountRatio float64 `json:"herbalismResourceCountRatio"`
	MiningResourceCountRatio    float64 `json:"miningResourceCountRatio"`
	ResourceCountRatio          float64 `json:"resourceCountRatio"`
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
	overlayColor := models.NormalizeHexColor(body.OverlayColor)
	if strings.TrimSpace(body.OverlayColor) != "" && overlayColor == "" {
		return nil, fmt.Errorf("overlayColor must be a valid hex color like #5f7d68")
	}
	herbalismRatio := body.HerbalismResourceCountRatio
	miningRatio := body.MiningResourceCountRatio
	if herbalismRatio == 0 && miningRatio == 0 && body.ResourceCountRatio > 0 {
		herbalismRatio = body.ResourceCountRatio
		miningRatio = body.ResourceCountRatio
	}
	legacyResourceRatio := (herbalismRatio + miningRatio) / 2

	return &models.ZoneKind{
		Name:                        name,
		Slug:                        slug,
		Description:                 strings.TrimSpace(body.Description),
		OverlayColor:                overlayColor,
		PlaceCountRatio:             body.PlaceCountRatio,
		MonsterCountRatio:           body.MonsterCountRatio,
		BossEncounterCountRatio:     body.BossEncounterCountRatio,
		RaidEncounterCountRatio:     body.RaidEncounterCountRatio,
		InputEncounterCountRatio:    body.InputEncounterCountRatio,
		OptionEncounterCountRatio:   body.OptionEncounterCountRatio,
		TreasureChestCountRatio:     body.TreasureChestCountRatio,
		HealingFountainCountRatio:   body.HealingFountainCountRatio,
		HerbalismResourceCountRatio: herbalismRatio,
		MiningResourceCountRatio:    miningRatio,
		ResourceCountRatio:          legacyResourceRatio,
	}, nil
}

type zoneKindPatternCue struct {
	label string
	value float64
}

func maxZoneKindPatternCueValue(a float64, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func zoneKindPatternCues(zoneKind models.ZoneKind) []string {
	candidates := []zoneKindPatternCue{
		{label: "place-rich", value: zoneKind.PlaceCountRatio},
		{label: "monster-heavy", value: zoneKind.MonsterCountRatio},
		{label: "boss-dangerous", value: zoneKind.BossEncounterCountRatio},
		{label: "raid-heavy", value: zoneKind.RaidEncounterCountRatio},
		{
			label: "scenario-rich",
			value: maxZoneKindPatternCueValue(
				zoneKind.InputEncounterCountRatio,
				zoneKind.OptionEncounterCountRatio,
			),
		},
		{label: "treasure-rich", value: zoneKind.TreasureChestCountRatio},
		{label: "restorative", value: zoneKind.HealingFountainCountRatio},
		{label: "herbalism-rich", value: zoneKind.HerbalismResourceCountRatio},
		{label: "mining-rich", value: zoneKind.MiningResourceCountRatio},
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].value > candidates[j].value
	})

	cues := make([]string, 0, 3)
	for _, candidate := range candidates {
		if candidate.value <= 1.05 {
			continue
		}
		cues = append(cues, candidate.label)
		if len(cues) == 3 {
			break
		}
	}
	if len(cues) == 0 {
		cues = append(cues, "mixed frontier")
	}
	return cues
}

func zoneKindPatternMotifs(zoneKind models.ZoneKind) string {
	switch models.NormalizeZoneKind(zoneKind.Slug) {
	case "forest":
		return "leaf clusters, canopy blotches, branch forks, trail scratches"
	case "swamp":
		return "reed strokes, puddle curves, marsh ripples, hanging moss"
	case "volcanic":
		return "lava cracks, ember seams, ash flecks, broken magma lines"
	case "graveyard":
		return "worn stone hash marks, grave glyphs, crosshatched weathering"
	case "desert":
		return "wind-carved dune lines, grit speckle, drifting wave contours"
	case "temple-grounds":
		return "sacred rings, shrine geometry, halo lines, ceremonial inlay"
	case "city":
		return "street grids, masonry blocks, alley runs, civic linework"
	case "industrial":
		return "rivet grids, pipe runs, hazard striping, forged plate seams"
	case "farmland":
		return "furrow stripes, field parcels, stitched paths, crop rows"
	case "academy":
		return "arcane circles, star-compass marks, inked diagram lines"
	case "village":
		return "cottage roof rhythms, lantern marks, fence lines, paths"
	case "badlands":
		return "sun-baked fractures, mesa ridges, eroded gullies"
	case "highlands":
		return "wind-swept contours, cairn-like marks, ridge bands"
	case "mountain":
		return "rock strata, sharp peak silhouettes, mineral seams"
	case "ruins":
		return "cracked tiles, broken arch fragments, chipped stone patterns"
	default:
		return "subtle fantasy map texture marks, organic linework, exploratory symbols"
	}
}

func zoneKindPatternPaletteAnchor(zoneKind models.ZoneKind) string {
	overlayColor := strings.TrimSpace(zoneKind.OverlayColor)
	if overlayColor != "" {
		return overlayColor
	}
	return "#5f7d68"
}

func defaultZoneKindPatternTilePrompt(zoneKind models.ZoneKind) string {
	name := strings.TrimSpace(zoneKind.Name)
	if name == "" {
		name = "Frontier"
	}
	slug := models.NormalizeZoneKind(zoneKind.Slug)
	if slug == "" {
		slug = models.NormalizeZoneKind(name)
	}
	description := strings.TrimSpace(zoneKind.Description)
	if description == "" {
		description = "A fantasy zone texture used as a readable map overlay."
	}

	return fmt.Sprintf(
		`Create a seamless repeating square texture tile for a fantasy RPG world map overlay.

Zone kind:
- name: %s
- slug: %s
- description: %s
- dominant gameplay cues: %s
- palette anchor: %s
- motif direction: %s

Requirements:
- The tile must repeat seamlessly on all four edges.
- This is not a full scene, landscape illustration, or diorama. It should read like an ornamental map texture.
- Use a bold, game-ready pattern with clear motif language that will still read on top of a watercolor fantasy map.
- Keep enough negative space for the basemap to show through, but do not make the marks timid or faint.
- Keep the motifs medium-scale, high-contrast, and clearly legible when repeated across a polygon.
- Aim for a retro fantasy RPG feel: adventure-manual map art, classic JRPG overworld texture language, hand-inked symbols, weathered parchment energy, and old-school exploratory charm.
- Favor stylized 16-bit / early-32-bit era fantasy sensibilities over modern glossy concept-art polish.
- No border, no frame, no text, no logos, no single centered subject.
- Square composition only.
- Top-down graphic texture language, never perspective or isometric.
- Fantasy RPG tone, handcrafted, slightly stylized, polished, tasteful.
- Avoid photorealism, modern UI gradients, and sterile vector-flat design.
`,
		name,
		slug,
		description,
		strings.Join(zoneKindPatternCues(zoneKind), ", "),
		zoneKindPatternPaletteAnchor(zoneKind),
		zoneKindPatternMotifs(zoneKind),
	)
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
	updated.PatternTileURL = existing.PatternTileURL
	updated.PatternTilePrompt = existing.PatternTilePrompt
	updated.PatternTileGenerationStatus = existing.PatternTileGenerationStatus
	updated.PatternTileGenerationError = existing.PatternTileGenerationError
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
		if _, err := s.dbClient.ZoneKind().ReplaceReferences(ctx, oldSlug, updated.Slug); err != nil {
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
	if _, err := s.dbClient.ZoneKind().ReplaceReferences(ctx, existing.Slug, ""); err != nil {
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

func (s *server) generateZoneKindPatternTile(ctx *gin.Context) {
	id, err := uuid.Parse(strings.TrimSpace(ctx.Param("id")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone kind ID"})
		return
	}

	zoneKind, err := s.dbClient.ZoneKind().FindByID(ctx, id)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "zone kind not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody struct {
		Prompt string `json:"prompt"`
	}
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prompt := strings.TrimSpace(requestBody.Prompt)
	if prompt == "" {
		prompt = defaultZoneKindPatternTilePrompt(*zoneKind)
	}
	if len(prompt) < 24 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "prompt must be at least 24 characters"})
		return
	}
	if len(prompt) > 8000 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "prompt must be at most 8000 characters"})
		return
	}

	zoneKind.PatternTilePrompt = prompt
	zoneKind.PatternTileGenerationStatus = models.ZoneKindPatternTileGenerationStatusQueued
	zoneKind.PatternTileGenerationError = ""
	if err := s.dbClient.ZoneKind().Update(ctx, zoneKind); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payloadBytes, err := json.Marshal(jobs.GenerateZoneKindPatternTileTaskPayload{
		ZoneKindID: zoneKind.ID,
		Prompt:     prompt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateZoneKindPatternTileTaskType, payloadBytes)); err != nil {
		errMsg := err.Error()
		zoneKind.PatternTileGenerationStatus = models.ZoneKindPatternTileGenerationStatusFailed
		zoneKind.PatternTileGenerationError = errMsg
		_ = s.dbClient.ZoneKind().Update(ctx, zoneKind)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	ctx.JSON(http.StatusAccepted, zoneKind)
}
