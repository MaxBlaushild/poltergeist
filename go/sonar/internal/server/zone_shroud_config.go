package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

func defaultZoneShroudPatternTilePrompt() string {
	return `Create a seamless repeating square texture tile for a fantasy RPG fog-of-war overlay.

Shroud treatment:
- purpose: conceal undiscovered map zones without revealing their biome or zone kind
- palette direction: cool slate, smoke-gray, muted blue-charcoal, faint parchment haze
- motif direction: mist bands, stipple grain, weathered cartographic haze, drifting fog curls, restrained exploration marks

Requirements:
- The tile must repeat seamlessly on all four edges.
- This is not a scene, landscape illustration, or full environment painting.
- It should read like a fog-of-war veil or shroud laid over a fantasy world map.
- Keep the pattern atmospheric, elegant, and readable on top of a watercolor fantasy basemap.
- Use a retro fantasy RPG feel: adventure-manual map art, classic JRPG overworld texture language, hand-inked fog motifs, weathered parchment energy.
- Keep enough negative space for the base shroud color to read through.
- Avoid revealing any specific biome, settlement, or zone identity.
- No border, no frame, no text, no logos, no centered emblem.
- Square composition only.
- Top-down graphic texture language, never perspective or isometric.
- Avoid photorealism, modern UI gradients, and sterile vector-flat design.`
}

func resolveZoneShroudPatternTilePrompt(config *models.ZoneShroudConfig) string {
	if config != nil {
		if prompt := strings.TrimSpace(config.PatternTilePrompt); prompt != "" {
			return prompt
		}
	}
	return defaultZoneShroudPatternTilePrompt()
}

func (s *server) getZoneShroudConfig(ctx *gin.Context) {
	config, err := s.dbClient.ZoneShroudConfig().Get(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(config.PatternTilePrompt) == "" {
		config.PatternTilePrompt = defaultZoneShroudPatternTilePrompt()
	}
	ctx.JSON(http.StatusOK, config)
}

func (s *server) generateZoneShroudPatternTile(ctx *gin.Context) {
	config, err := s.dbClient.ZoneShroudConfig().Get(ctx)
	if err != nil {
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
		prompt = resolveZoneShroudPatternTilePrompt(config)
	}
	if len(prompt) < 24 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "prompt must be at least 24 characters"})
		return
	}
	if len(prompt) > 8000 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "prompt must be at most 8000 characters"})
		return
	}

	config.PatternTilePrompt = prompt
	config.PatternTileGenerationStatus = models.ZoneKindPatternTileGenerationStatusQueued
	config.PatternTileGenerationError = ""
	if _, err := s.dbClient.ZoneShroudConfig().Upsert(ctx, config); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payloadBytes, err := json.Marshal(jobs.GenerateZoneShroudPatternTileTaskPayload{
		Prompt: prompt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateZoneShroudPatternTileTaskType, payloadBytes)); err != nil {
		errMsg := err.Error()
		config.PatternTileGenerationStatus = models.ZoneKindPatternTileGenerationStatusFailed
		config.PatternTileGenerationError = errMsg
		_, _ = s.dbClient.ZoneShroudConfig().Upsert(ctx, config)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	ctx.JSON(http.StatusAccepted, config)
}
