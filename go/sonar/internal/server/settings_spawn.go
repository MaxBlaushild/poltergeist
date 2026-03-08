package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

var nearbyScenarioPrompts = []string{
	"A faint shimmer appears between the cracks in the pavement. Describe how you investigate it.",
	"A local spirit leaves a coded message nearby. Explain how you decipher it.",
	"A hidden stash has been disturbed. Describe your approach to securing the area.",
	"You notice fresh signs of arcane activity. Explain how you track down the source.",
	"A whispering relic surfaces near you. Describe how you safely handle it.",
}

var fallbackNearbyMonsterNames = []string{
	"Restless Shade",
	"Street Warden",
	"Rift Stalker",
	"Moonlit Drifter",
	"Lantern Wraith",
}

func (s *server) spawnNearbyScenarioAndMonster(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userLat, userLng, err := s.getUserLatLng(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zones, err := s.dbClient.Zone().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	zone, err := selectZoneForCoordinates(zones, userLat, userLng)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scenarioLat, scenarioLng := randomPointNear(userLat, userLng, scenarioInteractRadiusMeters)
	scenario := &models.Scenario{
		ZoneID:                    zone.ID,
		Latitude:                  scenarioLat,
		Longitude:                 scenarioLng,
		Prompt:                    randomStringChoice(nearbyScenarioPrompts, "A strange opportunity appears nearby. Describe what you do."),
		ImageURL:                  poiPlaceholderImageURL,
		ThumbnailURL:              scenarioUndiscoveredIconKey,
		ScaleWithUserLevel:        true,
		RewardMode:                models.RewardModeRandom,
		RandomRewardSize:          models.RandomRewardSizeSmall,
		Difficulty:                scenarioDefaultDifficulty,
		RewardExperience:          0,
		RewardGold:                0,
		OpenEnded:                 true,
		FailurePenaltyMode:        models.ScenarioFailurePenaltyModeShared,
		FailureHealthDrainType:    models.ScenarioFailureDrainTypeNone,
		FailureHealthDrainValue:   0,
		FailureManaDrainType:      models.ScenarioFailureDrainTypeNone,
		FailureManaDrainValue:     0,
		FailureStatuses:           models.ScenarioFailureStatusTemplates{},
		SuccessRewardMode:         models.ScenarioSuccessRewardModeShared,
		SuccessHealthRestoreType:  models.ScenarioFailureDrainTypeNone,
		SuccessHealthRestoreValue: 0,
		SuccessManaRestoreType:    models.ScenarioFailureDrainTypeNone,
		SuccessManaRestoreValue:   0,
		SuccessStatuses:           models.ScenarioFailureStatusTemplates{},
	}
	if err := s.dbClient.Scenario().Create(ctx, scenario); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Scenario().ReplaceOptions(ctx, scenario.ID, []models.ScenarioOption{}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Scenario().ReplaceItemRewards(ctx, scenario.ID, []models.ScenarioItemReward{}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Scenario().ReplaceSpellRewards(ctx, scenario.ID, []models.ScenarioSpellReward{}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	scenarioImageQueued := true
	if err := s.enqueueScenarioImageGenerationTask(scenario.ID); err != nil {
		scenarioImageQueued = false
		log.Printf("spawnNearbyScenarioAndMonster: failed to queue scenario image generation for %s: %v", scenario.ID, err)
	}

	monsterSeed, err := s.randomMonsterSeedForZone(ctx, zone.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	monsterLat, monsterLng := randomPointNear(userLat, userLng, scenarioInteractRadiusMeters)
	monster := spawnedMonsterFromSeed(zone.ID, monsterLat, monsterLng, monsterSeed)
	if err := s.dbClient.Monster().Create(ctx, monster); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Monster().ReplaceItemRewards(ctx, monster.ID, []models.MonsterItemReward{}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	encounterLat, encounterLng := randomPointNear(userLat, userLng, scenarioInteractRadiusMeters)
	encounter := &models.MonsterEncounter{
		Name:               fmt.Sprintf("%s Encounter", strings.TrimSpace(monster.Name)),
		Description:        strings.TrimSpace(monster.Description),
		ImageURL:           strings.TrimSpace(monster.ImageURL),
		ThumbnailURL:       strings.TrimSpace(monster.ThumbnailURL),
		ScaleWithUserLevel: true,
		ZoneID:             zone.ID,
		Latitude:           encounterLat,
		Longitude:          encounterLng,
	}
	if encounter.Description == "" {
		encounter.Description = "A hostile presence manifests nearby."
	}
	if encounter.ImageURL == "" {
		encounter.ImageURL = poiPlaceholderImageURL
	}
	if encounter.ThumbnailURL == "" {
		encounter.ThumbnailURL = monsterUndiscoveredIconKey
	}
	if err := s.dbClient.MonsterEncounter().Create(ctx, encounter); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.MonsterEncounter().ReplaceMembers(ctx, encounter.ID, []models.MonsterEncounterMember{
		{
			MonsterID: monster.ID,
			Slot:      1,
		},
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"zoneId":   zone.ID,
		"zoneName": zone.Name,
		"scenario": gin.H{
			"id":                 scenario.ID,
			"latitude":           scenario.Latitude,
			"longitude":          scenario.Longitude,
			"distanceMeters":     util.HaversineDistance(userLat, userLng, scenario.Latitude, scenario.Longitude),
			"scaleWithUserLevel": scenario.ScaleWithUserLevel,
			"imageQueued":        scenarioImageQueued,
		},
		"monster": gin.H{
			"id":             monster.ID,
			"name":           monster.Name,
			"rewardMode":     monster.RewardMode,
			"rewardSize":     monster.RandomRewardSize,
			"distanceMeters": util.HaversineDistance(userLat, userLng, monster.Latitude, monster.Longitude),
		},
		"monsterEncounter": gin.H{
			"id":                 encounter.ID,
			"latitude":           encounter.Latitude,
			"longitude":          encounter.Longitude,
			"distanceMeters":     util.HaversineDistance(userLat, userLng, encounter.Latitude, encounter.Longitude),
			"scaleWithUserLevel": encounter.ScaleWithUserLevel,
		},
	})
}

func (s *server) enqueueScenarioImageGenerationTask(scenarioID uuid.UUID) error {
	payload := jobs.GenerateScenarioImageTaskPayload{
		ScenarioID: scenarioID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateScenarioImageTaskType, payloadBytes))
	return err
}

func (s *server) randomMonsterSeedForZone(ctx context.Context, zoneID uuid.UUID) (*models.Monster, error) {
	monsters, err := s.dbClient.Monster().FindByZoneIDExcludingQuestNodes(ctx, zoneID)
	if err != nil {
		return nil, err
	}
	if len(monsters) == 0 {
		monsters, err = s.dbClient.Monster().FindAll(ctx)
		if err != nil {
			return nil, err
		}
	}
	if len(monsters) == 0 {
		return nil, nil
	}
	choice := secureRandomIntBetween(0, len(monsters)-1)
	selected := monsters[choice]
	return &selected, nil
}

func spawnedMonsterFromSeed(zoneID uuid.UUID, latitude float64, longitude float64, seed *models.Monster) *models.Monster {
	name := ""
	description := ""
	imageURL := ""
	thumbnailURL := ""
	level := 1
	var templateID *uuid.UUID
	var dominantHandID *int
	var offHandID *int
	var weaponID *int

	if seed != nil {
		name = strings.TrimSpace(seed.Name)
		description = strings.TrimSpace(seed.Description)
		imageURL = strings.TrimSpace(seed.ImageURL)
		thumbnailURL = strings.TrimSpace(seed.ThumbnailURL)
		level = maxInt(1, seed.Level)
		if seed.TemplateID != nil {
			clonedTemplateID := *seed.TemplateID
			templateID = &clonedTemplateID
		}
		if seed.DominantHandInventoryItemID != nil {
			dominantHandID = intPtr(*seed.DominantHandInventoryItemID)
		}
		if seed.OffHandInventoryItemID != nil {
			offHandID = intPtr(*seed.OffHandInventoryItemID)
		}
		if seed.WeaponInventoryItemID != nil {
			weaponID = intPtr(*seed.WeaponInventoryItemID)
		}
	}

	if name == "" {
		name = randomStringChoice(fallbackNearbyMonsterNames, "Nearby Threat")
	}
	if description == "" {
		description = "A dangerous creature has emerged close to your position."
	}
	if imageURL == "" {
		imageURL = poiPlaceholderImageURL
	}
	if thumbnailURL == "" {
		thumbnailURL = monsterUndiscoveredIconKey
	}

	return &models.Monster{
		Name:                        name,
		Description:                 description,
		ImageURL:                    imageURL,
		ThumbnailURL:                thumbnailURL,
		ZoneID:                      zoneID,
		Latitude:                    latitude,
		Longitude:                   longitude,
		TemplateID:                  templateID,
		DominantHandInventoryItemID: dominantHandID,
		OffHandInventoryItemID:      offHandID,
		WeaponInventoryItemID:       weaponID,
		Level:                       level,
		RewardMode:                  models.RewardModeRandom,
		RandomRewardSize:            models.RandomRewardSizeSmall,
		RewardExperience:            0,
		RewardGold:                  0,
		ImageGenerationStatus:       models.MonsterImageGenerationStatusNone,
	}
}

func selectZoneForCoordinates(zones []*models.Zone, latitude float64, longitude float64) (*models.Zone, error) {
	var nearest *models.Zone
	nearestDistance := math.MaxFloat64

	for _, zone := range zones {
		if zone == nil {
			continue
		}
		if zone.IsPointInBoundary(latitude, longitude) {
			return zone, nil
		}
		distance := util.HaversineDistance(latitude, longitude, zone.Latitude, zone.Longitude)
		if distance < nearestDistance {
			nearest = zone
			nearestDistance = distance
		}
	}

	if nearest == nil {
		return nil, fmt.Errorf("no zones available")
	}
	return nearest, nil
}

func randomPointNear(latitude float64, longitude float64, maxDistanceMeters float64) (float64, float64) {
	if maxDistanceMeters <= 0 {
		return latitude, longitude
	}

	distanceMeters := float64(secureRandomIntBetween(0, int(math.Round(maxDistanceMeters))))
	bearingDegrees := float64(secureRandomIntBetween(0, 359))
	bearingRadians := bearingDegrees * math.Pi / 180.0

	const earthRadiusMeters = 6371000.0
	angularDistance := distanceMeters / earthRadiusMeters
	startLat := latitude * math.Pi / 180.0
	startLng := longitude * math.Pi / 180.0

	endLat := math.Asin(
		math.Sin(startLat)*math.Cos(angularDistance) +
			math.Cos(startLat)*math.Sin(angularDistance)*math.Cos(bearingRadians),
	)
	endLng := startLng + math.Atan2(
		math.Sin(bearingRadians)*math.Sin(angularDistance)*math.Cos(startLat),
		math.Cos(angularDistance)-math.Sin(startLat)*math.Sin(endLat),
	)
	endLng = math.Mod(endLng+3*math.Pi, 2*math.Pi) - math.Pi

	return endLat * 180.0 / math.Pi, endLng * 180.0 / math.Pi
}

func randomStringChoice(options []string, fallback string) string {
	if len(options) == 0 {
		return fallback
	}
	index := secureRandomIntBetween(0, len(options)-1)
	choice := strings.TrimSpace(options[index])
	if choice == "" {
		return fallback
	}
	return choice
}
