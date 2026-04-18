package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func chaosEngineRequiredInventoryItemID(config models.MetadataJSONB) (int, bool) {
	if len(config) == 0 {
		return 0, false
	}
	raw, ok := config["requiredInventoryItemId"]
	if !ok || raw == nil {
		return 0, false
	}
	switch value := raw.(type) {
	case int:
		if value > 0 {
			return value, true
		}
	case int32:
		if value > 0 {
			return int(value), true
		}
	case int64:
		if value > 0 {
			return int(value), true
		}
	case float64:
		if value > 0 {
			return int(value), true
		}
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err == nil && parsed > 0 {
			return parsed, true
		}
	}
	return 0, false
}

func (s *server) useBaseChaosEngine(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var body struct {
		ZoneID  string `json:"zoneId"`
		GenreID string `json:"genreId"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zoneID, err := uuid.Parse(strings.TrimSpace(body.ZoneID))
	if err != nil || zoneID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}
	genreID, err := uuid.Parse(strings.TrimSpace(body.GenreID))
	if err != nil || genreID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid genre ID"})
		return
	}

	base, err := s.dbClient.Base().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if base == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "you need a base before you can use the Chaos Engine"})
		return
	}

	definition, err := s.dbClient.BaseStructureDefinition().FindActiveByKey(ctx, "chaos_engine")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Chaos Engine Room definition not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	requiredInventoryItemID, ok := chaosEngineRequiredInventoryItemID(definition.EffectConfig)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Chaos Engine Room is not configured with a required item"})
		return
	}

	structures, err := s.dbClient.UserBaseStructure().FindByBaseID(ctx, base.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	chaosEngineBuilt := false
	for _, structure := range structures {
		if structure.StructureKey == "chaos_engine" && structure.Level > 0 {
			chaosEngineBuilt = true
			break
		}
	}
	if !chaosEngineBuilt {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "you need to build the Chaos Engine Room before you can use it"})
		return
	}

	zone, err := s.dbClient.Zone().FindByID(ctx, zoneID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	genre, err := s.dbClient.ZoneGenre().FindByID(ctx, genreID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "genre not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !genre.Active {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "that genre is not currently active"})
		return
	}

	requiredInventoryItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, requiredInventoryItemID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "the configured Chaos Engine item no longer exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if requiredInventoryItem == nil || requiredInventoryItem.Archived {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "the configured Chaos Engine item is unavailable"})
		return
	}

	updatedScore, err := s.dbClient.ZoneGenreScore().ConsumeUserItemAndIncrement(
		ctx,
		user.ID,
		requiredInventoryItemID,
		zoneID,
		genreID,
		1,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(strings.ToLower(err.Error()), "insufficient quantity") {
			ctx.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("you need a %s to fuel the Chaos Engine", strings.TrimSpace(requiredInventoryItem.Name))})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedZone, err := s.dbClient.Zone().FindByID(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	serializedZone, err := s.serializeSingleZoneWithGenres(ctx, updatedZone)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":                 fmt.Sprintf("%s gains 1 %s point.", strings.TrimSpace(zone.Name), strings.TrimSpace(genre.Name)),
		"zone":                    serializedZone,
		"genre":                   serializeZoneGenre(*genre),
		"newScore":                updatedScore.Score,
		"requiredInventoryItemId": requiredInventoryItemID,
		"requiredInventoryItem":   requiredInventoryItem,
	})
}
