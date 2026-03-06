package server

import (
	stdErrors "errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	healingFountainInteractRadiusMeters = 50.0
	healingFountainCooldownDuration     = 7 * 24 * time.Hour
	defaultHealingFountainImagePrompt   = "A discovered magical healing fountain in a retro 16-bit RPG style. Top-down map-ready icon art, luminous water, ancient stone basin, mystic runes, no text, no logos, centered composition, crisp outlines, limited palette."
)

type healingFountainWithUserStatus struct {
	models.HealingFountain
	Discovered               bool       `json:"discovered"`
	AvailableNow             bool       `json:"availableNow"`
	LastUsedAt               *time.Time `json:"lastUsedAt,omitempty"`
	NextAvailableAt          *time.Time `json:"nextAvailableAt,omitempty"`
	CooldownSecondsRemaining int        `json:"cooldownSecondsRemaining"`
}

type healingFountainUpsertRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	ThumbnailURL string   `json:"thumbnailUrl"`
	ZoneID       string   `json:"zoneId"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
}

func (s *server) getHealingFountains(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	fountains, err := s.dbClient.HealingFountain().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	latestVisitsByFountain, err := s.dbClient.HealingFountain().FindLatestVisitsByUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	discoveredByFountain, err := s.userHealingFountainDiscoveryMap(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]healingFountainWithUserStatus, 0, len(fountains))
	for _, fountain := range fountains {
		if fountain.Invalidated {
			continue
		}
		status := healingFountainCooldownStatusFromVisit(
			latestVisitsByFountain[fountain.ID],
			now,
		)
		response = append(
			response,
			healingFountainResponseWithStatus(
				fountain,
				status,
				discoveredByFountain[fountain.ID],
			),
		)
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) getHealingFountain(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	fountainID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid healing fountain ID"})
		return
	}
	fountain, err := s.dbClient.HealingFountain().FindByID(ctx, fountainID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "healing fountain not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if fountain == nil || fountain.Invalidated {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "healing fountain not found"})
		return
	}
	latestVisit, err := s.dbClient.HealingFountain().FindLatestVisitByUserAndFountain(ctx, user.ID, fountain.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	discovery, err := s.dbClient.HealingFountain().FindDiscoveryByUserAndFountain(ctx, user.ID, fountain.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	status := healingFountainCooldownStatusFromVisit(latestVisit, time.Now())
	ctx.JSON(http.StatusOK, healingFountainResponseWithStatus(*fountain, status, discovery != nil))
}

func (s *server) getHealingFountainsForZone(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	zoneID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}
	fountains, err := s.dbClient.HealingFountain().FindByZoneID(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	latestVisitsByFountain, err := s.dbClient.HealingFountain().FindLatestVisitsByUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	discoveredByFountain, err := s.userHealingFountainDiscoveryMap(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]healingFountainWithUserStatus, 0, len(fountains))
	for _, fountain := range fountains {
		if fountain.Invalidated {
			continue
		}
		status := healingFountainCooldownStatusFromVisit(
			latestVisitsByFountain[fountain.ID],
			now,
		)
		response = append(
			response,
			healingFountainResponseWithStatus(
				fountain,
				status,
				discoveredByFountain[fountain.ID],
			),
		)
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) createHealingFountain(ctx *gin.Context) {
	var body healingFountainUpsertRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Latitude == nil || body.Longitude == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "latitude and longitude are required"})
		return
	}
	zoneID, err := uuid.Parse(strings.TrimSpace(body.ZoneID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = "Healing Fountain"
	}
	description := strings.TrimSpace(body.Description)
	thumbnailURL := strings.TrimSpace(body.ThumbnailURL)
	if thumbnailURL == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "thumbnailUrl is required"})
		return
	}

	fountain := &models.HealingFountain{
		Name:         name,
		Description:  description,
		ThumbnailURL: thumbnailURL,
		ZoneID:       zoneID,
		Latitude:     *body.Latitude,
		Longitude:    *body.Longitude,
		Invalidated:  false,
	}
	if err := s.dbClient.HealingFountain().Create(ctx, fountain); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create healing fountain: " + err.Error()})
		return
	}
	created, err := s.dbClient.HealingFountain().FindByID(ctx, fountain.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch created healing fountain: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

func (s *server) updateHealingFountain(ctx *gin.Context) {
	fountainID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid healing fountain ID"})
		return
	}
	existing, err := s.dbClient.HealingFountain().FindByID(ctx, fountainID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "healing fountain not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "healing fountain not found"})
		return
	}

	var body healingFountainUpsertRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zoneID := existing.ZoneID
	if zoneRaw := strings.TrimSpace(body.ZoneID); zoneRaw != "" {
		parsedZoneID, err := uuid.Parse(zoneRaw)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
			return
		}
		zoneID = parsedZoneID
	}

	name := existing.Name
	if candidate := strings.TrimSpace(body.Name); candidate != "" {
		name = candidate
	}
	description := existing.Description
	if body.Description != "" {
		description = strings.TrimSpace(body.Description)
	}
	thumbnailURL := existing.ThumbnailURL
	if candidate := strings.TrimSpace(body.ThumbnailURL); candidate != "" {
		thumbnailURL = candidate
	}
	if strings.TrimSpace(thumbnailURL) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "thumbnailUrl is required"})
		return
	}

	latitude := existing.Latitude
	longitude := existing.Longitude
	if body.Latitude != nil {
		latitude = *body.Latitude
	}
	if body.Longitude != nil {
		longitude = *body.Longitude
	}

	updates := &models.HealingFountain{
		ID:           existing.ID,
		Name:         name,
		Description:  description,
		ThumbnailURL: thumbnailURL,
		ZoneID:       zoneID,
		Latitude:     latitude,
		Longitude:    longitude,
		Invalidated:  existing.Invalidated,
	}
	if err := s.dbClient.HealingFountain().Update(ctx, existing.ID, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update healing fountain: " + err.Error()})
		return
	}
	updated, err := s.dbClient.HealingFountain().FindByID(ctx, existing.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated healing fountain: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteHealingFountain(ctx *gin.Context) {
	fountainID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid healing fountain ID"})
		return
	}
	if _, err := s.dbClient.HealingFountain().FindByID(ctx, fountainID); err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "healing fountain not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.HealingFountain().Delete(ctx, fountainID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete healing fountain: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "healing fountain deleted successfully"})
}

func (s *server) unlockHealingFountain(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	fountainID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid healing fountain ID"})
		return
	}
	fountain, err := s.dbClient.HealingFountain().FindByID(ctx, fountainID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "healing fountain not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if fountain == nil || fountain.Invalidated {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "healing fountain not found"})
		return
	}

	existingDiscovery, err := s.dbClient.HealingFountain().FindDiscoveryByUserAndFountain(ctx, user.ID, fountain.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingDiscovery != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message":           "healing fountain already discovered",
			"healingFountainId": fountain.ID,
			"discovered":        true,
		})
		return
	}

	userLat, userLng, err := s.getUserLatLng(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	distance := util.HaversineDistance(userLat, userLng, fountain.Latitude, fountain.Longitude)
	if distance > healingFountainInteractRadiusMeters {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf(
				"you must be within %.0f meters of the healing fountain to discover it. Currently %.0f meters away",
				healingFountainInteractRadiusMeters,
				distance,
			),
		})
		return
	}

	discovery := &models.UserHealingFountainDiscovery{
		UserID:            user.ID,
		HealingFountainID: fountain.ID,
	}
	if err := s.dbClient.HealingFountain().CreateUserHealingFountainDiscovery(ctx, discovery); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":           "healing fountain discovered",
		"healingFountainId": fountain.ID,
		"discovered":        true,
	})
}

func (s *server) generateHealingFountainImage(ctx *gin.Context) {
	fountainID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid healing fountain ID"})
		return
	}
	fountain, err := s.dbClient.HealingFountain().FindByID(ctx, fountainID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "healing fountain not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if fountain == nil || fountain.Invalidated {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "healing fountain not found"})
		return
	}

	var requestBody struct {
		Prompt *string `json:"prompt"`
	}
	if err := ctx.ShouldBindJSON(&requestBody); err != nil && err != io.EOF {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prompt := strings.TrimSpace(defaultHealingFountainImagePrompt)
	if requestBody.Prompt != nil {
		customPrompt := strings.TrimSpace(*requestBody.Prompt)
		if customPrompt == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "prompt cannot be blank"})
			return
		}
		prompt = customPrompt
	} else {
		name := strings.TrimSpace(fountain.Name)
		description := strings.TrimSpace(fountain.Description)
		if name != "" || description != "" {
			prompt = fmt.Sprintf(
				"%s Name: %s. Description: %s.",
				defaultHealingFountainImagePrompt,
				name,
				description,
			)
		}
	}

	request := deep_priest.GenerateImageRequest{Prompt: prompt}
	deep_priest.ApplyGenerateImageDefaults(&request)
	generatedImageURL, err := s.deepPriest.GenerateImage(request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updates := &models.HealingFountain{
		ID:           fountain.ID,
		Name:         fountain.Name,
		Description:  fountain.Description,
		ThumbnailURL: generatedImageURL,
		ZoneID:       fountain.ZoneID,
		Latitude:     fountain.Latitude,
		Longitude:    fountain.Longitude,
		Invalidated:  fountain.Invalidated,
	}
	if err := s.dbClient.HealingFountain().Update(ctx, fountain.ID, updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update healing fountain image: " + err.Error()})
		return
	}

	updated, err := s.dbClient.HealingFountain().FindByID(ctx, fountain.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated healing fountain: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":          "completed",
		"healingFountain": updated,
		"thumbnailUrl":    generatedImageURL,
		"prompt":          prompt,
	})
}

func (s *server) useHealingFountain(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	fountainID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid healing fountain ID"})
		return
	}
	fountain, err := s.dbClient.HealingFountain().FindByID(ctx, fountainID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "healing fountain not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if fountain == nil || fountain.Invalidated {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "healing fountain not found"})
		return
	}
	discovery, err := s.dbClient.HealingFountain().FindDiscoveryByUserAndFountain(ctx, user.ID, fountain.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if discovery == nil {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "healing fountain must be discovered before use",
		})
		return
	}

	now := time.Now()
	latestVisit, err := s.dbClient.HealingFountain().FindLatestVisitByUserAndFountain(ctx, user.ID, fountain.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cooldownStatus := healingFountainCooldownStatusFromVisit(latestVisit, now)
	if !cooldownStatus.AvailableNow {
		ctx.JSON(http.StatusTooManyRequests, gin.H{
			"error":                    "healing fountain is on cooldown",
			"lastUsedAt":               cooldownStatus.LastUsedAt,
			"nextAvailableAt":          cooldownStatus.NextAvailableAt,
			"cooldownSecondsRemaining": cooldownStatus.CooldownSecondsRemaining,
		})
		return
	}

	userLat, userLng, err := s.getUserLatLng(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	distance := util.HaversineDistance(userLat, userLng, fountain.Latitude, fountain.Longitude)
	if distance > healingFountainInteractRadiusMeters {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf(
				"you must be within %.0f meters of the healing fountain. Currently %.0f meters away",
				healingFountainInteractRadiusMeters,
				distance,
			),
		})
		return
	}

	stats, maxHealth, maxMana, currentHealth, currentMana, err := s.getScenarioResourceState(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	healthRestored := stats.HealthDeficit
	manaRestored := stats.ManaDeficit
	if healthRestored > 0 || manaRestored > 0 {
		if _, err := s.dbClient.UserCharacterStats().AdjustResourceDeficits(ctx, user.ID, -healthRestored, -manaRestored); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		currentHealth = maxHealth
		currentMana = maxMana
	}

	visit := &models.UserHealingFountainVisit{
		UserID:            user.ID,
		HealingFountainID: fountain.ID,
		VisitedAt:         now,
	}
	if err := s.dbClient.HealingFountain().CreateUserHealingFountainVisit(ctx, visit); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	nextAvailableAt := now.Add(healingFountainCooldownDuration)
	ctx.JSON(http.StatusOK, gin.H{
		"message":                  "healing fountain used successfully",
		"healingFountainId":        fountain.ID,
		"healthRestored":           healthRestored,
		"manaRestored":             manaRestored,
		"currentHealth":            currentHealth,
		"maxHealth":                maxHealth,
		"currentMana":              currentMana,
		"maxMana":                  maxMana,
		"lastUsedAt":               now,
		"nextAvailableAt":          nextAvailableAt,
		"cooldownSecondsRemaining": int(healingFountainCooldownDuration.Seconds()),
		"availableNow":             false,
	})
}

type healingFountainCooldownInfo struct {
	AvailableNow             bool
	LastUsedAt               *time.Time
	NextAvailableAt          *time.Time
	CooldownSecondsRemaining int
}

func healingFountainCooldownStatusFromVisit(
	latestVisit *models.UserHealingFountainVisit,
	now time.Time,
) healingFountainCooldownInfo {
	if latestVisit == nil {
		return healingFountainCooldownInfo{AvailableNow: true}
	}
	lastUsed := latestVisit.VisitedAt
	nextAvailable := lastUsed.Add(healingFountainCooldownDuration)
	if !now.Before(nextAvailable) {
		return healingFountainCooldownInfo{
			AvailableNow: true,
			LastUsedAt:   &lastUsed,
		}
	}
	remaining := int(math.Ceil(nextAvailable.Sub(now).Seconds()))
	if remaining < 1 {
		remaining = 1
	}
	return healingFountainCooldownInfo{
		AvailableNow:             false,
		LastUsedAt:               &lastUsed,
		NextAvailableAt:          &nextAvailable,
		CooldownSecondsRemaining: remaining,
	}
}

func healingFountainResponseWithStatus(
	fountain models.HealingFountain,
	status healingFountainCooldownInfo,
	discovered bool,
) healingFountainWithUserStatus {
	var lastUsedAt *time.Time
	if status.LastUsedAt != nil {
		v := *status.LastUsedAt
		lastUsedAt = &v
	}
	var nextAvailableAt *time.Time
	if status.NextAvailableAt != nil {
		v := *status.NextAvailableAt
		nextAvailableAt = &v
	}
	return healingFountainWithUserStatus{
		HealingFountain:          fountain,
		Discovered:               discovered,
		AvailableNow:             status.AvailableNow,
		LastUsedAt:               lastUsedAt,
		NextAvailableAt:          nextAvailableAt,
		CooldownSecondsRemaining: status.CooldownSecondsRemaining,
	}
}

func (s *server) userHealingFountainDiscoveryMap(ctx *gin.Context, userID uuid.UUID) (map[uuid.UUID]bool, error) {
	rows, err := s.dbClient.HealingFountain().GetDiscoveriesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	discoveredByFountain := make(map[uuid.UUID]bool, len(rows))
	for _, row := range rows {
		discoveredByFountain[row.HealingFountainID] = true
	}
	return discoveredByFountain, nil
}
