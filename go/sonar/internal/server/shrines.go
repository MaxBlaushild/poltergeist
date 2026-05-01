package server

import (
	"context"
	stdErrors "errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	shrineInteractRadiusMeters = 50.0
	shrineBlessingDuration     = 24 * time.Hour
	shrineDefaultCooldown      = 7 * 24 * time.Hour
)

type shrineWithUserStatus struct {
	models.Shrine
	Name                     string                  `json:"name"`
	Description              string                  `json:"description"`
	BlessingName             string                  `json:"blessingName"`
	EffectDescription        string                  `json:"effectDescription"`
	EffectKind               models.ShrineEffectKind `json:"effectKind"`
	BaseMagnitude            int                     `json:"baseMagnitude"`
	AvailableNow             bool                    `json:"availableNow"`
	LastUsedAt               *time.Time              `json:"lastUsedAt,omitempty"`
	NextAvailableAt          *time.Time              `json:"nextAvailableAt,omitempty"`
	CooldownSecondsRemaining int                     `json:"cooldownSecondsRemaining"`
}

type shrineUpsertRequest struct {
	ShrineTemplateID string   `json:"shrineTemplateId"`
	ZoneID           string   `json:"zoneId"`
	ZoneKind         *string  `json:"zoneKind"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	CooldownSeconds  *int     `json:"cooldownSeconds"`
}

type shrineCooldownInfo struct {
	AvailableNow             bool
	LastUsedAt               *time.Time
	NextAvailableAt          *time.Time
	CooldownSecondsRemaining int
}

func (s *server) getShrines(ctx *gin.Context) {
	user, userErr := s.getAuthenticatedUser(ctx)
	var userID *uuid.UUID
	if userErr == nil {
		userID = &user.ID
	}
	shrines, err := s.dbClient.Shrine().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response, err := s.shrineResponsesForUser(ctx.Request.Context(), shrines, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) getShrine(ctx *gin.Context) {
	user, userErr := s.getAuthenticatedUser(ctx)
	var userID *uuid.UUID
	if userErr == nil {
		userID = &user.ID
	}
	shrineID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid shrine ID"})
		return
	}
	shrine, err := s.dbClient.Shrine().FindByID(ctx, shrineID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "shrine not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if shrine == nil || shrine.Invalidated {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "shrine not found"})
		return
	}
	response, err := s.shrineResponsesForUser(ctx.Request.Context(), []models.Shrine{*shrine}, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(response) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "shrine not found"})
		return
	}
	ctx.JSON(http.StatusOK, response[0])
}

func (s *server) getShrinesForZone(ctx *gin.Context) {
	user, userErr := s.getAuthenticatedUser(ctx)
	var userID *uuid.UUID
	if userErr == nil {
		userID = &user.ID
	}
	zoneID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}
	shrines, err := s.dbClient.Shrine().FindByZoneID(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response, err := s.shrineResponsesForUser(ctx.Request.Context(), shrines, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) createShrine(ctx *gin.Context) {
	var body shrineUpsertRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	shrine, err := s.parseShrineUpsertRequest(ctx, body, nil)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Shrine().Create(ctx, shrine); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	created, err := s.dbClient.Shrine().FindByID(ctx, shrine.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

func (s *server) updateShrine(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid shrine ID"})
		return
	}
	existing, err := s.dbClient.Shrine().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "shrine not found"})
		return
	}

	var body shrineUpsertRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	shrine, err := s.parseShrineUpsertRequest(ctx, body, existing)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Shrine().Update(ctx, id, shrine); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updated, err := s.dbClient.Shrine().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) updateShrineLocation(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid shrine ID"})
		return
	}
	existing, err := s.dbClient.Shrine().FindByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "shrine not found"})
		return
	}

	var body struct {
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Latitude == nil || body.Longitude == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "latitude and longitude are required"})
		return
	}
	updated := *existing
	updated.Latitude = *body.Latitude
	updated.Longitude = *body.Longitude
	if err := s.dbClient.Shrine().Update(ctx, id, &updated); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updated)
}

func (s *server) deleteShrine(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid shrine ID"})
		return
	}
	if err := s.dbClient.Shrine().Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "shrine deleted successfully"})
}

func (s *server) useShrine(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	shrineID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid shrine ID"})
		return
	}
	shrine, err := s.dbClient.Shrine().FindByID(ctx, shrineID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "shrine not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if shrine == nil || shrine.Invalidated {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "shrine not found"})
		return
	}

	now := time.Now()
	latestUse, err := s.dbClient.Shrine().FindLatestUseByUserAndShrine(ctx, user.ID, shrine.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cooldownStatus := shrineCooldownStatusFromUse(shrine, latestUse, now)
	if !cooldownStatus.AvailableNow {
		ctx.JSON(http.StatusTooManyRequests, gin.H{
			"error":                    "shrine is on cooldown",
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
	distance := util.HaversineDistance(userLat, userLng, shrine.Latitude, shrine.Longitude)
	if distance > shrineInteractRadiusMeters {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf(
				"you must be within %.0f meters of the shrine. Currently %.0f meters away",
				shrineInteractRadiusMeters,
				distance,
			),
		})
		return
	}

	userLevel, err := s.currentUserLevel(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	statusTemplates := shrineBlessingStatusTemplatesForUserLevel(&shrine.Template, userLevel)
	appliedStatuses, err := s.applyMonsterBattleUserStatuses(ctx, []uuid.UUID{user.ID}, statusTemplates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	use := &models.UserShrineUse{
		UserID:   user.ID,
		ShrineID: shrine.ID,
		UsedAt:   now,
	}
	if err := s.dbClient.Shrine().CreateUserShrineUse(ctx, use); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cooldown := shrineCooldownDuration(shrine)
	nextAvailableAt := now.Add(cooldown)
	ctx.JSON(http.StatusOK, gin.H{
		"message":                  "shrine invoked successfully",
		"shrineId":                 shrine.ID,
		"blessingName":             shrine.Template.BlessingName,
		"statusesApplied":          appliedStatuses,
		"lastUsedAt":               now,
		"nextAvailableAt":          nextAvailableAt,
		"cooldownSecondsRemaining": int(cooldown.Seconds()),
		"availableNow":             false,
	})
}

func (s *server) parseShrineUpsertRequest(
	ctx *gin.Context,
	body shrineUpsertRequest,
	existing *models.Shrine,
) (*models.Shrine, error) {
	if body.Latitude == nil || body.Longitude == nil {
		return nil, fmt.Errorf("latitude and longitude are required")
	}

	templateIDText := strings.TrimSpace(body.ShrineTemplateID)
	if templateIDText == "" && existing != nil {
		templateIDText = existing.ShrineTemplateID.String()
	}
	templateID, err := uuid.Parse(templateIDText)
	if err != nil {
		return nil, fmt.Errorf("invalid shrineTemplateId")
	}
	template, err := s.dbClient.ShrineTemplate().FindByID(ctx, templateID)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, fmt.Errorf("shrineTemplateId was not found")
	}

	zoneIDText := strings.TrimSpace(body.ZoneID)
	if zoneIDText == "" && existing != nil {
		zoneIDText = existing.ZoneID.String()
	}
	zoneID, err := uuid.Parse(zoneIDText)
	if err != nil {
		return nil, fmt.Errorf("invalid zoneId")
	}
	zone, err := s.dbClient.Zone().FindByID(ctx, zoneID)
	if err != nil {
		return nil, err
	}
	if zone == nil {
		return nil, fmt.Errorf("zoneId was not found")
	}

	cooldownSeconds := int(shrineDefaultCooldown.Seconds())
	if existing != nil && existing.CooldownSeconds > 0 {
		cooldownSeconds = existing.CooldownSeconds
	}
	if body.CooldownSeconds != nil {
		if *body.CooldownSeconds < 0 {
			return nil, fmt.Errorf("cooldownSeconds must be zero or greater")
		}
		cooldownSeconds = *body.CooldownSeconds
	}

	zoneKind := models.NormalizeZoneKind(zone.Kind)
	if body.ZoneKind != nil {
		zoneKind = models.NormalizeZoneKind(*body.ZoneKind)
	}
	if zoneKind == "" {
		zoneKind = models.NormalizeZoneKind(template.ZoneKind)
	}

	shrine := &models.Shrine{
		ShrineTemplateID: template.ID,
		ZoneID:           zone.ID,
		ZoneKind:         zoneKind,
		Latitude:         *body.Latitude,
		Longitude:        *body.Longitude,
		CooldownSeconds:  cooldownSeconds,
		Invalidated:      false,
	}
	if existing != nil {
		shrine.Invalidated = existing.Invalidated
	}
	return shrine, nil
}

func (s *server) shrineResponsesForUser(
	ctx context.Context,
	shrines []models.Shrine,
	userID *uuid.UUID,
) ([]shrineWithUserStatus, error) {
	now := time.Now()
	latestUsesByShrine := map[uuid.UUID]*models.UserShrineUse{}
	var err error
	if userID != nil {
		latestUsesByShrine, err = s.dbClient.Shrine().FindLatestUsesByUser(ctx, *userID)
		if err != nil {
			return nil, err
		}
	}

	response := make([]shrineWithUserStatus, 0, len(shrines))
	markerCache := contentMapMarkerExistenceCache{}
	for _, shrine := range shrines {
		if shrine.Invalidated {
			continue
		}
		status := shrineCooldownStatusFromUse(&shrine, latestUsesByShrine[shrine.ID], now)
		markerURL := s.resolveSharedContentMapMarkerURL(
			ctx,
			sharedContentMapMarkerDefinitions[10],
			effectiveContentMapMarkerZoneKind(shrine.ZoneKind, &shrine.Zone),
			"",
			markerCache,
		)
		response = append(response, shrineResponseWithStatus(shrine, status, markerURL))
	}
	return response, nil
}

func shrineResponseWithStatus(
	shrine models.Shrine,
	status shrineCooldownInfo,
	markerURL string,
) shrineWithUserStatus {
	shrine.MapMarkerURL = markerURL
	return shrineWithUserStatus{
		Shrine:                   shrine,
		Name:                     strings.TrimSpace(shrine.Template.Name),
		Description:              strings.TrimSpace(shrine.Template.Description),
		BlessingName:             strings.TrimSpace(shrine.Template.BlessingName),
		EffectDescription:        strings.TrimSpace(shrine.Template.EffectDescription),
		EffectKind:               shrine.Template.EffectKind,
		BaseMagnitude:            shrine.Template.BaseMagnitude,
		AvailableNow:             status.AvailableNow,
		LastUsedAt:               status.LastUsedAt,
		NextAvailableAt:          status.NextAvailableAt,
		CooldownSecondsRemaining: status.CooldownSecondsRemaining,
	}
}

func shrineCooldownDuration(shrine *models.Shrine) time.Duration {
	if shrine == nil || shrine.CooldownSeconds <= 0 {
		return shrineDefaultCooldown
	}
	return time.Duration(shrine.CooldownSeconds) * time.Second
}

func shrineCooldownStatusFromUse(
	shrine *models.Shrine,
	latestUse *models.UserShrineUse,
	now time.Time,
) shrineCooldownInfo {
	if latestUse == nil {
		return shrineCooldownInfo{AvailableNow: true}
	}
	nextAvailableAt := latestUse.UsedAt.Add(shrineCooldownDuration(shrine))
	if !now.Before(nextAvailableAt) {
		return shrineCooldownInfo{
			AvailableNow:    true,
			LastUsedAt:      &latestUse.UsedAt,
			NextAvailableAt: &nextAvailableAt,
		}
	}
	remaining := int(math.Ceil(nextAvailableAt.Sub(now).Seconds()))
	if remaining < 0 {
		remaining = 0
	}
	return shrineCooldownInfo{
		AvailableNow:             false,
		LastUsedAt:               &latestUse.UsedAt,
		NextAvailableAt:          &nextAvailableAt,
		CooldownSecondsRemaining: remaining,
	}
}

func shrineBlessingStatusTemplatesForUserLevel(
	template *models.ShrineTemplate,
	userLevel int,
) models.ScenarioFailureStatusTemplates {
	if template == nil {
		return models.ScenarioFailureStatusTemplates{}
	}
	base := shrineBaseStatusTemplate(template)
	scaled := scaleInventorySuggestionStatuses(
		models.ScenarioFailureStatusTemplates{base},
		shrineBlessingScaleRatio(userLevel),
	)
	if len(scaled) == 0 {
		return models.ScenarioFailureStatusTemplates{}
	}
	scaled[0].DurationSeconds = int(shrineBlessingDuration.Seconds())
	return scaled
}

func shrineBlessingScaleRatio(userLevel int) float64 {
	return float64(expectedScenarioStatForLevel(userLevel)) / float64(models.CharacterStatBaseValue)
}

func shrineBaseStatusTemplate(
	template *models.ShrineTemplate,
) models.ScenarioFailureStatusTemplate {
	magnitude := template.BaseMagnitude
	if magnitude <= 0 {
		magnitude = 1
	}
	status := models.ScenarioFailureStatusTemplate{
		Name:            strings.TrimSpace(template.BlessingName),
		Description:     strings.TrimSpace(template.EffectDescription),
		Effect:          strings.TrimSpace(template.EffectDescription),
		EffectType:      string(models.UserStatusEffectTypeStatModifier),
		Positive:        true,
		DurationSeconds: int(shrineBlessingDuration.Seconds()),
	}
	if status.Name == "" {
		status.Name = strings.TrimSpace(template.Name)
	}
	if status.Description == "" {
		status.Description = strings.TrimSpace(template.Description)
	}
	if status.Effect == "" {
		status.Effect = status.Description
	}

	switch models.NormalizeShrineEffectKind(string(template.EffectKind)) {
	case models.ShrineEffectKindDexterity:
		status.DexterityMod = magnitude
	case models.ShrineEffectKindConstitution:
		status.ConstitutionMod = magnitude
	case models.ShrineEffectKindIntelligence:
		status.IntelligenceMod = magnitude
	case models.ShrineEffectKindWisdom:
		status.WisdomMod = magnitude
	case models.ShrineEffectKindCharisma:
		status.CharismaMod = magnitude
	case models.ShrineEffectKindHealthRegen:
		status.EffectType = string(models.UserStatusEffectTypeHealthOverTime)
		status.HealthPerTick = magnitude
	case models.ShrineEffectKindManaRegen:
		status.EffectType = string(models.UserStatusEffectTypeManaOverTime)
		status.ManaPerTick = magnitude
	case models.ShrineEffectKindPhysicalDamage:
		status.PhysicalDamageBonusPercent = magnitude * 5
	case models.ShrineEffectKindArcaneDamage:
		status.ArcaneDamageBonusPercent = magnitude * 5
	case models.ShrineEffectKindHolyDamage:
		status.HolyDamageBonusPercent = magnitude * 5
	case models.ShrineEffectKindShadowDamage:
		status.ShadowDamageBonusPercent = magnitude * 5
	case models.ShrineEffectKindFireResistance:
		status.FireResistancePercent = magnitude * 6
	case models.ShrineEffectKindIceResistance:
		status.IceResistancePercent = magnitude * 6
	case models.ShrineEffectKindLightningResistance:
		status.LightningResistancePercent = magnitude * 6
	case models.ShrineEffectKindPoisonResistance:
		status.PoisonResistancePercent = magnitude * 6
	case models.ShrineEffectKindPhysicalResistance:
		status.PhysicalResistancePercent = magnitude * 6
	case models.ShrineEffectKindAllDamageResistance:
		all := magnitude * 3
		status.PhysicalResistancePercent = all
		status.FireResistancePercent = all
		status.IceResistancePercent = all
		status.LightningResistancePercent = all
		status.PoisonResistancePercent = all
		status.ArcaneResistancePercent = all
		status.HolyResistancePercent = all
		status.ShadowResistancePercent = all
	default:
		status.StrengthMod = magnitude
	}
	return status
}
